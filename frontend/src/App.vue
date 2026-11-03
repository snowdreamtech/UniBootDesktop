<template>
  <div class="app-container">
    <!-- Window drag strip for seamless macOS/frameless window dragging -->
    <div class="window-drag-strip" style="--wails-draggable: drag" @mousedown="handleWindowDrag" />

    <!-- Global Toast Notification -->
    <ToastNotification :message="toastMessage" :type="toastType" @close="dismissToast" />

    <!-- Header Component -->
    <AppHeader
      :currentLang="selectedLangSetting"
      :currentTheme="currentTheme"
      @toggle-theme="toggleTheme"
      @open-settings="openSettings"
      @open-about="openAbout"
      @open-logs="isLogViewerOpen = true"
      @select-lang="selectLanguage"
    />

    <!-- Main Content Area -->
    <main class="main-content">
      <HelloPanel :appConfig="appConfig" @open-settings="openSettings" @open-about="openAbout" />
    </main>

    <!-- Settings Modal -->
    <SettingsModal
      :isOpen="isSettingsOpen"
      :currentProxy="appConfig?.githubProxy"
      @close="isSettingsOpen = false"
      @save="onSaveSettings"
    />

    <!-- About Modal -->
    <AboutModal :show="isAboutOpen" @close="isAboutOpen = false" />

    <!-- Log Viewer Modal -->
    <LogViewerModal
      :isOpen="isLogViewerOpen"
      :logs="runtimeLogs"
      @close="isLogViewerOpen = false"
      @clear="handleClearLogs"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from "vue";
import AppHeader from "./components/AppHeader.vue";
import HelloPanel from "./components/HelloPanel.vue";
import SettingsModal from "./components/SettingsModal.vue";
import AboutModal from "./components/AboutModal.vue";
import LogViewerModal, { type LogItem } from "./components/LogViewerModal.vue";
import ToastNotification from "./components/ToastNotification.vue";
import { selectedLangSetting, setLanguage } from "./i18n";
import { GetConfig, SaveConfig, CheckUpdate } from "../wailsjs/go/main/App";
import { WindowShow } from "../wailsjs/runtime/runtime";
import { isWailsRuntime } from "./utils/wails";
import { useTheme } from "./composables/useTheme";
import { useToast } from "./composables/useToast";
import { useAppRuntimeEvents } from "./composables/useAppRuntimeEvents";
import type { config } from "../wailsjs/go/models";

type AppConfigType = config.AppConfig;

const { currentTheme, applyTheme, toggleTheme: baseToggleTheme } = useTheme();
const { toastMessage, toastType, showToast, dismissToast } = useToast();

const isSettingsOpen = ref(false);
const isAboutOpen = ref(false);
const isLogViewerOpen = ref(false);
const runtimeLogs = ref<LogItem[]>([]);
const appConfig = ref<AppConfigType | null>(null);

function handleClearLogs() {
  runtimeLogs.value = [];
  const w = window as any;
  if (w.go && w.go.main && w.go.main.App && typeof w.go.main.App.ClearLogs === "function") {
    w.go.main.App.ClearLogs().catch(() => {});
  }
}

import { isClickOnScrollbar, triggerNativeDrag } from "./utils/windowDrag";

function handleWindowDrag(e: MouseEvent) {
  if (e.buttons !== 1 || e.detail > 1) return;
  if (isClickOnScrollbar(e)) return;
  triggerNativeDrag();
}

function selectLanguage(langVal: string) {
  setLanguage(langVal);
  if (appConfig.value) {
    appConfig.value.language = langVal;
    saveConfigToBackend(appConfig.value);
  }
  const w = window as any;
  if (w.go && w.go.main && w.go.main.App && typeof w.go.main.App.ReloadAppMenu === "function") {
    w.go.main.App.ReloadAppMenu(langVal).catch((err: any) => {
      console.warn("Failed to reload app menu:", err);
    });
  }
}

function toggleTheme() {
  const nextTheme = baseToggleTheme();
  if (appConfig.value) {
    appConfig.value.theme = nextTheme;
    saveConfigToBackend(appConfig.value);
  }
}

function openSettings() {
  isSettingsOpen.value = true;
}

function openAbout() {
  isAboutOpen.value = true;
}

useAppRuntimeEvents({
  openAbout,
  openSettings,
  onThemeChanged: (theme) => applyTheme(theme),
});

function onSaveSettings(savedCfg: any) {
  appConfig.value = { ...(appConfig.value || {}), ...savedCfg } as any;
  if (savedCfg.theme) {
    applyTheme(savedCfg.theme);
  }
  if (savedCfg.language) {
    setLanguage(savedCfg.language);
    const w = window as any;
    if (w.go && w.go.main && w.go.main.App && typeof w.go.main.App.ReloadAppMenu === "function") {
      w.go.main.App.ReloadAppMenu(savedCfg.language).catch((err: any) => {
        console.warn("Failed to reload app menu:", err);
      });
    }
  }
  saveConfigToBackend(appConfig.value);
  showToast("Settings saved successfully", "success");
}

async function saveConfigToBackend(cfg: any) {
  try {
    if (isWailsRuntime()) {
      await SaveConfig(cfg);
    }
  } catch (err) {
    console.error("Failed to save config to backend:", err);
    showToast("Failed to save configuration", "error");
  }
}

async function loadConfig() {
  try {
    if (isWailsRuntime()) {
      const cfg = await GetConfig();
      if (cfg) {
        appConfig.value = cfg;
        if (cfg.theme) {
          applyTheme(cfg.theme);
        }
        if (cfg.language) {
          setLanguage(cfg.language);
        }
      }
    }
  } catch (err) {
    console.warn("Failed to load initial configuration:", err);
  }
}

let mediaQueryList: MediaQueryList | null = null;

function handleSystemThemeChange() {
  if (currentTheme.value === "system") {
    applyTheme("system");
  }
}

function handleGlobalKeydown(e: KeyboardEvent) {
  if (e.key === "Escape") {
    if (isAboutOpen.value) {
      isAboutOpen.value = false;
    } else if (isSettingsOpen.value) {
      isSettingsOpen.value = false;
    }
  }
}

onMounted(() => {
  loadConfig();

  const w = window as any;
  if (w.go && w.go.main && w.go.main.App && typeof w.go.main.App.GetRecentLogs === "function") {
    w.go.main.App.GetRecentLogs()
      .then((logs: any[]) => {
        if (Array.isArray(logs)) {
          runtimeLogs.value = logs;
        }
      })
      .catch(() => {});
  }

  if (isWailsRuntime() && window.runtime && typeof window.runtime.EventsOn === "function") {
    window.runtime.EventsOn("log:entry", (entry: any) => {
      runtimeLogs.value.push(entry);
      if (runtimeLogs.value.length > 500) {
        runtimeLogs.value.shift();
      }
    });

    try {
      WindowShow();
    } catch (e) {
      console.warn("WindowShow invocation failed:", e);
    }

    setTimeout(() => {
      CheckUpdate()
        .then((info) => {
          if (info && info.hasUpdate) {
            showToast(`New version ${info.latestTag} available!`, "info");
          }
        })
        .catch((err) => {
          console.warn("Background update check failed:", err);
        });
    }, 3000);
  }

  window.addEventListener("keydown", handleGlobalKeydown);

  if (window.matchMedia) {
    mediaQueryList = window.matchMedia("(prefers-color-scheme: dark)");
    mediaQueryList.addEventListener("change", handleSystemThemeChange);
  }
});

onUnmounted(() => {
  window.removeEventListener("keydown", handleGlobalKeydown);

  if (isWailsRuntime() && window.runtime && typeof window.runtime.EventsOff === "function") {
    window.runtime.EventsOff("log:entry");
  }

  if (mediaQueryList) {
    mediaQueryList.removeEventListener("change", handleSystemThemeChange);
    mediaQueryList = null;
  }
});
</script>

<style scoped>
.window-drag-strip {
  position: fixed;
  top: 0;
  left: 0;
  right: 18px;
  height: 24px;
  z-index: 999;
  pointer-events: auto;
  --wails-draggable: drag;
  -webkit-app-region: drag;
}

.app-container {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  width: 100vw;
  background-color: var(--bg-color);
  color: var(--text-main);
  padding: 2.25rem 2rem 1.5rem;
  box-sizing: border-box;
  overflow-x: hidden;
  cursor: default;
  --wails-draggable: no-drag;
  -webkit-app-region: no-drag;
}

.main-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  --wails-draggable: no-drag;
  -webkit-app-region: no-drag;
}

button,
input,
select,
textarea,
a,
pre,
code,
.settings-modal-card,
.about-modal-card {
  --wails-draggable: no-drag;
  -webkit-app-region: no-drag;
}
</style>

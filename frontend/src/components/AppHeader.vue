<template>
  <header
    class="app-header"
    style="--wails-draggable: drag"
    @mousedown="handleHeaderMouseDown"
    @dblclick="handleHeaderDblClick"
  >
    <div class="brand" style="--wails-draggable: drag">
      <img src="/logo.png" alt="UniGo" class="logo-img" />
      <div>
        <h1>{{ t("app.title") || "UniGoDesktop" }}</h1>
        <span class="sub-brand">{{ t("app.subtitle") || "Universal Cross-Platform Desktop Template" }}</span>
      </div>
    </div>

    <div class="header-actions">
      <!-- Quick Language Switcher Dropdown -->
      <div class="lang-selector-header" ref="langDropdownRef">
        <button class="lang-pill-btn" :title="t('settings.language') || 'Language'" @click.stop="toggleLangMenu">
          <span class="lang-icon">🌐</span>
          <span class="lang-label">{{ currentLangLabel }}</span>
          <span class="dropdown-caret">▾</span>
        </button>

        <transition name="dropdown-fade">
          <div v-if="isLangMenuOpen" class="lang-dropdown-menu" @click.stop>
            <button
              v-for="opt in langOptions"
              :key="opt.value"
              class="lang-option"
              :class="{ active: currentLang === opt.value }"
              @click="selectLanguage(opt.value)"
            >
              <span class="opt-text">{{ opt.label }}</span>
              <span v-if="currentLang === opt.value" class="opt-check">✓</span>
            </button>
          </div>
        </transition>
      </div>

      <!-- Quick Theme Switcher Button -->
      <button
        class="icon-action-btn"
        :title="
          isLight ? t('theme.toggleDark') || 'Switch to Dark Theme' : t('theme.toggleLight') || 'Switch to Light Theme'
        "
        @click="emit('toggle-theme')"
      >
        <span>{{ isLight ? "🌙" : "☀️" }}</span>
      </button>

      <!-- Log Viewer Button -->
      <button class="icon-action-btn" :title="t('log.title') || 'Logs'" @click="emit('open-logs')">
        <span>📜</span>
      </button>

      <!-- Settings Button -->
      <button class="icon-action-btn" :title="t('settings.title') || 'Settings'" @click="emit('open-settings')">
        <span>⚙️</span>
      </button>

      <!-- About Button -->
      <button class="icon-action-btn" :title="t('about.title') || 'About'" @click="emit('open-about')">
        <span>ℹ️</span>
      </button>
    </div>
  </header>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from "vue";
import { SUPPORTED_LANGUAGES, t } from "../i18n";
import { WindowToggleMaximise } from "../../wailsjs/runtime/runtime";
import { isWailsRuntime } from "../utils/wails";

const props = defineProps<{
  currentLang: string;
  currentTheme: string;
}>();

const emit = defineEmits<{
  (e: "toggle-theme"): void;
  (e: "open-settings"): void;
  (e: "open-about"): void;
  (e: "open-logs"): void;
  (e: "select-lang", lang: string): void;
}>();

const isLangMenuOpen = ref(false);
const langDropdownRef = ref<HTMLElement | null>(null);

const langOptions = computed(() => [
  { value: "auto", label: "🌐 " + (t("common.autoDetect") || "Auto Detect") },
  ...SUPPORTED_LANGUAGES.map((item) => ({
    value: item.code,
    label: item.nativeName,
  })),
]);

const currentLangLabel = computed(() => {
  if (props.currentLang === "auto") {
    return t("common.langAuto") || "Auto";
  }
  const opt = langOptions.value.find((o) => o.value === props.currentLang);
  return opt ? opt.label : t("common.lang") || "Language";
});

const isLight = computed(() => {
  if (props.currentTheme === "light") return true;
  if (props.currentTheme === "dark") return false;
  return typeof document !== "undefined" && document.documentElement.getAttribute("data-theme") === "light";
});

import { isClickOnScrollbar, triggerNativeDrag } from "../utils/windowDrag";

function handleHeaderMouseDown(e: MouseEvent) {
  if (e.buttons !== 1 || e.detail > 1) {
    return;
  }
  if (isClickOnScrollbar(e)) {
    return;
  }
  const target = e.target as HTMLElement | null;
  if (target && target.closest("button, input, select, a, .lang-dropdown-menu")) {
    return;
  }
  triggerNativeDrag();
}

function handleHeaderDblClick(e: MouseEvent) {
  const target = e.target as HTMLElement | null;
  if (target && target.closest("button, input, select, a, .lang-dropdown-menu")) {
    return;
  }
  if (isWailsRuntime()) {
    try {
      WindowToggleMaximise();
    } catch (err) {
      console.warn("Failed to toggle maximise:", err);
    }
  }
}

function toggleLangMenu() {
  isLangMenuOpen.value = !isLangMenuOpen.value;
}

function selectLanguage(langVal: string) {
  emit("select-lang", langVal);
  isLangMenuOpen.value = false;
}

function handleGlobalClick(event: MouseEvent) {
  if (langDropdownRef.value && !langDropdownRef.value.contains(event.target as Node)) {
    isLangMenuOpen.value = false;
  }
}

onMounted(() => {
  window.addEventListener("click", handleGlobalClick);
});

onUnmounted(() => {
  window.removeEventListener("click", handleGlobalClick);
});
</script>

<style scoped>
.app-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
  user-select: none;
  -webkit-user-select: none;
  --wails-draggable: drag;
  -webkit-app-region: drag;
}

.brand,
.brand * {
  --wails-draggable: drag;
  -webkit-app-region: drag;
}

.brand {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.logo-img {
  width: 48px;
  height: 48px;
  object-fit: contain;
  filter: drop-shadow(0 4px 6px rgba(0, 0, 0, 0.1));
}

.brand h1 {
  font-size: 1.5rem;
  font-weight: 700;
  margin: 0;
  line-height: 1.2;
  color: var(--text-main);
}

.sub-brand {
  font-size: 0.85rem;
  color: var(--text-muted);
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  position: relative;
}

.header-actions button,
.lang-selector-header,
.lang-dropdown-menu,
.lang-dropdown-menu *,
button,
input,
select,
a {
  --wails-draggable: no-drag;
  -webkit-app-region: no-drag;
}

.lang-selector-header {
  position: relative;
}

.lang-pill-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  color: var(--text-main);
  padding: 0.35rem 0.65rem;
  border-radius: 9999px;
  font-size: 0.8rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}

.lang-pill-btn:hover {
  border-color: var(--primary-color);
  background: var(--bg-hover);
}

.lang-dropdown-menu {
  position: absolute;
  top: calc(100% + 6px);
  right: 0;
  width: 200px;
  max-height: 280px;
  overflow-y: auto;
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  box-shadow: 0 10px 25px -5px rgba(0, 0, 0, 0.25);
  z-index: 1000;
  padding: 0.35rem;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.lang-option {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
  padding: 0.4rem 0.6rem;
  background: transparent;
  border: none;
  border-radius: 6px;
  color: var(--text-main);
  font-size: 0.8rem;
  cursor: pointer;
  text-align: left;
  transition: background 0.15s ease;
}

.lang-option:hover {
  background: var(--bg-hover);
  color: var(--primary-color);
}

.lang-option.active {
  background: rgba(59, 130, 246, 0.15);
  color: var(--primary-color);
  font-weight: 600;
}

.opt-check {
  font-size: 0.75rem;
  color: var(--primary-color);
}

.icon-action-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 50%;
  cursor: pointer;
  font-size: 0.9rem;
  transition: all 0.2s ease;
  color: var(--text-main);
}

.icon-action-btn:hover {
  border-color: var(--primary-color);
  background: var(--bg-hover);
  transform: translateY(-1px);
}

.dropdown-fade-enter-active,
.dropdown-fade-leave-active {
  transition:
    opacity 0.15s ease,
    transform 0.15s ease;
}

.dropdown-fade-enter-from,
.dropdown-fade-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}
</style>

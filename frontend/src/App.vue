<template>
  <div class="app-container">
    <!-- Global Toast Notification -->
    <transition name="toast-fade">
      <div v-if="toastMessage" class="global-toast" :class="toastType">
        <span class="toast-icon">
          <template v-if="toastType === 'warning'">⚠️</template>
          <template v-else-if="toastType === 'error'">❌</template>
          <template v-else-if="toastType === 'success'">🎉</template>
          <template v-else>ℹ️</template>
        </span>
        <span class="toast-text">{{ toastMessage }}</span>
        <button class="toast-close" @click="toastMessage = ''">✕</button>
      </div>
    </transition>

    <!-- Header -->
    <header class="app-header">
      <div class="brand">
        <span class="logo">🚀</span>
        <div>
          <h1>{{ t('app.title') || 'UniGoDesktop' }}</h1>
          <span class="sub-brand">{{ t('app.subtitle') || 'Universal Cross-Platform Desktop Template' }}</span>
        </div>
      </div>

      <div class="header-actions">
        <!-- Quick Language Switcher Dropdown -->
        <div class="lang-selector-header" ref="langDropdownRef">
          <button 
            class="lang-pill-btn" 
            :title="t('settings.language') || 'Language'"
            @click.stop="toggleLangMenu"
          >
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

        <!-- Quick Theme Toggle -->
        <button 
          class="icon-action-btn" 
          :title="isDarkTheme ? (t('theme.toggleLight') || 'Switch to Light Theme') : (t('theme.toggleDark') || 'Switch to Dark Theme')"
          @click="toggleTheme"
        >
          <span>{{ isDarkTheme ? '🌙' : '☀️' }}</span>
        </button>

        <!-- Settings Button -->
        <button 
          class="icon-action-btn" 
          :title="t('settings.title') || 'Settings'"
          @click="openSettings"
        >
          <span>⚙️</span>
        </button>

        <!-- About Button -->
        <button 
          class="icon-action-btn" 
          :title="t('about.title') || 'About'"
          @click="openAbout"
        >
          <span>ℹ️</span>
        </button>
      </div>
    </header>

    <!-- Main Content Area -->
    <main class="main-content">
      <HelloPanel 
        :appConfig="appConfig" 
        @open-settings="openSettings" 
        @open-about="openAbout" 
      />
    </main>

    <!-- Settings Modal -->
    <SettingsModal
      :isOpen="isSettingsOpen"
      :currentProxy="appConfig?.githubProxy"
      @close="isSettingsOpen = false"
      @save="onSaveSettings"
    />

    <!-- About Modal -->
    <AboutModal
      :show="isAboutOpen"
      @close="isAboutOpen = false"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue';
import HelloPanel from './components/HelloPanel.vue';
import SettingsModal from './components/SettingsModal.vue';
import AboutModal from './components/AboutModal.vue';
import { t, currentLang, setLanguage, SUPPORTED_LANGUAGES } from './i18n';
import type { config } from '../wailsjs/go/models';

type AppConfigType = config.AppConfig;

const isSettingsOpen = ref(false);
const isAboutOpen = ref(false);
const appConfig = ref<AppConfigType | null>(null);

const isLangMenuOpen = ref(false);
const langDropdownRef = ref<HTMLElement | null>(null);

const toastMessage = ref('');
const toastType = ref<'info' | 'success' | 'warning' | 'error'>('info');
let toastTimer: any = null;

const currentTheme = ref('dark');
const isDarkTheme = computed(() => currentTheme.value !== 'light');

const langOptions = computed(() => [
  { value: 'auto', label: '🌐 ' + (t('common.autoDetect') || 'Auto Detect') },
  ...SUPPORTED_LANGUAGES.map(item => ({
    value: item.code,
    label: `${item.nativeName} (${item.name})`
  }))
]);

const currentLangLabel = computed(() => {
  if (currentLang.value === 'auto') {
    return (t('common.langAuto') || 'Auto');
  }
  const opt = langOptions.value.find(o => o.value === currentLang.value);
  return opt ? opt.label.split(' ')[0] : 'Language';
});

function showToast(msg: string, type: 'info' | 'success' | 'warning' | 'error' = 'info') {
  toastMessage.value = msg;
  toastType.value = type;
  if (toastTimer) clearTimeout(toastTimer);
  toastTimer = setTimeout(() => {
    toastMessage.value = '';
  }, 4000);
}

function toggleLangMenu() {
  isLangMenuOpen.value = !isLangMenuOpen.value;
}

function selectLanguage(langVal: string) {
  setLanguage(langVal);
  isLangMenuOpen.value = false;
  if (appConfig.value) {
    appConfig.value.language = langVal;
    saveConfigToBackend(appConfig.value);
  }
}

function toggleTheme() {
  const nextTheme = currentTheme.value === 'light' ? 'dark' : 'light';
  applyTheme(nextTheme);
  if (appConfig.value) {
    appConfig.value.theme = nextTheme;
    saveConfigToBackend(appConfig.value);
  }
}

function applyTheme(themeName: string) {
  currentTheme.value = themeName;
  let applied = themeName;
  if (themeName === 'system') {
    const isDark = window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches;
    applied = isDark ? 'dark' : 'light';
  }
  document.documentElement.setAttribute('data-theme', applied);
}

function openSettings() {
  isSettingsOpen.value = true;
}

function openAbout() {
  isAboutOpen.value = true;
}

function onSaveSettings(savedCfg: any) {
  appConfig.value = { ...(appConfig.value || {}), ...savedCfg } as any;
  if (savedCfg.theme) {
    applyTheme(savedCfg.theme);
  }
  if (savedCfg.language) {
    setLanguage(savedCfg.language);
  }
}

async function saveConfigToBackend(cfg: any) {
  if ((window as any)?.go?.main?.App?.SaveConfig) {
    try {
      await (window as any).go.main.App.SaveConfig(cfg);
    } catch (e) {
      console.warn('Failed to save config:', e);
    }
  }
}

async function initApp() {
  if ((window as any)?.go?.main?.App?.GetConfig) {
    try {
      const cfg = await (window as any).go.main.App.GetConfig();
      if (cfg) {
        appConfig.value = cfg;
        if (cfg.language) {
          setLanguage(cfg.language);
        }
        if (cfg.theme) {
          applyTheme(cfg.theme);
        } else {
          applyTheme('dark');
        }

        // Auto check updates if enabled
        if (cfg.autoCheckUpdate !== false && (window as any)?.go?.main?.App?.CheckUpdate) {
          (window as any).go.main.App.CheckUpdate().then((res: any) => {
            if (res && res.hasUpdate) {
              showToast(`🚀 New version ${res.latestTag} is available!`, 'info');
            }
          }).catch(() => {});
        }
      }
    } catch (e) {
      console.warn('Failed to load initial config from backend:', e);
      applyTheme('dark');
    }
  } else {
    applyTheme('dark');
  }
}

function handleGlobalClick(e: MouseEvent) {
  if (isLangMenuOpen.value && langDropdownRef.value && !langDropdownRef.value.contains(e.target as Node)) {
    isLangMenuOpen.value = false;
  }
}

onMounted(() => {
  initApp();
  window.addEventListener('click', handleGlobalClick);
});

onUnmounted(() => {
  window.removeEventListener('click', handleGlobalClick);
});
</script>

<style scoped>
.app-container {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  width: 100vw;
  background-color: var(--bg-color);
  color: var(--text-main);
  padding: 1.5rem 2rem;
  box-sizing: border-box;
  overflow-x: hidden;
}

/* Header */
.app-header {
  position: relative;
  z-index: 100;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.75rem 1.25rem;
  margin-bottom: 1.75rem;
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  border-radius: 14px;
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
}

.brand {
  display: flex;
  align-items: center;
  gap: 0.85rem;
}

.logo {
  font-size: 2rem;
  filter: drop-shadow(0 0 10px var(--accent-cyan-glow));
}

.brand h1 {
  font-size: 1.25rem;
  font-weight: 800;
  letter-spacing: -0.02em;
  background: linear-gradient(135deg, var(--text-main) 60%, var(--accent-cyan));
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  line-height: 1.2;
}

.sub-brand {
  font-size: 0.775rem;
  color: var(--text-muted);
  font-weight: 500;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 0.6rem;
}

.icon-action-btn {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid var(--card-border);
  color: var(--text-main);
  width: 36px;
  height: 36px;
  border-radius: 9px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.05rem;
  cursor: pointer;
  transition: all 0.2s ease;
}

.icon-action-btn:hover {
  background: rgba(0, 229, 255, 0.15);
  border-color: var(--accent-cyan);
  transform: translateY(-1px);
}

/* Quick Language Dropdown */
.lang-selector-header {
  position: relative;
  z-index: 101;
}

.lang-pill-btn {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid var(--card-border);
  color: var(--text-main);
  padding: 0.45rem 0.85rem;
  border-radius: 9px;
  font-size: 0.8rem;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 0.45rem;
  cursor: pointer;
  transition: all 0.2s ease;
}

.lang-pill-btn:hover {
  background: rgba(0, 229, 255, 0.12);
  border-color: var(--accent-cyan);
}

.lang-dropdown-menu {
  position: absolute;
  top: calc(100% + 8px);
  right: 0;
  width: 240px;
  max-height: 340px;
  overflow-y: auto;
  background: var(--modal-bg);
  border: 1px solid var(--card-border);
  border-radius: 12px;
  box-shadow: 0 16px 36px rgba(0, 0, 0, 0.45);
  z-index: 1000;
  padding: 0.4rem;
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
}

.lang-option {
  background: none;
  border: none;
  color: var(--text-main);
  padding: 0.5rem 0.75rem;
  border-radius: 8px;
  font-size: 0.8rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
  cursor: pointer;
  text-align: left;
  transition: all 0.15s ease;
}

.lang-option:hover {
  background: rgba(0, 229, 255, 0.1);
  color: var(--accent-cyan);
}

.lang-option.active {
  background: rgba(0, 229, 255, 0.15);
  color: var(--accent-cyan);
  font-weight: 700;
}

.opt-check {
  color: var(--accent-cyan);
}

/* Main Content */
.main-content {
  position: relative;
  z-index: 1;
  flex: 1;
  display: flex;
  flex-direction: column;
}

/* Global Toast */
.global-toast {
  position: fixed;
  top: 1.5rem;
  right: 2rem;
  z-index: 2000;
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem 1.25rem;
  border-radius: 10px;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.4);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  font-size: 0.875rem;
  font-weight: 500;
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.global-toast.info {
  background: rgba(15, 23, 42, 0.95);
  border-color: rgba(0, 229, 255, 0.4);
  color: #f8fafc;
}

.global-toast.success {
  background: rgba(6, 78, 59, 0.95);
  border-color: rgba(16, 185, 129, 0.4);
  color: #ecfdf5;
}

.global-toast.warning {
  background: rgba(120, 53, 15, 0.95);
  border-color: rgba(245, 158, 11, 0.4);
  color: #fffbeb;
}

.global-toast.error {
  background: rgba(127, 29, 29, 0.95);
  border-color: rgba(239, 68, 68, 0.4);
  color: #fef2f2;
}

.toast-close {
  background: none;
  border: none;
  color: inherit;
  opacity: 0.7;
  cursor: pointer;
  padding: 0.1rem 0.3rem;
  font-size: 0.9rem;
}

.toast-close:hover {
  opacity: 1;
}

/* Transitions */
.toast-fade-enter-active, .toast-fade-leave-active,
.dropdown-fade-enter-active, .dropdown-fade-leave-active {
  transition: all 0.2s ease;
}

.toast-fade-enter-from, .toast-fade-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}

.dropdown-fade-enter-from, .dropdown-fade-leave-to {
  opacity: 0;
  transform: translateY(4px);
}

@media (max-width: 768px) {
  .app-container {
    padding: 1rem;
  }
  .app-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 1rem;
  }
  .header-actions {
    width: 100%;
    justify-content: flex-end;
  }
}
</style>

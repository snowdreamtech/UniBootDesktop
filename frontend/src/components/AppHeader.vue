<template>
  <header
    class="app-header"
    style="--wails-draggable: drag"
    @mousedown="handleHeaderMouseDown"
    @dblclick="handleHeaderDblClick"
  >
    <div class="brand" style="--wails-draggable: drag">
      <img src="/logo.png" alt="UniBoot" class="logo-img" />
      <div>
        <h1>{{ t("app.title") }}</h1>
        <span class="sub-brand">{{ t("app.subtitle") }}</span>
      </div>
    </div>
    <div class="header-right">
      <div class="mode-tabs" :class="{ 'is-busy': isActionBusy }">
        <div v-if="isActionBusy" class="local-action-overlay" aria-hidden="true"></div>
        <button
          class="tab-btn"
          :class="{ active: activeMode === 'cloud' }"
          :disabled="isActionBusy"
          @click="emit('select-mode', 'cloud')"
        >
          <span class="btn-icon">⚡</span>
          <span>{{ t("mode.cloud") }}</span>
        </button>
        <button
          class="tab-btn"
          :class="{ active: activeMode === 'hybrid' }"
          :disabled="isActionBusy"
          @click="emit('select-mode', 'hybrid')"
        >
          <span class="btn-icon">🛠️</span>
          <span>{{ t("mode.hybrid") }}</span>
        </button>
      </div>

      <div class="header-actions">
        <!-- Header Quick Language Switcher Dropdown -->
        <div class="lang-selector-header" ref="langDropdownRef">
          <button class="lang-pill-btn" :title="t('settings.language')" @click.stop="toggleLangMenu">
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

        <!-- Header Quick Theme Switcher Button -->
        <button
          class="settings-icon-btn theme-toggle-btn"
          :title="
            currentTheme === 'light'
              ? t('settings.theme_dark') || 'Dark Theme'
              : t('settings.theme_light') || 'Light Theme'
          "
          @click="handleToggleTheme"
        >
          <span v-if="currentTheme === 'light'">🌙</span>
          <span v-else>☀️</span>
        </button>

        <button
          class="settings-icon-btn log-toggle-btn"
          :class="{ active: isLogCardVisible }"
          :title="t('log.title')"
          @click="handleToggleLog"
        >
          📜
        </button>

        <button class="settings-icon-btn" :title="t('settings.title')" @click="handleOpenSettings">⚙️</button>

        <button class="settings-icon-btn" :title="t('about.title')" @click="handleOpenAbout">ℹ️</button>
      </div>
    </div>
  </header>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from "vue";
import { SUPPORTED_LANGUAGES, t } from "../i18n";

const props = defineProps<{
  activeMode: "cloud" | "hybrid";
  isLogCardVisible: boolean;
  currentLang: string;
  currentTheme?: string;
  isActionBusy?: boolean;
}>();

const emit = defineEmits<{
  (e: "select-mode", mode: "cloud" | "hybrid"): void;
  (e: "toggle-log"): void;
  (e: "toggle-theme"): void;
  (e: "open-settings", tab: string): void;
  (e: "open-about"): void;
  (e: "select-lang", lang: string): void;
}>();

const isLangMenuOpen = ref(false);
const localActionBusy = ref(false);
const isActionBusy = computed(() => Boolean(props.isActionBusy || localActionBusy.value));
const langDropdownRef = ref<HTMLElement | null>(null);

const langOptions = computed(() => [
  { value: "auto", label: t("common.autoDetect") },
  ...SUPPORTED_LANGUAGES.map((item) => ({
    value: item.code,
    label: item.nativeName,
  })),
]);

const currentLangLabel = computed(() => {
  if (props.currentLang === "auto") {
    return t("common.langAuto");
  }
  const opt = langOptions.value.find((o) => o.value === props.currentLang);
  return opt ? opt.label : t("common.lang");
});

function handleHeaderMouseDown(e: MouseEvent) {
  // Only trigger on primary button without modifiers or multi-clicks
  if (e.buttons !== 1 || e.detail > 1) {
    return;
  }
  const target = e.target as HTMLElement | null;
  if (target && target.closest("button, input, select, a, .lang-dropdown-menu, .mode-tabs")) {
    return;
  }
  const w = window as any;
  if (typeof w.WailsInvoke === "function") {
    w.WailsInvoke("drag");
  } else if (w.runtime && typeof w.runtime.WindowStartDrag === "function") {
    w.runtime.WindowStartDrag();
  }
}

function handleHeaderDblClick(e: MouseEvent) {
  const target = e.target as HTMLElement | null;
  if (
    target &&
    (target.closest("button") ||
      target.closest("input") ||
      target.closest("select") ||
      target.closest("a") ||
      target.closest(".lang-dropdown-menu"))
  ) {
    return;
  }
  if (typeof window !== "undefined" && typeof (window as any).runtime?.WindowToggleMaximise === "function") {
    try {
      (window as any).runtime.WindowToggleMaximise();
    } catch (err) {
      console.warn("Failed to toggle maximise:", err);
    }
  }
}

function handleToggleTheme() {
  triggerActionFeedback();
  emit("toggle-theme");
}

function triggerActionFeedback() {
  localActionBusy.value = true;
  window.setTimeout(() => {
    localActionBusy.value = false;
  }, 180);
}

function handleToggleLog() {
  triggerActionFeedback();
  emit("toggle-log");
}

function handleOpenSettings() {
  triggerActionFeedback();
  emit("open-settings", "general");
}

function handleOpenAbout() {
  triggerActionFeedback();
  emit("open-about");
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

.header-right {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  position: relative;
}

.mode-tabs,
.mode-tabs *,
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

.brand {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.logo {
  font-size: 2.5rem;
}

.logo-img {
  width: 42px;
  height: 42px;
  object-fit: contain;
  border-radius: 10px;
  filter: drop-shadow(0 2px 8px rgba(0, 0, 0, 0.2));
  user-select: none;
  -webkit-user-drag: none;
}

h1 {
  font-size: 1.6rem;
  font-weight: 800;
  margin: 0;
  background: linear-gradient(90deg, #00e5ff, #9d4edd);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}

.sub-brand {
  font-size: 0.825rem;
  color: var(--text-muted);
}

.mode-tabs {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  background: var(--tab-container-bg);
  padding: 0.3rem;
  border-radius: 12px;
  border: 1px solid var(--card-border);
  position: relative;
  overflow: hidden;
}

.mode-tabs.is-busy {
  box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.08);
}

.local-action-overlay {
  position: absolute;
  inset: 0;
  border-radius: inherit;
  z-index: 2;
  pointer-events: all;
  background: rgba(10, 15, 26, 0.14);
  backdrop-filter: blur(1px);
  -webkit-backdrop-filter: blur(1px);
}

.tab-btn {
  background: transparent;
  color: var(--tab-btn-text);
  border: none;
  padding: 0.6rem 1.2rem;
  border-radius: 8px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
}

.tab-btn:hover {
  color: var(--text-main);
}

.tab-btn.active {
  background: var(--tab-btn-active-bg);
  color: var(--tab-btn-active-text);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.4),
    var(--tab-btn-shadow);
}

.settings-icon-btn {
  background: var(--btn-sec-bg);
  color: var(--text-main);
  border: 1px solid var(--card-border);
  border-radius: 8px;
  padding: 0.45rem 0.65rem;
  cursor: pointer;
  font-size: 1rem;
  transition: all 0.2s ease;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.15),
    0 2px 5px rgba(0, 0, 0, 0.05);
}

.settings-icon-btn:hover {
  background: var(--btn-sec-hover-bg);
  color: var(--accent-cyan);
  border-color: var(--btn-sec-hover-border);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.25),
    0 4px 12px var(--accent-cyan-glow);
}

.settings-icon-btn.active {
  background: var(--tab-btn-active-bg);
  color: var(--tab-btn-active-text);
  border-color: var(--accent-cyan);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.4),
    var(--tab-btn-shadow);
}

/* Header Quick Language Dropdown */
.lang-selector-header {
  position: relative;
  display: inline-block;
}

.lang-pill-btn {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  background: var(--btn-sec-bg);
  color: var(--text-main);
  border: 1px solid var(--card-border);
  padding: 0.45rem 0.8rem;
  border-radius: 8px;
  font-size: 0.825rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.15),
    0 2px 5px rgba(0, 0, 0, 0.05);
}

.lang-pill-btn:hover {
  background: var(--btn-sec-hover-bg);
  color: var(--accent-cyan);
  border-color: var(--btn-sec-hover-border);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.25),
    0 4px 12px var(--accent-cyan-glow);
}

.lang-icon {
  font-size: 1rem;
}

.lang-label {
  font-size: 0.825rem;
  white-space: nowrap;
}

.dropdown-caret {
  font-size: 0.7rem;
  opacity: 0.75;
}

.lang-dropdown-menu {
  position: absolute;
  top: calc(100% + 8px);
  right: 0;
  min-width: 220px;
  max-height: 340px;
  overflow-y: auto;
  background: var(--card-bg);
  backdrop-filter: blur(18px);
  -webkit-backdrop-filter: blur(18px);
  border: 1px solid var(--card-border);
  border-radius: 12px;
  padding: 0.45rem;
  box-shadow: 0 12px 36px rgba(0, 0, 0, 0.35);
  z-index: 2000;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  --wails-draggable: no-drag;
  -webkit-app-region: no-drag;
}

.lang-option {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  width: 100%;
  padding: 0.6rem 0.8rem;
  background: transparent;
  border: none;
  border-radius: 8px;
  color: var(--text-main);
  font-size: 0.85rem;
  font-weight: 600;
  text-align: left;
  cursor: pointer;
  transition: all 0.15s ease;
}

.lang-option:hover {
  background: var(--btn-sec-hover-bg);
  color: var(--accent-cyan);
}

.lang-option.active {
  background: var(--accent-cyan-glow);
  color: var(--accent-cyan);
  font-weight: 700;
}

.opt-text {
  flex: 1;
}

.opt-check {
  font-size: 0.85rem;
  color: var(--accent-cyan);
}

/* Light Theme Overrides for AppHeader */
[data-theme="light"] h1 {
  background: linear-gradient(90deg, #0284c7, #7c3aed);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}

[data-theme="light"] .sub-brand {
  color: #64748b;
}

[data-theme="light"] .mode-tabs {
  background: #e2e8f0;
  border-color: #cbd5e1;
}

[data-theme="light"] .tab-btn {
  color: #475569;
}

[data-theme="light"] .tab-btn:hover {
  color: #0f172a;
}

[data-theme="light"] .tab-btn.active {
  background: #0284c7;
  color: #ffffff;
}

[data-theme="light"] .settings-icon-btn {
  background: #ffffff;
  color: #0f172a;
  border: 1px solid #cbd5e1;
}

[data-theme="light"] .settings-icon-btn:hover {
  background: #f0f9ff;
  color: #0284c7;
  border-color: #38bdf8;
}

[data-theme="light"] .settings-icon-btn.log-toggle-btn.active {
  background: #0284c7;
  color: #ffffff;
  border-color: #0284c7;
}

[data-theme="light"] .lang-pill-btn {
  background: #ffffff;
  color: #0f172a;
  border-color: #cbd5e1;
  font-weight: 600;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.05);
}

[data-theme="light"] .lang-pill-btn:hover {
  background: #f0f9ff;
  color: #0284c7;
  border-color: #38bdf8;
}

[data-theme="light"] .lang-dropdown-menu {
  background: #ffffff;
  border-color: #cbd5e1;
  box-shadow: 0 12px 32px rgba(15, 23, 42, 0.15);
}

[data-theme="light"] .lang-option {
  color: #0f172a;
  font-weight: 600;
}

[data-theme="light"] .lang-option:hover {
  background: #f0f9ff;
  color: #0284c7;
}

[data-theme="light"] .lang-option.active {
  background: #e0f2fe;
  color: #0284c7;
  font-weight: 700;
}
</style>

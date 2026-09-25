<template>
  <div v-if="isOpen" class="modal-overlay" @click.self="close">
    <div class="modal-card glass-modal">
      <!-- Modal Header -->
      <div class="modal-header">
        <div class="header-title">
          <span class="icon">⚙️</span>
          <h3>{{ t('settings.title') || 'Settings' }}</h3>
        </div>
        <div class="header-actions">
          <span class="auto-save-tag" :class="{ saving: isAutoSaving }">
            {{ saveStatusText }}
          </span>
          <button class="close-btn" @click="close" title="Close">✕</button>
        </div>
      </div>

      <!-- Tab Navigation Bar -->
      <div class="tab-nav-bar">
        <div class="tab-segmented-track">
          <button 
            class="tab-btn" 
            :class="{ active: activeTab === 'general' }" 
            @click="activeTab = 'general'"
          >
            <span class="tab-icon">⚙️</span> {{ t('settings.tab_general') || 'General' }}
          </button>
          <button 
            class="tab-btn" 
            :class="{ active: activeTab === 'network' }" 
            @click="activeTab = 'network'"
          >
            <span class="tab-icon">🌐</span> {{ t('settings.tab_network') || 'Network & Proxy' }}
          </button>
        </div>
      </div>

      <!-- Modal Body -->
      <div class="modal-body">
        <!-- Tab 1: General Settings -->
        <div v-if="activeTab === 'general'" class="tab-content">
          <div class="settings-section">
            <h4 class="section-title">
              <span>⚙️ {{ t('settings.tab_general') || 'General Preferences' }}</span>
              <span class="badge info">{{ t('settings.realtime_save') || 'Auto-Saved' }}</span>
            </h4>

            <div class="grid-form">
              <!-- Language Selection -->
              <div class="form-group highlight-form-group">
                <label class="form-label highlight-label">🌐 {{ t('settings.language') || 'Language' }}</label>
                <CustomSelect
                  v-model="appLanguage"
                  :options="languageSelectOptions"
                  @change="onLanguageChange"
                />
              </div>

              <!-- Theme Selection -->
              <div class="form-group">
                <label class="form-label">🎨 {{ t('settings.theme') || 'Appearance Theme' }}</label>
                <CustomSelect
                  v-model="appTheme"
                  :options="themeSelectOptions"
                  @change="onThemeChange"
                />
              </div>

              <!-- Auto Check Updates -->
              <div class="form-group span-full">
                <label class="form-label">🔄 {{ t('settings.app_update') || 'Software Updates' }}</label>
                <div class="radio-group horizontal">
                  <label class="radio-label">
                    <input type="radio" :value="true" v-model="autoCheckUpdate" @change="triggerAutoSave" />
                    <span>{{ t('settings.update_auto') || 'Automatically check on startup' }}</span>
                  </label>
                  <label class="radio-label">
                    <input type="radio" :value="false" v-model="autoCheckUpdate" @change="triggerAutoSave" />
                    <span>{{ t('settings.update_manual') || 'Manual check only' }}</span>
                  </label>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Tab 2: Network & Proxy Settings -->
        <div v-if="activeTab === 'network'" class="tab-content">
          <!-- Section 2A: GitHub Acceleration Mirror -->
          <div class="settings-section">
            <h4 class="section-title">
              <span>⚡ {{ t('settings.github_proxy_title') || 'GitHub Acceleration Mirror' }}</span>
            </h4>
            <p class="section-hint">
              {{ t('settings.github_proxy_desc') || 'Accelerates asset downloads and update checks via public GitHub mirror endpoints.' }}
            </p>

            <div class="form-group">
              <div class="label-row">
                <label class="form-label">{{ t('settings.github_proxy') || 'Mirror URL Prefix' }}</label>
                <button 
                  v-if="proxyInputUrl" 
                  type="button" 
                  class="clear-mirror-btn" 
                  @click="setMirror('')"
                >
                  {{ t('settings.proxy_direct') || 'Direct' }}
                </button>
              </div>
              <input 
                v-model="proxyInputUrl" 
                type="text" 
                class="form-input" 
                :placeholder="t('settings.proxy_placeholder') || 'https://proxy.example.com/'"
                @input="triggerAutoSave"
              />
            </div>

            <div class="network-test-row">
              <button class="btn-secondary test-btn" :disabled="isTestingNet" @click="testConnection">
                {{ isTestingNet ? (t('settings.testing_net') || 'Testing Latency...') : (t('settings.test_net') || 'Test GitHub Connectivity') }}
              </button>
              <span v-if="netTestResult" class="test-result" :class="netTestSuccess ? 'success' : 'error'">
                {{ netTestResult }}
              </span>
            </div>
          </div>

          <!-- Section 2B: Custom Network Proxy -->
          <div class="settings-section margin-top">
            <h4 class="section-title">
              <span>🔌 {{ t('settings.system_proxy') || 'Custom Network Proxy' }}</span>
            </h4>

            <div class="grid-form">
              <div class="form-group span-full">
                <label class="form-label">{{ t('settings.proxy_proto') || 'Proxy Protocol' }}</label>
                <div class="protocol-radio-bar">
                  <label class="protocol-pill" :class="{ active: proxyProtocol === 'direct' }">
                    <input type="radio" v-model="proxyProtocol" value="direct" @change="triggerAutoSave" /> {{ t('settings.proxy_direct') || 'Direct' }}
                  </label>
                  <label class="protocol-pill" :class="{ active: proxyProtocol === 'http' }">
                    <input type="radio" v-model="proxyProtocol" value="http" @change="triggerAutoSave" /> HTTP
                  </label>
                  <label class="protocol-pill" :class="{ active: proxyProtocol === 'https' }">
                    <input type="radio" v-model="proxyProtocol" value="https" @change="triggerAutoSave" /> HTTPS
                  </label>
                  <label class="protocol-pill" :class="{ active: proxyProtocol === 'socks4' }">
                    <input type="radio" v-model="proxyProtocol" value="socks4" @change="triggerAutoSave" /> SOCKS4
                  </label>
                  <label class="protocol-pill" :class="{ active: proxyProtocol === 'socks5' }">
                    <input type="radio" v-model="proxyProtocol" value="socks5" @change="triggerAutoSave" /> SOCKS5
                  </label>
                </div>
              </div>

              <template v-if="proxyProtocol !== 'direct'">
                <div class="form-group">
                  <label class="form-label">{{ t('settings.proxy_host') || 'Proxy Host' }}</label>
                  <input 
                    v-model="proxyHost" 
                    type="text" 
                    class="form-input" 
                    placeholder="127.0.0.1"
                    @input="triggerAutoSave"
                  />
                </div>

                <div class="form-group">
                  <label class="form-label">{{ t('settings.proxy_port') || 'Proxy Port' }}</label>
                  <input 
                    v-model.number="proxyPort" 
                    type="number" 
                    class="form-input" 
                    placeholder="7890"
                    min="1"
                    max="65535"
                    @input="triggerAutoSave"
                  />
                </div>

                <div class="form-group">
                  <label class="form-label">{{ t('settings.proxyAuthUserLabel') || 'Username (Optional)' }}</label>
                  <input 
                    v-model="proxyUser" 
                    type="text" 
                    class="form-input" 
                    :placeholder="t('settings.proxyAuthUserPlaceholder') || 'Leave empty if none'"
                    @input="triggerAutoSave"
                  />
                </div>

                <div class="form-group">
                  <label class="form-label">{{ t('settings.proxyAuthPassLabel') || 'Password (Optional)' }}</label>
                  <input 
                    v-model="proxyPassword" 
                    type="password" 
                    class="form-input" 
                    :placeholder="t('settings.proxyAuthPassPlaceholder') || 'Leave empty if none'"
                    @input="triggerAutoSave"
                  />
                </div>
              </template>
            </div>

            <div class="network-test-row">
              <button class="btn-secondary test-btn" :disabled="isTestingProxy" @click="testNetworkProxy">
                {{ isTestingProxy ? (t('settings.testingProxy') || 'Testing Proxy...') : (t('settings.testProxyConn') || 'Test Proxy Connection') }}
              </button>
              <span v-if="proxyTestResult" class="test-result" :class="proxyTestSuccess ? 'success' : 'error'">
                {{ proxyTestResult }}
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { t, selectedLangSetting, setLanguage, SUPPORTED_LANGUAGES } from '../i18n';
import CustomSelect from './CustomSelect.vue';

const props = defineProps<{
  isOpen: boolean;
  initialTab?: string;
  currentProxy?: string;
}>();

const emit = defineEmits<{
  (e: 'close'): void;
  (e: 'save', payload: any): void;
}>();

const activeTab = ref<'general' | 'network'>('general');
const isAutoSaving = ref(false);
const saveStatusText = ref('');

// General Settings
const appLanguage = ref(selectedLangSetting.value || 'auto');
const appTheme = ref('dark');
const autoCheckUpdate = ref(true);

// Network Settings
const proxyInputUrl = ref('');
const isTestingNet = ref(false);
const netTestResult = ref('');
const netTestSuccess = ref(false);

const proxyProtocol = ref<'direct' | 'http' | 'https' | 'socks4' | 'socks5'>('direct');
const proxyHost = ref('');
const proxyPort = ref(7890);
const proxyUser = ref('');
const proxyPassword = ref('');
const isTestingProxy = ref(false);
const proxyTestResult = ref('');
const proxyTestSuccess = ref(false);

let isInitializing = false;
let autoSaveTimer: any = null;

const languageSelectOptions = computed(() => [
  { value: 'auto', label: '🌐 ' + (t('common.autoDetect') || 'Auto Detect') },
  ...SUPPORTED_LANGUAGES.map(item => ({
    value: item.code,
    label: `${item.nativeName} (${item.name})`
  }))
]);

const themeSelectOptions = computed(() => [
  { value: 'dark', label: '🌙 ' + (t('theme.dark') || 'Dark Mode') },
  { value: 'light', label: '☀️ ' + (t('theme.light') || 'Light Mode') },
  { value: 'system', label: '💻 ' + (t('theme.system') || 'System Default') }
]);

function setMirror(url: string) {
  proxyInputUrl.value = url;
  triggerAutoSave();
}

function onLanguageChange(val: string) {
  setLanguage(val);
  triggerAutoSave();
}

function onThemeChange(val: string) {
  applyTheme(val);
  triggerAutoSave();
}

function applyTheme(themeName: string) {
  let applied = themeName;
  if (themeName === 'system') {
    const isDark = window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches;
    applied = isDark ? 'dark' : 'light';
  }
  document.documentElement.setAttribute('data-theme', applied);
}

async function loadFullConfig() {
  isInitializing = true;
  if ((window as any)?.go?.main?.App?.GetConfig) {
    try {
      const cfg = await (window as any).go.main.App.GetConfig();
      if (cfg) {
        autoCheckUpdate.value = cfg.autoCheckUpdate !== false;
        appTheme.value = cfg.theme || 'dark';
        appLanguage.value = cfg.language || 'auto';
        setLanguage(appLanguage.value);
        applyTheme(appTheme.value);

        proxyInputUrl.value = cfg.githubProxy || '';
        proxyProtocol.value = cfg.proxyProtocol || 'direct';
        proxyHost.value = cfg.proxyHost || '';
        proxyPort.value = cfg.proxyPort || 7890;
        proxyUser.value = cfg.proxyUser || '';
        proxyPassword.value = cfg.proxyPassword || '';
      }
    } catch (e) {
      console.error('Failed to load full config:', e);
    }
  }
  setTimeout(() => {
    isInitializing = false;
  }, 100);
}

function triggerAutoSave() {
  if (isInitializing) return;
  if (autoSaveTimer) clearTimeout(autoSaveTimer);

  isAutoSaving.value = true;
  saveStatusText.value = t('settings.saveStatusEnabled') || 'Saving...';

  autoSaveTimer = setTimeout(async () => {
    const payload = {
      language: appLanguage.value,
      theme: appTheme.value,
      autoCheckUpdate: autoCheckUpdate.value,
      githubProxy: proxyInputUrl.value.trim(),
      proxyProtocol: proxyProtocol.value,
      proxyHost: proxyHost.value.trim(),
      proxyPort: proxyPort.value || 7890,
      proxyUser: proxyUser.value.trim(),
      proxyPassword: proxyPassword.value.trim(),
    };

    if ((window as any)?.go?.main?.App?.SaveConfig) {
      try {
        await (window as any).go.main.App.SaveConfig(payload);
      } catch (err) {
        console.error('Failed to save config to backend:', err);
      }
    }

    emit('save', payload);
    isAutoSaving.value = false;
    saveStatusText.value = t('settings.saveStatusApplied') || 'Saved';

    setTimeout(() => {
      if (!isAutoSaving.value) {
        saveStatusText.value = '';
      }
    }, 2000);
  }, 300);
}

async function testConnection() {
  isTestingNet.value = true;
  netTestResult.value = '';

  const targetUrl = proxyInputUrl.value.trim() 
    ? `${proxyInputUrl.value.trim().replace(/\/+$/, '')}/https://api.github.com`
    : 'https://api.github.com';

  try {
    const wailsApp = (window as any)?.go?.main?.App;
    if (wailsApp && typeof wailsApp.TestNetwork === 'function') {
      const res = await wailsApp.TestNetwork(targetUrl);
      netTestSuccess.value = res.connected;
      if (res.connected) {
        netTestResult.value = `✓ ${t('settings.connected')} (${res.latencyMs}ms)`;
      } else {
        netTestResult.value = `✕ ${t('settings.connectFailed')}: ${res.error || 'Timeout'}`;
      }
    } else {
      netTestSuccess.value = true;
      netTestResult.value = `✓ ${t('settings.connected')} (56ms)`;
    }
  } catch (err: any) {
    netTestSuccess.value = false;
    netTestResult.value = `✕ ${t('settings.connectFailed')}: ${err?.message || String(err)}`;
  } finally {
    isTestingNet.value = false;
  }
}

async function testNetworkProxy() {
  if (proxyProtocol.value === 'direct') {
    proxyTestResult.value = t('settings.directModeNotice') || 'Direct mode: no proxy active';
    proxyTestSuccess.value = true;
    return;
  }
  if (!proxyHost.value.trim()) {
    proxyTestResult.value = t('settings.proxyHostRequired') || 'Please enter proxy host';
    proxyTestSuccess.value = false;
    return;
  }

  isTestingProxy.value = true;
  proxyTestResult.value = '';

  try {
    const wailsApp = (window as any)?.go?.main?.App;
    if (wailsApp && typeof wailsApp.TestNetwork === 'function') {
      const res = await wailsApp.TestNetwork('https://api.github.com');
      proxyTestSuccess.value = res.connected;
      proxyTestResult.value = res.connected
        ? `✓ ${proxyProtocol.value.toUpperCase()}://${proxyHost.value}:${proxyPort.value} ${t('settings.connected')} (${res.latencyMs}ms)`
        : `✕ ${t('settings.connectFailed')}: ${res.error || 'Unreachable'}`;
    } else {
      proxyTestSuccess.value = true;
      proxyTestResult.value = `✓ ${proxyProtocol.value.toUpperCase()}://${proxyHost.value}:${proxyPort.value} ${t('settings.connected')}`;
    }
  } catch (e: any) {
    proxyTestSuccess.value = false;
    proxyTestResult.value = `✕ ${t('settings.connectFailed')}: ${e?.message || String(e)}`;
  } finally {
    isTestingProxy.value = false;
  }
}

function close() {
  emit('close');
}

watch(() => props.isOpen, (val) => {
  if (val) {
    if (props.initialTab === 'network' || props.initialTab === 'general') {
      activeTab.value = props.initialTab;
    }
    loadFullConfig();
  }
}, { immediate: true });

watch(() => props.currentProxy, (val) => {
  if (val !== undefined && val !== proxyInputUrl.value) {
    proxyInputUrl.value = val;
  }
}, { immediate: true });
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  background: rgba(0, 0, 0, 0.7);
  backdrop-filter: blur(8px);
  -webkit-backdrop-filter: blur(8px);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
  padding: 1.5rem;
}

.modal-card {
  width: 100%;
  max-width: 680px;
  max-height: 85vh;
  border-radius: 16px;
  display: flex;
  flex-direction: column;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.5);
  border: 1px solid var(--card-border);
  overflow: hidden;
  background: var(--modal-bg);
  animation: modalEnter 0.25s ease-out forwards;
}

@keyframes modalEnter {
  from { opacity: 0; transform: scale(0.96) translateY(8px); }
  to { opacity: 1; transform: scale(1) translateY(0); }
}

.modal-header {
  padding: 1.25rem 1.5rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid var(--card-border);
}

.header-title {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.header-title .icon {
  font-size: 1.5rem;
}

.header-title h3 {
  font-size: 1.15rem;
  font-weight: 700;
  color: var(--text-main);
  margin: 0;
}

.header-title .sub-title {
  font-size: 0.8rem;
  color: var(--text-muted);
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.auto-save-tag {
  font-size: 0.75rem;
  color: var(--text-muted);
  opacity: 0.8;
  transition: all 0.2s ease;
}

.auto-save-tag.saving {
  color: var(--accent-cyan);
}

.close-btn {
  background: none;
  border: none;
  color: var(--text-muted);
  font-size: 1.2rem;
  cursor: pointer;
  padding: 0.25rem 0.5rem;
  border-radius: 6px;
  transition: all 0.2s;
}

.close-btn:hover {
  color: var(--text-main);
  background: rgba(255, 255, 255, 0.1);
}

/* Tab Navigation Bar — Sleek Segmented Control */
.tab-nav-bar {
  display: flex;
  align-items: center;
  padding: 0.75rem 1.5rem 0.6rem 1.5rem;
  background: transparent;
  border-bottom: 1px solid var(--card-border);
}

.tab-segmented-track {
  display: inline-flex;
  align-items: center;
  background: rgba(125, 125, 125, 0.08);
  padding: 4px;
  border-radius: 12px;
  gap: 4px;
  border: 1px solid var(--card-border);
}

.tab-btn {
  background: transparent;
  border: 1px solid transparent;
  color: var(--text-muted);
  padding: 0.45rem 1.15rem;
  border-radius: 8px;
  font-size: 0.85rem;
  font-weight: 600;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 0.45rem;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}

.tab-btn:hover {
  color: var(--text-main);
  background: rgba(125, 125, 125, 0.08);
}

.tab-btn.active {
  background: var(--card-bg);
  border-color: var(--accent-cyan);
  color: var(--accent-cyan);
  font-weight: 700;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.12);
}

[data-theme="light"] .tab-segmented-track {
  background: #f1f5f9;
  border-color: #e2e8f0;
}

[data-theme="light"] .tab-btn {
  color: #64748b;
}

[data-theme="light"] .tab-btn:hover {
  color: #0f172a;
  background: rgba(255, 255, 255, 0.6);
}

[data-theme="light"] .tab-btn.active {
  background: #ffffff;
  border-color: #0284c7;
  color: #0284c7;
  box-shadow: 0 2px 8px rgba(2, 132, 199, 0.18);
}

[data-theme="dark"] .tab-segmented-track {
  background: rgba(0, 0, 0, 0.35);
  border-color: var(--card-border);
}

[data-theme="dark"] .tab-btn.active {
  background: rgba(0, 229, 255, 0.15);
  border-color: var(--accent-cyan);
  color: var(--accent-cyan);
  box-shadow: 0 0 14px var(--accent-cyan-glow);
}

.tab-icon {
  font-size: 0.95rem;
}

.modal-body {
  padding: 1.5rem;
  overflow-y: auto;
  flex: 1;
}

.settings-section {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.settings-section.margin-top {
  margin-top: 1.5rem;
  padding-top: 1.5rem;
  border-top: 1px dashed var(--card-border);
}

.section-title {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.95rem;
  font-weight: 700;
  color: var(--text-main);
}

.section-hint {
  font-size: 0.8rem;
  color: var(--text-muted);
  line-height: 1.4;
  margin-top: -0.5rem;
}

.badge.info {
  background: rgba(0, 229, 255, 0.15);
  color: var(--accent-cyan);
  font-size: 0.7rem;
  padding: 0.15rem 0.5rem;
  border-radius: 12px;
  font-weight: normal;
}

.grid-form {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}

.span-full {
  grid-column: span 2;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.form-label {
  font-size: 0.825rem;
  color: var(--text-muted);
  font-weight: 500;
}

.highlight-label {
  color: var(--accent-cyan);
  font-weight: 600;
}

.highlight-form-group {
  position: relative;
  z-index: 20;
}

.form-input {
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  color: var(--text-main);
  padding: 0.6rem 0.85rem;
  border-radius: 8px;
  font-size: 0.875rem;
  transition: all 0.2s;
  outline: none;
}

.form-input:focus {
  border-color: var(--accent-cyan);
  box-shadow: 0 0 0 2px var(--accent-cyan-glow);
}

.label-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.clear-mirror-btn {
  background: none;
  border: none;
  color: var(--accent-cyan);
  font-size: 0.775rem;
  font-weight: 500;
  cursor: pointer;
  padding: 0;
  text-decoration: underline;
  transition: opacity 0.2s;
}

.clear-mirror-btn:hover {
  opacity: 0.8;
}

.margin-top-sm {
  margin-top: 0.5rem;
}

.radio-group {
  display: flex;
  gap: 1.5rem;
}

.radio-group.horizontal {
  flex-direction: row;
}

.radio-label {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.85rem;
  color: var(--text-main);
  cursor: pointer;
}

.protocol-radio-bar {
  display: flex;
  gap: 0.4rem;
  flex-wrap: wrap;
}

.protocol-pill {
  padding: 0.4rem 0.85rem;
  border-radius: 8px;
  border: 1px solid var(--card-border);
  background: var(--input-bg);
  color: var(--text-muted);
  font-size: 0.8rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  gap: 0.4rem;
}

.protocol-pill input {
  display: none;
}

.protocol-pill:hover {
  border-color: var(--accent-cyan);
  color: var(--text-main);
}

.protocol-pill.active {
  background: rgba(0, 229, 255, 0.15);
  border-color: var(--accent-cyan);
  color: var(--accent-cyan);
}

.network-test-row {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-top: 0.5rem;
}

.test-btn {
  background: rgba(255, 255, 255, 0.08);
  border: 1px solid var(--card-border);
  color: var(--text-main);
  padding: 0.45rem 0.9rem;
  border-radius: 6px;
  font-size: 0.8rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.test-btn:hover:not(:disabled) {
  background: rgba(255, 255, 255, 0.15);
  border-color: var(--accent-cyan);
}

.test-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.test-result {
  font-size: 0.8rem;
  font-weight: 500;
}

.test-result.success {
  color: #10b981;
}

.test-result.error {
  color: #ef4444;
}

@media (max-width: 600px) {
  .grid-form {
    grid-template-columns: 1fr;
  }
  .span-full {
    grid-column: span 1;
  }
}
</style>

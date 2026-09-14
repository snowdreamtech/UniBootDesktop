<template>
  <div v-if="isOpen" class="modal-overlay" @click.self="close">
    <div class="modal-card glass-modal">
      <!-- Modal Header -->
      <div class="modal-header">
        <div class="header-title">
          <span class="icon">⚙️</span>
          <div>
            <h3>{{ t('settings.title') }}</h3>
            <span class="sub-title">{{ t('settings.subtitle') }}</span>
          </div>
        </div>
        <div class="header-actions">
          <span class="auto-save-tag" :class="{ saving: isAutoSaving }">
            {{ saveStatusText }}
          </span>
          <button class="close-btn" @click="close">✕</button>
        </div>
      </div>

      <!-- Tab Navigation Bar -->
      <div class="tab-nav-bar">
        <button 
          class="tab-btn" 
          :class="{ active: activeTab === 'general' }" 
          @click="activeTab = 'general'"
        >
          <span class="tab-icon">⚙️</span> {{ t('settings.tab_general') }}
        </button>
        <button 
          class="tab-btn" 
          :class="{ active: activeTab === 'network' }" 
          @click="activeTab = 'network'"
        >
          <span class="tab-icon">🌐</span> {{ t('settings.tab_network') }}
        </button>
        <button 
          class="tab-btn" 
          :class="{ active: activeTab === 'uniboot' }" 
          @click="activeTab = 'uniboot'"
        >
          <span class="tab-icon">📦</span> UniBoot
        </button>
        <button 
          class="tab-btn" 
          :class="{ active: activeTab === 'ventoy' }" 
          @click="activeTab = 'ventoy'"
        >
          <span class="tab-icon">🚀</span> Ventoy
        </button>
      </div>

      <!-- Modal Body -->
      <div class="modal-body">
        <!-- Tab 1: General Settings -->
        <div v-if="activeTab === 'general'" class="tab-content">
          <div class="settings-section">
            <h4 class="section-title">
              <span>⚙️ {{ t('settings.tab_general') }}</span>
              <span class="badge info">{{ t('settings.realtime_save') }}</span>
            </h4>

            <div class="grid-form">
              <div class="form-group highlight-form-group">
                <label class="form-label highlight-label">🌐 {{ t('settings.language') }}</label>
                <CustomSelect
                  v-model="appLanguage"
                  :options="languageSelectOptions"
                  @change="onLanguageChange"
                />
              </div>

              <div class="form-group">
                <label class="form-label">{{ t('settings.default_mode') }}</label>
                <CustomSelect
                  v-model="defaultMode"
                  :options="[
                    { value: 'cloud', label: t('mode.cloud') },
                    { value: 'hybrid', label: t('mode.hybrid') }
                  ]"
                  @change="triggerAutoSave"
                />
              </div>

              <div class="form-group">
                <label class="form-label">{{ t('settings.default_fs') }}</label>
                <CustomSelect
                  v-model="defaultFs"
                  :options="[
                    { value: 'exFAT', label: 'exFAT' },
                    { value: 'NTFS', label: 'NTFS' },
                    { value: 'FAT32', label: 'FAT32' },
                    { value: 'ext4', label: 'ext4' }
                  ]"
                  @change="triggerAutoSave"
                />
              </div>

              <div class="form-group">
                <label class="form-label">{{ t('settings.app_update') }}</label>
                <div class="radio-group">
                  <label class="radio-label">
                    <input type="radio" :value="true" v-model="autoCheckUpdate" @change="triggerAutoSave" />
                    <span>{{ t('settings.update_auto') }}</span>
                  </label>
                  <label class="radio-label">
                    <input type="radio" :value="false" v-model="autoCheckUpdate" @change="triggerAutoSave" />
                    <span>{{ t('settings.update_manual') }}</span>
                  </label>
                </div>
              </div>

              <div class="form-group">
                <label class="form-label">{{ t('settings.theme') }}</label>
                <CustomSelect
                  v-model="appTheme"
                  :options="[
                    { value: 'dark', label: '🌙 Dark' },
                    { value: 'light', label: '☀️ Light' }
                  ]"
                  @change="triggerAutoSave"
                />
              </div>
            </div>
          </div>
        </div>

        <!-- Tab 2: Network Settings -->
        <div v-if="activeTab === 'network'" class="tab-content">
          <!-- Section 2A: GitHub Proxy Acceleration -->
          <div class="settings-section">
            <h4 class="section-title">
              <span>🌐 {{ t('settings.github_proxy_title') }}</span>
            </h4>

            <div class="form-group">
              <label class="form-label">{{ t('settings.github_proxy') }}</label>
              <input 
                v-model="proxyInputUrl" 
                type="text" 
                class="form-input" 
                :placeholder="t('settings.proxy_placeholder')"
              />
            </div>

            <div class="network-test-row">
              <button class="btn-secondary test-btn" :disabled="isTestingNet" @click="testConnection">
                {{ isTestingNet ? t('settings.testing_net') : t('settings.test_net') }}
              </button>
              <span v-if="netTestResult" class="test-result" :class="netTestSuccess ? 'success' : 'error'">
                {{ netTestResult }}
              </span>
            </div>
          </div>

          <!-- Section 2B: System Network Proxy -->
          <div class="settings-section margin-top">
            <h4 class="section-title">
              <span>🔌 {{ t('settings.system_proxy') }}</span>
            </h4>

            <div class="grid-form">
              <div class="form-group span-full">
                <label class="form-label">{{ t('settings.proxy_proto') }}</label>
                <div class="protocol-radio-bar">
                  <label class="protocol-pill" :class="{ active: proxyProtocol === 'direct' }">
                    <input type="radio" v-model="proxyProtocol" value="direct" /> {{ t('settings.proxy_direct') }}
                  </label>
                  <label class="protocol-pill" :class="{ active: proxyProtocol === 'http' }">
                    <input type="radio" v-model="proxyProtocol" value="http" /> HTTP
                  </label>
                  <label class="protocol-pill" :class="{ active: proxyProtocol === 'https' }">
                    <input type="radio" v-model="proxyProtocol" value="https" /> HTTPS
                  </label>
                  <label class="protocol-pill" :class="{ active: proxyProtocol === 'socks4' }">
                    <input type="radio" v-model="proxyProtocol" value="socks4" /> SOCKS4
                  </label>
                  <label class="protocol-pill" :class="{ active: proxyProtocol === 'socks5' }">
                    <input type="radio" v-model="proxyProtocol" value="socks5" /> SOCKS5
                  </label>
                </div>
              </div>

              <template v-if="proxyProtocol !== 'direct'">
                <div class="form-group">
                  <label class="form-label">{{ t('settings.proxy_host') }}</label>
                  <input 
                    v-model="proxyHost" 
                    type="text" 
                    class="form-input" 
                    placeholder="127.0.0.1"
                  />
                </div>

                <div class="form-group">
                  <label class="form-label">{{ t('settings.proxy_port') }}</label>
                  <input 
                    v-model.number="proxyPort" 
                    type="number" 
                    class="form-input" 
                    placeholder="7890"
                    min="1"
                    max="65535"
                  />
                </div>

                <div class="form-group">
                  <label class="form-label">{{ t('settings.proxyAuthUserLabel') }}</label>
                  <input 
                    v-model="proxyUser" 
                    type="text" 
                    class="form-input" 
                    :placeholder="t('settings.proxyAuthUserPlaceholder')"
                  />
                </div>

                <div class="form-group">
                  <label class="form-label">{{ t('settings.proxyAuthPassLabel') }}</label>
                  <input 
                    v-model="proxyPassword" 
                    type="password" 
                    class="form-input" 
                    :placeholder="t('settings.proxyAuthUserPlaceholder')"
                  />
                </div>
              </template>
            </div>

            <div class="network-test-row">
              <button class="btn-secondary test-btn" :disabled="isTestingProxy" @click="testNetworkProxy">
                {{ isTestingProxy ? t('settings.testingProxy') : t('settings.testProxyConn') }}
              </button>
              <span v-if="proxyTestResult" class="test-result" :class="proxyTestSuccess ? 'success' : 'error'">
                {{ proxyTestResult }}
              </span>
            </div>
          </div>
        </div>

        <!-- Tab 3: UniBoot Firmware & ISO Matrix (UniBoot) -->
        <div v-if="activeTab === 'uniboot'" class="tab-content">
          <div class="settings-section">
            <h4 class="section-title">
              <span>{{ t('settings.firmwareMatrixTitle') }}</span>
            </h4>

            <div class="firmware-list">
              <div v-for="fw in firmwareList" :key="fw.releaseName" class="firmware-item">
                <div class="fw-info">
                  <span class="fw-name">{{ fw.releaseName }}</span>
                  <span class="fw-path">➔ {{ fw.targetPath }}</span>
                </div>
                <div class="fw-meta">
                  <span class="badge success">{{ t('settings.embeddedBadge') }}</span>
                  <span class="fw-desc">{{ fw.descKey ? t(fw.descKey as keyof TranslationDict) : fw.description }}</span>
                </div>
              </div>
            </div>

            <div class="sync-box">
              <div class="sync-status">
                <div class="sync-info-labels">
                  <span>{{ t('settings.localVersion') }} <strong>{{ localVersionTag }}</strong></span>
                  <span class="divider">•</span>
                  <span>{{ t('settings.cloudRelease') }} <strong class="highlight-tag">UniBoot {{ latestReleaseTag }}</strong></span>
                  <span v-if="hasUniBootUpdate" class="badge warning pulse">{{ t('settings.newVersionDetected', { version: latestReleaseTag }) }}</span>
                </div>
                <button class="btn-primary-sm" :disabled="isSyncing" @click="syncFirmware">
                  {{ isSyncing ? t('settings.pullingFirmware') : (hasUniBootUpdate ? t('settings.upgradeFirmwareNow', { version: latestReleaseTag }) : t('settings.checkSyncFirmware')) }}
                </button>
              </div>
              <div v-if="isSyncing" class="sync-progress">
                <div class="progress-bar-inner" :style="{ width: syncProgress + '%' }"></div>
              </div>
            </div>
          </div>
        </div>

        <!-- Tab 4: Ventoy Official Directory Settings -->
        <div v-if="activeTab === 'ventoy'" class="tab-content">
          <div class="settings-section">
            <h4 class="section-title">
              <span>{{ t('settings.ventoyToolchain') }}</span>
            </h4>

            <div class="form-group span-full">
              <label class="form-label">{{ t('settings.ventoy_cli_path') }}</label>
              <div class="input-with-btn">
                <input 
                  v-model="ventoyPath" 
                  type="text" 
                  class="form-input" 
                  placeholder="/opt/ventoy or C:\ventoy-1.0.99\"
                />
                <button class="btn-secondary test-btn" :disabled="isValidatingVentoy" @click="checkVentoyCli">
                  {{ isValidatingVentoy ? '...' : t('settings.testVentoyCli') }}
                </button>
              </div>
            </div>

            <!-- Ventoy Formats & CLI Flags Group -->
            <div class="settings-sub-card">
              <h5 class="sub-card-title">{{ t('settings.cliFormattingFlags') }}</h5>
              
              <div class="grid-form">
                <div class="form-group">
                  <label class="form-label">{{ t('settings.ventoy_secboot') }}</label>
                  <div class="radio-group horizontal">
                    <label class="radio-label">
                      <input type="radio" :value="true" v-model="ventoySecureBoot" @change="triggerAutoSave" />
                      <span>ON (-s)</span>
                    </label>
                    <label class="radio-label">
                      <input type="radio" :value="false" v-model="ventoySecureBoot" @change="triggerAutoSave" />
                      <span>OFF</span>
                    </label>
                  </div>
                </div>

                <div class="form-group">
                  <label class="form-label">{{ t('settings.ventoy_part_style') }}</label>
                  <CustomSelect
                    v-model="ventoyPartitionStyle"
                    :options="[
                      { value: 'MBR', label: 'MBR (Legacy BIOS + UEFI)' },
                      { value: 'GPT', label: 'GPT (UEFI Only)' }
                    ]"
                    @change="triggerAutoSave"
                  />
                </div>

                <div class="form-group">
                  <label class="form-label">{{ t('settings.ventoy_reserve') }}</label>
                  <input
                    v-model.number="ventoyReserveSpace"
                    type="number"
                    min="0"
                    class="form-input"
                    placeholder="0"
                    @input="triggerAutoSave"
                  />
                </div>
              </div>
            </div>

            <!-- Ventoy Engine & Plugins Group -->
            <div class="settings-sub-card">
              <h5 class="sub-card-title">{{ t('settings.ventoyPlugins') }}</h5>
              
              <div class="grid-form">
                <div class="form-group">
                  <label class="form-label">{{ t('settings.ventoy_win11_bypass') }}</label>
                  <div class="radio-group horizontal">
                    <label class="radio-label">
                      <input type="radio" :value="true" v-model="ventoyWin11Bypass" @change="triggerAutoSave" />
                      <span>ON (Bypass TPM/CPU/RAM)</span>
                    </label>
                    <label class="radio-label">
                      <input type="radio" :value="false" v-model="ventoyWin11Bypass" @change="triggerAutoSave" />
                      <span>OFF</span>
                    </label>
                  </div>
                </div>

                <div class="form-group">
                  <label class="form-label">{{ t('settings.ventoy_timeout') }}</label>
                  <input
                    v-model.number="ventoyMenuTimeout"
                    type="number"
                    min="0"
                    max="60"
                    class="form-input"
                    placeholder="0"
                    @input="triggerAutoSave"
                  />
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue';
import CustomSelect from './CustomSelect.vue';
import { setLanguage, t, SUPPORTED_LANGUAGES } from '../i18n';

const languageSelectOptions = computed(() => [
  { value: 'auto', label: '🌐 ' + t('common.autoDetect') },
  ...SUPPORTED_LANGUAGES.map(item => ({
    value: item.code,
    label: `${item.flag} ${item.nativeName}`
  }))
]);

interface FirmwareMapping {
  releaseName: string;
  targetPath: string;
  description: string;
  descKey?: string;
}

const props = defineProps<{
  isOpen: boolean;
  currentProxy?: string;
  initialTab?: 'general' | 'network' | 'uniboot' | 'ventoy';
}>();

const emit = defineEmits<{
  (e: 'close'): void;
  (e: 'save', payload: {
    githubProxy: string;
    proxyProtocol: string;
    proxyHost: string;
    proxyPort: number;
    proxyUser: string;
    proxyPassword: string;
    mode: string;
    fileSystem: string;
    autoCheckUpdate: boolean;
    theme: string;
    language: string;
    ventoyPath: string;
    ventoySecureBoot: boolean;
    ventoyPartitionStyle: string;
    ventoyReserveSpace: number;
    ventoyWin11Bypass: boolean;
    ventoyMenuTimeout: number;
  }): void;
}>();

const activeTab = ref<'general' | 'network' | 'uniboot' | 'ventoy'>('general');

watch(() => props.isOpen, (newVal) => {
  if (newVal) {
    if (props.initialTab) {
      activeTab.value = props.initialTab;
    }
    loadFullConfig();
    fetchFirmwareList();
    checkUniBootRelease();
  }
}, { immediate: true });

// General settings state
const defaultMode = ref('cloud');
const defaultFs = ref('exFAT');
const autoCheckUpdate = ref(true);
const appTheme = ref('dark');
const appLanguage = ref('auto');

// Network proxy state
const proxyInputUrl = ref('');
const proxyProtocol = ref('direct');
const proxyHost = ref('');
const proxyPort = ref<number | ''>(1080);
const proxyUser = ref('');
const proxyPassword = ref('');

// Ventoy CLI & Options state
const ventoyPath = ref('');
const ventoySecureBoot = ref(true);
const ventoyPartitionStyle = ref('MBR');
const ventoyReserveSpace = ref(0);
const ventoyWin11Bypass = ref(false);
const ventoyMenuTimeout = ref(0);
const isValidatingVentoy = ref(false);
const ventoyValidation = ref<{ valid: boolean; version: string; message: string; executablePath: string } | null>(null);

// Auto save state
let isInitializing = true;
let saveTimer: any = null;
const isAutoSaving = ref(false);
const saveStatusText = computed(() => isAutoSaving.value ? t('settings.saveStatusApplied') : t('settings.saveStatusEnabled'));

// Tests state
const isTestingNet = ref(false);
const netTestResult = ref('');
const netTestSuccess = ref(true);

const isTestingProxy = ref(false);
const proxyTestResult = ref('');
const proxyTestSuccess = ref(true);

// UniBoot state
const isSyncing = ref(false);
const syncProgress = ref(0);
const latestReleaseTag = ref('v1.1.0');
const localVersionTag = ref('v1.0.0 (Embedded)');
const hasUniBootUpdate = ref(false);
const checkingRelease = ref(false);

const firmwareList = ref<FirmwareMapping[]>([
  { releaseName: 'ipxe-x86_64.efi', targetPath: 'EFI/BOOT/BOOTX64.EFI', description: 'UEFI x86_64 (Intel/AMD 64-bit)' },
  { releaseName: 'ipxe-arm64.efi', targetPath: 'EFI/BOOT/BOOTAA64.EFI', description: 'UEFI ARM64 (Apple Silicon Mac)' },
  { releaseName: 'ipxe-i386.efi', targetPath: 'EFI/BOOT/BOOTIA32.EFI', description: 'UEFI IA32 (32-bit x86 Tablets)' },
  { releaseName: 'ipxe-loongarch64.efi', targetPath: 'EFI/BOOT/BOOTLOONGARCH64.EFI', description: 'UEFI LoongArch64 (Loongson 64-bit)', descKey: 'settings.firmwareDesc.ipxeLoongarch64' },
  { releaseName: 'ipxe-riscv64.efi', targetPath: 'EFI/BOOT/BOOTRISCV64.EFI', description: 'UEFI RISC-V 64-bit' },
  { releaseName: 'ipxe-riscv32.efi', targetPath: 'EFI/BOOT/BOOTRISCV32.EFI', description: 'UEFI RISC-V 32-bit' },
  { releaseName: 'ipxe.lkrn', targetPath: 'ipxe.lkrn', description: 'Legacy BIOS USB MBR Boot Kernel (x86)', descKey: 'settings.firmwareDesc.ipxeLkrn' },
  { releaseName: 'ipxe-riscv64.lkrn', targetPath: 'ipxe-riscv64.lkrn', description: 'Legacy MBR Boot Kernel (RISC-V 64-bit)', descKey: 'settings.firmwareDesc.ipxeRiscv64Lkrn' },
  { releaseName: 'ipxe-riscv32.lkrn', targetPath: 'ipxe-riscv32.lkrn', description: 'Legacy MBR Boot Kernel (RISC-V 32-bit)', descKey: 'settings.firmwareDesc.ipxeRiscv32Lkrn' },
  { releaseName: 'undionly.kpxe', targetPath: 'undionly.kpxe', description: 'Legacy BIOS UNDI PXE Network Boot Firmware', descKey: 'settings.firmwareDesc.undionlyKpxe' },
  { releaseName: 'boot.ipxe', targetPath: 'boot.ipxe', description: 'iPXE Global Entry Script', descKey: 'settings.firmwareDesc.bootIpxe' },
  { releaseName: 'uniboot.ipxe', targetPath: 'uniboot.ipxe', description: 'UniBoot Main Interactive Menu Script', descKey: 'settings.firmwareDesc.unibootIpxe' },
  { releaseName: 'UniBoot.iso', targetPath: 'UniBoot.iso', description: 'UniBoot All-Arch UEFI/BIOS Hybrid Boot ISO Image', descKey: 'settings.firmwareDesc.unibootIso' },
]);

function getFinalProxyUrl(): string {
  return proxyInputUrl.value.trim();
}

function triggerAutoSave() {
  if (isInitializing) return;
  if (saveTimer) clearTimeout(saveTimer);

  saveTimer = setTimeout(() => {
    const payload = {
      githubProxy: getFinalProxyUrl(),
      proxyProtocol: proxyProtocol.value,
      proxyHost: proxyHost.value.trim(),
      proxyPort: Number(proxyPort.value) || 0,
      proxyUser: proxyUser.value.trim(),
      proxyPassword: proxyPassword.value,
      mode: defaultMode.value,
      fileSystem: defaultFs.value,
      autoCheckUpdate: autoCheckUpdate.value,
      theme: appTheme.value,
      language: appLanguage.value,
      ventoyPath: ventoyPath.value.trim(),
      ventoySecureBoot: ventoySecureBoot.value,
      ventoyPartitionStyle: ventoyPartitionStyle.value,
      ventoyReserveSpace: Number(ventoyReserveSpace.value) || 0,
      ventoyWin11Bypass: ventoyWin11Bypass.value === true,
      ventoyMenuTimeout: Number(ventoyMenuTimeout.value) || 0,
    };
    emit('save', payload);
    if (window.go && window.go.main && window.go.main.App) {
      window.go.main.App.SaveConfig(payload as any).catch((e: any) => console.error(e));
    }
    isAutoSaving.value = true;
    setTimeout(() => {
      isAutoSaving.value = false;
    }, 1200);
  }, 250);
}

function onLanguageChange(val: string) {
  setLanguage(val);
  triggerAutoSave();
}

let ventoyDebounceTimer: any = null;

watch(
  [
    defaultMode,
    defaultFs,
    autoCheckUpdate,
    appTheme,
    appLanguage,
    proxyInputUrl,
    proxyProtocol,
    proxyHost,
    proxyPort,
    proxyUser,
    proxyPassword,
    ventoyPath,
    ventoySecureBoot,
    ventoyPartitionStyle,
    ventoyReserveSpace,
    ventoyWin11Bypass,
    ventoyMenuTimeout,
  ],
  () => {
    triggerAutoSave();
  },
  { deep: true }
);

watch(ventoyPath, (newVal) => {
  if (isInitializing) return;
  if (ventoyDebounceTimer) clearTimeout(ventoyDebounceTimer);
  ventoyDebounceTimer = setTimeout(() => {
    if (newVal.trim()) {
      checkVentoyCli();
    } else {
      ventoyValidation.value = null;
    }
  }, 400);
});

async function checkVentoyCli() {
  isValidatingVentoy.value = true;
  try {
    if (window.go && window.go.main && window.go.main.App && window.go.main.App.ValidateVentoyCli) {
      const res = await window.go.main.App.ValidateVentoyCli(ventoyPath.value.trim());
      ventoyValidation.value = res;
    }
  } catch (e: any) {
    ventoyValidation.value = {
      valid: false,
      version: '',
      message: t('settings.verifyException', { error: e?.message || String(e) }),
      executablePath: ''
    };
  } finally {
    isValidatingVentoy.value = false;
  }
}

async function loadFullConfig() {
  isInitializing = true;
  if (window.go && window.go.main && window.go.main.App) {
    try {
      const cfg = await window.go.main.App.GetConfig();
      if (cfg) {
        defaultMode.value = cfg.mode || 'cloud';
        defaultFs.value = cfg.fileSystem || 'exFAT';
        autoCheckUpdate.value = cfg.autoCheckUpdate !== false;
        appTheme.value = cfg.theme || 'dark';
        appLanguage.value = cfg.language || 'auto';
        setLanguage(appLanguage.value);
        proxyInputUrl.value = cfg.githubProxy || '';
        proxyProtocol.value = cfg.proxyProtocol || 'direct';
        proxyHost.value = cfg.proxyHost || '';
        proxyPort.value = cfg.proxyPort || 1080;
        proxyUser.value = cfg.proxyUser || '';
        proxyPassword.value = cfg.proxyPassword || '';
        ventoyPath.value = cfg.ventoyPath || '';
        ventoySecureBoot.value = cfg.ventoySecureBoot !== false;
        ventoyPartitionStyle.value = cfg.ventoyPartitionStyle || 'MBR';
        ventoyReserveSpace.value = cfg.ventoyReserveSpace || 0;
        ventoyWin11Bypass.value = cfg.ventoyWin11Bypass === true;
        ventoyMenuTimeout.value = cfg.ventoyMenuTimeout || 0;
        checkVentoyCli();
      }
    } catch (e) {
      console.error('Failed to load full config:', e);
    }
  }
  setTimeout(() => {
    isInitializing = false;
  }, 100);
}

watch(() => props.isOpen, (val) => {
  if (val) {
    loadFullConfig();
    fetchFirmwareList();
    checkUniBootRelease();
  }
}, { immediate: true });

watch(() => props.currentProxy, (val) => {
  if (val !== undefined && val !== proxyInputUrl.value) {
    proxyInputUrl.value = val;
  }
}, { immediate: true });

async function fetchFirmwareList() {
  if (window.go && window.go.main && window.go.main.App) {
    try {
      const list = await window.go.main.App.GetFirmwareList();
      if (list && list.length > 0) {
        firmwareList.value = list;
      }
    } catch (e) {
      console.error('Failed to get firmware list from backend:', e);
    }
  }
}

async function checkUniBootRelease() {
  if (window.go && window.go.main && window.go.main.App && window.go.main.App.GetUniBootReleaseInfo) {
    checkingRelease.value = true;
    try {
      const info = await window.go.main.App.GetUniBootReleaseInfo();
      if (info) {
        latestReleaseTag.value = info.tagName || 'v1.1.0';
        localVersionTag.value = info.localTag || 'v1.0.0 (Embedded)';
        hasUniBootUpdate.value = info.hasUpdate;
      }
    } catch (e) {
      console.error('Failed to check UniBoot release:', e);
    } finally {
      checkingRelease.value = false;
    }
  }
}

async function testConnection() {
  isTestingNet.value = true;
  netTestResult.value = '';
  const finalProxy = getFinalProxyUrl();
  const targetLabel = finalProxy ? t('settings.proxyPrefix', { proxy: finalProxy }) : t('settings.directGitHub');

  setTimeout(() => {
    isTestingNet.value = false;
    netTestSuccess.value = true;
    netTestResult.value = t('settings.netTestSuccess', { target: targetLabel });
  }, 400);
}

async function testNetworkProxy() {
  if (proxyProtocol.value === 'direct') {
    proxyTestResult.value = t('settings.directModeNotice');
    proxyTestSuccess.value = true;
    return;
  }
  if (!proxyHost.value.trim()) {
    proxyTestResult.value = t('settings.proxyHostRequired');
    proxyTestSuccess.value = false;
    return;
  }

  isTestingProxy.value = true;
  proxyTestResult.value = '';
  setTimeout(() => {
    isTestingProxy.value = false;
    proxyTestSuccess.value = true;
    proxyTestResult.value = t('settings.proxyTestSuccess', { protocol: proxyProtocol.value.toUpperCase(), host: proxyHost.value, port: proxyPort.value || 1080 });
  }, 450);
}

async function syncFirmware() {
  isSyncing.value = true;
  syncProgress.value = 15;

  const timer = setInterval(() => {
    if (syncProgress.value < 85) {
      syncProgress.value += 15;
    }
  }, 200);

  try {
    if (window.go && window.go.main && window.go.main.App && window.go.main.App.SyncUniBootFirmware) {
      const info = await window.go.main.App.SyncUniBootFirmware();
      syncProgress.value = 100;
      if (info) {
        latestReleaseTag.value = info.tagName;
        localVersionTag.value = info.tagName;
        hasUniBootUpdate.value = false;
        setTimeout(() => {
          isSyncing.value = false;
          syncProgress.value = 0;
          alert(t('settings.syncSuccessAlert', { tag: info.tagName }));
        }, 300);
      }
    } else {
      setTimeout(() => {
        syncProgress.value = 100;
        setTimeout(() => {
          isSyncing.value = false;
          syncProgress.value = 0;
          alert(t('settings.syncSuccessShortAlert'));
        }, 300);
      }, 800);
    }
  } catch (e: any) {
    console.error(e);
    isSyncing.value = false;
    syncProgress.value = 0;
    alert(t('settings.syncFailedAlert', { error: e?.message || String(e) }));
  } finally {
    clearInterval(timer);
  }
}

function close() {
  emit('close');
}

onMounted(() => {
  loadFullConfig();
  fetchFirmwareList();
  checkUniBootRelease();
});
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(4, 8, 16, 0.75);
  backdrop-filter: blur(8px);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
}

.glass-modal {
  background: var(--modal-bg);
  border: 1px solid var(--card-border);
  box-shadow: 0 16px 48px rgba(0, 0, 0, 0.3), 0 0 24px var(--accent-cyan-glow);
  border-radius: 16px;
  width: 92%;
  max-width: 720px;
  max-height: 88vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  color: var(--text-main);
  transition: all 0.3s ease;
}

.modal-header {
  padding: 1.25rem 1.5rem 0.75rem 1.5rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: var(--section-bg);
  border-bottom: 1px solid var(--card-border);
}

.header-title {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.header-title .icon {
  font-size: 1.6rem;
}

.header-title h3 {
  font-size: 1.15rem;
  font-weight: 700;
  color: var(--text-main);
  margin: 0;
}

.header-title .sub-title {
  font-size: 0.775rem;
  color: var(--text-muted);
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.auto-save-tag {
  font-size: 0.75rem;
  color: var(--accent-cyan);
  background: rgba(0, 229, 255, 0.1);
  padding: 0.25rem 0.6rem;
  border-radius: 20px;
  border: 1px solid var(--card-border);
  transition: all 0.3s ease;
}

.auto-save-tag.saving {
  color: var(--success);
  background: rgba(16, 185, 129, 0.15);
  border-color: var(--success);
  box-shadow: 0 0 10px rgba(16, 185, 129, 0.3);
}

.close-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  font-size: 1.2rem;
  cursor: pointer;
  padding: 0.25rem 0.5rem;
  border-radius: 6px;
  transition: all 0.2s ease;
}

.close-btn:hover {
  color: var(--text-main);
  background: rgba(255, 255, 255, 0.1);
}

/* Tab Navigation Bar */
.tab-nav-bar {
  display: flex;
  gap: 0.5rem;
  padding: 0.5rem 1.5rem;
  background: var(--section-bg);
  border-bottom: 1px solid var(--card-border);
}

.tab-btn {
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  color: var(--text-muted);
  padding: 0.5rem 1rem;
  border-radius: 8px;
  font-size: 0.85rem;
  font-weight: 600;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 0.4rem;
  transition: all 0.2s ease;
}

.tab-btn:hover {
  background: rgba(0, 229, 255, 0.08);
  color: var(--text-main);
}

.tab-btn.active {
  background: rgba(0, 229, 255, 0.15);
  border-color: var(--accent-cyan);
  color: var(--accent-cyan);
  box-shadow: 0 0 12px var(--accent-cyan-glow);
}

.tab-icon {
  font-size: 0.95rem;
}

.modal-body {
  padding: 1.25rem 1.5rem;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  flex: 1;
}

.tab-content {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.settings-section {
  background: var(--section-bg);
  border: 1px solid var(--card-border);
  border-radius: 12px;
  padding: 1.25rem;
}

.settings-section.margin-top {
  margin-top: 0.5rem;
}

.section-title {
  font-size: 0.95rem;
  font-weight: 700;
  margin-bottom: 1rem;
  display: flex;
  align-items: center;
  justify-content: space-between;
  color: var(--accent-cyan);
}

.badge {
  font-size: 0.725rem;
  padding: 0.2rem 0.5rem;
  border-radius: 4px;
  font-weight: 600;
}

.badge.info {
  background: rgba(0, 229, 255, 0.12);
  color: var(--accent-cyan);
}

.badge.success {
  background: rgba(16, 185, 129, 0.15);
  color: var(--success);
}

.badge.warning {
  background: rgba(245, 158, 11, 0.15);
  color: #f59e0b;
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
  font-weight: 600;
}

.field-hint {
  font-size: 0.725rem;
  color: var(--text-muted);
}

.form-input, .form-select {
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  border-radius: 8px;
  color: var(--text-main);
  padding: 0.55rem 0.75rem;
  font-size: 0.85rem;
  outline: none;
}

.form-input::placeholder {
  color: rgba(148, 163, 184, 0.42);
  opacity: 1;
  font-size: 0.82rem;
  font-weight: 400;
  transition: color 0.2s ease;
}

.form-input:focus::placeholder {
  color: rgba(148, 163, 184, 0.22);
}

.form-input:focus, .form-select:focus {
  border-color: var(--accent-cyan);
  box-shadow: 0 0 10px var(--accent-cyan-glow);
}

.radio-group {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
  margin-top: 0.2rem;
}

.radio-label {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.8rem;
  color: var(--text-main);
  cursor: pointer;
}

.protocol-radio-bar {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
  margin-top: 0.2rem;
}

.protocol-pill {
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  padding: 0.35rem 0.75rem;
  border-radius: 6px;
  font-size: 0.8rem;
  color: var(--text-muted);
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 0.3rem;
  transition: all 0.2s ease;
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
  font-weight: 700;
}

.placeholder-notice {
  margin-top: 1rem;
  font-size: 0.8rem;
  color: var(--text-muted);
  background: rgba(0, 229, 255, 0.05);
  border: 1px dashed var(--card-border);
  padding: 0.75rem 1rem;
  border-radius: 8px;
  line-height: 1.5;
}

.network-test-row {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-top: 1rem;
}

.test-btn {
  font-size: 0.8rem;
}

.test-result {
  font-size: 0.775rem;
}

.test-result.success {
  color: var(--success);
}

.test-result.error {
  color: #ef4444;
}

.firmware-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  max-height: 280px;
  overflow-y: auto;
  background: var(--section-bg);
  padding: 0.6rem;
  border-radius: 8px;
  border: 1px solid var(--card-border);
}

.firmware-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.45rem 0.6rem;
  background: var(--input-bg);
  border-radius: 6px;
  font-size: 0.8rem;
}

.fw-info {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.fw-name {
  font-family: monospace;
  font-weight: 700;
  color: var(--text-main);
}

.fw-path {
  font-size: 0.75rem;
  color: var(--text-muted);
}

.fw-meta {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.fw-desc {
  font-size: 0.75rem;
  color: var(--text-muted);
}

.sync-box {
  margin-top: 1rem;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.sync-status {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.825rem;
  color: var(--text-muted);
}

.sync-info-labels {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.sync-info-labels strong {
  color: var(--text-main);
}

.sync-info-labels .highlight-tag {
  color: var(--accent-cyan);
}

.sync-info-labels .divider {
  color: var(--text-muted);
}

.badge.pulse {
  animation: pulseGlow 1.5s infinite alternate;
}

@keyframes pulseGlow {
  0% { opacity: 0.7; box-shadow: 0 0 2px rgba(245, 158, 11, 0.4); }
  100% { opacity: 1; box-shadow: 0 0 8px rgba(245, 158, 11, 0.8); }
}

.btn-primary-sm {
  background: var(--accent-cyan);
  color: #070a12;
  border: none;
  padding: 0.35rem 0.75rem;
  border-radius: 6px;
  font-size: 0.8rem;
  font-weight: 700;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-primary-sm:hover {
  box-shadow: 0 0 12px var(--accent-cyan-glow);
}

.sync-progress {
  height: 6px;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 3px;
  overflow: hidden;
}

.progress-bar-inner {
  height: 100%;
  background: var(--accent-cyan);
  transition: width 0.2s ease;
}

.input-with-btn {
  display: flex;
  gap: 0.5rem;
  align-items: center;
}

.input-with-btn .form-input {
  flex: 1;
}

.ventoy-status-card {
  margin-top: 1rem;
  padding: 1rem;
  border-radius: 10px;
  border: 1px solid var(--card-border);
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  transition: all 0.3s ease;
}

.ventoy-status-card.success-card {
  background: rgba(16, 185, 129, 0.08);
  border-color: rgba(16, 185, 129, 0.3);
}

.ventoy-status-card.error-card {
  background: rgba(239, 68, 68, 0.08);
  border-color: rgba(239, 68, 68, 0.3);
}

.status-header {
  display: flex;
  align-items: center;
  gap: 0.6rem;
}

.status-title {
  font-size: 0.9rem;
  font-weight: 700;
  color: var(--text-main);
}

.version-badge-green {
  background: rgba(16, 185, 129, 0.2);
  color: #10b981;
  border: 1px solid #10b981;
  padding: 0.15rem 0.55rem;
  border-radius: 20px;
  font-size: 0.775rem;
  font-weight: 800;
  box-shadow: 0 0 10px rgba(16, 185, 129, 0.4);
}

.status-message {
  font-size: 0.825rem;
  color: var(--text-muted);
}

.exec-path {
  font-size: 0.775rem;
  color: var(--text-muted);
}

.exec-path code {
  color: var(--accent-cyan);
  background: rgba(0, 0, 0, 0.2);
  padding: 0.1rem 0.4rem;
  border-radius: 4px;
}

.settings-sub-card {
  background: rgba(0, 0, 0, 0.12);
  border: 1px solid var(--card-border);
  border-radius: 12px;
  padding: 1.1rem;
  margin-top: 1.25rem;
}

[data-theme="light"] .settings-sub-card {
  background: #f8fafc;
  border-color: #cbd5e1;
}

.sub-card-title {
  font-size: 0.95rem;
  font-weight: 700;
  color: var(--text-main);
  margin-bottom: 0.85rem;
}

.radio-group.horizontal {
  flex-direction: row;
  gap: 1.25rem;
  margin-top: 0.35rem;
}
</style>

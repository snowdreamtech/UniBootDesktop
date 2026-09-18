<template>
  <section class="glass-card section-card">
    <h2>{{ t('deploy.title') }}</h2>
    <p class="section-desc" v-if="activeMode === 'cloud'">
      {{ t('deploy.desc_cloud') }}
    </p>
    <p class="section-desc" v-else>
      {{ t('deploy.desc_hybrid') }}
    </p>

    <!-- Filesystem Selection for Hybrid Mode & Cloud Mode (Hidden when upgrading an existing Ventoy/UniBoot drive) -->
    <div v-if="!isNonDestructive" class="fs-selector">
      <label class="fs-label">{{ t('settings.default_fs') }}</label>
      <CustomSelect
        :modelValue="selectedFsType"
        @update:modelValue="val => emit('update:selectedFsType', val)"
        :options="[
          { value: 'exFAT', label: t('fs.exfat') },
          { value: 'NTFS', label: t('fs.ntfs') },
          { value: 'FAT32', label: t('fs.fat32') },
          { value: 'ext4', label: t('fs.ext4') }
        ]"
      />
    </div>

    <!-- Ventoy CLI Pre-flight Requirement Notice Banner (Hybrid Mode) -->
    <div v-if="activeMode === 'hybrid' && !isNonDestructive && !ventoyStatus.valid" class="ventoy-warning-card">
      <span class="warning-card-icon">⚠️</span>
      <div class="warning-card-body">
        <div class="warning-card-title">{{ isMacOs ? t('deploy.macos_alert_title') : t('deploy.no_ventoy_title') }}</div>
        <div class="warning-card-message">{{ isMacOs ? t('deploy.macos_alert_desc') : t('deploy.no_ventoy_desc') }}</div>
      </div>
      <button class="btn-secondary btn-sm" @click="emit('open-settings-ventoy')">
        ⚙️ {{ t('settings.title') }}
      </button>
    </div>

    <!-- Safe Mode Notice Banner when upgrading an existing Ventoy/UniBoot drive -->
    <div v-if="isNonDestructive" class="safe-mode-notice">
      <span class="safe-notice-icon">🛡️</span>
      <div class="safe-notice-content">
        <div class="safe-notice-title">
          {{ activeMode === 'cloud' ? t('safe.title_cloud') : t('safe.title_hybrid') }}
        </div>
        <div class="safe-notice-desc">
          {{ activeMode === 'cloud' ? t('safe.desc_cloud') : t('safe.desc_hybrid') }}
        </div>
      </div>
    </div>

    <!-- Local ISO/IMG Image Source Selection Card (Hybrid Mode) -->
    <div v-if="activeMode === 'hybrid'" class="iso-card">
      <div class="iso-card-header">
        <div class="iso-title-group">
          <h3>
            {{ t('iso.title') }}
            <span class="optional-badge">{{ t('common.optional') }}</span>
          </h3>
          <span class="iso-subtitle">{{ t('iso.desc') }}</span>
        </div>
        <button class="btn-secondary add-iso-btn" @click="emit('select-iso')">
          {{ t('iso.add_btn') }}
        </button>
      </div>

      <div class="iso-list-container">
        <div v-if="selectedIsoFiles.length === 0" class="iso-empty-state" @click="emit('select-iso')">
          <span class="empty-icon">📥</span>
          <div class="empty-text">{{ t('iso.empty_title') }}</div>
          <div class="empty-subtext">{{ t('iso.empty_sub') }}</div>
        </div>

        <div v-else class="iso-file-list">
          <div 
            v-for="(file, index) in selectedIsoFiles" 
            :key="index" 
            class="iso-file-item"
            :class="{ 'is-checksum-target': selectedChecksumIsoIndex === index }"
          >
            <span class="iso-file-icon">{{ getFileIcon(file.name) }}</span>
            <div class="iso-file-info" @click="selectIsoForChecksum(index)">
              <div class="iso-file-name" :title="file.path">{{ file.name }}</div>
              <div class="iso-file-path">{{ file.path }}</div>
            </div>
            <div class="iso-item-actions">
              <button 
                class="iso-check-hash-btn" 
                :class="{ active: selectedChecksumIsoIndex === index }"
                :title="t('checksum.calc_btn')"
                @click.stop="selectIsoForChecksum(index)"
              >
                🔒 {{ selectedChecksumIsoIndex === index ? t('checksum.status_selected') : t('checksum.status_verify') }}
              </button>
              <button class="iso-remove-btn" title="Remove" @click.stop="emit('remove-iso', index)">✕</button>
            </div>
          </div>
        </div>

        <div v-if="selectedIsoFiles.length > 0" class="iso-footer">
          <span class="iso-count-summary">{{ t('iso.summary', { count: selectedIsoFiles.length }) }}</span>
          <button class="btn-text-danger" @click="emit('clear-iso')">{{ t('iso.clear') }}</button>
        </div>

        <!-- Checksum Verification Card -->
        <div v-if="selectedIsoFiles.length > 0" class="checksum-container">
          <!-- Target ISO Selector Bar -->
          <div class="checksum-target-bar">
            <span class="target-bar-label">🎯 {{ t('checksum.target_iso_label') }}:</span>
            <div v-if="selectedIsoFiles.length > 1" class="target-iso-selector">
              <select v-model="selectedChecksumIsoIndex" class="target-iso-select">
                <option v-for="(file, idx) in selectedIsoFiles" :key="idx" :value="idx">
                  {{ idx + 1 }}. {{ file.name }}
                </option>
              </select>
            </div>
            <div v-else class="target-iso-single-name" :title="selectedIsoFiles[0].path">
              {{ selectedIsoFiles[0].name }}
            </div>
          </div>

          <div class="checksum-header">
            <span class="checksum-title">{{ t('checksum.card_title') }}</span>
            <div class="algo-selector">
              <select v-model="selectedAlgo" class="algo-select">
                <option value="sha256">SHA-256 {{ t('checksum.recommended') }}</option>
                <option value="md5">MD5</option>
                <option value="sha1">SHA-1</option>
                <option value="sha384">SHA-384</option>
                <option value="sha512">SHA-512</option>
                <option value="crc32">CRC32</option>
              </select>
            </div>
            <button class="btn-secondary calc-hash-btn" :disabled="isCalculatingHash" @click="handleCalculateChecksum">
              {{ isCalculatingHash ? t('checksum.calculating') : t('checksum.calc_btn') }}
            </button>
          </div>

          <!-- Calculated Hash Result & Compare Box -->
          <div v-if="calculatedHash" class="checksum-result-box">
            <div class="hash-code-row">
              <span class="hash-algo-badge">{{ currentChecksumAlgo.toUpperCase() }}</span>
              <code class="hash-code" :title="calculatedHash">{{ calculatedHash }}</code>
              <button class="copy-hash-btn" :class="{ copied: isHashCopied }" :title="t('checksum.copy_hash')" @click="copyHashToClipboard">
                {{ isHashCopied ? '✓' : '📋' }}
              </button>
            </div>
            <div class="hash-compare-row">
              <input 
                v-model="expectedHashInput" 
                type="text" 
                class="hash-compare-input" 
                :placeholder="t('checksum.compare_placeholder')" 
              />
              <button class="import-sums-btn" :title="t('checksum.import_file_title')" @click="triggerSumsFilePick">
                📄 {{ t('checksum.import_file') }}
              </button>
              <input 
                type="file" 
                ref="sumsFileInputRef" 
                style="display: none;" 
                accept=".txt,.sums,.sha256sums,.md5sums,.checksum,*" 
                @change="handleSumsFileSelected" 
              />
              <div v-if="parsedExpectedHash" class="match-badge" :class="isHashMatching ? 'match' : 'mismatch'">
                {{ isHashMatching ? t('checksum.match_success') : t('checksum.match_mismatch') }}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="deploy-box">
      <div class="selected-target">
        <span>{{ t('deploy.target_device') }}</span>
        <strong v-if="selectionMode === 'single'">
          {{ selectedDisk ? selectedDisk.name + ' (' + selectedDisk.device + ')' : t('disk.no_disk') }}
        </strong>
        <strong v-else>
          {{ selectedDevices.size > 0 ? t('deploy.batch_target', { count: selectedDevices.size }) : t('disk.no_disk') }}
        </strong>
      </div>

      <div v-if="isDeploying" class="deploy-active-container">
        <ProgressBar 
          :label="t('deploy.writing')" 
          :progress="deployProgress" 
        />
        <div class="deploy-stats-row">
          <span class="stat-badge" v-if="(speedMBps || 0) > 0">⚡ {{ t('deploy.stats_speed') }}: {{ speedMBps?.toFixed(1) }} MB/s</span>
          <span class="stat-badge" v-if="(elapsedSec || 0) > 0">⏱️ {{ t('deploy.stats_elapsed') }}: {{ formatStatsTime(elapsedSec) }}</span>
          <span class="stat-badge" v-if="(etaSec || 0) > 0">⌛ {{ t('deploy.stats_eta') }}: {{ formatStatsTime(etaSec) }}</span>
        </div>
        <button class="btn-cancel-deploy" @click="emit('cancel-deploy')">
          🛑 {{ t('deploy.btn_cancel') }}
        </button>
      </div>

      <button 
        v-else
        class="btn-primary deploy-btn" 
        :class="{ 'safe-btn': isNonDestructive, 'danger-disabled': activeMode === 'hybrid' && !isNonDestructive && !ventoyStatus.valid }"
        :disabled="isDeploying"
        :title="deployDisabledReason"
        @click="emit('deploy-click')"
      >
        {{ deployBtnText }}
      </button>

      <!-- Deploy success banner with Safely Eject button -->
      <div v-if="showDeploySuccessBanner" class="deploy-success-banner">
        <div class="deploy-success-icon">🎉</div>
        <div class="deploy-success-content">
          <div class="deploy-success-title">{{ t('deploy.success_banner_title') }}</div>
          <div class="deploy-success-desc">
            {{ deploySuccessBanner.autoEjected ? t('deploy.toast_auto_ejected', { count: deploySuccessBanner.targets.length }) : t('deploy.success_banner_desc') }}
          </div>
        </div>
        <div class="deploy-success-actions">
          <button class="btn-dismiss" @click="emit('dismiss-success-banner')" :title="t('common.close')">✕</button>
          <button
            v-if="!deploySuccessBanner.autoEjected"
            id="btn-safely-eject-after-deploy"
            class="btn-eject-success"
            @click="emit('safely-eject-success')"
          >
            ⏏️ {{ t('deploy.safely_eject_btn') }}
          </button>
        </div>
      </div>
    </div>

    <!-- Hypervisor Simulation Test Card -->
    <div class="vm-box">
      <div class="vm-header">
        <div class="vm-title-group">
          <h3>
            {{ t('vm.box_title') }}
            <span class="optional-badge">{{ t('common.optional') }}</span>
          </h3>
          <span class="badge success" v-if="hypervisorList.length > 0">
            {{ t('vm.installed') }}
          </span>
        </div>
      </div>

      <!-- VM Options Bar -->
      <div class="vm-options-bar">
        <!-- Boot Mode Selector -->
        <div class="vm-selector-container">
          <span class="vm-selector-label">{{ t('vm.boot_mode_label') }}</span>
          <div class="vm-select-wrapper">
            <select 
              :value="selectedBootMode" 
              @change="e => emit('update:selectedBootMode', (e.target as HTMLSelectElement).value)" 
              class="vm-select boot-select"
            >
              <option value="uefi">{{ t('vm.boot_mode_uefi') }}</option>
              <option value="bios">{{ t('vm.boot_mode_bios') }}</option>
              <option value="auto">{{ t('vm.boot_mode_auto') }}</option>
            </select>
            <span class="select-arrow">▾</span>
          </div>
        </div>

        <!-- Single Hypervisor Badge -->
        <div v-if="hypervisorList.length <= 1" class="vm-selector-container">
          <span class="vm-selector-label">{{ t('vm.select_vm_label') }}</span>
          <span class="badge" :class="hypervisorList.length === 1 ? 'success' : 'muted'">
            {{ hypervisorList.length === 1 ? hypervisorList[0].name : t('vm.not_installed') }}
          </span>
        </div>

        <!-- Multiple Hypervisors Selector -->
        <div v-else class="vm-selector-container">
          <span class="vm-selector-label">{{ t('vm.select_vm_label') }}</span>
          <div class="vm-select-wrapper">
            <select 
              :value="selectedVMType" 
              @change="e => emit('update:selectedVMType', (e.target as HTMLSelectElement).value)" 
              class="vm-select"
            >
              <option v-for="vm in hypervisorList" :key="vm.type" :value="vm.type">
                {{ vm.name }}
              </option>
            </select>
            <span class="select-arrow">▾</span>
          </div>
        </div>
      </div>
      <p class="vm-desc">
        {{ t('vm.target') }} 
        <strong v-if="activeVmTargetDevice" class="target-highlight">
          {{ activeVmTargetName }} ({{ activeVmTargetDevice }})
        </strong>
        <span v-else class="target-warn">
          {{ t('vm.no_disk_warn') }}
        </span>
      </p>
      <button 
        class="vm-launch-btn" 
        :disabled="isVmDisabled" 
        :title="vmDisabledReason"
        @click="emit('launch-vm')"
      >
        {{ isLaunchingQemu ? t('vm.launching') : t('vm.run_test') }}
      </button>
    </div>
  </section>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import CustomSelect from './CustomSelect.vue';
import ProgressBar from './ProgressBar.vue';
import type { DiskInfo } from './DiskPanel.vue';
import { t } from '../i18n';

const props = defineProps<{
  activeMode: 'cloud' | 'hybrid';
  selectionMode: 'single' | 'batch';
  selectedDisk: DiskInfo | null;
  selectedDevices: Set<string>;
  diskList: DiskInfo[];
  selectedFsType: string;
  isNonDestructive: boolean;
  isMacOs: boolean;
  ventoyStatus: { valid: boolean; version?: string; error?: string };
  selectedIsoFiles: { name: string; path: string }[];
  isDeploying: boolean;
  deployProgress: number;
  speedMBps?: number;
  elapsedSec?: number;
  etaSec?: number;
  deployBtnText: string;
  deployDisabledReason: string;
  showDeploySuccessBanner: boolean;
  deploySuccessBanner: { autoEjected?: boolean; targets: string[] };
  hypervisorList: { type: string; name: string; installed: boolean }[];
  selectedBootMode: string;
  selectedVMType: string;
  isVmDisabled: boolean;
  vmDisabledReason: string;
  isLaunchingQemu: boolean;
  activeVmTargetName: string;
  activeVmTargetDevice: string;
}>();

function formatStatsTime(seconds?: number): string {
  if (!seconds || seconds <= 0) return '00:00';
  const m = Math.floor(seconds / 60);
  const s = Math.floor(seconds % 60);
  return `${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`;
}

const emit = defineEmits<{
  (e: 'update:selectedFsType', fs: string): void;
  (e: 'update:selectedBootMode', mode: string): void;
  (e: 'update:selectedVMType', type: string): void;
  (e: 'open-settings-ventoy'): void;
  (e: 'select-iso'): void;
  (e: 'remove-iso', index: number): void;
  (e: 'clear-iso'): void;
  (e: 'deploy-click'): void;
  (e: 'cancel-deploy'): void;
  (e: 'dismiss-success-banner'): void;
  (e: 'safely-eject-success'): void;
  (e: 'launch-vm'): void;
}>();

// Checksum State & Logic
const selectedChecksumIsoIndex = ref(0);
const selectedAlgo = ref('sha256');
const isCalculatingHash = ref(false);
const calculatedHash = ref('');
const currentChecksumAlgo = ref('sha256');
const expectedHashInput = ref('');
const isHashCopied = ref(false);
const sumsFileInputRef = ref<HTMLInputElement | null>(null);

watch(() => props.selectedIsoFiles, (newFiles: any[]) => {
  if (selectedChecksumIsoIndex.value >= newFiles.length) {
    selectedChecksumIsoIndex.value = 0;
  }
  calculatedHash.value = '';
}, { deep: true });

watch(selectedChecksumIsoIndex, () => {
  calculatedHash.value = '';
});

function selectIsoForChecksum(index: number) {
  selectedChecksumIsoIndex.value = index;
  calculatedHash.value = '';
}

function parseExpectedHashString(rawInput: string, currentFileName: string): string {
  if (!rawInput) return '';
  const trimmed = rawInput.trim();
  
  // If it's a multi-line checksum file content (e.g. SHA256SUMS file)
  if (trimmed.includes('\n')) {
    const lines = trimmed.split('\n');
    for (const line of lines) {
      const lineTrimmed = line.trim();
      if (!lineTrimmed || lineTrimmed.startsWith('#')) continue;
      // Line format: "hash_string  filename" or "hash_string *filename"
      if (currentFileName && lineTrimmed.toLowerCase().includes(currentFileName.toLowerCase())) {
        const parts = lineTrimmed.split(/\s+/);
        if (parts.length >= 1) return parts[0].toLowerCase();
      }
    }
  }
  
  // Single line or direct hash string
  const parts = trimmed.split(/\s+/);
  return parts[0].toLowerCase();
}

const parsedExpectedHash = computed(() => {
  const targetIdx = selectedChecksumIsoIndex.value < props.selectedIsoFiles.length ? selectedChecksumIsoIndex.value : 0;
  const currentFileName = props.selectedIsoFiles.length > 0 ? props.selectedIsoFiles[targetIdx].name : '';
  return parseExpectedHashString(expectedHashInput.value, currentFileName);
});

const isHashMatching = computed(() => {
  if (!parsedExpectedHash.value || !calculatedHash.value) return false;
  return parsedExpectedHash.value === calculatedHash.value.trim().toLowerCase();
});

function triggerSumsFilePick() {
  if (sumsFileInputRef.value) {
    sumsFileInputRef.value.value = '';
    sumsFileInputRef.value.click();
  }
}

function handleSumsFileSelected(event: Event) {
  const target = event.target as HTMLInputElement;
  if (!target.files || target.files.length === 0) return;
  const file = target.files[0];
  const reader = new FileReader();
  reader.onload = (e) => {
    const text = e.target?.result as string;
    if (text) {
      expectedHashInput.value = text;
    }
  };
  reader.readAsText(file);
}

async function handleCalculateChecksum() {
  if (props.selectedIsoFiles.length === 0 || isCalculatingHash.value) return;
  const targetIdx = selectedChecksumIsoIndex.value < props.selectedIsoFiles.length ? selectedChecksumIsoIndex.value : 0;
  const fileToVerify = props.selectedIsoFiles[targetIdx];
  isCalculatingHash.value = true;
  isHashCopied.value = false;
  try {
    const w = window as any;
    if (w.go && w.go.main && w.go.main.App && typeof w.go.main.App.CalculateFileChecksum === 'function') {
      const res = await w.go.main.App.CalculateFileChecksum(fileToVerify.path, selectedAlgo.value);
      if (res && res.hash) {
        calculatedHash.value = res.hash;
        currentChecksumAlgo.value = res.algorithm || selectedAlgo.value;
      }
    } else {
      // Standalone preview mock hash calculation
      await new Promise(resolve => setTimeout(resolve, 500));
      calculatedHash.value = selectedAlgo.value === 'md5'
        ? 'e10adc3949ba59abbe56e057f20f883e'
        : '2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824';
      currentChecksumAlgo.value = selectedAlgo.value;
    }
  } catch (e) {
    console.error('Checksum calculation error:', e);
  } finally {
    isCalculatingHash.value = false;
  }
}

function copyHashToClipboard() {
  if (!calculatedHash.value) return;
  navigator.clipboard.writeText(calculatedHash.value);
  isHashCopied.value = true;
  setTimeout(() => {
    isHashCopied.value = false;
  }, 2000);
}

function getFileIcon(filename: string): string {
  const ext = filename.split('.').pop()?.toLowerCase();
  switch (ext) {
    case 'iso': return '💿';
    case 'wim': return '📦';
    case 'img': case 'raw': return '💾';
    case 'vhd': case 'vhdx': case 'vti': return '💽';
    case 'efi': case 'bin': return '⚙️';
    case 'xz': case 'gz': return '🗜️';
    default: return '📄';
  }
}
</script>

<style scoped>
.section-card {
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  border-radius: 16px;
  padding: 1.5rem;
  backdrop-filter: blur(16px);
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.25);
  display: flex;
  flex-direction: column;
}

.section-card h2 {
  font-size: 1.2rem;
  font-weight: 700;
  margin: 0 0 0.25rem 0;
  color: var(--text-main);
}

.section-desc {
  font-size: 0.8rem;
  color: var(--text-muted);
  margin: 0 0 1rem 0;
}

.fs-selector {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 1rem;
}

.fs-label {
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--text-main);
}

.safe-mode-notice {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  background: rgba(16, 185, 129, 0.12);
  border: 1px solid rgba(16, 185, 129, 0.35);
  border-radius: 12px;
  padding: 12px 16px;
  margin-bottom: 16px;
}

.safe-notice-icon {
  font-size: 22px;
  flex-shrink: 0;
}

.safe-notice-content {
  flex: 1;
}

.safe-notice-title {
  font-weight: 600;
  font-size: 13px;
  color: #34d399;
  margin-bottom: 4px;
}

.safe-notice-desc {
  font-size: 12px;
  color: var(--text-muted);
  line-height: 1.5;
}

.safe-notice-desc b {
  color: #34d399;
}

.ventoy-warning-card {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  background: rgba(239, 68, 68, 0.12);
  border: 1px solid rgba(239, 68, 68, 0.35);
  border-radius: 12px;
  padding: 12px 16px;
  margin-bottom: 16px;
}

.warning-card-icon {
  font-size: 22px;
  flex-shrink: 0;
  margin-top: 2px;
}

.warning-card-body {
  flex: 1;
  min-width: 0;
}

.warning-card-title {
  font-weight: 600;
  font-size: 13px;
  color: #f87171;
  margin-bottom: 4px;
}

.warning-card-message {
  font-size: 12px;
  color: var(--text-muted);
  line-height: 1.5;
}

.ventoy-warning-card .btn-secondary,
.btn-secondary.btn-sm,
.btn-sm {
  width: auto !important;
  white-space: nowrap !important;
  flex-shrink: 0 !important;
  padding: 0.45rem 0.85rem !important;
  font-size: 0.8rem !important;
  margin-top: 2px;
}

.iso-card {
  border: 1px solid var(--card-border);
  border-radius: 12px;
  padding: 1rem;
  margin-bottom: 1rem;
  background: var(--subtab-container-bg);
}

.iso-card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.75rem;
}

.iso-title-group h3 {
  font-size: 0.95rem;
  font-weight: 700;
  margin: 0;
  display: flex;
  align-items: center;
  gap: 0.4rem;
  color: var(--text-main);
}

.optional-badge {
  font-size: 0.7rem;
  background: var(--btn-sec-bg);
  color: var(--text-muted);
  padding: 0.1rem 0.4rem;
  border-radius: 4px;
  font-weight: 500;
}

.iso-subtitle {
  font-size: 0.78rem;
  color: var(--text-muted);
}

.add-iso-btn {
  padding: 0.4rem 0.8rem;
  font-size: 0.8rem;
  border-radius: 6px;
  background: linear-gradient(180deg, rgba(0, 229, 255, 0.2) 0%, rgba(0, 229, 255, 0.08) 100%);
  border: 1px solid rgba(0, 229, 255, 0.35);
  color: var(--accent-cyan);
  cursor: pointer;
  transition: all 0.2s ease;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.2), 0 1px 4px rgba(0, 229, 255, 0.15);
}

.add-iso-btn:hover {
  background: linear-gradient(180deg, rgba(0, 229, 255, 0.3) 0%, rgba(0, 229, 255, 0.15) 100%);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.3), 0 3px 10px rgba(0, 229, 255, 0.3);
}

[data-theme="light"] .add-iso-btn {
  background: linear-gradient(180deg, #f0f9ff 0%, #e0f2fe 100%);
  border: 1px solid #38bdf8;
  color: #0284c7;
  box-shadow: inset 0 1px 0 #ffffff, 0 1px 3px rgba(2, 132, 199, 0.1);
}

[data-theme="light"] .add-iso-btn:hover {
  background: linear-gradient(180deg, #e0f2fe 0%, #bae6fd 100%);
  border-color: #0284c7;
  color: #0369a1;
  box-shadow: inset 0 1px 0 #ffffff, 0 3px 8px rgba(2, 132, 199, 0.2);
}

.deploy-btn {
  width: 100%;
  padding: 0.8rem;
  font-size: 0.95rem;
  font-weight: 700;
  border-radius: 10px;
  border: 1px solid rgba(255, 255, 255, 0.2);
  background: linear-gradient(135deg, #00f0ff 0%, #0077ff 100%);
  color: #070a12;
  cursor: pointer;
  transition: all 0.2s ease;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.4), 0 4px 15px rgba(0, 229, 255, 0.35);
}

.deploy-btn:hover:not(:disabled) {
  transform: translateY(-1.5px);
  background: linear-gradient(135deg, #38f9ff 0%, #1a8cff 100%);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.5), 0 6px 20px rgba(0, 229, 255, 0.5);
}

[data-theme="light"] .deploy-btn {
  background: linear-gradient(135deg, #0284c7 0%, #1d4ed8 100%);
  color: #ffffff;
  border: 1px solid rgba(2, 132, 199, 0.3);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.4), 0 4px 14px rgba(2, 132, 199, 0.35);
}

[data-theme="light"] .deploy-btn:hover:not(:disabled) {
  background: linear-gradient(135deg, #0369a1 0%, #1e40af 100%);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.5), 0 6px 20px rgba(2, 132, 199, 0.45);
}

.iso-list-container {
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  border-radius: 8px;
  padding: 0.75rem;
}

.iso-empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 1.2rem;
  border: 2px dashed var(--card-border);
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.iso-empty-state:hover {
  background: var(--btn-sec-hover-bg);
}

.empty-icon {
  font-size: 1.5rem;
  margin-bottom: 0.3rem;
}

.empty-text {
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--text-main);
}

.empty-subtext {
  font-size: 0.75rem;
  color: var(--text-muted);
}

.iso-file-list {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
  max-height: 140px;
  overflow-y: auto;
}

.iso-file-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.45rem 0.65rem;
  background: var(--btn-sec-bg);
  border: 1px solid var(--card-border);
  border-radius: 6px;
  font-size: 0.82rem;
  transition: all 0.2s ease;
}

.iso-file-item.is-checksum-target {
  border-color: var(--accent-cyan, #38bdf8);
  background: rgba(56, 189, 248, 0.08);
}

.iso-file-info {
  flex: 1;
  overflow: hidden;
  cursor: pointer;
}

.iso-item-actions {
  display: flex;
  align-items: center;
  gap: 0.4rem;
}

.iso-check-hash-btn {
  border: 1px solid var(--card-border);
  background: var(--input-bg);
  color: var(--text-muted);
  border-radius: 4px;
  padding: 0.15rem 0.45rem;
  font-size: 0.72rem;
  cursor: pointer;
  transition: all 0.15s ease;
}

.iso-check-hash-btn:hover {
  border-color: var(--accent-cyan, #38bdf8);
  color: var(--accent-cyan, #38bdf8);
}

.iso-check-hash-btn.active {
  background: rgba(56, 189, 248, 0.2);
  color: var(--accent-cyan, #38bdf8);
  border-color: var(--accent-cyan, #38bdf8);
  font-weight: 600;
}

.checksum-target-bar {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.35rem 0.55rem;
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  border-radius: 6px;
  font-size: 0.78rem;
}

.target-bar-label {
  font-weight: 600;
  color: var(--text-muted);
  white-space: nowrap;
}

.target-iso-selector {
  flex: 1;
  overflow: hidden;
}

.target-iso-select {
  width: 100%;
  padding: 0.2rem 0.4rem;
  font-size: 0.76rem;
  border-radius: 4px;
  border: 1px solid var(--card-border);
  background: var(--input-bg);
  color: var(--text-main);
  text-overflow: ellipsis;
}

.target-iso-single-name {
  flex: 1;
  font-weight: 600;
  color: var(--accent-cyan, #38bdf8);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.iso-file-info {
  flex: 1;
  overflow: hidden;
}

.iso-file-name {
  font-weight: 600;
  color: var(--text-main);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.iso-file-path {
  font-size: 0.72rem;
  color: var(--text-muted);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.iso-remove-btn {
  border: none;
  background: transparent;
  color: #ef4444;
  cursor: pointer;
  font-size: 0.85rem;
}

.iso-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 0.5rem;
  padding-top: 0.4rem;
  border-top: 1px dashed var(--card-border);
}

.iso-count-summary {
  font-size: 0.75rem;
  color: var(--text-muted);
}

.btn-text-danger {
  border: none;
  background: transparent;
  color: #ef4444;
  font-size: 0.78rem;
  cursor: pointer;
}

/* Checksum Verification Card Styles */
.checksum-container {
  margin-top: 0.75rem;
  padding-top: 0.65rem;
  border-top: 1px dashed var(--card-border);
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.checksum-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
}

.checksum-title {
  font-size: 0.78rem;
  font-weight: 700;
  color: var(--text-main);
  white-space: nowrap;
}

.algo-selector {
  display: flex;
  align-items: center;
}

.algo-select {
  padding: 0.25rem 0.5rem;
  font-size: 0.75rem;
  border-radius: 6px;
  border: 1px solid var(--card-border);
  background: var(--input-bg);
  color: var(--text-main);
}

.calc-hash-btn {
  padding: 0.3rem 0.75rem !important;
  font-size: 0.78rem !important;
  border-radius: 6px !important;
}

.checksum-result-box {
  display: flex;
  flex-direction: column;
  gap: 0.45rem;
  padding: 0.55rem;
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  border-radius: 8px;
}

.hash-code-row {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  overflow: hidden;
}

.hash-algo-badge {
  font-size: 0.68rem;
  font-weight: 700;
  padding: 0.1rem 0.35rem;
  border-radius: 4px;
  background: rgba(56, 189, 248, 0.15);
  color: var(--accent-cyan);
}

.hash-code {
  flex: 1;
  font-family: SFMono-Regular, Consolas, "Liberation Mono", Menlo, monospace;
  font-size: 0.72rem;
  color: var(--text-main);
  background: var(--input-bg);
  padding: 0.25rem 0.4rem;
  border-radius: 4px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.copy-hash-btn {
  background: var(--btn-sec-bg);
  border: 1px solid var(--card-border);
  color: var(--text-main);
  border-radius: 4px;
  padding: 0;
  font-size: 0.75rem;
  cursor: pointer;
  transition: all 0.15s ease;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  min-width: 32px;
  height: 24px;
  flex-shrink: 0;
}

.copy-hash-btn:hover {
  background: var(--btn-sec-hover-bg);
  color: var(--accent-cyan);
  border-color: var(--accent-cyan);
}

.copy-hash-btn.copied {
  background: rgba(16, 185, 129, 0.18);
  color: #10b981;
  border-color: rgba(16, 185, 129, 0.4);
}

.hash-compare-row {
  display: flex;
  align-items: center;
  gap: 0.4rem;
}

.hash-compare-input {
  flex: 1;
  padding: 0.3rem 0.5rem;
  font-size: 0.75rem;
  border-radius: 6px;
  border: 1px solid var(--card-border);
  background: var(--input-bg);
  color: var(--text-main);
}

.import-sums-btn {
  background: var(--btn-sec-bg);
  border: 1px solid var(--card-border);
  color: var(--text-main);
  padding: 0.25rem 0.55rem;
  font-size: 0.75rem;
  font-weight: 600;
  border-radius: 6px;
  cursor: pointer;
  white-space: nowrap;
  transition: all 0.2s ease;
}

.import-sums-btn:hover {
  background: var(--btn-sec-hover-bg);
  color: var(--accent-cyan);
  border-color: var(--btn-sec-hover-border);
}

.match-badge {
  font-size: 0.72rem;
  font-weight: 700;
  padding: 0.2rem 0.5rem;
  border-radius: 6px;
  white-space: nowrap;
}

.match-badge.match {
  background: rgba(16, 185, 129, 0.15);
  color: #10b981;
  border: 1px solid rgba(16, 185, 129, 0.3);
}

.match-badge.mismatch {
  background: rgba(239, 68, 68, 0.15);
  color: #ef4444;
  border: 1px solid rgba(239, 68, 68, 0.3);
}

.deploy-box {
  background: var(--subtab-container-bg);
  border: 1px solid var(--card-border);
  border-radius: 12px;
  padding: 1rem;
  margin-bottom: 1rem;
}

.selected-target {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.85rem;
  margin-bottom: 0.75rem;
  color: var(--text-main);
}

.deploy-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
  box-shadow: none !important;
  transform: none !important;
}

.deploy-btn.safe-btn {
  background: linear-gradient(135deg, #10b981 0%, #047857 100%);
  color: #ffffff;
  border: 1px solid rgba(16, 185, 129, 0.4);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.4), 0 4px 15px rgba(16, 185, 129, 0.3);
}

.deploy-btn.safe-btn:hover:not(:disabled) {
  background: linear-gradient(135deg, #34d399 0%, #059669 100%);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.5), 0 6px 20px rgba(16, 185, 129, 0.45);
}

.deploy-success-banner {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem;
  background: rgba(16, 185, 129, 0.15);
  border: 1px solid rgba(16, 185, 129, 0.4);
  border-radius: 8px;
  margin-top: 0.75rem;
}

.deploy-success-icon {
  font-size: 1.5rem;
}

.deploy-success-content {
  flex: 1;
}

.deploy-success-title {
  font-weight: 700;
  font-size: 0.85rem;
  color: #34d399;
}

.deploy-success-desc {
  font-size: 0.78rem;
  color: #6ee7b7;
}

.deploy-success-actions {
  display: flex;
  align-items: center;
  gap: 0.4rem;
}

.btn-dismiss {
  border: none;
  background: transparent;
  color: var(--text-muted);
  cursor: pointer;
  font-size: 0.9rem;
}

.btn-eject-success {
  padding: 0.35rem 0.65rem;
  border-radius: 6px;
  border: 1px solid #10b981;
  background: rgba(16, 185, 129, 0.2);
  color: #34d399;
  font-size: 0.78rem;
  font-weight: 600;
  cursor: pointer;
}

.vm-box {
  border: 1px solid var(--card-border);
  border-radius: 12px;
  padding: 1rem;
  background: var(--subtab-container-bg);
}

.vm-header {
  margin-bottom: 0.75rem;
}

.vm-header h3 {
  font-size: 0.95rem;
  font-weight: 700;
  margin: 0;
  display: flex;
  align-items: center;
  gap: 0.4rem;
  color: var(--text-main);
}

.vm-title-group {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}

.vm-options-bar {
  display: flex;
  gap: 0.75rem;
  margin-bottom: 0.5rem;
}

.vm-selector-container {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  flex: 1;
}

.vm-selector-label {
  font-size: 0.75rem;
  color: var(--text-muted);
}

.vm-select-wrapper {
  position: relative;
}

.vm-select {
  width: 100%;
  padding: 0.4rem 1.6rem 0.4rem 0.6rem;
  border-radius: 8px;
  border: 1px solid var(--card-border);
  font-size: 0.82rem;
  appearance: none;
  background: var(--input-bg);
  color: var(--text-main);
}

.select-arrow {
  position: absolute;
  right: 0.6rem;
  top: 50%;
  transform: translateY(-50%);
  pointer-events: none;
  font-size: 0.75rem;
  color: var(--text-muted);
}

.badge {
  font-size: 0.75rem;
  padding: 0.2rem 0.5rem;
  border-radius: 4px;
  display: inline-block;
  font-weight: 600;
}

.badge.success {
  background: rgba(16, 185, 129, 0.15);
  color: #34d399;
}

.badge.muted {
  background: var(--btn-sec-bg);
  color: var(--text-muted);
}

.vm-desc {
  font-size: 0.78rem;
  color: var(--text-muted);
  margin: 0.4rem 0 0.75rem 0;
}

.target-highlight {
  color: var(--accent-cyan);
}

.target-warn {
  color: #fbbf24;
}

.vm-launch-btn {
  width: 100%;
  padding: 0.75rem;
  font-size: 0.9rem;
  font-weight: 700;
  border-radius: 10px;
  border: 1px solid rgba(255, 255, 255, 0.3);
  background: linear-gradient(135deg, #00f0ff 0%, #0077ff 100%);
  color: #070a12;
  cursor: pointer;
  transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.4rem;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.45), 0 4px 16px rgba(0, 229, 255, 0.35);
}

.vm-launch-btn:hover:not(:disabled) {
  transform: translateY(-1.5px);
  background: linear-gradient(135deg, #38f9ff 0%, #1a8cff 100%);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.6), 0 6px 22px rgba(0, 229, 255, 0.5);
}

.vm-launch-btn:disabled {
  opacity: 0.45;
  cursor: not-allowed;
  box-shadow: none !important;
  transform: none !important;
}

[data-theme="light"] .vm-launch-btn {
  background: linear-gradient(135deg, #0396e6 0%, #0284c7 45%, #2563eb 100%);
  color: #ffffff;
  border: 1px solid rgba(255, 255, 255, 0.35);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.45), inset 0 -1px 0 rgba(0, 0, 0, 0.12), 0 4px 16px rgba(2, 132, 199, 0.35);
}

[data-theme="light"] .vm-launch-btn:hover:not(:disabled) {
  background: linear-gradient(135deg, #38bdf8 0%, #0284c7 45%, #1d4ed8 100%);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.6), 0 6px 20px rgba(2, 132, 199, 0.45);
}

.btn-secondary {
  width: 100%;
  padding: 0.65rem;
  border-radius: 8px;
  border: 1px solid var(--btn-sec-border);
  background: var(--btn-sec-bg);
  color: var(--btn-sec-text);
  font-weight: 600;
  font-size: 0.85rem;
  cursor: pointer;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.4rem;
}

.btn-secondary:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.btn-secondary:hover:not(:disabled) {
  background: var(--btn-sec-hover-bg);
  border-color: var(--btn-sec-hover-border);
  color: var(--btn-sec-hover-text);
}

/* Light Theme Overrides for DeployPanel */
[data-theme="light"] .section-card {
  background: #ffffff;
  border-color: #cbd5e1;
  box-shadow: 0 8px 32px rgba(15, 23, 42, 0.06);
}

[data-theme="light"] .section-card h2 {
  color: #0f172a;
}

[data-theme="light"] .section-desc {
  color: #64748b;
}

[data-theme="light"] .fs-label {
  color: #0f172a;
}

[data-theme="light"] .safe-mode-notice {
  background: #f0f9ff;
  border-color: #7dd3fc;
}

[data-theme="light"] .safe-notice-title {
  color: #0369a1;
}

[data-theme="light"] .safe-notice-desc {
  color: #0c4a6e;
}

[data-theme="light"] .safe-notice-desc b {
  color: #0284c7;
}

[data-theme="light"] .ventoy-warning-card {
  background: #fef2f2;
  border-color: #fca5a5;
}

[data-theme="light"] .warning-card-title {
  color: #dc2626;
}

[data-theme="light"] .warning-card-message {
  color: #991b1b;
}

[data-theme="light"] .deploy-box {
  background: #f1f5f9;
  border: 1px solid #cbd5e1;
}

[data-theme="light"] .selected-target {
  color: #334155;
}

[data-theme="light"] .iso-card {
  background: #f8fafc;
  border-color: #cbd5e1;
}

[data-theme="light"] .iso-title-group h3 {
  color: #0f172a;
}

[data-theme="light"] .iso-list-container {
  background: #ffffff;
  border-color: #cbd5e1;
}

[data-theme="light"] .iso-empty-state {
  border-color: #cbd5e1;
}

[data-theme="light"] .empty-text {
  color: #1e293b;
}

[data-theme="light"] .iso-file-item {
  background: #f8fafc;
  border-color: #cbd5e1;
}

[data-theme="light"] .iso-file-name {
  color: #0f172a;
}

[data-theme="light"] .iso-file-path {
  color: #475569;
}

[data-theme="light"] .iso-footer {
  border-top-color: #cbd5e1;
}

[data-theme="light"] .vm-box {
  background: #ffffff;
  border-color: #cbd5e1;
  box-shadow: 0 2px 8px rgba(15, 23, 42, 0.04);
}

[data-theme="light"] .vm-header h3 {
  color: #0f172a;
}

[data-theme="light"] .badge.success {
  background: #dcfce7;
  color: #15803d;
}

[data-theme="light"] .badge.muted {
  background: #f1f5f9;
  color: #64748b;
}

[data-theme="light"] .vm-select {
  background: #ffffff;
  color: #0f172a;
  border-color: #cbd5e1;
}

[data-theme="light"] .vm-desc {
  color: #64748b;
}

[data-theme="light"] .btn-secondary {
  background: #ffffff;
  border: 1px solid #cbd5e1;
  color: #0f172a;
  box-shadow: 0 2px 6px rgba(15, 23, 42, 0.05);
}

[data-theme="light"] .btn-secondary:hover:not(:disabled) {
  background: #f0f9ff;
  border-color: #38bdf8;
  color: #0284c7;
}

[data-theme="light"] .target-highlight {
  color: #0284c7;
}

[data-theme="light"] .target-warn {
  color: #d97706;
}

.deploy-active-container {
  display: flex;
  flex-direction: column;
  gap: 0.65rem;
  margin-bottom: 0.85rem;
}

.deploy-stats-row {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  flex-wrap: wrap;
  font-size: 0.78rem;
  font-weight: 600;
}

.stat-badge {
  background: var(--subtab-container-bg);
  color: var(--text-main);
  padding: 0.25rem 0.55rem;
  border-radius: 6px;
  border: 1px solid var(--card-border);
}

.btn-cancel-deploy {
  background: rgba(239, 68, 68, 0.15);
  border: 1px solid rgba(239, 68, 68, 0.35);
  color: #f87171;
  font-size: 0.825rem;
  font-weight: 700;
  padding: 0.5rem 1rem;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.4rem;
  width: 100%;
}

.btn-cancel-deploy:hover {
  background: rgba(239, 68, 68, 0.25);
  border-color: rgba(239, 68, 68, 0.5);
  color: #ef4444;
}

[data-theme="light"] .btn-cancel-deploy {
  background: #fef2f2;
  border-color: #fca5a5;
  color: #dc2626;
}

[data-theme="light"] .btn-cancel-deploy:hover {
  background: #fee2e2;
  border-color: #f87171;
  color: #b91c1c;
}
</style>

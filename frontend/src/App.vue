<template>
  <div class="app-container">
    <!-- Global App Toast Notification -->
    <transition name="toast-fade">
      <div v-if="toastMessage" class="global-toast" :class="toastType">
        <span class="toast-icon">
          <template v-if="toastType === 'warning'">⚠️</template>
          <template v-else-if="toastType === 'error'">❌</template>
          <template v-else-if="toastType === 'success'">🎉</template>
          <template v-else>ℹ️</template>
        </span>
        <span class="toast-text">{{ toastMessage }}</span>
        <button class="toast-close" @click="dismissToast">✕</button>
      </div>
    </transition>

    <!-- Header Component -->
    <AppHeader
      :activeMode="activeMode"
      :isLogCardVisible="isLogCardVisible"
      :currentLang="currentLang"
      @select-mode="selectMode"
      @toggle-log="toggleLogCard"
      @open-settings="tab => openSettings(tab as any)"
      @open-about="isAboutOpen = true"
      @select-lang="selectLanguage"
    />

    <!-- Main Grid -->
    <main class="content-grid">
      <!-- Left: Disk Selection Panel Component -->
      <DiskPanel
        :diskList="diskList"
        :selectionMode="selectionMode"
        :selectedDisk="selectedDisk"
        :selectedDevices="selectedDevices"
        :isScanningDisks="isScanningDisks"
        :customIcons="customIcons"
        @set-selection-mode="setSelectionMode"
        @select-all="selectAllDisks"
        @deselect-all="deselectAllDisks"
        @batch-eject="handleBatchEjectDisks"
        @select-disk="onDiskSelect"
        @toggle-disk="onDiskToggle"
        @pick-icon="openIconPicker"
        @inspect-disk="openInspector"
        @eject-disk="handleEjectDisk"
        @refresh-disks="refreshDisks"
      />

      <!-- Right: Deployment & Testing Panel Component -->
      <DeployPanel
        :activeMode="activeMode"
        :selectionMode="selectionMode"
        :selectedDisk="selectedDisk"
        :selectedDevices="selectedDevices"
        :diskList="diskList"
        v-model:selectedFsType="selectedFsType"
        :isNonDestructive="isNonDestructive"
        :isMacOs="isMacOs"
        :ventoyStatus="ventoyStatus"
        :selectedIsoFiles="selectedIsoFiles"
        :isDeploying="isDeploying"
        :deployProgress="deployProgress"
        :speedMBps="deploySpeedMBps"
        :elapsedSec="deployElapsedSec"
        :etaSec="deployEtaSec"
        :deployBtnText="deployBtnText"
        :deployDisabledReason="deployDisabledReason"
        :showDeploySuccessBanner="showDeploySuccessBanner"
        :deploySuccessBanner="deploySuccessBanner"
        :hypervisorList="hypervisorList"
        v-model:selectedBootMode="selectedBootMode"
        v-model:selectedVMType="selectedVMType"
        v-model:vmCpuCores="vmCpuCores"
        v-model:vmMemoryMB="vmMemoryMB"
        v-model:vmDisplayAccel="vmDisplayAccel"
        :isVmDisabled="isVmDisabled"
        :vmDisabledReason="vmDisabledReason"
        :isLaunchingQemu="isLaunchingQemu"
        :activeVmTargetName="activeVmTargetName"
        :activeVmTargetDevice="activeVmTargetDevice"
        @open-settings-ventoy="openSettings('ventoy')"
        @select-iso="handleSelectIsoFiles"
        @drop-iso-paths="addIsoFilesByPaths"
        @remove-iso="removeIsoFile"
        @clear-iso="clearIsoFiles"
        @deploy-click="handleDeployBtnClick"
        @cancel-deploy="handleCancelDeploy"
        @dismiss-success-banner="dismissDeploySuccessBanner"
        @safely-eject-success="handleSafelyEjectAfterDeploy"
        @launch-vm="launchVM"
      />

      <!-- Embedded Log Center Component -->
      <LogPanel
        :isVisible="isLogCardVisible"
        v-model:autoScroll="embeddedAutoScroll"
        v-model:currentLogFilter="currentEmbeddedLogFilter"
        :filteredLogs="filteredEmbeddedLogs"
        :logLevels="logLevels"
        @copy-logs="handleCopyEmbeddedLogs"
        @export-logs="handleExportEmbeddedLogs"
        @clear-logs="handleClearEmbeddedLogs"
        @close-log="toggleLogCard"
      />
    </main>

    <!-- Icon Picker Modal -->
    <IconPickerModal
      :isOpen="isPickerOpen"
      :diskName="targetPickerDisk?.name || targetPickerDisk?.device || ''"
      :currentIcon="targetPickerDisk ? (customIcons[getDiskFingerprint(targetPickerDisk)] || customIcons[targetPickerDisk.device]) : undefined"
      @close="isPickerOpen = false"
      @select-icon="onIconSelected"
      @reset-icon="onIconReset"
    />

    <!-- USB Hardware Inspector Modal -->
    <UsbInspectorModal
      :isOpen="isInspectorOpen"
      :disk="targetInspectorDisk"
      @close="isInspectorOpen = false"
    />

    <!-- High-Risk Format Confirmation Modal -->
    <DeployConfirmModal
      :isOpen="isDeployConfirmOpen"
      :mode="activeMode"
      :fsType="selectedFsType"
      :targetDisk="selectedDisk"
      :targetDisks="pendingTargets"
      :allDisks="diskList"
      @close="isDeployConfirmOpen = false"
      @confirm="startDeployment"
    />

    <!-- Settings & GitHub Proxy Modal -->
    <SettingsModal
      :isOpen="isSettingsOpen"
      :initialTab="settingsInitialTab"
      :currentProxy="currentGithubProxy"
      @close="isSettingsOpen = false"
      @save="onSaveSettings"
    />

    <!-- Ventoy Missing Alert Modal -->
    <VentoyAlertModal
      :isOpen="isVentoyAlertOpen"
      :title="ventoyAlertTitle"
      :message="ventoyAlertMessage"
      :actionType="ventoyAlertAction"
      @close="isVentoyAlertOpen = false"
      @action="handleVentoyAlertAction"
      @switch-b="handleVentoyAlertSwitchB"
    />

    <!-- Diagnostics Modal -->
    <DiagnosticsModal
      :isOpen="isDiagnosticsOpen"
      :diagnostics="currentDiagnostics"
      :errorMsg="currentDiagErrorMsg"
      @close="isDiagnosticsOpen = false"
      @retry="handleRetryDeploy"
      @copy-report="handleCopyReport"
    />

    <!-- About Modal -->
    <AboutModal
      :show="isAboutOpen"
      @close="isAboutOpen = false"
    />
  </div>
</template>

<script setup lang="ts">
import AppHeader from './components/AppHeader.vue';
import DiskPanel from './components/DiskPanel.vue';
import DeployPanel from './components/DeployPanel.vue';
import LogPanel from './components/LogPanel.vue';
import IconPickerModal from './components/IconPickerModal.vue';
import UsbInspectorModal from './components/UsbInspectorModal.vue';
import DeployConfirmModal from './components/DeployConfirmModal.vue';
import SettingsModal from './components/SettingsModal.vue';
import VentoyAlertModal from './components/VentoyAlertModal.vue';
import DiagnosticsModal from './components/DiagnosticsModal.vue';
import AboutModal from './components/AboutModal.vue';
import { t, currentLang } from './i18n';
import { useToast } from './composables/useToast';
import { useIsoManager } from './composables/useIsoManager';
import { useVirtualMachine } from './composables/useVirtualMachine';
import { useDiskSelection, getDiskFingerprint } from './composables/useDiskSelection';
import { useDeployment } from './composables/useDeployment';
import { useLogPanel } from './composables/useLogPanel';
import { useAppSettings } from './composables/useAppSettings';
import { useAppRuntimeEvents } from './composables/useAppRuntimeEvents';

// 1. Global Toast
const { toastMessage, toastType, showToast, dismissToast } = useToast();

// 2. ISO Manager
const {
  selectedIsoFiles,
  isoCopyStatus,
  addIsoFilesByPaths,
  handleSelectIsoFiles,
  removeIsoFile,
  clearIsoFiles,
} = useIsoManager(showToast);

// 3. Disk Selection
const {
  selectionMode,
  diskList,
  selectedDisk,
  selectedDevices,
  customIcons,
  isPickerOpen,
  targetPickerDisk,
  isInspectorOpen,
  targetInspectorDisk,
  isScanningDisks,
  setSelectionMode,
  selectAllDisks,
  deselectAllDisks,
  onDiskSelect,
  onDiskToggle,
  openInspector,
  openIconPicker,
  onIconSelected,
  onIconReset,
  handleEjectDisk,
  handleBatchEjectDisks,
  handleSafelyEjectAfterDeploy,
  refreshDisks,
} = useDiskSelection({
  t,
  showToast,
  getTargetsToEject: () => deploySuccessBanner.value.targets,
  onEjectSuccess: (device) => {
    if (deploySuccessBanner.value.targets.includes(device)) {
      deploySuccessBanner.value.dismissed = true;
      deploySuccessBanner.value.visible = false;
    }
  },
  onSafelyEjectSuccess: () => {
    deploySuccessBanner.value.dismissed = true;
    deploySuccessBanner.value.visible = false;
  },
});

// 4. App Settings & Theme (Declared early so openSettings callback can be passed to useDeployment)
let openSettingsFn: (tab?: 'general' | 'network' | 'uniboot' | 'ventoy') => void;

// 5. Deployment Engine
const {
  activeMode,
  selectedFsType,
  autoEjectAfterDeploy,
  isDeploying,
  deployProgress,
  deploySpeedMBps,
  deployElapsedSec,
  deployEtaSec,
  isDeployConfirmOpen,
  pendingTargets,
  isVentoyAlertOpen,
  ventoyAlertTitle,
  ventoyAlertMessage,
  ventoyAlertAction,
  ventoyStatus,
  isDiagnosticsOpen,
  currentDiagnostics,
  currentDiagErrorMsg,
  deploySuccessBanner,
  isMacOs,
  isNonDestructive,
  deployBtnText,
  deployDisabledReason,
  activeVmTargetDevice,
  activeVmTargetName,
  showDeploySuccessBanner,
  checkVentoyStatus,
  selectMode,
  handleVentoyAlertAction,
  handleVentoyAlertSwitchB,
  handleCopyReport,
  handleRetryDeploy,
  handleDeployBtnClick,
  startDeployment,
  handleCancelDeploy,
  dismissDeploySuccessBanner,
} = useDeployment({
  t,
  showToast,
  selectionMode,
  selectedDisk,
  selectedDevices,
  diskList,
  selectedIsoFiles,
  refreshDisks,
  openSettings: (tab) => openSettingsFn(tab),
});

// 6. Virtual Machine / Hypervisor
const {
  hypervisorList,
  selectedBootMode,
  selectedVMType,
  vmCpuCores,
  vmMemoryMB,
  vmDisplayAccel,
  isLaunchingQemu,
  isVmDisabled,
  vmDisabledReason,
  checkQemu,
  launchVM,
} = useVirtualMachine({
  activeVmTargetDevice,
  diskList,
  isDeploying,
  showToast,
});

// 7. Log Center Panel
const {
  isLogCardVisible,
  embeddedAutoScroll,
  currentEmbeddedLogFilter,
  logLevels,
  filteredEmbeddedLogs,
  toggleLogCard,
  handleCopyEmbeddedLogs,
  handleExportEmbeddedLogs,
  handleClearEmbeddedLogs,
  appendLogEntry,
  setInitialLogs,
} = useLogPanel({
  t,
  showToast,
});

// 8. App Settings & Theme
const {
  settingsInitialTab,
  isSettingsOpen,
  isAboutOpen,
  currentGithubProxy,
  openSettings,
  selectLanguage,
  loadConfig,
  onSaveSettings,
} = useAppSettings({
  selectedFsType,
  activeMode,
  autoEjectAfterDeploy,
});

openSettingsFn = openSettings;

// 9. Wails Global Events & Lifecycle
useAppRuntimeEvents({
  t,
  loadConfig,
  refreshDisks,
  checkQemu,
  checkVentoyStatus,
  setInitialLogs,
  appendLogEntry,
  isoCopyStatus,
  deployProgress,
  deploySpeedMBps,
  deployElapsedSec,
  deployEtaSec,
  isDeploying,
  isAboutOpen,
  addIsoFilesByPaths,
});
</script>

<style scoped>
.app-container {
  max-width: 1280px;
  width: 95%;
  margin: 0 auto;
  padding: 2.2rem;
}

.content-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
  gap: 1.75rem;
}

@media (max-width: 960px) {
  .content-grid {
    grid-template-columns: minmax(0, 1fr);
  }
}

/* Global App Toast Notification Styles */
.global-toast {
  position: fixed;
  top: 24px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 99999;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 24px;
  border-radius: 12px;
  font-size: 0.9rem;
  font-weight: 600;
  color: var(--text-main);
  box-shadow: 0 12px 32px var(--modal-backdrop);
  backdrop-filter: blur(16px);
  -webkit-backdrop-filter: blur(16px);
  border: 1px solid var(--card-border);
  background: var(--card-bg);
}

.global-toast.warning {
  background: var(--alert-warning-bg);
  border-color: var(--alert-warning-border);
  color: var(--alert-warning-title);
}

.global-toast.error {
  background: var(--alert-danger-bg);
  border-color: var(--alert-danger-border);
  color: var(--alert-danger-title);
}

.global-toast.success {
  background: var(--alert-success-bg);
  border-color: var(--alert-success-border);
  color: var(--alert-success-title);
}

.global-toast.info {
  background: var(--alert-info-bg);
  border-color: var(--alert-info-border);
  color: var(--alert-info-title);
}

.toast-close {
  background: transparent;
  border: none;
  color: currentColor;
  font-size: 1rem;
  cursor: pointer;
  opacity: 0.8;
  margin-left: 8px;
  transition: opacity 0.2s ease;
}

.toast-close:hover {
  opacity: 1;
}

.toast-fade-enter-active,
.toast-fade-leave-active {
  transition: opacity 0.3s cubic-bezier(0.16, 1, 0.3, 1), transform 0.3s cubic-bezier(0.16, 1, 0.3, 1);
}

.toast-fade-enter-from,
.toast-fade-leave-to {
  opacity: 0;
  transform: translate(-50%, -20px);
}
</style>

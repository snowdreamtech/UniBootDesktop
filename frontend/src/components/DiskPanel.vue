<template>
  <section class="glass-card section-card">
    <div class="section-header-row">
      <div>
        <h2>{{ t('disk.select_title') }}</h2>
        <p class="section-desc">{{ t('disk.select_desc') }}</p>
      </div>

      <!-- Privilege Status Shield Badge -->
      <div class="privilege-badge-wrapper">
        <button
          class="privilege-status-badge"
          :class="{ elevated: isPrivileged, standard: !isPrivileged }"
          :title="isPrivileged ? t('privilege.status_elevated') : t('privilege.btn_elevate')"
          @click="handlePrivilegeBadgeClick"
        >
          <svg class="badge-shield-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
          </svg>
          <span>{{ isPrivileged ? t('privilege.status_elevated') : t('privilege.btn_elevate') }}</span>
        </button>
      </div>
    </div>

    <!-- Mode controls -->
    <div class="selection-controls">
      <div class="selection-mode-toggle">
        <button
          class="sub-tab-btn"
          :class="{ active: selectionMode === 'single' }"
          @click="emit('set-selection-mode', 'single')"
        >
          {{ t('disk.single_mode') }}
        </button>
        <button
          class="sub-tab-btn"
          :class="{ active: selectionMode === 'batch' }"
          @click="emit('set-selection-mode', 'batch')"
        >
          {{ t('disk.batch_mode') }}
        </button>
      </div>
    </div>

    <!-- Batch Actions Bar (另起一行，位于模式切换栏下方，防止窄窗口撑爆布局) -->
    <transition name="slide-fade">
      <div v-if="selectionMode === 'batch'" class="batch-actions-bar">
        <div class="batch-btn-group">
          <button
            class="batch-btn"
            :disabled="diskList.length === 0 || selectedDevices.size === diskList.length"
            @click="emit('select-all')"
          >
            {{ t('disk.select_all') }}
          </button>
          <button
            class="batch-btn"
            :disabled="selectedDevices.size === 0"
            @click="emit('deselect-all')"
          >
            {{ t('disk.clear_select') }}
          </button>
          <button
            class="batch-btn btn-eject"
            :disabled="selectedDevices.size === 0"
            :title="t('disk.batch_eject')"
            @click="emit('batch-eject')"
          >
            ⏏️ {{ t('disk.batch_eject') }}
          </button>
        </div>
        <div class="selection-count-badge">
          {{ t('disk.selected_count', { count: selectedDevices.size, total: diskList.length }) }}
        </div>
      </div>
    </transition>

    <transition-group name="disk-item" tag="div" class="disk-list">
      <DiskCard
        v-for="disk in diskList"
        :key="disk.device"
        :disk="disk"
        :isBatchMode="selectionMode === 'batch'"
        :isSelected="selectionMode === 'single' ? selectedDisk?.device === disk.device : selectedDevices.has(disk.device)"
        :customIcon="getCustomIcon(disk)"
        @select="emit('select-disk', disk)"
        @toggle="emit('toggle-disk', disk)"
        @pick-icon="emit('pick-icon', disk)"
        @inspect="emit('inspect-disk', disk)"
        @eject="emit('eject-disk', disk)"
      />
    </transition-group>

    <div v-if="diskList.length === 0" class="empty-state">
      <div v-if="isScanningDisks" class="scanning-state">
        <span class="spin-icon">🔄</span>
        <span>{{ t('disk.scanning') }}</span>
      </div>
      <span v-else>⚠️ {{ t('disk.empty_list') }}</span>
    </div>

    <button
      class="btn-secondary refresh-btn"
      :disabled="isScanningDisks"
      @click="emit('refresh-disks')"
    >
      <span class="refresh-icon">🔄</span>
      <span>{{ isScanningDisks ? t('disk.scanning') : t('disk.rescan') }}</span>
    </button>

    <!-- Privilege Trust Modal -->
    <PrivilegeTrustModal
      v-model:visible="showPrivilegeModal"
      @authorized="handlePrivilegeAuthorized"
    />
  </section>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import DiskCard from './DiskCard.vue';
import PrivilegeTrustModal from './PrivilegeTrustModal.vue';
import type { DiskIconType } from './IconPickerModal.vue';
import { t } from '../i18n';
import type { disk } from '../../wailsjs/go/models';

export type DiskInfo = disk.DiskInfo;

const props = defineProps<{
  diskList: DiskInfo[];
  selectionMode: 'single' | 'batch';
  selectedDisk: DiskInfo | null;
  selectedDevices: Set<string>;
  isScanningDisks: boolean;
  customIcons: Record<string, DiskIconType>;
}>();

const emit = defineEmits<{
  (e: 'set-selection-mode', mode: 'single' | 'batch'): void;
  (e: 'select-all'): void;
  (e: 'deselect-all'): void;
  (e: 'batch-eject'): void;
  (e: 'select-disk', disk: DiskInfo): void;
  (e: 'toggle-disk', disk: DiskInfo): void;
  (e: 'pick-icon', disk: DiskInfo): void;
  (e: 'inspect-disk', disk: DiskInfo): void;
  (e: 'eject-disk', disk: DiskInfo): void;
  (e: 'refresh-disks'): void;
}>();

const isPrivileged = ref(false);
const showPrivilegeModal = ref(false);

const checkPrivilegeStatus = async (retryCount = 0) => {
  try {
    const wailsAny = window as any;
    if (wailsAny.go?.main?.App?.IsPrivileged) {
      isPrivileged.value = await wailsAny.go.main.App.IsPrivileged();
      if (!isPrivileged.value) {
        showPrivilegeModal.value = true;
      }
    } else if (retryCount < 10) {
      setTimeout(() => checkPrivilegeStatus(retryCount + 1), 200);
    }
  } catch (e) {
    console.debug('Privilege check error:', e);
  }
};

onMounted(() => {
  checkPrivilegeStatus();
});

const handlePrivilegeBadgeClick = () => {
  showPrivilegeModal.value = true;
};

const handlePrivilegeAuthorized = () => {
  isPrivileged.value = true;
  emit('refresh-disks');
};

function getDiskFingerprint(disk: DiskInfo): string {
  if (disk.serialNumber && disk.serialNumber.trim() !== '') {
    return `sn:${disk.serialNumber.trim()}`;
  }
  return `dev:${disk.name}_${disk.size}`;
}

function getCustomIcon(disk: DiskInfo): DiskIconType | undefined {
  const fp = getDiskFingerprint(disk);
  return props.customIcons[fp] ?? props.customIcons[disk.device];
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

.section-header-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.5rem;
}

.privilege-status-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.35rem 0.75rem;
  border-radius: 9999px;
  font-size: 0.75rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
  border: 1px solid transparent;
}

.privilege-status-badge.elevated {
  background: rgba(16, 185, 129, 0.12);
  color: #10b981;
  border-color: rgba(16, 185, 129, 0.3);
}

.privilege-status-badge.standard {
  background: rgba(245, 158, 11, 0.12);
  color: #f59e0b;
  border-color: rgba(245, 158, 11, 0.3);
}

.privilege-status-badge:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.badge-shield-icon {
  width: 14px;
  height: 14px;
}

.section-header-row h2 {
  font-size: 1.2rem;
  font-weight: 700;
  margin: 0 0 0.25rem 0;
  color: var(--text-main);
}

.section-desc {
  font-size: 0.8rem;
  color: var(--text-muted);
  margin: 0;
}

.selection-controls {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin: 1rem 0;
  padding: 0.4rem;
  background: var(--subtab-container-bg);
  border-radius: 10px;
  border: 1px solid var(--card-border);
}

.selection-mode-toggle {
  display: flex;
  gap: 0.3rem;
}

.sub-tab-btn {
  background: transparent;
  border: 1px solid transparent;
  color: var(--subtab-btn-text);
  padding: 0.45rem 1rem;
  font-size: 0.825rem;
  font-weight: 600;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.sub-tab-btn.active {
  background: var(--subtab-btn-active-bg);
  color: var(--subtab-btn-active-text);
  border-color: var(--subtab-btn-active-border);
}

.batch-actions-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 0.5rem;
  margin: -0.4rem 0 0.85rem 0;
  padding: 0.45rem 0.65rem;
  background: var(--subtab-container-bg);
  border: 1px solid var(--card-border);
  border-radius: 8px;
}

.batch-btn-group {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 0.4rem;
}

.batch-btn {
  background: rgba(255, 255, 255, 0.08);
  border: 1px solid var(--card-border);
  color: var(--text-main);
  padding: 0.3rem 0.65rem;
  font-size: 0.78rem;
  font-weight: 600;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s ease;
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
}

.batch-btn:hover:not(:disabled) {
  background: var(--accent-cyan);
  color: #fff;
  border-color: var(--accent-cyan);
}

.batch-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.batch-btn.btn-eject {
  background: rgba(239, 68, 68, 0.12);
  border-color: rgba(239, 68, 68, 0.25);
  color: #f87171;
}

.batch-btn.btn-eject:hover:not(:disabled) {
  background: rgba(239, 68, 68, 0.25);
  border-color: rgba(239, 68, 68, 0.4);
  color: #ef4444;
}

.selection-count-badge {
  font-size: 0.78rem;
  font-weight: 600;
  color: var(--text-muted);
}

.slide-fade-enter-active,
.slide-fade-leave-active {
  transition: all 0.2s cubic-bezier(0.16, 1, 0.3, 1);
}

.slide-fade-enter-from,
.slide-fade-leave-to {
  opacity: 0;
  transform: translateY(-6px);
}

.disk-list {
  display: flex;
  flex-direction: column;
  gap: 0.85rem;
  flex: 1;
  min-height: 200px;
  max-height: 480px;
  overflow-y: auto;
  margin-bottom: 1rem;
}

.empty-state {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 180px;
  background: var(--btn-sec-bg);
  border: 2px dashed var(--card-border);
  border-radius: 12px;
  color: var(--text-muted);
  font-size: 0.85rem;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.spin-icon {
  display: inline-block;
  animation: spin 0.8s linear infinite;
}

.scanning-state {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: var(--accent-cyan);
  font-weight: 600;
}

.refresh-btn {
  width: 100%;
  padding: 0.75rem;
  font-size: 0.9rem;
  font-weight: 700;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.4rem;
}

/* Light Theme Overrides for DiskPanel */
[data-theme="light"] .section-card {
  background: #ffffff;
  border-color: #cbd5e1;
  box-shadow: 0 8px 32px rgba(15, 23, 42, 0.06);
}

[data-theme="light"] .section-header-row h2 {
  color: #0f172a;
}

[data-theme="light"] .section-desc {
  color: #64748b;
}

[data-theme="light"] .selection-controls {
  background: #f1f5f9;
  border-color: #cbd5e1;
}

[data-theme="light"] .sub-tab-btn {
  background: linear-gradient(180deg, #ffffff 0%, #f8fafc 100%);
  color: #475569;
  border: 1px solid #cbd5e1;
  box-shadow: inset 0 1px 0 #ffffff, 0 1px 3px rgba(15, 23, 42, 0.05);
}

[data-theme="light"] .sub-tab-btn:hover {
  color: #0f172a;
  background: linear-gradient(180deg, #f8fafc 0%, #f1f5f9 100%);
  border-color: #94a3b8;
}

[data-theme="light"] .sub-tab-btn.active {
  background: linear-gradient(135deg, #e0f2fe 0%, #bae6fd 100%);
  color: #0284c7;
  border-color: #38bdf8;
  font-weight: 700;
  box-shadow: inset 0 1px 0 #ffffff, 0 2px 8px rgba(2, 132, 199, 0.2);
}

[data-theme="light"] .btn-text {
  color: #0284c7;
}

[data-theme="light"] .selection-count {
  color: #475569;
}

[data-theme="light"] .empty-state {
  background: #f8fafc;
  border-color: #cbd5e1;
  color: #64748b;
}

/* Smooth transition for disk cards */
.disk-item-move,
.disk-item-enter-active,
.disk-item-leave-active {
  transition: all 0.35s cubic-bezier(0.4, 0, 0.2, 1);
}
.disk-item-enter-from,
.disk-item-leave-to {
  opacity: 0;
  transform: translateY(8px);
}
</style>

<template>
  <section class="glass-card section-card">
    <div class="section-header-row">
      <div>
        <h2>{{ t('disk.select_title') }}</h2>
        <p class="section-desc">{{ t('disk.select_desc') }}</p>
      </div>
    </div>

    <!-- Mode & Selection controls -->
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

      <div v-if="selectionMode === 'batch'" class="batch-actions">
        <button 
          class="btn-text" 
          :disabled="diskList.length === 0 || selectedDevices.size === diskList.length" 
          @click="emit('select-all')"
        >
          {{ t('disk.select_all') }}
        </button>
        <button 
          class="btn-text" 
          :disabled="selectedDevices.size === 0" 
          @click="emit('deselect-all')"
        >
          {{ t('disk.clear_select') }}
        </button>
        <button 
          class="btn-eject" 
          :disabled="selectedDevices.size === 0" 
          :title="t('disk.batch_eject')" 
          @click="emit('batch-eject')"
        >
          ⏏️ {{ t('disk.batch_eject') }}
        </button>
        <span class="selection-count">{{ t('disk.selected_count', { count: selectedDevices.size, total: diskList.length }) }}</span>
      </div>
    </div>

    <div class="disk-list">
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
      <div v-if="diskList.length === 0" class="empty-state">
        <div v-if="isScanningDisks" class="scanning-state">
          <span class="spin-icon">🔄</span>
          <span>{{ t('disk.scanning') }}</span>
        </div>
        <span v-else>⚠️ {{ t('disk.empty_list') }}</span>
      </div>
    </div>

    <button 
      class="btn-secondary refresh-btn" 
      :disabled="isScanningDisks"
      @click="emit('refresh-disks')"
    >
      <span class="refresh-icon">🔄</span>
      <span>{{ isScanningDisks ? t('disk.scanning') : t('disk.rescan') }}</span>
    </button>
  </section>
</template>

<script setup lang="ts">
import DiskCard from './DiskCard.vue';
import type { DiskIconType } from './IconPickerModal.vue';
import { t } from '../i18n';

export interface DiskInfo {
  device: string;
  name: string;
  size: number;
  formatted: string;
  freeSpace?: number;
  freeFormatted?: string;
  isRemovable: boolean;
  isSystem: boolean;
  usbVersion?: string;
  usbSpeed?: string;
  vendor?: string;
  fileSystem?: string;
  partitionScheme?: string;
  writable?: boolean;
  serialNumber?: string;
  vendorId?: string;
  productId?: string;
  smartStatus?: string;
  busPower?: string;
  busPowerUsed?: string;
  sectorSize?: string;
  transportProtocol?: string;
  bootStatus?: string;
}

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

function getCustomIcon(disk: DiskInfo): DiskIconType | undefined {
  return props.customIcons[disk.device];
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

.batch-actions {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.btn-text {
  background: transparent;
  border: none;
  color: var(--accent-cyan);
  font-size: 0.78rem;
  font-weight: 600;
  cursor: pointer;
}

.btn-text:hover:not(:disabled) {
  text-decoration: underline;
}

.btn-text:disabled {
  opacity: 0.4;
  cursor: not-allowed;
  pointer-events: none;
  text-decoration: none;
}

.btn-eject {
  background: rgba(239, 68, 68, 0.15);
  border: 1px solid rgba(239, 68, 68, 0.3);
  color: #f87171;
  padding: 0.3rem 0.6rem;
  border-radius: 6px;
  font-size: 0.78rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-eject:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.selection-count {
  font-size: 0.78rem;
  color: var(--text-muted);
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
</style>

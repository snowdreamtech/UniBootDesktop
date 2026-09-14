<template>
  <div v-if="isOpen" class="modal-overlay" @click="close">
    <div class="glass-modal confirm-card" :class="{ 'safe-card': isAllVentoy, 'mixed-card': isMixed }" @click.stop>
      <div class="modal-header" :class="isAllVentoy ? 'safe-header' : (isMixed ? 'mixed-header' : 'danger-header')">
        <div class="header-title">
          <span class="warning-icon">{{ isAllVentoy ? '🛡️' : (isMixed ? '⚡' : '⚠️') }}</span>
          <h3>
            {{ isAllVentoy ? t('confirm.title_safe') : (isMixed ? t('confirm.title_mixed') : t('confirm.title_danger')) }}
          </h3>
        </div>
        <button class="close-btn" @click="close">✕</button>
      </div>

      <div class="modal-body">
        <!-- Safe Info Banner for ALL Ventoy Disks -->
        <div v-if="isAllVentoy" class="safe-banner">
          <div class="banner-title">{{ t('confirm.safe_banner_title') }}</div>
          <div class="banner-desc">
            {{ t('confirm.safe_banner_desc') }}
          </div>
        </div>

        <!-- Mixed Mode Info Banner for Mixed Selections -->
        <div v-else-if="isMixed" class="mixed-banner">
          <div class="banner-title">{{ t('confirm.mixed_banner_title') }}</div>
          <div class="banner-desc">
            {{ t('confirm.mixed_banner_desc', { ventoyCount: ventoyDisks.length, blankCount: blankDisks.length }) }}
          </div>
        </div>

        <!-- Danger Warning Alert Banner for Pure Blank Disks -->
        <div v-else class="danger-banner">
          <div class="banner-title">{{ t('confirm.danger_banner_title') }}</div>
          <div class="banner-desc">
            {{ t('confirm.danger_banner_desc') }}
          </div>
        </div>

        <!-- Target Devices Summary Box -->
        <div class="target-summary-box">
          <div class="summary-label">{{ t('confirm.summary_title') }}</div>
          
          <!-- Single Disk Summary -->
          <div v-if="targetDisks.length === 1 && targetDisk" class="target-disk-item">
            <div class="disk-main-info">
              <span class="disk-name">{{ targetDisk.name }}</span>
              <span class="disk-path">{{ targetDisk.device }}</span>
            </div>
            <div class="disk-meta-pills">
              <span class="pill-tag">{{ targetDisk.formatted }}</span>
              <span class="pill-tag">{{ targetDisk.fileSystem || 'FAT32' }}</span>
              <span class="pill-tag accent" v-if="mode === 'hybrid'">{{ t('confirm.fs_format', { fs: fsType }) }}</span>
              <span class="pill-tag highlight">{{ mode === 'cloud' ? t('mode.cloud') : t('mode.hybrid') }}</span>
              <span class="pill-tag safe-tag" v-if="isAllVentoy">{{ t('confirm.smart_safe_tag') }}</span>
            </div>
          </div>

          <!-- Batch Disks Mixed Summary -->
          <div v-else-if="isMixed" class="batch-summary">
            <div class="mixed-group" v-if="ventoyDisks.length > 0">
              <div class="group-title safe-title">{{ t('confirm.ventoy_group_title') }}</div>
              <div class="batch-tags">
                <span v-for="dev in ventoyDisks" :key="dev" class="batch-dev-tag safe-dev-tag">🛡️ {{ dev }}</span>
              </div>
            </div>
            <div class="mixed-group" v-if="blankDisks.length > 0">
              <div class="group-title danger-title">{{ t('confirm.blank_group_title') }}</div>
              <div class="batch-tags">
                <span v-for="dev in blankDisks" :key="dev" class="batch-dev-tag danger-dev-tag">💾 {{ dev }}</span>
              </div>
            </div>
          </div>

          <!-- Batch Disks Pure Summary -->
          <div v-else class="batch-summary">
            <div class="batch-count">{{ t('confirm.batch_summary_title', { count: targetDisks.length }) }}</div>
            <div class="batch-tags">
              <span v-for="dev in targetDisks" :key="dev" class="batch-dev-tag">💾 {{ dev }}</span>
            </div>
          </div>
        </div>
      </div>

      <div class="modal-footer">
        <button class="btn-cancel" @click="close">{{ t('confirm.cancel_btn') }}</button>
        <button :class="isAllVentoy || isMixed ? 'btn-safe-confirm' : 'btn-danger-confirm'" @click="confirm">
          {{ isAllVentoy ? t('deploy.start_update') : (isMixed ? '🚀 ' + t('confirm.confirm_btn') : t('confirm.confirm_btn')) }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { t } from '../i18n';

interface DiskInfo {
  device: string;
  name: string;
  size: number;
  formatted: string;
  fileSystem?: string;
  bootStatus?: string;
}

const props = defineProps<{
  isOpen: boolean;
  mode: 'cloud' | 'hybrid';
  fsType?: string;
  targetDisk: DiskInfo | null;
  targetDisks: string[];
  allDisks?: DiskInfo[];
}>();

const emit = defineEmits(['close', 'confirm']);

const ventoyDisks = computed(() => {
  if (!props.allDisks || props.allDisks.length === 0) {
    if (props.targetDisk) {
      const name = (props.targetDisk.name || '').toUpperCase();
      const status = (props.targetDisk.bootStatus || '').toUpperCase();
      if (name.includes('VENTOY') || status.includes('VENTOY') || status.includes('UNIBOOT')) {
        return [props.targetDisk.device];
      }
    }
    return [];
  }
  return props.targetDisks.filter(dev => {
    const found = props.allDisks?.find(d => d.device === dev);
    if (found) {
      const name = (found.name || '').toUpperCase();
      const status = (found.bootStatus || '').toUpperCase();
      return name.includes('VENTOY') || status.includes('VENTOY') || status.includes('UNIBOOT');
    }
    return dev.toUpperCase().includes('VENTOY');
  });
});

const blankDisks = computed(() => {
  return props.targetDisks.filter(dev => !ventoyDisks.value.includes(dev));
});

const isAllVentoy = computed(() => {
  if (props.targetDisks.length === 0) return false;
  return ventoyDisks.value.length === props.targetDisks.length;
});

const isMixed = computed(() => {
  return ventoyDisks.value.length > 0 && blankDisks.value.length > 0;
});

function close() {
  emit('close');
}

function confirm() {
  emit('confirm');
  close();
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(7, 10, 18, 0.82);
  backdrop-filter: blur(10px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1100;
  padding: 1rem;
}

.glass-modal {
  background: rgba(18, 24, 38, 0.96);
  border: 1px solid rgba(239, 68, 68, 0.4);
  box-shadow: 0 20px 50px rgba(0, 0, 0, 0.7), 0 0 25px rgba(239, 68, 68, 0.2);
  border-radius: 16px;
  width: 100%;
  max-width: 580px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.25rem 1.5rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
}

.glass-modal.safe-card {
  border: 1px solid rgba(0, 229, 255, 0.4);
  box-shadow: 0 20px 50px rgba(0, 0, 0, 0.7), 0 0 25px rgba(0, 229, 255, 0.2);
}

.glass-modal.mixed-card {
  border: 1px solid rgba(168, 85, 247, 0.4);
  box-shadow: 0 20px 50px rgba(0, 0, 0, 0.7), 0 0 25px rgba(168, 85, 247, 0.2);
}

.modal-header.safe-header {
  background: rgba(0, 229, 255, 0.08);
}

.modal-header.safe-header h3 {
  color: var(--accent-cyan);
}

.modal-header.mixed-header {
  background: rgba(168, 85, 247, 0.08);
}

.modal-header.mixed-header h3 {
  color: #d8b4fe;
}

.header-title {
  display: flex;
  align-items: center;
  gap: 0.6rem;
}

.warning-icon {
  font-size: 1.3rem;
}

.modal-header h3 {
  font-size: 1.15rem;
  font-weight: 700;
  color: #ef4444;
  margin: 0;
}

.close-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  font-size: 1.2rem;
  cursor: pointer;
  padding: 0.2rem 0.5rem;
  transition: color 0.2s;
}

.close-btn:hover {
  color: #fff;
}

.modal-body {
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.safe-banner {
  background: rgba(0, 229, 255, 0.08);
  border: 1px solid rgba(0, 229, 255, 0.35);
  border-radius: 12px;
  padding: 1rem 1.25rem;
}

.safe-banner .banner-title {
  color: var(--accent-cyan);
  font-weight: 700;
  font-size: 0.95rem;
  margin-bottom: 0.4rem;
}

.safe-banner .banner-desc {
  color: #a5f3fc;
  font-size: 0.825rem;
  line-height: 1.5;
}

.mixed-banner {
  background: rgba(168, 85, 247, 0.1);
  border: 1px solid rgba(168, 85, 247, 0.35);
  border-radius: 12px;
  padding: 1rem 1.25rem;
}

.mixed-banner .banner-title {
  color: #d8b4fe;
  font-weight: 700;
  font-size: 0.95rem;
  margin-bottom: 0.4rem;
}

.mixed-banner .banner-desc {
  color: #e9d5ff;
  font-size: 0.825rem;
  line-height: 1.5;
}

.mixed-group {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
  margin-bottom: 0.75rem;
}

.group-title {
  font-size: 0.8rem;
  font-weight: 600;
}

.group-title.safe-title {
  color: #4ade80;
}

.group-title.danger-title {
  color: #f87171;
}

.batch-dev-tag.safe-dev-tag {
  background: rgba(34, 197, 94, 0.12);
  border-color: rgba(34, 197, 94, 0.3);
  color: #4ade80;
}

.batch-dev-tag.danger-dev-tag {
  background: rgba(239, 68, 68, 0.12);
  border-color: rgba(239, 68, 68, 0.3);
  color: #fca5a5;
}

.danger-banner {
  background: rgba(239, 68, 68, 0.12);
  border: 1px solid rgba(239, 68, 68, 0.35);
  border-radius: 12px;
  padding: 1rem 1.25rem;
}

.banner-title {
  color: #ef4444;
  font-weight: 700;
  font-size: 0.95rem;
  margin-bottom: 0.4rem;
}

.banner-desc {
  color: #fca5a5;
  font-size: 0.825rem;
  line-height: 1.5;
}

.target-summary-box {
  background: rgba(255, 255, 255, 0.02);
  border: 1px solid var(--card-border);
  border-radius: 12px;
  padding: 1.25rem;
}

.summary-label {
  font-size: 0.825rem;
  color: var(--text-muted);
  margin-bottom: 0.75rem;
}

.target-disk-item {
  display: flex;
  flex-direction: column;
  gap: 0.6rem;
}

.disk-main-info {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
}

.disk-name {
  font-size: 1rem;
  font-weight: 700;
  color: #fff;
}

.disk-path {
  font-family: monospace;
  font-size: 0.8rem;
  color: var(--accent-cyan);
}

.disk-meta-pills {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.pill-tag {
  background: rgba(255, 255, 255, 0.06);
  border: 1px solid rgba(255, 255, 255, 0.1);
  padding: 0.2rem 0.6rem;
  border-radius: 6px;
  font-size: 0.75rem;
  color: var(--text-color);
}

.pill-tag.accent {
  background: rgba(0, 229, 255, 0.12);
  border-color: rgba(0, 229, 255, 0.3);
  color: var(--accent-cyan);
  font-weight: 600;
}

.pill-tag.highlight {
  background: rgba(168, 85, 247, 0.15);
  border-color: rgba(168, 85, 247, 0.3);
  color: #d8b4fe;
  font-weight: 600;
}

.batch-summary {
  display: flex;
  flex-direction: column;
  gap: 0.6rem;
}

.batch-count {
  font-size: 0.9rem;
  color: #fff;
}

.batch-tags {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.batch-dev-tag {
  background: rgba(0, 229, 255, 0.1);
  border: 1px solid rgba(0, 229, 255, 0.2);
  padding: 0.25rem 0.6rem;
  border-radius: 6px;
  font-family: monospace;
  font-size: 0.8rem;
  color: var(--accent-cyan);
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.8rem;
  padding: 1rem 1.5rem;
  background: rgba(0, 0, 0, 0.2);
  border-top: 1px solid var(--card-border);
}

.btn-cancel {
  background: rgba(255, 255, 255, 0.06);
  border: 1px solid var(--card-border);
  color: var(--text-color);
  padding: 0.6rem 1.25rem;
  border-radius: 8px;
  font-size: 0.875rem;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-cancel:hover {
  background: rgba(255, 255, 255, 0.12);
}

.pill-tag.safe-tag {
  background: rgba(34, 197, 94, 0.15);
  border-color: rgba(34, 197, 94, 0.3);
  color: #4ade80;
  font-weight: 600;
}

.btn-safe-confirm {
  background: linear-gradient(135deg, #00e5ff 0%, #0284c7 100%);
  border: none;
  color: #070a12;
  padding: 0.6rem 1.25rem;
  border-radius: 8px;
  font-size: 0.875rem;
  font-weight: 700;
  cursor: pointer;
  box-shadow: 0 4px 14px rgba(0, 229, 255, 0.35);
  transition: all 0.2s;
}

.btn-safe-confirm:hover {
  background: linear-gradient(135deg, #38bdf8 0%, #00e5ff 100%);
  box-shadow: 0 6px 20px rgba(0, 229, 255, 0.5);
  transform: translateY(-1px);
}

.btn-danger-confirm {
  background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
  border: none;
  color: #fff;
  padding: 0.6rem 1.25rem;
  border-radius: 8px;
  font-size: 0.875rem;
  font-weight: 700;
  cursor: pointer;
  box-shadow: 0 4px 14px rgba(239, 68, 68, 0.4);
  transition: all 0.2s;
}

.btn-danger-confirm:hover {
  background: linear-gradient(135deg, #f87171 0%, #ef4444 100%);
  box-shadow: 0 6px 20px rgba(239, 68, 68, 0.6);
  transform: translateY(-1px);
}
</style>

<template>
  <div v-if="isOpen" class="modal-overlay" @click.self="close">
    <div class="glass-modal confirm-card" @click.stop>
      <div class="modal-header safe-header">
        <div class="header-title">
          <span class="eject-icon">⏏️</span>
          <h3>{{ t('deploy.confirm_auto_eject_title') }}</h3>
        </div>
        <button class="close-btn" @click="close">✕</button>
      </div>

      <div class="modal-body">
        <div class="alert-banner safe-banner">
          <div class="banner-title">🎉 {{ t('deploy.alert_success') }}</div>
          <div class="banner-desc">{{ t('deploy.confirm_auto_eject_desc') }}</div>
        </div>

        <div v-if="targets && targets.length > 0" class="targets-summary">
          <span class="summary-label">U 盘目标 / Targets:</span>
          <div class="targets-tags">
            <span v-for="target in targets" :key="target" class="target-tag">
              💾 {{ target }}
            </span>
          </div>
        </div>

        <div class="action-buttons-group">
          <button class="btn-primary flex-btn eject-btn" @click="onConfirm">
            ⏏️ {{ t('deploy.confirm_auto_eject_yes') }}
          </button>
          <button class="btn-secondary flex-btn" @click="close">
            {{ t('deploy.confirm_auto_eject_no') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { t } from '../i18n';

defineProps<{
  isOpen: boolean;
  targets: string[];
}>();

const emit = defineEmits(['close', 'confirm']);

function close() {
  emit('close');
}

function onConfirm() {
  emit('confirm');
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(15, 23, 42, 0.65);
  backdrop-filter: blur(8px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
}

.glass-modal {
  background: var(--modal-bg, #1e293b);
  border: 1px solid var(--card-border, rgba(255, 255, 255, 0.1));
  border-radius: 16px;
  width: 90%;
  max-width: 480px;
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.35);
  overflow: hidden;
  color: var(--text-main, #f8fafc);
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  border-bottom: 1px solid var(--card-border, rgba(255, 255, 255, 0.1));
  background: var(--modal-header-bg, rgba(30, 41, 59, 0.8));
}

.safe-header {
  border-left: 4px solid var(--success, #10b981);
}

.header-title {
  display: flex;
  align-items: center;
  gap: 10px;
}

.header-title h3 {
  margin: 0;
  font-size: 1.1rem;
  font-weight: 600;
}

.eject-icon {
  font-size: 1.2rem;
}

.close-btn {
  background: transparent;
  border: none;
  color: var(--text-muted, #94a3b8);
  font-size: 1.2rem;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 6px;
  transition: all 0.2s;
}

.close-btn:hover {
  background: rgba(255, 255, 255, 0.1);
  color: var(--text-main, #f8fafc);
}

.modal-body {
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.alert-banner {
  padding: 14px 16px;
  border-radius: 10px;
  background: rgba(16, 185, 129, 0.1);
  border: 1px solid rgba(16, 185, 129, 0.3);
}

.banner-title {
  font-weight: 600;
  font-size: 0.95rem;
  color: var(--success, #10b981);
  margin-bottom: 4px;
}

.banner-desc {
  font-size: 0.85rem;
  color: var(--text-muted, #cbd5e1);
  line-height: 1.4;
}

.targets-summary {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.summary-label {
  font-size: 0.8rem;
  color: var(--text-muted, #94a3b8);
  font-weight: 500;
}

.targets-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.target-tag {
  background: rgba(255, 255, 255, 0.08);
  border: 1px solid rgba(255, 255, 255, 0.15);
  padding: 4px 10px;
  border-radius: 6px;
  font-size: 0.82rem;
  font-family: monospace;
  color: var(--text-main, #f8fafc);
}

.action-buttons-group {
  display: flex;
  gap: 12px;
  margin-top: 8px;
}

.flex-btn {
  flex: 1;
  padding: 10px 16px;
  border-radius: 10px;
  font-weight: 600;
  font-size: 0.9rem;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-primary.eject-btn {
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  color: white;
  border: none;
  box-shadow: 0 4px 12px rgba(16, 185, 129, 0.3);
}

.btn-primary.eject-btn:hover {
  background: linear-gradient(135deg, #059669 0%, #047857 100%);
  transform: translateY(-1px);
}

.btn-secondary {
  background: rgba(255, 255, 255, 0.08);
  color: var(--text-main, #cbd5e1);
  border: 1px solid var(--card-border, rgba(255, 255, 255, 0.15));
}

.btn-secondary:hover {
  background: rgba(255, 255, 255, 0.15);
  color: #fff;
}
</style>

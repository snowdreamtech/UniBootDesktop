<template>
  <div v-if="isOpen" class="modal-overlay" @click.self="close">
    <div class="glass-modal alert-card" @click.stop>
      <div class="modal-header danger-header">
        <div class="header-title">
          <span class="warning-icon">⚠️</span>
          <h3>{{ title || t('ventoy_alert.default_title') }}</h3>
        </div>
        <button class="close-btn" @click="close">✕</button>
      </div>

      <div class="modal-body">
        <div class="alert-banner">
          <div class="banner-title">{{ t('ventoy_alert.banner_title') }}</div>
          <div class="banner-desc">{{ message }}</div>
        </div>

        <div class="action-buttons-group">
          <button v-if="actionType === 'open_settings'" class="btn-primary flex-btn" @click="onAction">
            {{ t('ventoy_alert.goto_settings') }}
          </button>
          <button class="btn-accent flex-btn" @click="onSwitchB">
            {{ t('ventoy_alert.switch_b') }}
          </button>
          <button class="btn-secondary flex-btn" @click="close">
            {{ t('ventoy_alert.close') }}
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
  title: string;
  message: string;
  actionType: 'open_settings' | 'switch_b';
}>();

const emit = defineEmits(['close', 'action', 'switch-b']);

function close() {
  emit('close');
}

function onAction() {
  emit('action');
}

function onSwitchB() {
  emit('switch-b');
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(15, 23, 42, 0.75);
  backdrop-filter: blur(8px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
}

.glass-modal {
  background: var(--modal-bg, rgba(30, 41, 59, 0.95));
  border: 1px solid rgba(239, 68, 68, 0.3);
  border-radius: 16px;
  width: 90%;
  max-width: 520px;
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.6);
  overflow: hidden;
  color: #f8fafc;
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.danger-header {
  background: linear-gradient(135deg, rgba(239, 68, 68, 0.2), rgba(185, 28, 28, 0.1));
}

.header-title {
  display: flex;
  align-items: center;
  gap: 10px;
}

.header-title h3 {
  margin: 0;
  font-size: 1.15rem;
  font-weight: 600;
  color: #fef2f2;
}

.close-btn {
  background: transparent;
  border: none;
  color: #94a3b8;
  font-size: 1.2rem;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 6px;
  transition: all 0.2s;
}

.close-btn:hover {
  background: rgba(255, 255, 255, 0.1);
  color: #fff;
}

.modal-body {
  padding: 20px;
}

.alert-banner {
  background: rgba(239, 68, 68, 0.1);
  border: 1px solid rgba(239, 68, 68, 0.25);
  border-radius: 10px;
  padding: 14px 16px;
  margin-bottom: 20px;
}

.banner-title {
  font-weight: 600;
  color: #fca5a5;
  margin-bottom: 6px;
  font-size: 0.95rem;
}

.banner-desc {
  color: #cbd5e1;
  font-size: 0.9rem;
  line-height: 1.5;
}

.action-buttons-group {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.flex-btn {
  width: 100%;
  padding: 11px 16px;
  border-radius: 10px;
  font-weight: 600;
  font-size: 0.95rem;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.btn-primary {
  background: linear-gradient(135deg, #3b82f6, #2563eb);
  color: white;
  border: none;
}

.btn-primary:hover {
  background: linear-gradient(135deg, #60a5fa, #3b82f6);
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.35);
}

.btn-accent {
  background: linear-gradient(135deg, #10b981, #059669);
  color: white;
  border: none;
}

.btn-accent:hover {
  background: linear-gradient(135deg, #34d399, #10b981);
  box-shadow: 0 4px 12px rgba(16, 185, 129, 0.35);
}

.btn-secondary {
  background: rgba(51, 65, 85, 0.8);
  color: #cbd5e1;
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.btn-secondary:hover {
  background: rgba(71, 85, 105, 0.9);
  color: #fff;
}

/* Light Theme Overrides */
[data-theme="light"] .glass-modal {
  background: #ffffff;
  border-color: #fca5a5;
  color: #0f172a;
  box-shadow: 0 25px 50px -12px rgba(15, 23, 42, 0.2);
}

[data-theme="light"] .btn-secondary {
  background: #f1f5f9;
  color: #0f172a;
  border-color: #cbd5e1;
}

[data-theme="light"] .btn-secondary:hover {
  background: #e2e8f0;
}
</style>

<template>
  <div v-if="isOpen" class="modal-overlay" @click.self="close">
    <div class="glass-modal alert-card" @click.stop>
      <div class="modal-header danger-header">
        <div class="header-title">
          <span class="warning-icon">⚠️</span>
          <h3>{{ title || t("ventoy_alert.default_title") }}</h3>
        </div>
        <button class="close-btn" @click="close">✕</button>
      </div>

      <div class="modal-body">
        <div class="alert-banner">
          <div class="banner-title"><span class="banner-icon">💡</span> {{ t("ventoy_alert.banner_title") }}</div>
          <div class="banner-desc">{{ message }}</div>
        </div>

        <div class="action-buttons-group">
          <button v-if="actionType === 'open_settings'" class="btn-primary flex-btn" @click="onAction">
            <span class="btn-icon">⚙️</span>
            <span>{{ t("ventoy_alert.goto_settings") }}</span>
          </button>
          <button class="btn-accent flex-btn" @click="onSwitchB">
            <span class="btn-icon">🚀</span>
            <span>{{ t("ventoy_alert.switch_b") }}</span>
          </button>
          <button class="btn-secondary flex-btn" @click="close">
            <span>{{ t("ventoy_alert.close") }}</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { t } from "../i18n";

defineProps<{
  isOpen: boolean;
  title: string;
  message: string;
  actionType: "open_settings" | "switch_b";
}>();

const emit = defineEmits(["close", "action", "switch-b"]);

function close() {
  emit("close");
}

function onAction() {
  emit("action");
}

function onSwitchB() {
  emit("switch-b");
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
  background: var(--modal-bg);
  border: 1px solid var(--card-border);
  border-radius: 16px;
  width: 90%;
  max-width: 520px;
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
  overflow: hidden;
  color: var(--text-main);
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  border-bottom: 1px solid var(--card-border);
  background: var(--modal-header-bg);
}

.danger-header {
  border-left: 4px solid var(--warning);
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
  color: var(--text-main);
}

.close-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  font-size: 1.2rem;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 6px;
  transition: all 0.2s;
}

.close-btn:hover {
  background: var(--section-bg);
  color: var(--text-main);
}

.modal-body {
  padding: 20px;
}

.alert-banner {
  background: var(--alert-warning-bg);
  border: 1px solid var(--alert-warning-border);
  border-radius: 10px;
  padding: 14px 16px;
  margin-bottom: 20px;
}

.banner-title {
  font-weight: 600;
  color: var(--alert-warning-title);
  margin-bottom: 6px;
  font-size: 0.95rem;
}

.banner-desc {
  color: var(--alert-warning-text);
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
  background: #2563eb;
  color: #ffffff;
  border: none;
}

.btn-primary:hover {
  background: #1d4ed8;
  box-shadow: 0 4px 12px rgba(37, 99, 235, 0.25);
}

.btn-accent {
  background: #059669;
  color: #ffffff;
  border: none;
}

.btn-accent:hover {
  background: #047857;
  box-shadow: 0 4px 12px rgba(5, 150, 105, 0.25);
}

.btn-secondary {
  background: var(--section-bg);
  color: var(--text-main);
  border: 1px solid var(--card-border);
}

.btn-secondary:hover {
  background: var(--card-border);
}
</style>

<template>
  <Teleport to="body">
    <div v-if="visible" class="privilege-modal-overlay" @click.self="handleCancel">
      <div class="privilege-modal-card">
        <!-- Modal Header -->
        <div class="modal-header">
          <div class="shield-badge">
            <svg class="shield-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor">
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"
              />
            </svg>
          </div>
          <div class="header-text">
            <h2 class="title">{{ t("privilege.modal_title") }}</h2>
            <p class="subtitle">{{ t("privilege.modal_subtitle") }}</p>
          </div>
          <button class="close-btn" @click="handleCancel" aria-label="Close">
            <svg viewBox="0 0 24 24" width="18" height="18" stroke="currentColor" stroke-width="2" fill="none">
              <line x1="18" y1="6" x2="6" y2="18"></line>
              <line x1="6" y1="6" x2="18" y2="18"></line>
            </svg>
          </button>
        </div>

        <!-- Trust Pillars -->
        <div class="trust-pillars">
          <!-- Pillar 1: Why Needed -->
          <div class="pillar-card">
            <div class="pillar-icon-box blue">
              <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2">
                <rect x="2" y="2" width="20" height="8" rx="2" ry="2"></rect>
                <rect x="2" y="14" width="20" height="8" rx="2" ry="2"></rect>
                <line x1="6" y1="6" x2="6.01" y2="6"></line>
                <line x1="6" y1="18" x2="6.01" y2="18"></line>
              </svg>
            </div>
            <div class="pillar-content">
              <h4>{{ t("privilege.reason_title") }}</h4>
              <p>{{ t("privilege.reason_desc") }}</p>
            </div>
          </div>

          <!-- Pillar 2: Restricted Scope -->
          <div class="pillar-card">
            <div class="pillar-icon-box green">
              <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"></path>
              </svg>
            </div>
            <div class="pillar-content">
              <h4>{{ t("privilege.scope_title") }}</h4>
              <p>{{ t("privilege.scope_desc") }}</p>
            </div>
          </div>

          <!-- Pillar 3: Read-Only Safety -->
          <div class="pillar-card">
            <div class="pillar-icon-box amber">
              <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2">
                <circle cx="12" cy="12" r="10"></circle>
                <polyline points="12 6 12 12 14 14"></polyline>
              </svg>
            </div>
            <div class="pillar-content">
              <h4>{{ t("privilege.safety_title") }}</h4>
              <p>{{ t("privilege.safety_desc") }}</p>
            </div>
          </div>
        </div>

        <!-- Action Buttons -->
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="handleCancel" :disabled="authorizing">
            {{ t("privilege.cancel_btn") }}
          </button>
          <button class="btn btn-primary" @click="handleConfirm" :disabled="authorizing">
            <span v-if="authorizing" class="spinner"></span>
            <span v-else>{{ t("privilege.confirm_btn") }}</span>
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { t } from "../i18n";

defineProps<{
  visible: boolean;
}>();

const emit = defineEmits<{
  (e: "update:visible", val: boolean): void;
  (e: "authorized"): void;
}>();

const authorizing = ref(false);

const handleCancel = () => {
  if (authorizing.value) return;
  emit("update:visible", false);
};

const handleConfirm = async () => {
  authorizing.value = true;
  try {
    const wailsAny = window as any;
    if (wailsAny.go && wailsAny.go.main && wailsAny.go.main.App && wailsAny.go.main.App.RequestPrivilegeElevation) {
      const promptText = t("privilege.elevation_prompt");
      const ok = await wailsAny.go.main.App.RequestPrivilegeElevation(promptText);
      if (ok) {
        emit("authorized");
        emit("update:visible", false);
      }
    } else {
      // Fallback for standalone mock/dev
      emit("authorized");
      emit("update:visible", false);
    }
  } catch (err) {
    console.error("Elevation request error:", err);
  } finally {
    authorizing.value = false;
  }
};
</script>

<style scoped>
.privilege-modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(10, 14, 23, 0.75);
  backdrop-filter: blur(8px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100000;
  padding: 1.5rem;
  animation: fadeIn 0.2s ease-out;
}

.privilege-modal-card {
  width: 100%;
  max-width: 580px;
  max-height: 90vh;
  background: var(--modal-bg, #131c2e);
  border: 1px solid var(--card-border, rgba(255, 255, 255, 0.12));
  border-radius: 1rem;
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.5);
  overflow: hidden;
  display: flex;
  flex-direction: column;
  transition:
    background 0.3s ease,
    border-color 0.3s ease;
}

.modal-header {
  padding: 1.5rem;
  background: var(--section-bg, rgba(255, 255, 255, 0.03));
  border-bottom: 1px solid var(--card-border, rgba(255, 255, 255, 0.08));
  display: flex;
  align-items: center;
  gap: 1rem;
  position: relative;
}

.shield-badge {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  background: linear-gradient(135deg, rgba(16, 185, 129, 0.2), rgba(6, 95, 70, 0.3));
  border: 1px solid rgba(16, 185, 129, 0.4);
  color: #10b981;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.shield-icon {
  width: 28px;
  height: 28px;
}

.header-text {
  flex: 1;
}

.title {
  margin: 0;
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--text-main, #ffffff);
}

.subtitle {
  margin: 0.25rem 0 0 0;
  font-size: 0.85rem;
  color: var(--text-muted, #94a3b8);
}

.close-btn {
  background: transparent;
  border: none;
  color: var(--text-muted, #94a3b8);
  cursor: pointer;
  padding: 0.5rem;
  border-radius: 6px;
  transition: all 0.15s;
}

.close-btn:hover {
  background: rgba(255, 255, 255, 0.08);
  color: var(--text-main, #ffffff);
}

.trust-pillars {
  padding: 1.25rem 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
  overflow-y: auto;
  max-height: calc(90vh - 160px);
}

.pillar-card {
  display: flex;
  gap: 1rem;
  align-items: flex-start;
  padding: 0.85rem 1rem;
  border-radius: 0.75rem;
  background: rgba(255, 255, 255, 0.02);
  border: 1px solid rgba(255, 255, 255, 0.06);
  transition:
    background 0.2s ease,
    border-color 0.2s ease;
}

.pillar-icon-box {
  width: 36px;
  height: 36px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.pillar-icon-box.blue {
  background: rgba(59, 130, 246, 0.15);
  color: #60a5fa;
  border: 1px solid rgba(59, 130, 246, 0.3);
}

.pillar-icon-box.green {
  background: rgba(16, 185, 129, 0.15);
  color: #34d399;
  border: 1px solid rgba(16, 185, 129, 0.3);
}

.pillar-icon-box.amber {
  background: rgba(245, 158, 11, 0.15);
  color: #fbbf24;
  border: 1px solid rgba(245, 158, 11, 0.3);
}

.pillar-content h4 {
  margin: 0;
  font-size: 0.95rem;
  font-weight: 600;
  color: var(--text-main, #ffffff);
}

.pillar-content p {
  margin: 0.25rem 0 0 0;
  font-size: 0.82rem;
  line-height: 1.4;
  color: var(--text-muted, #94a3b8);
}

.modal-footer {
  padding: 1.25rem 1.5rem;
  background: var(--section-bg, rgba(255, 255, 255, 0.02));
  border-top: 1px solid var(--card-border, rgba(255, 255, 255, 0.08));
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
}

.btn {
  padding: 0.6rem 1.25rem;
  border-radius: 0.5rem;
  font-size: 0.875rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
  border: 1px solid transparent;
}

.btn-secondary {
  background: var(--btn-sec-bg, rgba(255, 255, 255, 0.06));
  color: var(--btn-sec-text, #94a3b8);
  border: 1px solid var(--btn-sec-border, rgba(255, 255, 255, 0.1));
}

.btn-secondary:hover:not(:disabled) {
  background: var(--btn-sec-hover-bg, rgba(255, 255, 255, 0.1));
  color: var(--btn-sec-hover-text, #ffffff);
  border-color: var(--btn-sec-hover-border, rgba(255, 255, 255, 0.2));
}

.btn-primary {
  background: linear-gradient(135deg, #0396e6 0%, #0284c7 45%, #2563eb 100%);
  color: #ffffff;
  border: 1px solid rgba(255, 255, 255, 0.35);
  box-shadow: 0 4px 14px rgba(2, 132, 199, 0.35);
}

.btn-primary:hover:not(:disabled) {
  background: linear-gradient(135deg, #38bdf8 0%, #0284c7 45%, #1d4ed8 100%);
  box-shadow: 0 6px 18px rgba(2, 132, 199, 0.45);
  transform: translateY(-1px);
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.spinner {
  width: 18px;
  height: 18px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-radius: 50%;
  border-top-color: #ffffff;
  animation: spin 0.8s linear infinite;
  display: inline-block;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: scale(0.98);
  }
  to {
    opacity: 1;
    transform: scale(1);
  }
}

/* Light Theme Overrides */
[data-theme="light"] .privilege-modal-overlay {
  background: rgba(15, 23, 42, 0.45);
}

[data-theme="light"] .privilege-modal-card {
  background: #ffffff;
  border-color: #cbd5e1;
  box-shadow: 0 25px 50px -12px rgba(15, 23, 42, 0.18);
}

[data-theme="light"] .modal-header {
  background: #f8fafc;
  border-bottom-color: #e2e8f0;
}

[data-theme="light"] .shield-badge {
  background: #dcfce7;
  border-color: #86efac;
  color: #15803d;
}

[data-theme="light"] .title {
  color: #0f172a;
}

[data-theme="light"] .subtitle {
  color: #475569;
}

[data-theme="light"] .close-btn {
  color: #64748b;
}

[data-theme="light"] .close-btn:hover {
  background: #f1f5f9;
  color: #0f172a;
}

[data-theme="light"] .pillar-card {
  background: #f8fafc;
  border-color: #e2e8f0;
}

[data-theme="light"] .pillar-icon-box.blue {
  background: #e0f2fe;
  color: #0284c7;
  border-color: #bae6fd;
}

[data-theme="light"] .pillar-icon-box.green {
  background: #dcfce7;
  color: #15803d;
  border-color: #bbf7d0;
}

[data-theme="light"] .pillar-icon-box.amber {
  background: #fef3c7;
  color: #92400e;
  border-color: #fde68a;
}

[data-theme="light"] .pillar-content h4 {
  color: #0f172a;
}

[data-theme="light"] .pillar-content p {
  color: #475569;
}

[data-theme="light"] .modal-footer {
  background: #f8fafc;
  border-top-color: #e2e8f0;
}

[data-theme="light"] .btn-secondary {
  background: #f1f5f9;
  color: #334155;
  border-color: #cbd5e1;
}

[data-theme="light"] .btn-secondary:hover:not(:disabled) {
  background: #e2e8f0;
  color: #0f172a;
  border-color: #94a3b8;
}
</style>

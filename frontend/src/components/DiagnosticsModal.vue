<template>
  <div v-if="isOpen" class="modal-overlay" @click.self="close">
    <div class="glass-modal diag-card" @click.stop>
      <div class="modal-header danger-header">
        <div class="header-title">
          <span class="warning-icon">🚨</span>
          <h3>{{ t("diag.title") }}</h3>
        </div>
        <button class="close-btn" @click="close">✕</button>
      </div>

      <div class="modal-body">
        <!-- Failure Alert Banner -->
        <div class="alert-banner">
          <div class="banner-title">
            <span class="stage-badge" v-if="diagnostics?.failedStage">
              {{ diagnostics.failedStage }}
            </span>
            <span>{{ t("diag.task_terminated") }}</span>
          </div>
          <div class="banner-desc">{{ diagnostics?.errorCause || errorMsg }}</div>
        </div>

        <!-- Diagnostics Status Grid -->
        <div class="status-grid" v-if="diagnostics">
          <div class="grid-item">
            <span class="item-label">{{ t("diag.task_id") }}</span>
            <span class="item-value code-font">{{ diagnostics.taskId }}</span>
          </div>
          <div class="grid-item">
            <span class="item-label">{{ t("diag.target_device") }}</span>
            <span class="item-value">{{ diagnostics.deviceSummary || diagnostics.target }}</span>
          </div>
          <div class="grid-item">
            <span class="item-label">{{ t("diag.format_status") }}</span>
            <span class="item-value" :class="diagnostics.isFormatted ? 'text-warning' : 'text-muted'">
              {{ diagnostics.isFormatted ? t("diag.formatted") : t("diag.not_formatted") }}
            </span>
          </div>
          <div class="grid-item">
            <span class="item-label">{{ t("diag.safe_unplug_status") }}</span>
            <span class="item-value" :class="diagnostics.safeToUnplug ? 'text-success' : 'text-danger'">
              <span class="val-icon">{{ diagnostics.safeToUnplug ? "✅" : "❌" }}</span>
              <span>{{ diagnostics.safeToUnplug ? t("diag.safe_to_unplug") : t("diag.not_safe_to_unplug") }}</span>
            </span>
          </div>
        </div>

        <!-- Recommended Action Card -->
        <div class="recommend-card" :class="actionClass" v-if="diagnostics">
          <div class="recommend-title">
            <span class="recommend-icon">{{ actionIcon }}</span>
            <span>{{ t("diag.suggested_action") }}: {{ actionTitle }}</span>
          </div>
          <div class="recommend-desc">{{ actionDescription }}</div>
        </div>

        <!-- Written Files Collapsible List -->
        <div class="files-collapsible" v-if="diagnostics?.writtenFiles && diagnostics.writtenFiles.length > 0">
          <div class="files-header" @click="showFiles = !showFiles">
            <span
              ><span class="icon">📁</span>
              {{ t("diag.written_files", { count: diagnostics.writtenFiles.length }) }}</span
            >
            <span class="arrow">{{ showFiles ? "▲" : "▼" }}</span>
          </div>
          <div class="files-body" v-if="showFiles">
            <ul>
              <li v-for="(file, idx) in diagnostics.writtenFiles" :key="idx" class="code-font">
                {{ file }}
              </li>
            </ul>
          </div>
        </div>

        <!-- Action Buttons -->
        <div class="action-buttons-group">
          <button class="btn-primary flex-btn" @click="onRetry">
            <span class="btn-icon">🔄</span>
            <span>{{ t("diag.btn_retry") }}</span>
          </button>
          <button class="btn-secondary flex-btn" @click="onCopyReport">
            <span class="btn-icon">📋</span>
            <span>{{ t("diag.btn_copy_report") }}</span>
          </button>
          <button class="btn-outline flex-btn" @click="close">
            <span>{{ t("common.close") }}</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import { t } from "../i18n";

export interface InstallDiagnosticsData {
  taskId: string;
  target: string;
  deviceSummary: string;
  mode: string;
  failedStage: string;
  failedStepCode: string;
  errorCause: string;
  isFormatted: boolean;
  writtenFiles: string[];
  safeToUnplug: boolean;
  recommendedAction: string;
  reportSummary: string;
}

const props = defineProps<{
  isOpen: boolean;
  diagnostics?: InstallDiagnosticsData | null;
  errorMsg: string;
}>();

const emit = defineEmits(["close", "retry", "copy-report"]);

const showFiles = ref(false);

const actionIcon = computed(() => {
  switch (props.diagnostics?.recommendedAction) {
    case "reformat":
      return "🧹";
    case "remount":
      return "🔌";
    case "retry":
    default:
      return "🔄";
  }
});

const actionTitle = computed(() => {
  switch (props.diagnostics?.recommendedAction) {
    case "reformat":
      return t("diag.action_reformat_title");
    case "remount":
      return t("diag.action_remount_title");
    case "retry":
    default:
      return t("diag.action_retry_title");
  }
});

const actionDescription = computed(() => {
  switch (props.diagnostics?.recommendedAction) {
    case "reformat":
      return t("diag.action_reformat_desc");
    case "remount":
      return t("diag.action_remount_desc");
    case "retry":
    default:
      return t("diag.action_retry_desc");
  }
});

const actionClass = computed(() => {
  switch (props.diagnostics?.recommendedAction) {
    case "reformat":
      return "card-reformat";
    case "remount":
      return "card-remount";
    case "retry":
    default:
      return "card-retry";
  }
});

function close() {
  emit("close");
}

function onRetry() {
  emit("retry");
}

function onCopyReport() {
  emit("copy-report");
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;
  background: var(--modal-backdrop);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  backdrop-filter: blur(10px);
}

.diag-card {
  background: var(--modal-bg);
  border: 1px solid var(--alert-danger-border);
  border-radius: 16px;
  box-shadow:
    0 25px 60px -12px rgba(0, 0, 0, 0.45),
    0 0 0 1px var(--card-border);
  width: 90%;
  max-width: 580px;
  max-height: 85vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

[data-theme="light"] .diag-card {
  box-shadow:
    0 20px 50px -10px rgba(15, 23, 42, 0.22),
    0 0 0 1px #cbd5e1;
}

.danger-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  background: var(--alert-danger-bg);
  border-bottom: 1px solid var(--alert-danger-border);
}

.header-title {
  display: flex;
  align-items: center;
  gap: 10px;
}

.header-title h3 {
  margin: 0;
  font-size: 1.1rem;
  color: var(--alert-danger-title);
  font-weight: 700;
}

.close-btn {
  background: none;
  border: none;
  color: var(--text-muted);
  font-size: 1.2rem;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 6px;
  transition: all 0.2s ease;
}

.close-btn:hover {
  background: rgba(255, 255, 255, 0.1);
  color: var(--text-main);
}

[data-theme="light"] .close-btn:hover {
  background: rgba(15, 23, 42, 0.08);
}

.modal-body {
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 16px;
  overflow-y: auto;
}

.alert-banner {
  background: var(--alert-danger-bg);
  border-left: 4px solid var(--danger);
  border-top: 1px solid var(--alert-danger-border);
  border-right: 1px solid var(--alert-danger-border);
  border-bottom: 1px solid var(--alert-danger-border);
  padding: 12px 16px;
  border-radius: 8px;
}

.banner-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 700;
  color: var(--alert-danger-title);
  margin-bottom: 6px;
  font-size: 0.95rem;
}

.stage-badge {
  background: var(--badge-danger-bg);
  color: var(--badge-danger-text);
  font-size: 0.75rem;
  padding: 2px 8px;
  border-radius: 4px;
  font-family: monospace;
  font-weight: 600;
  border: 1px solid var(--alert-danger-border);
}

.banner-desc {
  font-size: 0.88rem;
  color: var(--alert-danger-text);
  word-break: break-word;
  line-height: 1.5;
  font-weight: 500;
}

.status-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
  background: var(--subtab-container-bg);
  padding: 14px;
  border-radius: 10px;
  border: 1px solid var(--card-border);
}

[data-theme="light"] .status-grid {
  background: #f8fafc;
  border-color: #cbd5e1;
}

.grid-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.item-label {
  font-size: 0.78rem;
  color: var(--text-muted);
  font-weight: 600;
}

[data-theme="light"] .item-label {
  color: #475569;
}

.item-value {
  font-size: 0.92rem;
  color: var(--text-main);
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 6px;
}

[data-theme="light"] .item-value {
  color: #0f172a;
}

.code-font {
  font-family: "JetBrains Mono", Consolas, monospace;
}

.text-warning {
  color: var(--warning);
  font-weight: 600;
}
.text-muted {
  color: var(--text-muted);
}
.text-success {
  color: var(--success);
  font-weight: 600;
}
.text-danger {
  color: var(--danger);
  font-weight: 600;
}

.recommend-card {
  padding: 14px 16px;
  border-radius: 10px;
  border: 1px solid;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.card-retry {
  background: rgba(59, 130, 246, 0.12);
  border-color: rgba(59, 130, 246, 0.35);
}

.card-remount {
  background: rgba(245, 158, 11, 0.12);
  border-color: rgba(245, 158, 11, 0.35);
}

.card-reformat {
  background: rgba(239, 68, 68, 0.12);
  border-color: rgba(239, 68, 68, 0.35);
}

[data-theme="light"] .card-retry {
  background: #eff6ff;
  border-color: #93c5fd;
}

[data-theme="light"] .card-remount {
  background: #fffbeb;
  border-color: #fcd34d;
}

[data-theme="light"] .card-reformat {
  background: #fef2f2;
  border-color: #fca5a5;
}

.recommend-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 700;
  font-size: 0.96rem;
  color: #f8fafc;
}

.recommend-icon {
  font-size: 1.15rem;
}

.recommend-desc {
  font-size: 0.88rem;
  color: #cbd5e1;
  line-height: 1.5;
  font-weight: 500;
}

[data-theme="light"] .card-retry .recommend-title {
  color: #1e40af;
}

[data-theme="light"] .card-remount .recommend-title {
  color: #92400e;
}

[data-theme="light"] .card-reformat .recommend-title {
  color: #991b1b;
}

[data-theme="light"] .recommend-desc {
  color: #334155;
}

.files-collapsible {
  background: var(--subtab-container-bg);
  border-radius: 8px;
  border: 1px solid var(--card-border);
  overflow: hidden;
}

[data-theme="light"] .files-collapsible {
  background: #f1f5f9;
  border-color: #cbd5e1;
}

.files-header {
  padding: 10px 14px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.85rem;
  color: var(--text-main);
  font-weight: 600;
  cursor: pointer;
  user-select: none;
  transition: background 0.2s ease;
}

.files-header:hover {
  background: var(--item-hover-bg);
}

[data-theme="light"] .files-header {
  color: #334155;
}

.files-body {
  padding: 10px 14px;
  border-top: 1px solid var(--card-border);
  max-height: 120px;
  overflow-y: auto;
}

.files-body ul {
  margin: 0;
  padding-left: 18px;
  font-size: 0.8rem;
  color: var(--text-muted);
}

[data-theme="light"] .files-body ul {
  color: #475569;
}

.action-buttons-group {
  display: flex;
  gap: 12px;
  margin-top: 6px;
}

.flex-btn {
  flex: 1;
  padding: 11px 16px;
  border-radius: 10px;
  font-size: 0.92rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.btn-primary {
  background: linear-gradient(135deg, #0284c7 0%, #2563eb 100%);
  color: #ffffff;
  border: none;
  box-shadow: 0 3px 10px rgba(2, 132, 199, 0.35);
}

.btn-primary:hover {
  filter: brightness(1.08);
  box-shadow: 0 5px 16px rgba(2, 132, 199, 0.45);
}

.btn-secondary {
  background: var(--btn-sec-bg);
  color: var(--btn-sec-text);
  border: 1px solid var(--btn-sec-border);
}

.btn-secondary:hover {
  background: var(--btn-sec-hover-bg);
  color: var(--btn-sec-hover-text);
  border-color: var(--btn-sec-hover-border);
}

[data-theme="light"] .btn-secondary {
  background: #ffffff;
  color: #0f172a;
  border: 1px solid #cbd5e1;
  box-shadow: 0 2px 6px rgba(15, 23, 42, 0.06);
}

[data-theme="light"] .btn-secondary:hover {
  background: #f0f9ff;
  border-color: #38bdf8;
  color: #0284c7;
}

.btn-outline {
  background: transparent;
  color: var(--text-muted);
  border: 1px solid var(--btn-sec-border);
}

.btn-outline:hover {
  background: var(--item-hover-bg);
  color: var(--text-main);
  border-color: var(--card-border);
}

[data-theme="light"] .btn-outline {
  background: #f8fafc;
  color: #475569;
  border: 1px solid #cbd5e1;
  box-shadow: 0 1px 3px rgba(15, 23, 42, 0.04);
}

[data-theme="light"] .btn-outline:hover {
  background: #f1f5f9;
  color: #0f172a;
  border-color: #94a3b8;
}
</style>

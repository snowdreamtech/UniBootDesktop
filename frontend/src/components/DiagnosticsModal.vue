<template>
  <div v-if="isOpen" class="modal-overlay" @click.self="close">
    <div class="glass-modal diag-card" @click.stop>
      <div class="modal-header danger-header">
        <div class="header-title">
          <span class="warning-icon">🚨</span>
          <h3>{{ t('diag.title') }}</h3>
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
            <span>{{ t('diag.task_terminated') }}</span>
          </div>
          <div class="banner-desc">{{ diagnostics?.errorCause || errorMsg }}</div>
        </div>

        <!-- Diagnostics Status Grid -->
        <div class="status-grid" v-if="diagnostics">
          <div class="grid-item">
            <span class="item-label">{{ t('diag.task_id') }}</span>
            <span class="item-value code-font">{{ diagnostics.taskId }}</span>
          </div>
          <div class="grid-item">
            <span class="item-label">{{ t('diag.target_device') }}</span>
            <span class="item-value">{{ diagnostics.deviceSummary || diagnostics.target }}</span>
          </div>
          <div class="grid-item">
            <span class="item-label">{{ t('diag.format_status') }}</span>
            <span class="item-value" :class="diagnostics.isFormatted ? 'text-warning' : 'text-muted'">
              {{ diagnostics.isFormatted ? t('diag.formatted') : t('diag.not_formatted') }}
            </span>
          </div>
          <div class="grid-item">
            <span class="item-label">{{ t('diag.safe_unplug_status') }}</span>
            <span class="item-value" :class="diagnostics.safeToUnplug ? 'text-success' : 'text-danger'">
              <span class="val-icon">{{ diagnostics.safeToUnplug ? '✅' : '❌' }}</span>
              <span>{{ diagnostics.safeToUnplug ? t('diag.safe_to_unplug') : t('diag.not_safe_to_unplug') }}</span>
            </span>
          </div>
        </div>

        <!-- Recommended Action Card -->
        <div class="recommend-card" :class="actionClass" v-if="diagnostics">
          <div class="recommend-title">
            <span class="recommend-icon">{{ actionIcon }}</span>
            <span>{{ t('diag.suggested_action') }}: {{ actionTitle }}</span>
          </div>
          <div class="recommend-desc">{{ actionDescription }}</div>
        </div>

        <!-- Written Files Collapsible List -->
        <div class="files-collapsible" v-if="diagnostics?.writtenFiles && diagnostics.writtenFiles.length > 0">
          <div class="files-header" @click="showFiles = !showFiles">
            <span>📁 {{ t('diag.written_files', { count: diagnostics.writtenFiles.length }) }}</span>
            <span class="arrow">{{ showFiles ? '▲' : '▼' }}</span>
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
            🔄 {{ t('diag.btn_retry') }}
          </button>
          <button class="btn-secondary flex-btn" @click="onCopyReport">
            📋 {{ t('diag.btn_copy_report') }}
          </button>
          <button class="btn-outline flex-btn" @click="close">
            {{ t('common.close') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';
import { t } from '../i18n';

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

const emit = defineEmits(['close', 'retry', 'copy-report']);

const showFiles = ref(false);

const actionIcon = computed(() => {
  switch (props.diagnostics?.recommendedAction) {
    case 'reformat': return '🧹';
    case 'remount': return '🔌';
    case 'retry': default: return '🔄';
  }
});

const actionTitle = computed(() => {
  switch (props.diagnostics?.recommendedAction) {
    case 'reformat': return t('diag.action_reformat_title');
    case 'remount': return t('diag.action_remount_title');
    case 'retry': default: return t('diag.action_retry_title');
  }
});

const actionDescription = computed(() => {
  switch (props.diagnostics?.recommendedAction) {
    case 'reformat':
      return t('diag.action_reformat_desc');
    case 'remount':
      return t('diag.action_remount_desc');
    case 'retry':
    default:
      return t('diag.action_retry_desc');
  }
});

const actionClass = computed(() => {
  switch (props.diagnostics?.recommendedAction) {
    case 'reformat': return 'card-reformat';
    case 'remount': return 'card-remount';
    case 'retry': default: return 'card-retry';
  }
});

function close() {
  emit('close');
}

function onRetry() {
  emit('retry');
}

function onCopyReport() {
  emit('copy-report');
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
  backdrop-filter: blur(8px);
}

.diag-card {
  background: var(--modal-bg);
  border: 1px solid var(--alert-danger-border);
  border-radius: 16px;
  box-shadow: 0 20px 50px var(--modal-backdrop);
  width: 90%;
  max-width: 580px;
  max-height: 85vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
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
  font-weight: 600;
}

.close-btn {
  background: none;
  border: none;
  color: #9ca3af;
  font-size: 1.2rem;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 6px;
}

.close-btn:hover {
  background: rgba(255, 255, 255, 0.1);
  color: #fff;
}

.modal-body {
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.alert-banner {
  background: var(--alert-danger-bg);
  border-left: 4px solid var(--danger);
  padding: 12px 16px;
  border-radius: 8px;
}

.banner-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
  color: var(--alert-danger-title);
  margin-bottom: 4px;
}

.stage-badge {
  background: var(--badge-danger-bg);
  color: var(--badge-danger-text);
  font-size: 0.75rem;
  padding: 2px 8px;
  border-radius: 4px;
  font-family: monospace;
}

.banner-desc {
  font-size: 0.88rem;
  color: var(--text-main);
  word-break: break-word;
}

.status-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
  background: var(--subtab-container-bg);
  padding: 12px;
  border-radius: 10px;
  border: 1px solid var(--card-border);
}

.grid-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.item-label {
  font-size: 0.75rem;
  color: var(--text-muted);
}

.item-value {
  font-size: 0.9rem;
  color: var(--text-main);
}

.code-font {
  font-family: 'JetBrains Mono', Consolas, monospace;
}

.text-warning { color: var(--warning); }
.text-muted { color: var(--text-muted); }
.text-success { color: var(--success); }
.text-danger { color: var(--danger); }

.recommend-card {
  padding: 14px 16px;
  border-radius: 10px;
  border: 1px solid;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.card-retry {
  background: rgba(59, 130, 246, 0.1);
  border-color: rgba(59, 130, 246, 0.3);
}

.card-remount {
  background: rgba(245, 158, 11, 0.1);
  border-color: rgba(245, 158, 11, 0.3);
}

.card-reformat {
  background: rgba(239, 68, 68, 0.1);
  border-color: rgba(239, 68, 68, 0.3);
}

.recommend-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
  font-size: 0.95rem;
  color: #f3f4f6;
}

.recommend-desc {
  font-size: 0.85rem;
  color: #9ca3af;
  line-height: 1.4;
}

.files-collapsible {
  background: rgba(0, 0, 0, 0.2);
  border-radius: 8px;
  border: 1px solid rgba(255, 255, 255, 0.05);
  overflow: hidden;
}

.files-header {
  padding: 10px 14px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.85rem;
  color: #9ca3af;
  cursor: pointer;
  user-select: none;
}

.files-header:hover {
  background: rgba(255, 255, 255, 0.03);
}

.files-body {
  padding: 10px 14px;
  border-top: 1px solid rgba(255, 255, 255, 0.05);
  max-height: 120px;
  overflow-y: auto;
}

.files-body ul {
  margin: 0;
  padding-left: 18px;
  font-size: 0.8rem;
  color: #6b7280;
}

.action-buttons-group {
  display: flex;
  gap: 10px;
  margin-top: 8px;
}

.flex-btn {
  flex: 1;
  padding: 10px 14px;
  border-radius: 8px;
  font-size: 0.88rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  border: none;
}

.btn-primary {
  background: linear-gradient(135deg, #3b82f6, #2563eb);
  color: #fff;
}

.btn-primary:hover {
  filter: brightness(1.1);
  box-shadow: 0 0 12px rgba(59, 130, 246, 0.4);
}

.btn-secondary {
  background: rgba(255, 255, 255, 0.08);
  color: #e5e7eb;
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.btn-secondary:hover {
  background: rgba(255, 255, 255, 0.15);
}

.btn-outline {
  background: transparent;
  color: #9ca3af;
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.btn-outline:hover {
  background: rgba(255, 255, 255, 0.05);
  color: #fff;
}
</style>

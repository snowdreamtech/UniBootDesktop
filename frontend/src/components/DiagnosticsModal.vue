<template>
  <div v-if="isOpen" class="modal-overlay" @click.self="close">
    <div class="glass-modal diag-card" @click.stop>
      <div class="modal-header danger-header">
        <div class="header-title">
          <span class="warning-icon">🚨</span>
          <h3>写盘失败诊断与恢复方案</h3>
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
            <span>写盘任务异常终止</span>
          </div>
          <div class="banner-desc">{{ diagnostics?.errorCause || errorMsg }}</div>
        </div>

        <!-- Diagnostics Status Grid -->
        <div class="status-grid" v-if="diagnostics">
          <div class="grid-item">
            <span class="item-label">任务 ID</span>
            <span class="item-value code-font">{{ diagnostics.taskId }}</span>
          </div>
          <div class="grid-item">
            <span class="item-label">目标设备</span>
            <span class="item-value">{{ diagnostics.deviceSummary || diagnostics.target }}</span>
          </div>
          <div class="grid-item">
            <span class="item-label">格式化状态</span>
            <span class="item-value" :class="diagnostics.isFormatted ? 'text-warning' : 'text-muted'">
              {{ diagnostics.isFormatted ? '已格式化' : '未格式化' }}
            </span>
          </div>
          <div class="grid-item">
            <span class="item-label">拔盘安全状态</span>
            <span class="item-value" :class="diagnostics.safeToUnplug ? 'text-success' : 'text-danger'">
              {{ diagnostics.safeToUnplug ? '✅ 可安全拔盘' : '❌ 暂不可拔盘' }}
            </span>
          </div>
        </div>

        <!-- Recommended Action Card -->
        <div class="recommend-card" :class="actionClass" v-if="diagnostics">
          <div class="recommend-title">
            <span class="recommend-icon">{{ actionIcon }}</span>
            <span>建议下一步操作：{{ actionTitle }}</span>
          </div>
          <div class="recommend-desc">{{ actionDescription }}</div>
        </div>

        <!-- Written Files Collapsible List -->
        <div class="files-collapsible" v-if="diagnostics?.writtenFiles && diagnostics.writtenFiles.length > 0">
          <div class="files-header" @click="showFiles = !showFiles">
            <span>📁 已写入文件 ({{ diagnostics.writtenFiles.length }} 个文件)</span>
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
            🔄 重新尝试写盘
          </button>
          <button class="btn-secondary flex-btn" @click="onCopyReport">
            📋 复制完整诊断报告
          </button>
          <button class="btn-outline flex-btn" @click="close">
            关闭
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';

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
    case 'reformat': return '重新格式化 (reformat)';
    case 'remount': return '重新挂载 (remount)';
    case 'retry': default: return '重试 (retry)';
  }
});

const actionDescription = computed(() => {
  switch (props.diagnostics?.recommendedAction) {
    case 'reformat':
      return '写盘中断导致磁盘分区结构或文件系统不完整，建议重新初始化格式化后写入。';
    case 'remount':
      return '设备挂载路径在写盘过程中掉盘或掉挂，请重新插拔 U 盘或重新挂载卷。';
    case 'retry':
    default:
      return '写盘环境与设备状态完好，可直接选择重试继续进行部署。';
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
  const reportText = props.diagnostics?.reportSummary || props.errorMsg;
  navigator.clipboard.writeText(reportText);
  emit('copy-report', reportText);
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
  background: var(--modal-bg, #1e293b);
  border: 1px solid var(--card-border, #334155);
  border-radius: 16px;
  width: 92%;
  max-width: 580px;
  max-height: 90vh;
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.4);
  overflow-y: auto;
  color: var(--text-main, #f8fafc);
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  border-bottom: 1px solid var(--card-border, #334155);
  background: var(--modal-header-bg, #0f172a);
}

.danger-header {
  border-left: 4px solid #ef4444;
}

.header-title {
  display: flex;
  align-items: center;
  gap: 10px;
}

.header-title h3 {
  margin: 0;
  font-size: 1.15rem;
  font-weight: 700;
  color: #f8fafc;
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
  background: #334155;
  color: #f8fafc;
}

.modal-body {
  padding: 20px;
}

.alert-banner {
  background: rgba(239, 68, 68, 0.1);
  border: 1px solid rgba(239, 68, 68, 0.3);
  border-radius: 10px;
  padding: 14px 16px;
  margin-bottom: 16px;
}

.banner-title {
  font-weight: 700;
  color: #fca5a5;
  margin-bottom: 6px;
  font-size: 1rem;
  display: flex;
  align-items: center;
  gap: 8px;
}

.stage-badge {
  background: #ef4444;
  color: #ffffff;
  font-size: 0.75rem;
  padding: 2px 8px;
  border-radius: 12px;
  font-weight: 600;
}

.banner-desc {
  color: #f8fafc;
  font-size: 0.9rem;
  line-height: 1.5;
  word-break: break-word;
}

.status-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 10px;
  margin-bottom: 16px;
  background: rgba(15, 23, 42, 0.5);
  padding: 12px;
  border-radius: 10px;
  border: 1px solid var(--card-border, #334155);
}

.grid-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.item-label {
  font-size: 0.75rem;
  color: #94a3b8;
}

.item-value {
  font-size: 0.85rem;
  font-weight: 600;
  color: #f1f5f9;
}

.code-font {
  font-family: monospace;
}

.text-warning { color: #f59e0b; }
.text-muted { color: #94a3b8; }
.text-success { color: #10b981; }
.text-danger { color: #ef4444; }

.recommend-card {
  border-radius: 10px;
  padding: 14px 16px;
  margin-bottom: 16px;
  border: 1px solid;
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
  background: rgba(168, 85, 247, 0.1);
  border-color: rgba(168, 85, 247, 0.3);
}

.recommend-title {
  font-weight: 700;
  font-size: 0.95rem;
  margin-bottom: 4px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.recommend-desc {
  font-size: 0.85rem;
  color: #cbd5e1;
  line-height: 1.4;
}

.files-collapsible {
  margin-bottom: 16px;
  border: 1px solid var(--card-border, #334155);
  border-radius: 8px;
  overflow: hidden;
}

.files-header {
  background: rgba(15, 23, 42, 0.6);
  padding: 10px 14px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  cursor: pointer;
  font-size: 0.85rem;
  font-weight: 600;
  color: #cbd5e1;
}

.files-body {
  padding: 10px 14px;
  max-height: 120px;
  overflow-y: auto;
  background: rgba(0, 0, 0, 0.2);
}

.files-body ul {
  margin: 0;
  padding-left: 18px;
  font-size: 0.8rem;
  color: #94a3b8;
}

.files-body li {
  margin-bottom: 3px;
  word-break: break-all;
}

.action-buttons-group {
  display: flex;
  gap: 10px;
  margin-top: 16px;
}

.flex-btn {
  flex: 1;
  padding: 11px 14px;
  border-radius: 10px;
  font-weight: 600;
  font-size: 0.9rem;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
}

.btn-primary {
  background: #2563eb;
  color: #ffffff;
  border: none;
}

.btn-primary:hover {
  background: #1d4ed8;
  box-shadow: 0 4px 12px rgba(37, 99, 235, 0.3);
}

.btn-secondary {
  background: #334155;
  color: #ffffff;
  border: none;
}

.btn-secondary:hover {
  background: #475569;
}

.btn-outline {
  background: transparent;
  color: #94a3b8;
  border: 1px solid #334155;
}

.btn-outline:hover {
  background: #1e293b;
  color: #ffffff;
}
</style>

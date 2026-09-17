<template>
  <div v-if="isOpen" class="modal-overlay" @click.self="$emit('close')">
    <div class="modal-content glass-card log-modal">
      <div class="modal-header">
        <div class="title-with-badge">
          <h3>📋 {{ t('log.title') }}</h3>
          <span class="badge live-badge">● {{ t('log.live') }}</span>
        </div>
        <button class="btn-close" @click="$emit('close')">✕</button>
      </div>

      <div class="modal-body">
        <!-- Log Filter & Search Bar -->
        <div class="log-controls">
          <div class="filter-tabs">
            <button 
              v-for="level in logLevels" 
              :key="level.key"
              class="btn-tab"
              :class="{ active: currentFilter === level.key, [level.key.toLowerCase()]: true }"
              @click="currentFilter = level.key"
            >
              {{ level.label }} ({{ getLevelCount(level.key) }})
            </button>
          </div>

          <div class="search-box">
            <input 
              type="text" 
              v-model="searchQuery" 
              :placeholder="t('log.search_placeholder')"
              class="search-input"
            />
          </div>
        </div>

        <!-- Terminal Log Window -->
        <div class="terminal-window" ref="terminalRef">
          <div v-if="filteredLogs.length === 0" class="empty-logs">
            {{ t('log.empty') }}
          </div>
          <div 
            v-for="log in filteredLogs" 
            :key="log.id || String(log.timestamp)"
            class="log-row"
            :class="log.level.toLowerCase()"
          >
            <span class="log-time">{{ formatLogTime(log.timestamp) }}</span>
            <span class="log-level-badge" :class="log.level.toLowerCase()">[{{ log.level }}]</span>
            <span class="log-msg">{{ log.message }}</span>
            <span v-if="log.details" class="log-details">{{ log.details }}</span>
          </div>
        </div>
      </div>

      <div class="modal-footer">
        <div class="footer-left">
          <label class="auto-scroll-label">
            <input type="checkbox" v-model="autoScroll" />
            {{ t('log.auto_scroll') }}
          </label>
        </div>
        <div class="footer-actions">
          <button class="btn btn-secondary" @click="copyAllLogs">
            📋 {{ t('log.copy') }}
          </button>
          <button class="btn btn-secondary" @click="exportLogFile">
            📥 {{ t('log.export') }}
          </button>
          <button class="btn btn-danger" @click="$emit('clear')">
            🗑️ {{ t('log.clear') }}
          </button>
          <button class="btn btn-primary" @click="$emit('close')">
            {{ t('common.close') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue';
import { t } from '../i18n';

export interface LogItem {
  id?: number;
  timestamp: string | Date;
  level: string;
  message: string;
  details?: string;
}

const props = defineProps<{
  isOpen: boolean;
  logs: LogItem[];
}>();

defineEmits(['close', 'clear']);

const currentFilter = ref('ALL');
const searchQuery = ref('');
const autoScroll = ref(true);
const terminalRef = ref<HTMLDivElement | null>(null);

const logLevels = computed(() => [
  { key: 'ALL', label: t('log.level_all') },
  { key: 'INFO', label: t('log.level_info') },
  { key: 'WARN', label: t('log.level_warn') },
  { key: 'ERROR', label: t('log.level_error') },
  { key: 'DEBUG', label: t('log.level_debug') }
]);

function getLevelCount(level: string): number {
  if (level === 'ALL') return props.logs.length;
  return props.logs.filter(l => (l.level || '').toUpperCase() === level).length;
}

const filteredLogs = computed(() => {
  return props.logs.filter(log => {
    const matchesLevel = currentFilter.value === 'ALL' || (log.level || '').toUpperCase() === currentFilter.value;
    const query = searchQuery.value.trim().toLowerCase();
    const matchesQuery = !query || 
      log.message.toLowerCase().includes(query) || 
      (log.details && log.details.toLowerCase().includes(query));
    return matchesLevel && matchesQuery;
  });
});

function formatLogTime(ts: string | Date): string {
  const date = new Date(ts);
  const hours = date.getHours().toString().padStart(2, '0');
  const minutes = date.getMinutes().toString().padStart(2, '0');
  const seconds = date.getSeconds().toString().padStart(2, '0');
  const ms = date.getMilliseconds().toString().padStart(3, '0');
  return `${hours}:${minutes}:${seconds}.${ms}`;
}

function scrollToBottom() {
  if (autoScroll.value && terminalRef.value) {
    nextTick(() => {
      if (terminalRef.value) {
        terminalRef.value.scrollTop = terminalRef.value.scrollHeight;
      }
    });
  }
}

watch(() => props.logs.length, () => {
  scrollToBottom();
});

function copyAllLogs() {
  const text = filteredLogs.value.map(l => `[${formatLogTime(l.timestamp)}] [${l.level}] ${l.message} ${l.details || ''}`).join('\n');
  navigator.clipboard.writeText(text);
  alert(t('log.copied_toast'));
}

function exportLogFile() {
  const text = filteredLogs.value.map(l => `[${formatLogTime(l.timestamp)}] [${l.level}] ${l.message} ${l.details || ''}`).join('\n');
  const blob = new Blob([text], { type: 'text/plain;charset=utf-8' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = `uniboot_log_${new Date().toISOString().slice(0, 10)}.log`;
  a.click();
  URL.revokeObjectURL(url);
}
</script>

<style scoped>
.log-modal {
  max-width: 900px;
  width: 90vw;
  max-height: 85vh;
  display: flex;
  flex-direction: column;
}

.title-with-badge {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.live-badge {
  background: rgba(16, 185, 129, 0.15);
  color: #10b981;
  font-size: 0.75rem;
  padding: 0.2rem 0.6rem;
  border-radius: 12px;
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0% { opacity: 1; }
  50% { opacity: 0.4; }
  100% { opacity: 1; }
}

.log-controls {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.85rem;
  gap: 1rem;
}

.filter-tabs {
  display: flex;
  gap: 0.4rem;
}

.btn-tab {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  color: var(--text-secondary, #9ca3af);
  padding: 0.3rem 0.7rem;
  border-radius: 6px;
  font-size: 0.8rem;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-tab.active {
  background: var(--accent-cyan, #00e5ff);
  color: #000;
  border-color: var(--accent-cyan, #00e5ff);
  font-weight: 600;
}

.search-input {
  background: rgba(0, 0, 0, 0.2);
  border: 1px solid rgba(255, 255, 255, 0.15);
  color: #fff;
  padding: 0.35rem 0.75rem;
  border-radius: 6px;
  font-size: 0.85rem;
  width: 220px;
}

.terminal-window {
  background: #0f172a;
  border-radius: 8px;
  padding: 0.85rem;
  font-family: 'Fira Code', 'Courier New', monospace;
  font-size: 0.82rem;
  line-height: 1.5;
  height: 420px;
  overflow-y: auto;
  border: 1px solid rgba(255, 255, 255, 0.08);
}

.empty-logs {
  color: #64748b;
  text-align: center;
  padding: 4rem 0;
}

.log-row {
  display: flex;
  align-items: flex-start;
  gap: 0.6rem;
  padding: 0.2rem 0;
  border-bottom: 1px solid rgba(255, 255, 255, 0.03);
}

.log-time {
  color: #64748b;
  font-size: 0.78rem;
  flex-shrink: 0;
}

.log-level-badge {
  font-weight: 600;
  font-size: 0.75rem;
  padding: 0 0.3rem;
  border-radius: 3px;
  flex-shrink: 0;
}

.log-level-badge.info { color: #38bdf8; }
.log-level-badge.warn { color: #fbbf24; }
.log-level-badge.error { color: #f87171; }
.log-level-badge.debug { color: #a78bfa; }

.log-msg {
  color: #e2e8f0;
  word-break: break-word;
}

.log-details {
  color: #94a3b8;
  font-size: 0.78rem;
  opacity: 0.8;
}

.footer-left {
  display: flex;
  align-items: center;
}

.auto-scroll-label {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  font-size: 0.85rem;
  cursor: pointer;
}

.footer-actions {
  display: flex;
  gap: 0.5rem;
}

/* Light Mode Overrides for LogViewerModal */
[data-theme="light"] .log-modal {
  background: #ffffff;
  border-color: #cbd5e1;
  box-shadow: 0 16px 48px rgba(15, 23, 42, 0.2);
}

[data-theme="light"] .btn-tab {
  background: #f1f5f9;
  border-color: #cbd5e1;
  color: #475569;
  font-weight: 600;
}

[data-theme="light"] .btn-tab:hover {
  background: #e2e8f0;
  color: #0f172a;
}

[data-theme="light"] .btn-tab.active {
  background: #0284c7;
  color: #ffffff;
  border-color: #0284c7;
}

[data-theme="light"] .search-input {
  background: #ffffff;
  border-color: #cbd5e1;
  color: #0f172a;
}

[data-theme="light"] .auto-scroll-label {
  color: #334155;
  font-weight: 500;
}

/* Light Mode Terminal Window & Log Row Colors for Modal */
[data-theme="light"] .terminal-window {
  background: #f8fafc;
  border-color: #cbd5e1;
  box-shadow: inset 0 2px 4px rgba(15, 23, 42, 0.04);
}

[data-theme="light"] .terminal-window .empty-logs {
  color: #94a3b8;
}

[data-theme="light"] .terminal-window .log-row {
  border-bottom-color: #e2e8f0;
}

[data-theme="light"] .terminal-window .log-time {
  color: #64748b;
}

[data-theme="light"] .terminal-window .log-msg {
  color: #0f172a;
}

[data-theme="light"] .terminal-window .log-row.info .log-msg {
  color: #0f172a;
}

[data-theme="light"] .terminal-window .log-row.warn .log-msg {
  color: #b45309;
}

[data-theme="light"] .terminal-window .log-row.error .log-msg {
  color: #dc2626;
}

[data-theme="light"] .terminal-window .log-row.debug .log-msg {
  color: #7e22ce;
}

[data-theme="light"] .terminal-window .log-level-badge.info {
  background: #e0f2fe;
  color: #0284c7;
  border: 1px solid #bae6fd;
}

[data-theme="light"] .terminal-window .log-level-badge.warn {
  background: #fef3c7;
  color: #d97706;
  border: 1px solid #fde68a;
}

[data-theme="light"] .terminal-window .log-level-badge.error {
  background: #fee2e2;
  color: #dc2626;
  border: 1px solid #fca5a5;
}

[data-theme="light"] .terminal-window .log-level-badge.debug {
  background: #f3e8ff;
  color: #7e22ce;
  border: 1px solid #e9d5ff;
}
</style>

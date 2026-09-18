<template>
  <transition name="card-fade">
    <section v-if="isVisible" class="glass-card log-section-card">
      <!-- Row 1: Title & Top Control Buttons -->
      <div class="log-section-header">
        <div class="log-title-group">
          <h2>📜 {{ t('log.title') }}</h2>
          <span class="badge live-badge">● {{ t('log.live') }}</span>
        </div>

        <div class="log-section-controls">
          <label class="auto-scroll-label-sm">
            <input
              type="checkbox"
              :checked="autoScroll"
              @change="e => emit('update:autoScroll', (e.target as HTMLInputElement).checked)"
            />
            {{ t('log.auto_scroll') }}
          </label>

          <button class="btn-text-sm" @click="emit('copy-logs')">📋 {{ t('log.copy') }}</button>
          <button class="btn-text-sm" @click="emit('export-logs')">📥 {{ t('log.export') }}</button>
          <button class="btn-text-danger-sm" @click="emit('clear-logs')">🗑️ {{ t('log.clear') }}</button>
          <button
            class="btn-text-sm btn-close-log"
            :title="t('common.close')"
            @click="emit('close-log')"
          >
            ✕
          </button>
        </div>
      </div>

      <!-- Row 2: Filter Tabs -->
      <div class="log-sub-header">
        <div class="filter-tabs-sm">
          <button
            v-for="level in logLevels"
            :key="level.key"
            class="btn-tab-sm"
            :class="{ active: currentLogFilter === level.key }"
            @click="emit('update:currentLogFilter', level.key)"
          >
            {{ level.label }}
          </button>
        </div>
      </div>

      <div class="embedded-terminal-window" ref="embeddedTerminalRef">
        <div v-if="filteredLogs.length === 0" class="empty-logs">
          {{ t('log.empty') }}
        </div>
        <div
          v-for="log in filteredLogs"
          :key="log.id || String(log.timestamp)"
          class="log-row"
          :class="log.level.toLowerCase()"
        >
          <span class="log-time"><bdi>{{ formatLogTime(log.timestamp) }}</bdi></span>
          <span class="log-level-badge" :class="log.level.toLowerCase()"><bdi>[{{ log.level }}]</bdi></span>
          <div class="log-content">
            <span class="log-msg"><bdi>{{ log.message }}</bdi></span>
            <span v-if="log.details" class="log-details"><bdi>{{ log.details }}</bdi></span>
          </div>
        </div>
      </div>
    </section>
  </transition>
</template>

<script setup lang="ts">
import { ref, watch, nextTick, onMounted } from 'vue';
import type { LogItem } from './LogViewerModal.vue';
import { t } from '../i18n';
import { formatLogTime } from '../utils/logFormatter';

const props = defineProps<{
  isVisible: boolean;
  autoScroll: boolean;
  currentLogFilter: string;
  filteredLogs: LogItem[];
  logLevels: { key: string; label: string }[];
}>();

const emit = defineEmits<{
  (e: 'update:autoScroll', scroll: boolean): void;
  (e: 'update:currentLogFilter', filter: string): void;
  (e: 'copy-logs'): void;
  (e: 'export-logs'): void;
  (e: 'clear-logs'): void;
  (e: 'close-log'): void;
}>();

const embeddedTerminalRef = ref<HTMLElement | null>(null);

function scrollToBottom() {
  if (props.autoScroll && embeddedTerminalRef.value) {
    nextTick(() => {
      if (embeddedTerminalRef.value) {
        embeddedTerminalRef.value.scrollTop = embeddedTerminalRef.value.scrollHeight;
      }
    });
  }
}

watch(() => props.filteredLogs.length, () => {
  scrollToBottom();
});

watch(() => props.filteredLogs, () => {
  scrollToBottom();
}, { deep: true });

watch(() => props.autoScroll, (val) => {
  if (val) {
    scrollToBottom();
  }
});

watch(() => props.isVisible, (val) => {
  if (val) {
    scrollToBottom();
  }
});

watch(() => props.currentLogFilter, () => {
  scrollToBottom();
});

onMounted(() => {
  scrollToBottom();
});
</script>

<style scoped>
.log-section-card {
  grid-column: 1 / -1;
  display: flex;
  flex-direction: column;
  padding: 1.25rem 1.5rem;
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  border-radius: 16px;
  backdrop-filter: blur(16px);
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.25);
  min-height: 500px;
}

.log-section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.75rem;
  flex-wrap: wrap;
  gap: 0.75rem;
}

.log-sub-header {
  display: flex;
  align-items: center;
  margin-bottom: 0.75rem;
}

.log-title-group {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.log-title-group h2 {
  font-size: 1.15rem;
  font-weight: 700;
  margin: 0;
  color: var(--text-main);
}

.live-badge {
  background: rgba(16, 185, 129, 0.15);
  color: #10b981;
  border: 1px solid rgba(16, 185, 129, 0.3);
  font-size: 0.75rem;
  padding: 0.2rem 0.6rem;
  border-radius: 12px;
  font-weight: 600;
  letter-spacing: 0.5px;
}

.log-section-controls {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  flex-wrap: wrap;
}

.filter-tabs-sm {
  display: flex;
  gap: 0.3rem;
  background: var(--subtab-container-bg);
  padding: 3px;
  border-radius: 8px;
  border: 1px solid var(--card-border);
}

.btn-tab-sm {
  background: transparent;
  border: none;
  color: var(--subtab-btn-text);
  padding: 0.25rem 0.55rem;
  font-size: 0.75rem;
  font-weight: 600;
  border-radius: 5px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-tab-sm:hover {
  color: var(--text-main);
  background: var(--btn-sec-hover-bg);
}

.btn-tab-sm.active {
  background: var(--tab-btn-active-bg);
  color: var(--tab-btn-active-text);
}

.btn-text-sm {
  background: var(--btn-sec-bg);
  border: 1px solid var(--btn-sec-border);
  color: var(--btn-sec-text);
  font-size: 0.78rem;
  font-weight: 600;
  padding: 0.35rem 0.7rem;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s ease;
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
}

.btn-text-sm:hover {
  background: var(--btn-sec-hover-bg);
  border-color: var(--btn-sec-hover-border);
  color: var(--btn-sec-hover-text);
}

.btn-text-danger-sm {
  background: rgba(239, 68, 68, 0.15);
  border: 1px solid rgba(239, 68, 68, 0.35);
  color: #f87171;
  font-size: 0.78rem;
  font-weight: 600;
  padding: 0.35rem 0.7rem;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s ease;
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
}

.btn-text-danger-sm:hover {
  background: rgba(239, 68, 68, 0.25);
  border-color: rgba(239, 68, 68, 0.5);
  color: #ef4444;
}

.auto-scroll-label-sm {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  font-size: 0.78rem;
  font-weight: 600;
  color: var(--btn-sec-text);
  cursor: pointer;
  user-select: none;
  padding: 0.35rem 0.65rem;
  border-radius: 6px;
  background: var(--btn-sec-bg);
  border: 1px solid var(--btn-sec-border);
  transition: all 0.2s ease;
}

.auto-scroll-label-sm input[type="checkbox"] {
  accent-color: var(--accent-cyan);
  width: 14px;
  height: 14px;
  cursor: pointer;
}

.auto-scroll-label-sm:hover {
  background: var(--btn-sec-hover-bg);
  border-color: var(--btn-sec-hover-border);
  color: var(--btn-sec-hover-text);
}

.embedded-terminal-window {
  background: var(--terminal-bg);
  border: 1px solid var(--terminal-border);
  border-radius: 10px;
  padding: 1rem 1.15rem;
  min-height: 440px;
  max-height: 560px;
  overflow-y: auto;
  overflow-x: hidden;
  font-family: 'JetBrains Mono', 'Fira Code', 'Cascadia Code', Consolas, monospace;
  font-size: 0.82rem;
  line-height: 1.6;
  flex: 1;
}

.embedded-terminal-window .empty-logs {
  color: var(--text-muted);
  text-align: center;
  padding-top: 60px;
}

.log-row {
  display: flex;
  gap: 0.75rem;
  padding: 0.2rem 0;
  border-bottom: 1px solid var(--card-border);
}

.log-time {
  color: var(--text-muted);
  flex-shrink: 0;
}

.log-level-badge {
  font-weight: 700;
  font-size: 0.75rem;
  padding: 0.05rem 0.35rem;
  border-radius: 4px;
  flex-shrink: 0;
}

.log-level-badge.info { color: #38bdf8; }
.log-level-badge.warn, .log-level-badge.warning { color: #fbbf24; }
.log-level-badge.error { color: #f87171; }
.log-level-badge.debug { color: #c084fc; }

.log-content {
  flex: 1;
  word-break: break-word;
}

.log-msg {
  color: var(--text-main);
}

.log-details {
  display: inline-block;
  font-size: 0.78rem;
  color: var(--text-muted);
  margin-left: 0.5rem;
  opacity: 0.85;
}

/* Light Mode Overrides for LogPanel */
[data-theme="light"] .log-section-card {
  background: #ffffff;
  border-color: #cbd5e1;
  box-shadow: 0 4px 20px rgba(15, 23, 42, 0.06);
}

[data-theme="light"] .log-title-group h2 {
  color: #0f172a;
}

[data-theme="light"] .filter-tabs-sm {
  background: #f1f5f9;
  border-color: #cbd5e1;
}

[data-theme="light"] .btn-tab-sm {
  color: #475569;
  font-weight: 600;
}

[data-theme="light"] .btn-tab-sm:hover {
  color: #0f172a;
  background: #e2e8f0;
}

[data-theme="light"] .btn-tab-sm.active {
  background: linear-gradient(135deg, #0396e6 0%, #0284c7 45%, #2563eb 100%);
  color: #ffffff;
  font-weight: 700;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.4), 0 2px 8px rgba(2, 132, 199, 0.3);
}

[data-theme="light"] .btn-text-sm {
  background: linear-gradient(180deg, #ffffff 0%, #f8fafc 60%, #f1f5f9 100%);
  border: 1px solid rgba(15, 23, 42, 0.12);
  color: #0f172a;
  font-weight: 600;
  box-shadow: inset 0 1px 0 #ffffff, 0 1px 3px rgba(15, 23, 42, 0.05);
}

[data-theme="light"] .btn-text-sm:hover {
  background: linear-gradient(180deg, #f0f9ff 0%, #e0f2fe 100%);
  border-color: rgba(2, 132, 199, 0.35);
  color: #0284c7;
  box-shadow: inset 0 1px 0 #ffffff, 0 3px 8px rgba(2, 132, 199, 0.15);
}

[data-theme="light"] .auto-scroll-label-sm {
  background: linear-gradient(180deg, #ffffff 0%, #f8fafc 60%, #f1f5f9 100%);
  border: 1px solid rgba(15, 23, 42, 0.12);
  color: #0f172a;
  font-weight: 600;
  box-shadow: inset 0 1px 0 #ffffff, 0 1px 3px rgba(15, 23, 42, 0.05);
}

[data-theme="light"] .auto-scroll-label-sm input[type="checkbox"] {
  accent-color: #0284c7;
}

[data-theme="light"] .auto-scroll-label-sm:hover {
  background: linear-gradient(180deg, #f0f9ff 0%, #e0f2fe 100%);
  border-color: rgba(2, 132, 199, 0.35);
  color: #0284c7;
  box-shadow: inset 0 1px 0 #ffffff, 0 3px 8px rgba(2, 132, 199, 0.15);
}

[data-theme="light"] .btn-text-danger-sm {
  background: linear-gradient(180deg, #ffffff 0%, #fff5f5 60%, #fef2f2 100%);
  border: 1px solid rgba(239, 68, 68, 0.25);
  color: #dc2626;
  font-weight: 600;
  box-shadow: inset 0 1px 0 #ffffff, 0 1px 3px rgba(239, 68, 68, 0.08);
}

[data-theme="light"] .btn-text-danger-sm:hover {
  background: linear-gradient(180deg, #fef2f2 0%, #fee2e2 100%);
  border-color: rgba(239, 68, 68, 0.45);
  color: #b91c1c;
  box-shadow: inset 0 1px 0 #ffffff, 0 3px 8px rgba(239, 68, 68, 0.15);
}

[data-theme="light"] .embedded-terminal-window {
  background: #f8fafc;
  border-color: #cbd5e1;
  box-shadow: inset 0 2px 4px rgba(15, 23, 42, 0.04);
}

[data-theme="light"] .embedded-terminal-window .empty-logs {
  color: #94a3b8;
}

[data-theme="light"] .embedded-terminal-window .log-row {
  border-bottom-color: #e2e8f0;
}

[data-theme="light"] .embedded-terminal-window .log-time {
  color: #64748b;
}

[data-theme="light"] .embedded-terminal-window .log-msg {
  color: #0f172a;
}

[data-theme="light"] .embedded-terminal-window .log-details {
  color: #64748b;
}

[data-theme="light"] .embedded-terminal-window .log-row.info .log-msg {
  color: #0f172a;
}

[data-theme="light"] .embedded-terminal-window .log-row.warn .log-msg {
  color: #b45309;
}

[data-theme="light"] .embedded-terminal-window .log-row.error .log-msg {
  color: #dc2626;
}

[data-theme="light"] .embedded-terminal-window .log-row.debug .log-msg {
  color: #7e22ce;
}

[data-theme="light"] .embedded-terminal-window .log-level-badge.info {
  background: #ecfdf5;
  color: #047857;
  border: 1px solid #a7f3d0;
}

[data-theme="light"] .embedded-terminal-window .log-level-badge.warn {
  background: #fffbeb;
  color: #b45309;
  border: 1px solid #fde68a;
}

[data-theme="light"] .embedded-terminal-window .log-level-badge.error {
  background: #fef2f2;
  color: #b91c1c;
  border: 1px solid #fca5a5;
}

[data-theme="light"] .embedded-terminal-window .log-level-badge.debug {
  background: #f3e8ff;
  color: #6b21a8;
  border: 1px solid #e9d5ff;
}
</style>

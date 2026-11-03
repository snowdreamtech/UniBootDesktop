<template>
  <transition name="card-fade">
    <section v-if="isVisible" class="glass-card log-section-card">
      <!-- Row 1: Title & Top Control Buttons -->
      <div class="log-section-header">
        <div class="log-title-group">
          <h2>📜 {{ t("log.title") }}</h2>
          <span class="badge live-badge">● {{ t("log.live") }}</span>
        </div>

        <div class="log-section-controls">
          <label class="auto-scroll-label-sm">
            <input
              type="checkbox"
              :checked="autoScroll"
              @change="(e) => emit('update:autoScroll', (e.target as HTMLInputElement).checked)"
            />
            {{ t("log.auto_scroll") }}
          </label>

          <button class="btn-text-sm" @click="emit('copy-logs')">📋 {{ t("log.copy") }}</button>
          <button class="btn-text-sm" @click="emit('export-logs')">📥 {{ t("log.export") }}</button>
          <button class="btn-text-danger-sm" @click="emit('clear-logs')">🗑️ {{ t("log.clear") }}</button>
          <button class="btn-text-sm btn-close-log" :title="t('common.close')" @click="emit('close-log')">✕</button>
        </div>
      </div>

      <!-- Row 2: Filter Tabs & Search Box -->
      <div class="log-sub-header">
        <div class="filter-tabs-sm">
          <button
            v-for="level in logLevels"
            :key="level.key"
            class="btn-tab-sm"
            :class="{ active: currentLogFilter === level.key, [level.key.toLowerCase()]: true }"
            @click="emit('update:currentLogFilter', level.key)"
          >
            {{ level.label }} ({{ getLevelCount(level.key) }})
          </button>
        </div>

        <div class="search-box-sm">
          <input v-model="searchQuery" type="text" :placeholder="t('log.search_placeholder')" class="search-input-sm" />
        </div>
      </div>

      <!-- Terminal Log Window -->
      <div class="embedded-terminal-window" ref="embeddedTerminalRef">
        <div v-if="displayedLogs.length === 0" class="empty-logs">
          {{ t("log.empty") }}
        </div>
        <div
          v-for="log in displayedLogs"
          :key="log.id || String(log.timestamp)"
          class="log-row"
          :class="(log.level || 'info').toLowerCase()"
        >
          <span class="log-time"
            ><bdi>{{ formatLogTime(log.timestamp) }}</bdi></span
          >
          <span class="log-level-badge" :class="(log.level || 'info').toLowerCase()"
            ><bdi>[{{ log.level }}]</bdi></span
          >
          <div class="log-content">
            <span class="log-msg"
              ><bdi>{{ log.message }}</bdi></span
            >
            <span v-if="log.details" class="log-details"
              ><bdi>{{ log.details }}</bdi></span
            >
          </div>
        </div>
      </div>
    </section>
  </transition>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick } from "vue";
import { t } from "../i18n";
import { formatLogTime } from "../utils/logFormatter";

export interface LogItem {
  id?: number;
  timestamp: string | Date;
  level: string;
  message: string;
  details?: string;
}

const props = defineProps<{
  isVisible: boolean;
  autoScroll: boolean;
  currentLogFilter: string;
  filteredLogs: LogItem[];
  allLogs?: LogItem[];
  logLevels: { key: string; label: string }[];
}>();

const emit = defineEmits<{
  (e: "update:autoScroll", scroll: boolean): void;
  (e: "update:currentLogFilter", filter: string): void;
  (e: "copy-logs"): void;
  (e: "export-logs"): void;
  (e: "clear-logs"): void;
  (e: "close-log"): void;
}>();

const searchQuery = ref("");
const embeddedTerminalRef = ref<HTMLElement | null>(null);

function getLevelCount(levelKey: string): number {
  const source = props.allLogs || props.filteredLogs;
  if (levelKey === "ALL") return source.length;
  return source.filter((l) => (l.level || "").toUpperCase() === levelKey).length;
}

const displayedLogs = computed(() => {
  const query = searchQuery.value.trim().toLowerCase();
  if (!query) return props.filteredLogs;
  return props.filteredLogs.filter((log) => {
    const msg = (log.message || "").toLowerCase();
    const details = (log.details || "").toLowerCase();
    return msg.includes(query) || details.includes(query);
  });
});

function scrollToBottom() {
  if (props.autoScroll && embeddedTerminalRef.value) {
    embeddedTerminalRef.value.scrollTop = embeddedTerminalRef.value.scrollHeight;
  }
}

watch(
  () => props.filteredLogs.length,
  () => {
    nextTick(scrollToBottom);
  }
);
</script>

<style scoped>
.log-section-card {
  width: 100%;
  margin-top: 1.5rem;
  display: flex;
  flex-direction: column;
  padding: 1.25rem 1.5rem;
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  border-radius: 16px;
  backdrop-filter: blur(16px);
  -webkit-backdrop-filter: blur(16px);
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.25);
  box-sizing: border-box;
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
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.85rem;
  gap: 1rem;
  flex-wrap: wrap;
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
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0% {
    opacity: 1;
  }
  50% {
    opacity: 0.4;
  }
  100% {
    opacity: 1;
  }
}

.log-section-controls {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  flex-wrap: wrap;
}

.filter-tabs-sm {
  display: flex;
  gap: 0.35rem;
  background: var(--subtab-container-bg, rgba(255, 255, 255, 0.04));
  padding: 3px;
  border-radius: 8px;
  border: 1px solid var(--card-border);
}

.btn-tab-sm {
  background: transparent;
  border: 1px solid transparent;
  color: var(--subtab-btn-text, var(--text-muted));
  padding: 0.3rem 0.65rem;
  border-radius: 6px;
  font-size: 0.8rem;
  cursor: pointer;
  transition: all 0.2s ease;
  font-weight: 500;
}

.btn-tab-sm:hover {
  color: var(--text-main);
  background: rgba(255, 255, 255, 0.06);
}

.btn-tab-sm.active {
  background: var(--tab-btn-active-bg, #0284c7);
  color: var(--tab-btn-active-text, #ffffff);
  border-color: var(--subtab-btn-active-border, #0284c7);
  font-weight: 600;
}

.search-box-sm {
  display: flex;
  align-items: center;
}

.search-input-sm {
  background: var(--input-bg, rgba(0, 0, 0, 0.2));
  border: 1px solid var(--card-border);
  color: var(--text-main);
  padding: 0.35rem 0.75rem;
  border-radius: 6px;
  font-size: 0.82rem;
  width: 220px;
  transition: all 0.2s ease;
}

.search-input-sm:focus {
  outline: none;
  border-color: var(--accent-cyan, #0284c7);
  box-shadow: 0 0 8px rgba(2, 132, 199, 0.25);
}

.btn-text-sm {
  background: var(--btn-sec-bg, rgba(255, 255, 255, 0.06));
  border: 1px solid var(--card-border);
  color: var(--text-main);
  padding: 0.35rem 0.75rem;
  border-radius: 6px;
  font-size: 0.82rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
}

.btn-text-sm:hover {
  background: var(--btn-sec-hover-bg, rgba(255, 255, 255, 0.12));
  transform: translateY(-1px);
}

.btn-text-danger-sm {
  background: rgba(239, 68, 68, 0.12);
  border: 1px solid rgba(239, 68, 68, 0.3);
  color: #ef4444;
  padding: 0.35rem 0.75rem;
  border-radius: 6px;
  font-size: 0.82rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
}

.btn-text-danger-sm:hover {
  background: rgba(239, 68, 68, 0.22);
  transform: translateY(-1px);
}

.btn-close-log {
  padding: 0.35rem 0.55rem;
  font-size: 0.85rem;
  line-height: 1;
}

.btn-close-log:hover {
  color: #ef4444;
  background: rgba(239, 68, 68, 0.15);
}

.auto-scroll-label-sm {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  font-size: 0.82rem;
  cursor: pointer;
  padding: 0.35rem 0.65rem;
  border-radius: 6px;
  background: var(--btn-sec-bg, rgba(255, 255, 255, 0.04));
  border: 1px solid var(--card-border);
  color: var(--text-main);
  transition: all 0.2s ease;
}

.auto-scroll-label-sm input[type="checkbox"] {
  accent-color: var(--accent-cyan, #0284c7);
  width: 14px;
  height: 14px;
  cursor: pointer;
}

.auto-scroll-label-sm:hover {
  background: var(--btn-sec-hover-bg, rgba(255, 255, 255, 0.1));
}

.embedded-terminal-window {
  background: var(--terminal-bg, #0b1120);
  border: 1px solid var(--terminal-border, rgba(255, 255, 255, 0.08));
  border-radius: 10px;
  padding: 0.9rem 1.15rem;
  min-height: 280px;
  max-height: 460px;
  overflow-y: auto;
  overflow-x: hidden;
  font-family: "JetBrains Mono", "Fira Code", "Cascadia Code", Consolas, monospace;
  font-size: 0.82rem;
  line-height: 1.6;
}

.embedded-terminal-window .empty-logs {
  color: var(--text-muted, #64748b);
  text-align: center;
  padding: 4rem 0;
}

.log-row {
  display: flex;
  align-items: flex-start;
  gap: 0.85rem;
  padding: 0.3rem 0;
  border-bottom: 1px dashed rgba(255, 255, 255, 0.06);
  box-sizing: border-box;
  width: 100%;
}

.log-time {
  color: var(--text-muted, #64748b);
  font-size: 0.78rem;
  font-family: "JetBrains Mono", monospace;
  white-space: nowrap;
  flex-shrink: 0;
  width: 100px;
  min-width: 100px;
}

.log-level-badge {
  font-size: 0.7rem;
  font-weight: 700;
  padding: 0.05rem 0.45rem;
  border-radius: 4px;
  flex-shrink: 0;
  width: 58px;
  min-width: 58px;
  text-align: center;
  letter-spacing: 0.5px;
}

.log-level-badge.info {
  background: rgba(16, 185, 129, 0.18);
  color: #34d399;
  border: 1px solid rgba(16, 185, 129, 0.3);
}

.log-level-badge.warn,
.log-level-badge.warning {
  background: rgba(245, 158, 11, 0.18);
  color: #fbbf24;
  border: 1px solid rgba(245, 158, 11, 0.3);
}

.log-level-badge.error {
  background: rgba(239, 68, 68, 0.2);
  color: #f87171;
  border: 1px solid rgba(239, 68, 68, 0.35);
}

.log-level-badge.debug {
  background: rgba(168, 85, 247, 0.18);
  color: #c084fc;
  border: 1px solid rgba(168, 85, 247, 0.3);
}

.log-content {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  gap: 0.5rem;
}

.log-msg {
  color: var(--text-main, #e2e8f0);
  font-size: 0.82rem;
  line-height: 20px;
  white-space: pre-wrap;
  word-break: break-word;
  overflow-wrap: anywhere;
}

.log-row.info .log-msg {
  color: #f1f5f9;
}
.log-row.warn .log-msg {
  color: #fde047;
}
.log-row.error .log-msg {
  color: #fca5a5;
}
.log-row.debug .log-msg {
  color: #c084fc;
}

.log-details {
  color: var(--text-muted, #94a3b8);
  font-size: 0.78rem;
  line-height: 20px;
  white-space: pre-wrap;
  word-break: break-word;
  overflow-wrap: anywhere;
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
  background: #0284c7;
  color: #ffffff;
  border-color: #0284c7;
  font-weight: 700;
}

[data-theme="light"] .search-input-sm {
  background: #ffffff;
  border-color: #cbd5e1;
  color: #0f172a;
}

[data-theme="light"] .search-input-sm:focus {
  border-color: #0284c7;
}

[data-theme="light"] .btn-text-sm {
  background: #ffffff;
  border: 1px solid #cbd5e1;
  color: #0f172a;
  font-weight: 600;
  box-shadow: 0 1px 3px rgba(15, 23, 42, 0.08);
}

[data-theme="light"] .btn-text-sm:hover {
  background: #f0f9ff;
  border-color: #0284c7;
  color: #0284c7;
}

[data-theme="light"] .auto-scroll-label-sm {
  background: #ffffff;
  border: 1px solid #cbd5e1;
  color: #0f172a;
  font-weight: 600;
  box-shadow: 0 1px 3px rgba(15, 23, 42, 0.08);
}

[data-theme="light"] .auto-scroll-label-sm input[type="checkbox"] {
  accent-color: #0284c7;
}

[data-theme="light"] .auto-scroll-label-sm:hover {
  background: #f0f9ff;
  border-color: #0284c7;
  color: #0284c7;
}

[data-theme="light"] .btn-text-danger-sm {
  background: #fef2f2;
  border: 1px solid #fca5a5;
  color: #dc2626;
  font-weight: 600;
  box-shadow: 0 1px 3px rgba(239, 68, 68, 0.08);
}

[data-theme="light"] .btn-text-danger-sm:hover {
  background: #fee2e2;
  border-color: #ef4444;
  color: #b91c1c;
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

/* Card fade transition */
.card-fade-enter-active,
.card-fade-leave-active {
  transition: all 0.25s ease-out;
}

.card-fade-enter-from,
.card-fade-leave-to {
  opacity: 0;
  transform: translateY(12px);
}
</style>

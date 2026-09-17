<template>
  <div v-if="isOpen" class="modal-overlay" @click="close">
    <div class="glass-modal confirm-card" :class="{ 'safe-card': isAllVentoy, 'mixed-card': isMixed }" @click.stop>
      <div class="modal-header" :class="isAllVentoy ? 'safe-header' : isMixed ? 'mixed-header' : 'danger-header'">
        <div class="header-title">
          <span class="warning-icon">{{ isAllVentoy ? "🛡️" : isMixed ? "⚡" : "⚠️" }}</span>
          <h3>
            {{ isAllVentoy ? t("confirm.title_safe") : isMixed ? t("confirm.title_mixed") : t("confirm.title_danger") }}
          </h3>
        </div>
        <button class="close-btn" @click="close">✕</button>
      </div>

      <div class="modal-body">
        <!-- Safe Info Banner for ALL Ventoy Disks -->
        <div v-if="isAllVentoy" class="safe-banner">
          <div class="banner-title">{{ t("confirm.safe_banner_title") }}</div>
          <div class="banner-desc">
            {{ t("confirm.safe_banner_desc") }}
          </div>
        </div>

        <!-- Mixed Mode Info Banner for Mixed Selections -->
        <div v-else-if="isMixed" class="mixed-banner">
          <div class="banner-title">{{ t("confirm.mixed_banner_title") }}</div>
          <div class="banner-desc">
            {{ t("confirm.mixed_banner_desc", { ventoyCount: ventoyDisks.length, blankCount: blankDisks.length }) }}
          </div>
        </div>

        <!-- Danger Warning Alert Banner for Pure Blank Disks -->
        <div v-else class="danger-banner">
          <div class="banner-title">{{ t("confirm.danger_banner_title") }}</div>
          <div class="banner-desc">
            {{ t("confirm.danger_banner_desc") }}
          </div>
        </div>

        <!-- Target Devices Summary Box -->
        <div class="target-summary-box">
          <div class="summary-label">{{ t("confirm.summary_title") }}</div>

          <!-- Single Disk Summary -->
          <div v-if="targetDisks.length === 1 && targetDisk" class="target-disk-item">
            <div class="disk-main-info">
              <span class="disk-name">{{ targetDisk.name }}</span>
              <span class="disk-path">{{ targetDisk.device }}</span>
            </div>
            <div class="disk-meta-pills">
              <span class="pill-tag">{{ formatDiskCapacity(targetDisk.formatted) }}</span>
              <span class="pill-tag">{{ targetDisk.fileSystem || "FAT32" }}</span>
              <span class="pill-tag">{{ targetDisk.partitionScheme || "Unknown partition table" }}</span>
              <span class="pill-tag accent" v-if="mode === 'hybrid'">{{
                t("confirm.fs_format", { fs: fsType || "" })
              }}</span>
              <span class="pill-tag highlight">{{ mode === "cloud" ? t("mode.cloud") : t("mode.hybrid") }}</span>
              <span class="pill-tag safe-tag" v-if="isAllVentoy">{{ t("confirm.smart_safe_tag") }}</span>
            </div>
            <div class="target-disk-details">
              <div class="detail-item"><strong>{{ t("confirm.vendor") }}:</strong> {{ targetDisk.vendor || t("confirm.unknown") }}</div>
              <div class="detail-item"><strong>{{ t("confirm.serial") }}:</strong> {{ targetDisk.serialNumber || t("confirm.unavailable") }}</div>
              <div class="detail-item"><strong>{{ t("confirm.system_disk") }}:</strong> {{ targetDisk.isSystem ? t("confirm.yes") : t("confirm.no") }}</div>
              <div class="detail-item"><strong>{{ t("confirm.mounted") }}:</strong> {{ targetDisk.mountPoint || t("confirm.unavailable") }}</div>
              <div class="detail-item full-width"><strong>{{ t("confirm.format_action") }}:</strong> {{ isAllVentoy ? t("confirm.no_format") : t("confirm.will_format") }}</div>
            </div>
          </div>

          <!-- Batch Disks Mixed Summary -->
          <div v-else-if="isMixed" class="batch-summary">
            <div class="mixed-group" v-if="ventoyDisks.length > 0">
              <div class="group-title safe-title">{{ t("confirm.ventoy_group_title") }}</div>
              <div class="batch-device-list">
                <div v-for="disk in ventoyDiskDetails" :key="disk.device" class="batch-device-detail safe-dev-tag">
                  <div class="batch-dev-header">
                    <strong>🛡️ {{ disk.name || disk.device }}</strong>
                    <span class="batch-dev-path">{{ disk.device }}</span>
                  </div>
                  <div class="batch-dev-info-grid">
                    <div><span>{{ formatDiskCapacity(disk.formatted) || t("confirm.unavailable") }}</span> · <span>{{ disk.vendor || t("confirm.unknown") }}</span></div>
                    <div><strong>{{ t("confirm.serial") }}:</strong> {{ disk.serialNumber || t("confirm.unavailable") }}</div>
                    <div><strong>{{ t("confirm.system_disk") }}:</strong> {{ disk.isSystem ? t("confirm.yes") : t("confirm.no") }}</div>
                    <div><strong>{{ t("confirm.mounted") }}:</strong> {{ disk.mountPoint || t("confirm.unavailable") }}</div>
                    <div class="full-width action-text safe-action"><strong>{{ t("confirm.format_action") }}:</strong> {{ t("confirm.no_format") }}</div>
                  </div>
                </div>
              </div>
            </div>
            <div class="mixed-group" v-if="blankDisks.length > 0">
              <div class="group-title danger-title">{{ t("confirm.blank_group_title") }}</div>
              <div class="batch-device-list">
                <div v-for="disk in blankDiskDetails" :key="disk.device" class="batch-device-detail danger-dev-tag">
                  <div class="batch-dev-header">
                    <strong>💾 {{ disk.name || disk.device }}</strong>
                    <span class="batch-dev-path">{{ disk.device }}</span>
                  </div>
                  <div class="batch-dev-info-grid">
                    <div><span>{{ formatDiskCapacity(disk.formatted) || t("confirm.unavailable") }}</span> · <span>{{ disk.vendor || t("confirm.unknown") }}</span></div>
                    <div><strong>{{ t("confirm.serial") }}:</strong> {{ disk.serialNumber || t("confirm.unavailable") }}</div>
                    <div><strong>{{ t("confirm.system_disk") }}:</strong> {{ disk.isSystem ? t("confirm.yes") : t("confirm.no") }}</div>
                    <div><strong>{{ t("confirm.mounted") }}:</strong> {{ disk.mountPoint || t("confirm.unavailable") }}</div>
                    <div class="full-width action-text danger-action"><strong>{{ t("confirm.format_action") }}:</strong> {{ t("confirm.will_format") }}</div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Batch Disks Pure Summary -->
          <div v-else class="batch-summary">
            <div class="batch-count">{{ t("confirm.batch_summary_title", { count: targetDisks.length }) }}</div>
            <div class="batch-device-list">
              <div v-for="disk in targetDiskDetails" :key="disk.device" class="batch-device-detail danger-dev-tag">
                <div class="batch-dev-header">
                  <strong>💾 {{ disk.name || disk.device }}</strong>
                  <span class="batch-dev-path">{{ disk.device }}</span>
                </div>
                <div class="batch-dev-info-grid">
                  <div><span>{{ formatDiskCapacity(disk.formatted) || t("confirm.unavailable") }}</span> · <span>{{ disk.vendor || t("confirm.unknown") }}</span></div>
                  <div><strong>{{ t("confirm.serial") }}:</strong> {{ disk.serialNumber || t("confirm.unavailable") }}</div>
                  <div><strong>{{ t("confirm.system_disk") }}:</strong> {{ disk.isSystem ? t("confirm.yes") : t("confirm.no") }}</div>
                  <div><strong>{{ t("confirm.mounted") }}:</strong> {{ disk.mountPoint || t("confirm.unavailable") }}</div>
                  <div class="full-width action-text danger-action"><strong>{{ t("confirm.format_action") }}:</strong> {{ t("confirm.will_format") }}</div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- ESP Partition Note Banner -->
        <div class="esp-note-banner">
          {{ t("confirm.esp_partition_note") }}
        </div>
      </div>

      <div class="modal-footer">
        <button class="btn-cancel" @click="close">{{ t("confirm.cancel_btn") }}</button>
        <button :class="isAllVentoy || isMixed ? 'btn-safe-confirm' : 'btn-danger-confirm'" @click="confirm">
          {{ confirmBtnText }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { t, formatDiskCapacity } from "../i18n";

interface DiskInfo {
  device: string;
  name: string;
  size: number;
  formatted: string;
  fileSystem?: string;
  bootStatus?: string;
  partitionScheme?: string;
  vendor?: string;
  serialNumber?: string;
  isSystem?: boolean;
  mountPoint?: string;
  isVentoy?: boolean;
  isRealVentoy?: boolean;
}

const props = defineProps<{
  isOpen: boolean;
  mode: "cloud" | "hybrid";
  fsType?: string;
  targetDisk: DiskInfo | null;
  targetDisks: string[];
  allDisks?: DiskInfo[];
}>();

const emit = defineEmits(["close", "confirm"]);

function checkIsExistingBootDisk(disk: any): boolean {
  if (!disk) return false;
  return Boolean(disk.isRealVentoy || disk.isCloudMode);
}

const ventoyDisks = computed(() => {
  if (!props.allDisks || props.allDisks.length === 0) {
    if (props.targetDisk && checkIsExistingBootDisk(props.targetDisk)) {
      return [props.targetDisk.device];
    }
    return [];
  }
  return props.targetDisks.filter((dev) => {
    const found = props.allDisks?.find((d) => d.device === dev);
    return found ? checkIsExistingBootDisk(found) : false;
  });
});

const blankDisks = computed(() => {
  return props.targetDisks.filter((dev) => !ventoyDisks.value.includes(dev));
});

const targetDiskDetails = computed(() =>
  props.targetDisks.map((device) => {
    return (
      props.allDisks?.find((disk) => disk.device === device) || {
        device,
        name: "",
        size: 0,
        formatted: "",
      }
    );
  })
);

const ventoyDiskDetails = computed(() =>
  targetDiskDetails.value.filter((disk) => ventoyDisks.value.includes(disk.device))
);
const blankDiskDetails = computed(() =>
  targetDiskDetails.value.filter((disk) => blankDisks.value.includes(disk.device))
);

const isAllVentoy = computed(() => {
  if (props.targetDisks.length === 0) return false;
  return ventoyDisks.value.length === props.targetDisks.length;
});

const isMixed = computed(() => {
  return ventoyDisks.value.length > 0 && blankDisks.value.length > 0;
});

const confirmBtnText = computed(() => {
  if (props.targetDisks.length === 1) {
    return isAllVentoy.value
      ? t("deploy.start_update")
      : t("confirm.confirm_btn");
  }

  if (isAllVentoy.value) {
    return t("confirm.batch_safe_confirm", { count: props.targetDisks.length });
  } else if (isMixed.value) {
    return t("confirm.batch_mixed_confirm", {
      ventoy: ventoyDisks.value.length,
      blank: blankDisks.value.length,
    });
  } else {
    return t("confirm.batch_danger_confirm", { count: props.targetDisks.length });
  }
});

function close() {
  emit("close");
}

function confirm() {
  emit("confirm");
  close();
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
  z-index: 1100;
  padding: 1rem;
}

.glass-modal {
  background: var(--modal-bg);
  border: 1px solid var(--card-border);
  box-shadow: 0 20px 50px rgba(0, 0, 0, 0.25);
  border-radius: 16px;
  width: 100%;
  max-width: 580px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  color: var(--text-main);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.25rem 1.5rem;
  border-bottom: 1px solid var(--card-border);
  background: var(--modal-header-bg);
}

.glass-modal.safe-card {
  border-left: 4px solid var(--accent-cyan);
}

.glass-modal.mixed-card {
  border-left: 4px solid var(--accent-purple);
}

.modal-header.safe-header h3 {
  color: var(--accent-cyan);
}

.modal-header.mixed-header h3 {
  color: var(--accent-purple);
}

.modal-header h3 {
  font-size: 1.1rem;
  font-weight: 700;
  color: var(--danger);
  margin: 0;
}

.header-title {
  display: flex;
  align-items: center;
  gap: 0.6rem;
}

.warning-icon {
  font-size: 1.3rem;
}

.close-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  font-size: 1.2rem;
  cursor: pointer;
  padding: 0.2rem 0.5rem;
  border-radius: 6px;
  transition: all 0.2s;
}

.close-btn:hover {
  background: var(--section-bg);
  color: var(--text-main);
}

.modal-body {
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.safe-banner {
  background: var(--alert-info-bg);
  border: 1px solid var(--alert-info-border);
  border-radius: 12px;
  padding: 1rem 1.25rem;
}

.safe-banner .banner-title {
  color: var(--alert-info-title);
  font-weight: 700;
  font-size: 0.95rem;
  margin-bottom: 0.4rem;
}

.safe-banner .banner-desc {
  color: var(--alert-info-text);
  font-size: 0.85rem;
  line-height: 1.5;
}

.mixed-banner {
  background: var(--alert-warning-bg);
  border: 1px solid var(--alert-warning-border);
  border-radius: 12px;
  padding: 1rem 1.25rem;
}

.mixed-banner .banner-title {
  color: var(--alert-warning-title);
  font-weight: 700;
  font-size: 0.95rem;
  margin-bottom: 0.4rem;
}

.mixed-banner .banner-desc {
  color: var(--alert-warning-text);
  font-size: 0.85rem;
  line-height: 1.5;
}

.danger-banner {
  background: var(--alert-danger-bg);
  border: 1px solid var(--alert-danger-border);
  border-radius: 12px;
  padding: 1rem 1.25rem;
}

.banner-title {
  color: var(--alert-danger-title);
  font-weight: 700;
  font-size: 0.95rem;
  margin-bottom: 0.4rem;
}

.banner-desc {
  color: var(--alert-danger-text);
  font-size: 0.85rem;
  line-height: 1.5;
}

.mixed-group {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
  margin-bottom: 0.75rem;
}

.group-title {
  font-size: 0.8rem;
  font-weight: 600;
}

.group-title.safe-title {
  color: var(--success);
}

.group-title.danger-title {
  color: var(--danger);
}

.batch-dev-tag.safe-dev-tag {
  background: var(--alert-success-bg);
  border: 1px solid var(--alert-success-border);
  color: var(--alert-success-title);
}

.batch-dev-tag.danger-dev-tag {
  background: var(--alert-danger-bg);
  border: 1px solid var(--alert-danger-border);
  color: var(--alert-danger-title);
}

.target-summary-box {
  background: var(--section-bg);
  border: 1px solid var(--card-border);
  border-radius: 12px;
  padding: 1.25rem;
}

.summary-label {
  font-size: 0.85rem;
  font-weight: 700;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin-bottom: 0.75rem;
}

.target-disk-item {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.disk-main-info {
  display: flex;
  align-items: baseline;
  gap: 0.75rem;
}

.disk-name {
  font-size: 1.1rem;
  font-weight: 700;
  color: var(--text-main);
}

.disk-path {
  font-family: monospace;
  font-size: 0.9rem;
  color: var(--accent-cyan);
}

.disk-meta-pills {
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem;
}

.pill-tag {
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  color: var(--text-main);
  font-size: 0.75rem;
  padding: 0.2rem 0.55rem;
  border-radius: 6px;
  font-weight: 500;
}

.pill-tag.accent {
  border-color: var(--accent-cyan);
  color: var(--accent-cyan);
}

.pill-tag.highlight {
  border-color: var(--accent-purple);
  color: var(--accent-purple);
}

.pill-tag.safe-tag {
  border-color: var(--success);
  color: var(--success);
}

.esp-note-banner {
  margin-top: 0.85rem;
  padding: 0.65rem 0.85rem;
  background: rgba(59, 130, 246, 0.08);
  border: 1px solid rgba(59, 130, 246, 0.25);
  border-radius: 8px;
  font-size: 0.78rem;
  color: var(--text-main);
  line-height: 1.45;
}

.target-disk-details {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.5rem 1rem;
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  border-radius: 8px;
  padding: 0.75rem;
  margin-top: 0.25rem;
}

.detail-item {
  font-size: 0.8rem;
  color: var(--text-muted);
}

.detail-item strong {
  color: var(--text-main);
}

.detail-item.full-width {
  grid-column: span 2;
  margin-top: 0.25rem;
  padding-top: 0.5rem;
  border-top: 1px dashed var(--card-border);
}

.batch-summary {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.batch-count {
  font-size: 0.9rem;
  font-weight: 700;
  color: var(--text-main);
  margin-bottom: 0.5rem;
}

.batch-device-list {
  display: flex;
  flex-direction: column;
  gap: 0.6rem;
  max-height: 220px;
  overflow-y: auto;
  padding-right: 0.2rem;
}

.batch-device-detail {
  border-radius: 8px;
  padding: 0.6rem 0.8rem;
  font-size: 0.8rem;
}

.batch-dev-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.85rem;
  font-weight: 600;
  margin-bottom: 0.35rem;
}

.batch-dev-path {
  font-family: monospace;
  font-size: 0.8rem;
  opacity: 0.85;
}

.batch-dev-info-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.2rem 0.75rem;
  font-size: 0.75rem;
  opacity: 0.9;
}

.batch-dev-info-grid .full-width {
  grid-column: span 2;
  margin-top: 0.2rem;
}

.action-text.safe-action {
  color: var(--success);
  font-weight: 600;
}

.action-text.danger-action {
  color: var(--danger);
  font-weight: 600;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  padding: 1rem 1.5rem;
  background: var(--modal-header-bg);
  border-top: 1px solid var(--card-border);
}

.btn-cancel {
  background: var(--section-bg);
  color: var(--text-main);
  border: 1px solid var(--card-border);
  padding: 0.6rem 1.25rem;
  border-radius: 8px;
  font-weight: 600;
  font-size: 0.875rem;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-cancel:hover {
  background: var(--card-border);
}

.btn-safe-confirm {
  background: #059669;
  color: #ffffff;
  border: none;
  padding: 0.6rem 1.25rem;
  border-radius: 8px;
  font-weight: 600;
  font-size: 0.875rem;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-safe-confirm:hover {
  background: #047857;
  box-shadow: 0 4px 12px rgba(5, 150, 105, 0.25);
}

.btn-danger-confirm {
  background: #dc2626;
  color: #ffffff;
  border: none;
  padding: 0.6rem 1.25rem;
  border-radius: 8px;
  font-weight: 600;
  font-size: 0.875rem;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-danger-confirm:hover {
  background: #b91c1c;
  box-shadow: 0 4px 12px rgba(220, 38, 38, 0.25);
}
</style>

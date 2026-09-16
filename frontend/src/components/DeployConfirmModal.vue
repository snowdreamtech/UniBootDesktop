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
              <span class="pill-tag">{{ targetDisk.formatted }}</span>
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
                    <div><span>{{ disk.formatted || t("confirm.unavailable") }}</span> · <span>{{ disk.vendor || t("confirm.unknown") }}</span></div>
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
                    <div><span>{{ disk.formatted || t("confirm.unavailable") }}</span> · <span>{{ disk.vendor || t("confirm.unknown") }}</span></div>
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
                  <div><span>{{ disk.formatted || t("confirm.unavailable") }}</span> · <span>{{ disk.vendor || t("confirm.unknown") }}</span></div>
                  <div><strong>{{ t("confirm.serial") }}:</strong> {{ disk.serialNumber || t("confirm.unavailable") }}</div>
                  <div><strong>{{ t("confirm.system_disk") }}:</strong> {{ disk.isSystem ? t("confirm.yes") : t("confirm.no") }}</div>
                  <div><strong>{{ t("confirm.mounted") }}:</strong> {{ disk.mountPoint || t("confirm.unavailable") }}</div>
                  <div class="full-width action-text danger-action"><strong>{{ t("confirm.format_action") }}:</strong> {{ t("confirm.will_format") }}</div>
                </div>
              </div>
            </div>
          </div>
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
import { t } from "../i18n";

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
  if (disk.isRealVentoy || disk.isModeB) return true;
  const name = (disk.name || "").toUpperCase();
  const status = (disk.bootStatus || "").toUpperCase();
  const rawStatus = disk.bootStatus || "";
  if (name.includes("VENTOY") || status.includes("VENTOY") || name.includes("UNIBOOT") || status.includes("UNIBOOT")) {
    return true;
  }
  if (status.includes("MODE A") || status.includes("MODE B") || rawStatus.includes("模式 A") || rawStatus.includes("模式 B")) {
    return true;
  }
  return false;
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
  background: rgba(7, 10, 18, 0.82);
  backdrop-filter: blur(10px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1100;
  padding: 1rem;
}

.glass-modal {
  background: var(--modal-bg, rgba(18, 24, 38, 0.96));
  border: 1px solid rgba(239, 68, 68, 0.4);
  box-shadow:
    0 20px 50px rgba(0, 0, 0, 0.7),
    0 0 25px rgba(239, 68, 68, 0.2);
  border-radius: 16px;
  width: 100%;
  max-width: 580px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.25rem 1.5rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
}

.glass-modal.safe-card {
  border: 1px solid rgba(0, 229, 255, 0.4);
  box-shadow:
    0 20px 50px rgba(0, 0, 0, 0.7),
    0 0 25px rgba(0, 229, 255, 0.2);
}

.glass-modal.mixed-card {
  border: 1px solid rgba(168, 85, 247, 0.4);
  box-shadow:
    0 20px 50px rgba(0, 0, 0, 0.7),
    0 0 25px rgba(168, 85, 247, 0.2);
}

.modal-header.safe-header {
  background: rgba(0, 229, 255, 0.08);
}

.modal-header.safe-header h3 {
  color: var(--accent-cyan);
}

.modal-header.mixed-header {
  background: rgba(168, 85, 247, 0.08);
}

.modal-header.mixed-header h3 {
  color: #d8b4fe;
}

.header-title {
  display: flex;
  align-items: center;
  gap: 0.6rem;
}

.warning-icon {
  font-size: 1.3rem;
}

.modal-header h3 {
  font-size: 1.15rem;
  font-weight: 700;
  color: #ef4444;
  margin: 0;
}

.close-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  font-size: 1.2rem;
  cursor: pointer;
  padding: 0.2rem 0.5rem;
  transition: color 0.2s;
}

.close-btn:hover {
  color: #fff;
}

.modal-body {
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.safe-banner {
  background: rgba(0, 229, 255, 0.08);
  border: 1px solid rgba(0, 229, 255, 0.35);
  border-radius: 12px;
  padding: 1rem 1.25rem;
}

.safe-banner .banner-title {
  color: var(--accent-cyan);
  font-weight: 700;
  font-size: 0.95rem;
  margin-bottom: 0.4rem;
}

.safe-banner .banner-desc {
  color: #a5f3fc;
  font-size: 0.825rem;
  line-height: 1.5;
}

.mixed-banner {
  background: rgba(168, 85, 247, 0.1);
  border: 1px solid rgba(168, 85, 247, 0.35);
  border-radius: 12px;
  padding: 1rem 1.25rem;
}

.mixed-banner .banner-title {
  color: #d8b4fe;
  font-weight: 700;
  font-size: 0.95rem;
  margin-bottom: 0.4rem;
}

.mixed-banner .banner-desc {
  color: #e9d5ff;
  font-size: 0.825rem;
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
  color: #4ade80;
}

.group-title.danger-title {
  color: #f87171;
}

.batch-dev-tag.safe-dev-tag {
  background: rgba(34, 197, 94, 0.12);
  border-color: rgba(34, 197, 94, 0.3);
  color: #4ade80;
}

.batch-dev-tag.danger-dev-tag {
  background: rgba(239, 68, 68, 0.12);
  border-color: rgba(239, 68, 68, 0.3);
  color: #fca5a5;
}

.danger-banner {
  background: rgba(239, 68, 68, 0.12);
  border: 1px solid rgba(239, 68, 68, 0.35);
  border-radius: 12px;
  padding: 1rem 1.25rem;
}

.banner-title {
  color: #ef4444;
  font-weight: 700;
  font-size: 0.95rem;
  margin-bottom: 0.4rem;
}

.banner-desc {
  color: #fca5a5;
  font-size: 0.825rem;
  line-height: 1.5;
}

.target-summary-box {
  background: rgba(255, 255, 255, 0.02);
  border: 1px solid var(--card-border);
  border-radius: 12px;
  padding: 1.25rem;
}

.summary-label {
  font-size: 0.825rem;
  color: var(--text-muted);
  margin-bottom: 0.75rem;
}

.target-disk-item {
  display: flex;
  flex-direction: column;
  gap: 0.6rem;
}

.disk-main-info {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
}

.disk-name {
  font-size: 1rem;
  font-weight: 700;
  color: #fff;
}

.disk-path {
  font-family: monospace;
  font-size: 0.8rem;
  color: var(--accent-cyan);
}

.disk-meta-pills {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.pill-tag {
  background: rgba(255, 255, 255, 0.06);
  border: 1px solid rgba(255, 255, 255, 0.1);
  padding: 0.2rem 0.6rem;
  border-radius: 6px;
  font-size: 0.75rem;
  color: var(--text-main);
}

.pill-tag.accent {
  background: rgba(0, 229, 255, 0.12);
  border-color: rgba(0, 229, 255, 0.3);
  color: var(--accent-cyan);
  font-weight: 600;
}

.pill-tag.highlight {
  background: rgba(168, 85, 247, 0.15);
  border-color: rgba(168, 85, 247, 0.3);
  color: #d8b4fe;
  font-weight: 600;
}

.batch-summary {
  display: flex;
  flex-direction: column;
  gap: 0.6rem;
}

.batch-count {
  font-size: 0.9rem;
  color: #fff;
}

.batch-tags {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.batch-dev-tag {
  background: rgba(0, 229, 255, 0.1);
  border: 1px solid rgba(0, 229, 255, 0.2);
  padding: 0.25rem 0.6rem;
  border-radius: 6px;
  font-family: monospace;
  font-size: 0.8rem;
  color: var(--accent-cyan);
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.8rem;
  padding: 1rem 1.5rem;
  background: rgba(0, 0, 0, 0.2);
  border-top: 1px solid var(--card-border);
}

.btn-cancel {
  background: rgba(255, 255, 255, 0.06);
  border: 1px solid var(--card-border);
  color: var(--text-main);
  padding: 0.6rem 1.25rem;
  border-radius: 8px;
  font-size: 0.875rem;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-cancel:hover {
  background: rgba(255, 255, 255, 0.12);
}

.pill-tag.safe-tag {
  background: rgba(34, 197, 94, 0.15);
  border-color: rgba(34, 197, 94, 0.3);
  color: #4ade80;
  font-weight: 600;
}

.btn-safe-confirm {
  background: linear-gradient(135deg, #00e5ff 0%, #0284c7 100%);
  border: none;
  color: #070a12;
  padding: 0.6rem 1.25rem;
  border-radius: 8px;
  font-size: 0.875rem;
  font-weight: 700;
  cursor: pointer;
  box-shadow: 0 4px 14px rgba(0, 229, 255, 0.35);
  transition: all 0.2s;
}

.btn-safe-confirm:hover {
  background: linear-gradient(135deg, #38bdf8 0%, #00e5ff 100%);
  box-shadow: 0 6px 20px rgba(0, 229, 255, 0.5);
  transform: translateY(-1px);
}

.btn-danger-confirm {
  background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
  border: none;
  color: #fff;
  padding: 0.6rem 1.25rem;
  border-radius: 8px;
  font-size: 0.875rem;
  font-weight: 700;
  cursor: pointer;
  box-shadow: 0 4px 14px rgba(239, 68, 68, 0.4);
  transition: all 0.2s;
}

.btn-danger-confirm:hover {
  background: linear-gradient(135deg, #f87171 0%, #ef4444 100%);
  box-shadow: 0 6px 20px rgba(239, 68, 68, 0.6);
  transform: translateY(-1px);
}

.target-disk-details {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 0.4rem 1rem;
  background: rgba(0, 0, 0, 0.25);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 8px;
  padding: 0.75rem 1rem;
  margin-top: 0.25rem;
}

.detail-item {
  font-size: 0.8rem;
  color: var(--text-muted);
}

.detail-item strong {
  color: var(--text-main);
  font-weight: 600;
}

.detail-item.full-width {
  grid-column: span 2;
  border-top: 1px dashed rgba(255, 255, 255, 0.1);
  padding-top: 0.4rem;
  margin-top: 0.2rem;
}

.batch-device-list {
  display: flex;
  flex-direction: column;
  gap: 0.6rem;
}

.batch-device-detail {
  padding: 0.75rem 0.9rem;
  border-radius: 8px;
  border: 1px solid transparent;
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.batch-dev-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.875rem;
}

.batch-dev-path {
  font-family: monospace;
  font-size: 0.775rem;
  opacity: 0.9;
}

.batch-dev-info-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 0.3rem 0.8rem;
  font-size: 0.775rem;
  line-height: 1.4;
}

.batch-dev-info-grid .full-width {
  grid-column: span 2;
}

.batch-dev-info-grid .action-text {
  font-weight: 600;
  margin-top: 0.1rem;
}

.safe-action {
  color: #4ade80;
}

.danger-action {
  color: #f87171;
}

/* Light Theme Overrides */
[data-theme="light"] .glass-modal {
  background: #ffffff;
  border-color: #cbd5e1;
  box-shadow: 0 25px 60px rgba(15, 23, 42, 0.22);
}

[data-theme="light"] .glass-modal.safe-card {
  border-color: #06b6d4;
  box-shadow: 0 25px 60px rgba(6, 182, 212, 0.18);
}

[data-theme="light"] .glass-modal.mixed-card {
  border-color: #a855f7;
  box-shadow: 0 25px 60px rgba(168, 85, 247, 0.18);
}

[data-theme="light"] .modal-header {
  border-bottom-color: #e2e8f0;
}

[data-theme="light"] .modal-header.safe-header {
  background: #ecfeff;
}

[data-theme="light"] .modal-header.safe-header h3 {
  color: #0891b2;
}

[data-theme="light"] .modal-header.mixed-header {
  background: #faf5ff;
}

[data-theme="light"] .modal-header.mixed-header h3 {
  color: #7e22ce;
}

[data-theme="light"] .modal-header.danger-header {
  background: #fef2f2;
}

[data-theme="light"] .modal-header.danger-header h3 {
  color: #dc2626;
}

[data-theme="light"] .close-btn {
  color: #64748b;
}

[data-theme="light"] .close-btn:hover {
  color: #0f172a;
}

[data-theme="light"] .safe-banner {
  background: #ecfeff;
  border-color: #67e8f9;
}

[data-theme="light"] .safe-banner .banner-title {
  color: #0891b2;
}

[data-theme="light"] .safe-banner .banner-desc {
  color: #0e7490;
  font-weight: 500;
}

[data-theme="light"] .mixed-banner {
  background: #faf5ff;
  border-color: #d8b4fe;
}

[data-theme="light"] .mixed-banner .banner-title {
  color: #7e22ce;
}

[data-theme="light"] .mixed-banner .banner-desc {
  color: #6b21a8;
  font-weight: 500;
}

[data-theme="light"] .danger-banner {
  background: #fef2f2;
  border-color: #fca5a5;
}

[data-theme="light"] .danger-banner .banner-title {
  color: #dc2626;
}

[data-theme="light"] .danger-banner .banner-desc {
  color: #991b1b;
  font-weight: 500;
}

[data-theme="light"] .target-summary-box {
  background: #f8fafc;
  border-color: #e2e8f0;
}

[data-theme="light"] .summary-label {
  color: #475569;
  font-weight: 600;
}

[data-theme="light"] .disk-name {
  color: #0f172a;
}

[data-theme="light"] .disk-path {
  color: #0284c7;
}

[data-theme="light"] .target-disk-details {
  background: #f1f5f9;
  border-color: #cbd5e1;
}

[data-theme="light"] .detail-item {
  color: #475569;
}

[data-theme="light"] .detail-item strong {
  color: #0f172a;
}

[data-theme="light"] .detail-item.full-width {
  border-top-color: #cbd5e1;
}

[data-theme="light"] .pill-tag {
  background: #e2e8f0;
  border-color: #cbd5e1;
  color: #1e293b;
}

[data-theme="light"] .pill-tag.accent {
  background: #e0f2fe;
  border-color: #7dd3fc;
  color: #0369a1;
}

[data-theme="light"] .pill-tag.highlight {
  background: #f3e8ff;
  border-color: #d8b4fe;
  color: #6b21a8;
}

[data-theme="light"] .pill-tag.safe-tag {
  background: #dcfce7;
  border-color: #86efac;
  color: #15803d;
}

[data-theme="light"] .batch-count {
  color: #0f172a;
}

[data-theme="light"] .batch-dev-tag.safe-dev-tag {
  background: #f0fdf4;
  border-color: #86efac;
  color: #166534;
}

[data-theme="light"] .batch-dev-tag.safe-dev-tag .safe-action {
  color: #15803d;
}

[data-theme="light"] .batch-dev-tag.danger-dev-tag {
  background: #fef2f2;
  border-color: #fca5a5;
  color: #991b1b;
}

[data-theme="light"] .batch-dev-tag.danger-dev-tag .danger-action {
  color: #b91c1c;
}

[data-theme="light"] .modal-footer {
  background: #f8fafc;
  border-top-color: #e2e8f0;
}

[data-theme="light"] .btn-cancel {
  background: #f1f5f9;
  color: #0f172a;
  border-color: #cbd5e1;
  font-weight: 600;
}

[data-theme="light"] .btn-cancel:hover {
  background: #e2e8f0;
}
</style>

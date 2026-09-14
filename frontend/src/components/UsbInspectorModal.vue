<template>
  <div v-if="isOpen" class="modal-overlay" @click.self="closeModal">
    <div class="modal-card glass-modal">
      <div class="modal-header">
        <div class="header-title">
          <span class="modal-icon">ℹ️</span>
          <div>
            <h3>{{ t('inspector.title') }}</h3>
            <p class="subtitle">PHY Protocol, Speed & Hardware Analysis</p>
          </div>
        </div>
        <button class="close-btn" @click="closeModal">✕</button>
      </div>

      <div class="modal-body" v-if="disk">
        <!-- Fake USB 3.0 Alert Banner -->
        <div v-if="disk.isFakeUsb3" class="audit-banner fake-alert">
          <div class="banner-icon">⚠️</div>
          <div class="banner-text">
            <h4>{{ t('inspector.fake_title') }}</h4>
            <p>{{ t('inspector.fake_desc', { speed: disk.usbSpeed || '480 Mb/s' }) }}</p>
          </div>
        </div>

        <div v-else-if="disk.protocolCode === 'usb3_0' || disk.protocolCode === 'usb3_1' || disk.protocolCode === 'usb3_2' || disk.protocolCode === 'usb4'" class="audit-banner genuine-pass">
          <div class="banner-icon">✅</div>
          <div class="banner-text">
            <h4>{{ t('inspector.genuine_title') }}</h4>
            <p>{{ t('inspector.genuine_desc', { speed: disk.usbSpeed || '5 Gb/s' }) }}</p>
          </div>
        </div>

        <div v-else class="audit-banner usb2-info">
          <div class="banner-icon">ℹ️</div>
          <div class="banner-text">
            <h4>{{ t('inspector.usb2_title') }}</h4>
            <p>{{ t('inspector.usb2_desc') }}</p>
          </div>
        </div>

        <!-- Basic Device Info Header & Grid -->
        <div class="section-divider">
          <span>{{ t('inspector.section_basic') }}</span>
        </div>

        <div class="spec-grid">
          <div class="spec-item">
            <span class="spec-label">{{ t('inspector.lbl_name') }}</span>
            <span class="spec-val highlight">{{ disk.name }}</span>
          </div>

          <div class="spec-item">
            <span class="spec-label">{{ t('inspector.lbl_vendor') }}</span>
            <span class="spec-val">{{ disk.vendor || 'Generic USB Device' }}</span>
          </div>

          <div class="spec-item">
            <span class="spec-label">{{ t('inspector.lbl_size') }}</span>
            <span class="spec-val">{{ disk.formatted }}</span>
          </div>

          <div class="spec-item">
            <span class="spec-label">{{ t('inspector.lbl_free') }}</span>
            <span class="spec-val highlight">{{ disk.freeFormatted || 'N/A' }}</span>
          </div>

          <div class="spec-item">
            <span class="spec-label">{{ t('inspector.lbl_protocol') }}</span>
            <span class="spec-val badge-val" :class="disk.protocolCode || 'usb2'">
              {{ disk.usbVersion || 'USB 2.0' }}
            </span>
          </div>

          <div class="spec-item">
            <span class="spec-label">{{ t('inspector.lbl_speed') }}</span>
            <span class="spec-val speed-val" :class="{ 'slow-speed': disk.isFakeUsb3 }">
              ⚡ {{ disk.usbSpeed || '480 Mb/s' }}
            </span>
          </div>

          <div class="spec-item">
            <span class="spec-label">{{ t('inspector.lbl_fs') }}</span>
            <span class="spec-val highlight">{{ disk.fileSystem || 'ExFAT / FAT32' }}</span>
          </div>

          <div class="spec-item">
            <span class="spec-label">{{ t('inspector.lbl_scheme') }}</span>
            <span class="spec-val highlight">{{ disk.partitionScheme || 'GPT / MBR' }}</span>
          </div>

          <div class="spec-item">
            <span class="spec-label">{{ t('inspector.lbl_mount') }}</span>
            <span class="spec-val code">{{ disk.device }}</span>
          </div>

          <div class="spec-item">
            <span class="spec-label">{{ t('inspector.lbl_perm') }}</span>
            <span class="spec-val" :class="disk.writable !== false ? 'pass-val' : 'warn-val'">
              {{ disk.writable !== false ? t('inspector.val_rw') : t('inspector.val_ro') }}
            </span>
          </div>

          <div class="spec-item spec-full">
            <span class="spec-label">{{ t('inspector.lbl_boot_status') }}</span>
            <span class="spec-val highlight">{{ disk.bootStatus || t('inspector.val_data_disk') }}</span>
          </div>
        </div>

        <!-- Hardware Details Header & Grid -->
        <div class="section-divider">
          <span>{{ t('inspector.section_hw') }}</span>
        </div>

        <div class="spec-grid advanced-grid">
          <div class="spec-item">
            <span class="spec-label">{{ t('inspector.lbl_smart') }}</span>
            <span class="spec-val" :class="disk.smartStatus === 'Verified' ? 'pass-val' : 'highlight'">
              {{ disk.smartStatus === 'Verified' ? t('inspector.val_smart_good') : (disk.smartStatus || 'ℹ️ N/A') }}
            </span>
          </div>

          <div class="spec-item" v-if="disk.busPower || disk.busPowerUsed">
            <span class="spec-label">{{ t('inspector.lbl_bus_power') }}</span>
            <span class="spec-val highlight">
              {{ formatPower(disk.busPowerUsed || disk.busPower) }} {{ t('inspector.val_power_limit', { limit: formatPower(disk.busPower) }) }}
            </span>
          </div>

          <div class="spec-item">
            <span class="spec-label">{{ t('inspector.lbl_transport') }}</span>
            <span class="spec-val highlight">{{ disk.transportProtocol || 'BOT (Bulk-Only Transport)' }}</span>
          </div>

          <div class="spec-item">
            <span class="spec-label">{{ t('inspector.lbl_sector') }}</span>
            <span class="spec-val highlight">{{ disk.sectorSize || '512 Bytes (512n/512e)' }}</span>
          </div>

          <div class="spec-item">
            <span class="spec-label">{{ t('inspector.lbl_controller') }}</span>
            <span class="spec-val highlight">{{ disk.controllerVendor || t('inspector.std_controller') }}</span>
          </div>

          <div class="spec-item" v-if="disk.vendorId || disk.productId">
            <span class="spec-label">{{ t('inspector.lbl_vid_pid') }}</span>
            <span class="spec-val code">VID: {{ disk.vendorId || 'N/A' }} | PID: {{ disk.productId || 'N/A' }}</span>
          </div>

          <div class="spec-item spec-full" v-if="disk.serialNumber">
            <span class="spec-label">{{ t('inspector.lbl_serial') }}</span>
            <span class="spec-val code">{{ disk.serialNumber }}</span>
          </div>
        </div>

        <!-- Protocol Compatibility Matrix -->
        <div class="protocol-matrix">
          <h4>{{ t('inspector.ext_protocols') }}</h4>
          <div class="matrix-pills">
            <span class="matrix-pill" :class="{ active: disk.protocolCode === 'usb2' }">USB 2.0 (480 Mbps)</span>
            <span class="matrix-pill" :class="{ active: disk.protocolCode === 'usb3_0' }">USB 3.0 / 3.2 Gen 1 (5 Gbps)</span>
            <span class="matrix-pill" :class="{ active: disk.protocolCode === 'usb3_1' }">USB 3.1 / 3.2 Gen 2 (10 Gbps)</span>
            <span class="matrix-pill" :class="{ active: disk.protocolCode === 'usb3_2' }">USB 3.2 Gen 2x2 (20 Gbps)</span>
            <span class="matrix-pill" :class="{ active: disk.protocolCode === 'usb4' }">USB4 / Thunderbolt 4 (40 Gbps)</span>
          </div>
        </div>
      </div>

      <div class="modal-footer">
        <button class="btn-primary" @click="closeModal">{{ t('inspector.close') }}</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
interface DiskInfo {
  device: string;
  name: string;
  size: number;
  formatted: string;
  freeSpace?: number;
  freeFormatted?: string;
  isRemovable: boolean;
  isSystem: boolean;
  usbVersion?: string;
  usbSpeed?: string;
  vendor?: string;
  fileSystem?: string;
  partitionScheme?: string;
  writable?: boolean;
  serialNumber?: string;
  vendorId?: string;
  productId?: string;
  smartStatus?: string;
  busPower?: string;
  busPowerUsed?: string;
  sectorSize?: string;
  transportProtocol?: string;
  bootStatus?: string;
  controllerVendor?: string;
  isFakeUsb3?: boolean;
  protocolCode?: string;
}

defineProps<{
  isOpen: boolean;
  disk: DiskInfo | null;
}>();

import { t } from '../i18n';

const emit = defineEmits(['close']);

function formatPower(val?: string): string {
  if (!val) return '500 mA';
  const str = val.trim();
  if (!str.toLowerCase().includes('ma') && !str.toLowerCase().includes('a')) {
    return `${str} mA`;
  }
  return str;
}

function closeModal() {
  emit('close');
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(7, 10, 18, 0.75);
  backdrop-filter: blur(8px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.glass-modal {
  background: var(--modal-bg, rgba(18, 24, 38, 0.95));
  border: 1px solid rgba(0, 229, 255, 0.25);
  box-shadow: 0 16px 40px rgba(0, 0, 0, 0.6), 0 0 20px rgba(0, 229, 255, 0.1);
  border-radius: 16px;
  width: 90%;
  max-width: 650px;
  max-height: 85vh;
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

.header-title {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.modal-icon {
  font-size: 1.6rem;
}

.modal-header h3 {
  font-size: 1.15rem;
  font-weight: 700;
  margin: 0;
  color: var(--text-main);
}

.subtitle {
  font-size: 0.775rem;
  color: var(--text-muted);
  margin: 0.15rem 0 0 0;
}

.close-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  font-size: 1.2rem;
  cursor: pointer;
  padding: 0.25rem 0.5rem;
}

.close-btn:hover {
  color: #fff;
}

.modal-body {
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  overflow-y: auto;
  flex: 1;
}

.modal-body::-webkit-scrollbar {
  width: 6px;
}

.modal-body::-webkit-scrollbar-track {
  background: rgba(0, 0, 0, 0.15);
  border-radius: 4px;
}

.modal-body::-webkit-scrollbar-thumb {
  background: rgba(0, 229, 255, 0.3);
  border-radius: 4px;
}

.modal-body::-webkit-scrollbar-thumb:hover {
  background: var(--accent-cyan);
}

.audit-banner {
  display: flex;
  gap: 1rem;
  padding: 1rem 1.25rem;
  border-radius: 12px;
  align-items: flex-start;
}

.audit-banner.fake-alert {
  background: rgba(239, 68, 68, 0.15);
  border: 1px solid rgba(239, 68, 68, 0.4);
}

.audit-banner.fake-alert h4 {
  color: #f87171;
  margin: 0 0 0.3rem 0;
}

.audit-banner.fake-alert p {
  color: #fca5a5;
  font-size: 0.825rem;
  margin: 0;
  line-height: 1.45;
}

.audit-banner.genuine-pass {
  background: rgba(16, 185, 129, 0.15);
  border: 1px solid rgba(16, 185, 129, 0.4);
}

.audit-banner.genuine-pass h4 {
  color: #34d399;
  margin: 0 0 0.3rem 0;
}

.audit-banner.genuine-pass p {
  color: #a7f3d0;
  font-size: 0.825rem;
  margin: 0;
  line-height: 1.45;
}

.audit-banner.usb2-info {
  background: rgba(59, 130, 246, 0.15);
  border: 1px solid rgba(59, 130, 246, 0.3);
}

.audit-banner.usb2-info h4 {
  color: #60a5fa;
  margin: 0 0 0.3rem 0;
}

.audit-banner.usb2-info p {
  color: #bfdbfe;
  font-size: 0.825rem;
  margin: 0;
}

.banner-icon {
  font-size: 1.5rem;
}

.spec-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.75rem;
  background: rgba(0, 0, 0, 0.2);
  padding: 1rem;
  border-radius: 12px;
  border: 1px solid rgba(255, 255, 255, 0.05);
}

.section-divider {
  display: flex;
  align-items: center;
  padding: 0.4rem 0.2rem;
  font-size: 0.8rem;
  font-weight: 700;
  color: var(--accent-cyan);
  border-bottom: 1px dashed rgba(0, 229, 255, 0.2);
}

.spec-item {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.spec-item.spec-full {
  grid-column: 1 / -1;
}

.spec-label {
  font-size: 0.75rem;
  color: var(--text-muted);
}

.spec-val {
  font-size: 0.9rem;
  font-weight: 600;
  color: #e2e8f0;
}

.spec-val.highlight {
  color: #fff;
  font-weight: 700;
}

.spec-val.pass-val {
  color: #34d399;
}

.spec-val.warn-val {
  color: #ef4444;
  font-weight: 700;
}

.spec-val.code {
  font-family: monospace;
  font-size: 0.825rem;
  color: var(--accent-cyan);
}

.badge-val {
  display: inline-block;
  padding: 0.2rem 0.5rem;
  border-radius: 6px;
  font-size: 0.8rem;
  width: fit-content;
}

.badge-val.usb2 {
  background: rgba(148, 163, 184, 0.2);
  color: #cbd5e1;
}

.badge-val.usb3_0 {
  background: rgba(0, 229, 255, 0.2);
  color: #00e5ff;
}

.badge-val.usb3_1, .badge-val.usb3_2 {
  background: rgba(157, 78, 221, 0.25);
  color: #c084fc;
}

.badge-val.usb4 {
  background: rgba(245, 158, 11, 0.25);
  color: #fbbf24;
}

.speed-val {
  color: var(--accent-cyan);
}

.speed-val.slow-speed {
  color: #ef4444;
  font-weight: 700;
}

.protocol-matrix h4 {
  font-size: 0.85rem;
  color: var(--text-muted);
  margin: 0 0 0.6rem 0;
}

.matrix-pills {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.matrix-pill {
  font-size: 0.75rem;
  padding: 0.3rem 0.6rem;
  border-radius: 6px;
  background: rgba(255, 255, 255, 0.05);
  color: var(--text-muted);
  border: 1px solid transparent;
}

.matrix-pill.active {
  background: rgba(0, 229, 255, 0.15);
  color: var(--accent-cyan);
  border-color: rgba(0, 229, 255, 0.4);
  font-weight: 700;
}

.modal-footer {
  padding: 1rem 1.5rem;
  border-top: 1px solid rgba(255, 255, 255, 0.08);
  display: flex;
  justify-content: flex-end;
}

/* Light Theme Overrides */
[data-theme="light"] .glass-modal {
  background: #ffffff;
  border-color: #cbd5e1;
  box-shadow: 0 25px 60px rgba(15, 23, 42, 0.18);
}

[data-theme="light"] .modal-header h3 {
  color: #0f172a;
}

[data-theme="light"] .subtitle {
  color: #475569;
}

[data-theme="light"] .close-btn {
  color: #64748b;
}

[data-theme="light"] .close-btn:hover {
  color: #0f172a;
}

[data-theme="light"] .spec-card {
  background: #f8fafc;
  border-color: #cbd5e1;
}

[data-theme="light"] .spec-label {
  color: #475569;
}

[data-theme="light"] .spec-val {
  color: #0f172a;
}

[data-theme="light"] .matrix-pill {
  background: #f1f5f9;
  color: #475569;
}

[data-theme="light"] .matrix-pill.active {
  background: #e0f2fe;
  color: #0284c7;
  border-color: #38bdf8;
}

[data-theme="light"] .modal-footer {
  background: #f8fafc;
  border-top-color: #e2e8f0;
}
</style>

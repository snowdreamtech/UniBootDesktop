<template>
  <div v-if="isOpen" class="modal-overlay" @click.self="closeModal">
    <div class="modal-card glass-modal">
      <div class="modal-header">
        <div class="header-title">
          <span class="modal-icon">ℹ️</span>
          <div>
            <h3>USB 设备硬件详情</h3>
            <p class="subtitle">物理 PHY 协议识别、底层传输速率与硬件数据</p>
          </div>
        </div>
        <button class="close-btn" @click="closeModal">✕</button>
      </div>

      <div class="modal-body" v-if="disk">
        <!-- Fake USB 3.0 Alert Banner -->
        <div v-if="disk.isFakeUsb3" class="audit-banner fake-alert">
          <div class="banner-icon">⚠️</div>
          <div class="banner-text">
            <h4>伪造 USB 3.0 预警！(Fake USB 3.0 Alert)</h4>
            <p>设备宣传名称带有 USB 3.0 / 3.1 标识，但操作系统物理层实际协商速率仅为 <strong>{{ disk.usbSpeed || '480 Mb/s' }}</strong> (USB 2.0 High-Speed PHY)。此 U 盘疑似被刷固件或使用虚假蓝色接口。</p>
          </div>
        </div>

        <div v-else-if="disk.protocolCode === 'usb3_0' || disk.protocolCode === 'usb3_1' || disk.protocolCode === 'usb3_2' || disk.protocolCode === 'usb4'" class="audit-banner genuine-pass">
          <div class="banner-icon">✅</div>
          <div class="banner-text">
            <h4>物理硬件校验通过 (Genuine USB 3.0+ Device)</h4>
            <p>硬件物理层已成功建立 SuperSpeed/SuperSpeed+ 高速通道，实测协商速率为 <strong>{{ disk.usbSpeed || '5 Gb/s' }}</strong>。</p>
          </div>
        </div>

        <div v-else class="audit-banner usb2-info">
          <div class="banner-icon">ℹ️</div>
          <div class="banner-text">
            <h4>标准 USB 2.0 传输接口</h4>
            <p>设备硬件版本为 USB 2.0，理论物理最高速率 480 Mb/s (High-Speed)。</p>
          </div>
        </div>

        <!-- Spec Data Table Grid -->
        <div class="spec-grid">
          <div class="spec-item">
            <span class="spec-label">设备名称 (Device Name)</span>
            <span class="spec-val highlight">{{ disk.name }}</span>
          </div>

          <div class="spec-item">
            <span class="spec-label">厂商/制造商 (Vendor)</span>
            <span class="spec-val">{{ disk.vendor || 'Generic USB Device' }}</span>
          </div>

          <div class="spec-item">
            <span class="spec-label">系统挂载路径 (Mount Path)</span>
            <span class="spec-val code">{{ disk.device }}</span>
          </div>

          <div class="spec-item">
            <span class="spec-label">设备容量 (Storage Size)</span>
            <span class="spec-val">{{ disk.formatted }} ({{ disk.size.toLocaleString() }} Bytes)</span>
          </div>

          <div class="spec-item">
            <span class="spec-label">USB 协议版本 (Protocol Version)</span>
            <span class="spec-val badge-val" :class="disk.protocolCode || 'usb2'">
              {{ disk.usbVersion || 'USB 2.0' }}
            </span>
          </div>

          <div class="spec-item">
            <span class="spec-label">物理 PHY 速率 (Negotiated Speed)</span>
            <span class="spec-val speed-val" :class="{ 'slow-speed': disk.isFakeUsb3 }">
              ⚡ {{ disk.usbSpeed || '480 Mb/s' }}
            </span>
          </div>
        </div>

        <!-- Protocol Compatibility Matrix -->
        <div class="protocol-matrix">
          <h4>USB 协议支持库扩展 (Extensible Protocol Standards)</h4>
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
        <button class="btn-primary" @click="closeModal">确定 / 关闭</button>
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
  isRemovable: boolean;
  isSystem: boolean;
  usbVersion?: string;
  usbSpeed?: string;
  vendor?: string;
  isFakeUsb3?: boolean;
  protocolCode?: string;
}

defineProps<{
  isOpen: boolean;
  disk: DiskInfo | null;
}>();

const emit = defineEmits(['close']);

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
  background: rgba(18, 24, 38, 0.95);
  border: 1px solid rgba(0, 229, 255, 0.25);
  box-shadow: 0 16px 40px rgba(0, 0, 0, 0.6), 0 0 20px rgba(0, 229, 255, 0.1);
  border-radius: 16px;
  width: 90%;
  max-width: 620px;
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
  color: #fff;
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

.spec-item {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
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
</style>

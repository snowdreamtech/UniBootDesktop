<template>
  <div v-if="isOpen" class="modal-overlay" @click="close">
    <div class="glass-modal confirm-card" @click.stop>
      <div class="modal-header danger-header">
        <div class="header-title">
          <span class="warning-icon">⚠️</span>
          <h3>高危操作确认：即将抹除 U 盘数据</h3>
        </div>
        <button class="close-btn" @click="close">✕</button>
      </div>

      <div class="modal-body">
        <!-- Danger Warning Alert Banner -->
        <div class="danger-banner">
          <div class="banner-title">💥 警告：格式化过程不可逆！</div>
          <div class="banner-desc">
            部署写入将对目标设备进行<strong>底层重新分区与格式化</strong>，改写主引导记录 (MBR/GPT)。<strong>所选 U 盘上的全部现有数据、文档与资料将被彻底永久清空</strong>。
          </div>
        </div>

        <!-- Target Devices Summary Box -->
        <div class="target-summary-box">
          <div class="summary-label">本次将部署写入的目标设备：</div>
          
          <!-- Single Disk Summary -->
          <div v-if="targetDisks.length === 1 && targetDisk" class="target-disk-item">
            <div class="disk-main-info">
              <span class="disk-name">{{ targetDisk.name }}</span>
              <span class="disk-path">{{ targetDisk.device }}</span>
            </div>
            <div class="disk-meta-pills">
              <span class="pill-tag">{{ targetDisk.formatted }}</span>
              <span class="pill-tag">{{ targetDisk.fileSystem || 'FAT32' }}</span>
              <span class="pill-tag accent" v-if="mode === 'hybrid'">{{ fsType }} 格式</span>
              <span class="pill-tag highlight">{{ mode === 'cloud' ? '模式 B (1秒云端)' : '模式 A (混合双模)' }}</span>
            </div>
          </div>

          <!-- Batch Disks Summary -->
          <div v-else class="batch-summary">
            <div class="batch-count">已选中 <strong>{{ targetDisks.length }}</strong> 块 U 盘独立并行写入：</div>
            <div class="batch-tags">
              <span v-for="dev in targetDisks" :key="dev" class="batch-dev-tag">💾 {{ dev }}</span>
            </div>
          </div>
        </div>
      </div>

      <div class="modal-footer">
        <button class="btn-cancel" @click="close">取消</button>
        <button class="btn-danger-confirm" @click="confirm">
          ⚠️ 确认数据已备份，开始格式化写入
        </button>
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
  fileSystem?: string;
}

defineProps<{
  isOpen: boolean;
  mode: 'cloud' | 'hybrid';
  fsType?: string;
  targetDisk: DiskInfo | null;
  targetDisks: string[];
}>();

const emit = defineEmits(['close', 'confirm']);

function close() {
  emit('close');
}

function confirm() {
  emit('confirm');
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
  background: rgba(18, 24, 38, 0.96);
  border: 1px solid rgba(239, 68, 68, 0.4);
  box-shadow: 0 20px 50px rgba(0, 0, 0, 0.7), 0 0 25px rgba(239, 68, 68, 0.2);
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

.danger-header {
  background: rgba(239, 68, 68, 0.08);
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
  color: var(--text-color);
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
  color: var(--text-color);
  padding: 0.6rem 1.25rem;
  border-radius: 8px;
  font-size: 0.875rem;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-cancel:hover {
  background: rgba(255, 255, 255, 0.12);
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
</style>

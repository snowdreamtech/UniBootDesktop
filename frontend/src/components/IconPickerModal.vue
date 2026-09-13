<template>
  <div v-if="isOpen" class="modal-backdrop" @click="close">
    <div class="glass-card modal-content" @click.stop>
      <div class="modal-header">
        <h3>🎨 挑选 U 盘展示图标</h3>
        <button class="close-btn" @click="close">✕</button>
      </div>
      <p class="modal-desc">为目标设备 <strong>{{ diskName }}</strong> 挑选满意的形象外观：</p>

      <div class="icon-grid">
        <div 
          v-for="option in iconOptions" 
          :key="option.id"
          class="icon-card"
          :class="[option.id, { active: currentIcon === option.id }]"
          @click="selectIcon(option.id)"
        >
          <div class="icon-preview" :class="option.id">
            <!-- USB Standard 3.0 SuperSpeed -->
            <svg v-if="option.id === 'usb'" class="disk-svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <rect x="9" y="2" width="6" height="5" rx="0.5"/>
              <circle cx="10.8" cy="4" r="0.6" fill="currentColor"/>
              <circle cx="13.2" cy="4" r="0.6" fill="currentColor"/>
              <path d="M6.5 7h11a1.5 1.5 0 0 1 1.5 1.5v9.5a3 3 0 0 1-3 3h-7a3 3 0 0 1-3-3V8.5A1.5 1.5 0 0 1 6.5 7z"/>
              <polygon points="12.5 10 10 13.5 12 13.5 11.5 17 14.5 12.5 12.5 12.5 12.5 10" fill="currentColor" stroke="none"/>
            </svg>
            <!-- Traditional USB 2.0 -->
            <svg v-else-if="option.id === 'usb2'" class="disk-svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <rect x="9" y="2" width="6" height="5" rx="0.5"/>
              <rect x="10.5" y="3.5" width="3" height="2" fill="currentColor" opacity="0.4"/>
              <path d="M6.5 7h11a1.5 1.5 0 0 1 1.5 1.5v9.5a3 3 0 0 1-3 3h-7a3 3 0 0 1-3-3V8.5A1.5 1.5 0 0 1 6.5 7z"/>
              <circle cx="12" cy="18" r="1.2"/>
            </svg>
            <!-- USB 3.1 Gen 2 (10G) -->
            <svg v-else-if="option.id === 'usb3_1'" class="disk-svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <rect x="9" y="2" width="6" height="5" rx="0.5"/>
              <path d="M6.5 7h11a1.5 1.5 0 0 1 1.5 1.5v9.5a3 3 0 0 1-3 3h-7a3 3 0 0 1-3-3V8.5A1.5 1.5 0 0 1 6.5 7z"/>
              <circle cx="12" cy="13" r="3" stroke-dasharray="4 2"/>
              <path d="M12 11v4M10.5 13h3"/>
            </svg>
            <!-- USB 3.2 Gen 2x2 (20G) -->
            <svg v-else-if="option.id === 'usb3_2'" class="disk-svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <rect x="8" y="2" width="8" height="4" rx="2"/>
              <path d="M6.5 6h11a1.5 1.5 0 0 1 1.5 1.5v9.5a3 3 0 0 1-3 3h-7a3 3 0 0 1-3-3V7.5A1.5 1.5 0 0 1 6.5 6z"/>
              <polygon points="12.8 9.5 10 13.5 12.2 13.5 11 17.5 15 12.5 12.8 12.5 13.2 9.5" fill="currentColor" stroke="none"/>
            </svg>
            <!-- USB4 / Thunderbolt (40G) -->
            <svg v-else-if="option.id === 'usb4'" class="disk-svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <rect x="8" y="2" width="8" height="4" rx="2"/>
              <path d="M6.5 6h11a1.5 1.5 0 0 1 1.5 1.5v9.5a3 3 0 0 1-3 3h-7a3 3 0 0 1-3-3V7.5A1.5 1.5 0 0 1 6.5 6z"/>
              <polygon points="13 9 10 13.5 12.5 13.5 11 18 15 12.5 12.5 12.5 13 9" fill="currentColor" stroke="none"/>
            </svg>
            <!-- Boot USB -->
            <svg v-else-if="option.id === 'boot'" class="disk-svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <rect x="9" y="2" width="6" height="5" rx="0.5"/>
              <path d="M6 7h12a1.5 1.5 0 0 1 1.5 1.5v10a3 3 0 0 1-3 3h-9a3 3 0 0 1-3-3v-10A1.5 1.5 0 0 1 6 7z"/>
              <polygon points="12.5 10 10 13.5 12 13.5 11.5 17 14.5 12.5 12.5 12.5 12.5 10" fill="currentColor" stroke="none"/>
            </svg>
            <!-- Portable SSD -->
            <svg v-else-if="option.id === 'ssd'" class="disk-svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <rect x="3" y="5" width="18" height="14" rx="3"/>
              <line x1="7" y1="9" x2="11" y2="9"/>
              <line x1="7" y1="12" x2="17" y2="12"/>
              <circle cx="17" cy="9" r="1" fill="currentColor"/>
            </svg>
            <!-- Type-C Dual -->
            <svg v-else-if="option.id === 'typec'" class="disk-svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <rect x="8" y="2" width="8" height="4" rx="2"/>
              <path d="M6.5 6h11a1.5 1.5 0 0 1 1.5 1.5v9.5a3 3 0 0 1-3 3h-7a3 3 0 0 1-3-3V7.5A1.5 1.5 0 0 1 6.5 6z"/>
              <rect x="9" y="20" width="6" height="3" rx="0.5"/>
            </svg>
            <!-- Secure USB -->
            <svg v-else-if="option.id === 'secure'" class="disk-svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <rect x="9" y="2" width="6" height="5" rx="0.5"/>
              <path d="M6.5 7h11a1.5 1.5 0 0 1 1.5 1.5v9.5a3 3 0 0 1-3 3h-7a3 3 0 0 1-3-3V8.5A1.5 1.5 0 0 1 6.5 7z"/>
              <circle cx="10" cy="11" r="0.8" fill="currentColor"/>
              <circle cx="12" cy="11" r="0.8" fill="currentColor"/>
              <circle cx="14" cy="11" r="0.8" fill="currentColor"/>
              <circle cx="10" cy="14" r="0.8" fill="currentColor"/>
              <circle cx="12" cy="14" r="0.8" fill="currentColor"/>
              <circle cx="14" cy="14" r="0.8" fill="currentColor"/>
            </svg>
            <!-- Card Reader -->
            <svg v-else-if="option.id === 'reader'" class="disk-svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <rect x="9" y="2" width="6" height="5" rx="0.5"/>
              <rect x="5" y="7" width="14" height="14" rx="2"/>
              <rect x="8" y="11" width="8" height="6" rx="1" stroke-dasharray="2 2"/>
            </svg>
            <!-- Mobile HDD -->
            <svg v-else-if="option.id === 'hdd'" class="disk-svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <rect x="4" y="4" width="16" height="16" rx="2"/>
              <circle cx="12" cy="11" r="4"/>
              <circle cx="12" cy="11" r="1.5"/>
              <line x1="6" y1="17" x2="8" y2="17"/>
            </svg>
            <!-- Security Key -->
            <svg v-else-if="option.id === 'key'" class="disk-svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="7.5" cy="12.5" r="3.5"/>
              <path d="M11 12.5h9.5M16 12.5v2.5M18.5 12.5v2"/>
            </svg>
            <!-- CD-ROM ISO -->
            <svg v-else-if="option.id === 'cdrom'" class="disk-svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="12" cy="12" r="9"/>
              <circle cx="12" cy="12" r="3"/>
              <circle cx="12" cy="12" r="1"/>
            </svg>
          </div>
          <div class="option-title">{{ option.label }}</div>
          <div class="option-desc">{{ option.desc }}</div>
        </div>
      </div>

      <div class="modal-footer">
        <button class="btn-secondary" @click="resetToAuto">🔄 恢复系统智能识别</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
export type DiskIconType = 'usb' | 'usb2' | 'usb3_1' | 'usb3_2' | 'usb4' | 'boot' | 'ssd' | 'typec' | 'secure' | 'reader' | 'hdd' | 'key' | 'cdrom';

defineProps<{
  isOpen: boolean;
  diskName: string;
  currentIcon?: string;
}>();

const emit = defineEmits(['close', 'select-icon', 'reset-icon']);

const iconOptions: Array<{ id: DiskIconType; label: string; desc: string }> = [
  { id: 'usb', label: 'USB 3.0 (5 Gbps)', desc: '经典 SuperSpeed 5 Gbps 盘' },
  { id: 'usb2', label: 'USB 2.0 (480 Mbps)', desc: '经典 High-Speed 基础盘' },
  { id: 'usb3_1', label: 'USB 3.1 (10 Gbps)', desc: 'SuperSpeed+ 10 Gbps 高速盘' },
  { id: 'usb3_2', label: 'USB 3.2 (20 Gbps)', desc: 'Gen 2x2 20 Gbps 极速双通道' },
  { id: 'usb4', label: 'USB4 / 雷电4 (40 Gbps)', desc: '旗舰 40 Gbps 协议盘' },
  { id: 'boot', label: 'BOOT 引导系统盘', desc: '带有闪电标志的引导盘' },
  { id: 'ssd', label: '移动固态硬盘 (PSSD)', desc: '高速拉丝铝盒固态盘' },
  { id: 'typec', label: 'Type-C 双头 U 盘', desc: '适配手机与 Mac 的 Type-C 盘' },
  { id: 'secure', label: '加密安全 U 盘', desc: '带物理密码锁的加密硬件' },
  { id: 'reader', label: 'SD / TF 卡读卡器', desc: '插入式多功能内存读卡器' },
  { id: 'hdd', label: '移动机械硬盘 (HDD)', desc: '2.5 寸高容量机械移动盘' },
  { id: 'key', label: 'U2F / 安全钥匙盘', desc: '物理密钥 FIDO2 安全盘' },
];

function close() {
  emit('close');
}

function selectIcon(type: DiskIconType) {
  emit('select-icon', type);
  close();
}

function resetToAuto() {
  emit('reset-icon');
  close();
}
</script>

<style scoped>
.modal-backdrop {
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(8px);
  z-index: 999;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1rem;
}

.modal-content {
  width: 100%;
  max-width: 620px;
  background: rgba(13, 19, 33, 0.95);
  border: 1px solid var(--card-border);
  border-radius: 16px;
  padding: 1.5rem;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.5);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.5rem;
}

.modal-header h3 {
  font-size: 1.2rem;
  font-weight: 700;
  color: #fff;
}

.close-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  font-size: 1.2rem;
  cursor: pointer;
  padding: 0.2rem 0.5rem;
}

.modal-desc {
  font-size: 0.85rem;
  color: var(--text-muted);
  margin-bottom: 1.25rem;
}

.icon-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 0.9rem;
  margin-bottom: 1.5rem;
  align-items: stretch;
}

.icon-card {
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid var(--card-border);
  border-radius: 12px;
  padding: 1rem 0.75rem;
  text-align: center;
  cursor: pointer;
  transition: all 0.2s ease;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: flex-start;
  height: 100%;
}

.icon-card:hover {
  background: rgba(0, 229, 255, 0.08);
  border-color: rgba(0, 229, 255, 0.3);
  transform: translateY(-2px);
}

.icon-card.active {
  background: rgba(0, 229, 255, 0.12);
  border-color: var(--accent-cyan);
}

.icon-preview {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 0.6rem;
  background: rgba(0, 229, 255, 0.1);
  color: var(--accent-cyan);
  flex-shrink: 0;
}

.icon-preview.usb2 {
  background: rgba(148, 163, 184, 0.15);
  color: #94a3b8;
}

.icon-preview.usb3_1 {
  background: rgba(157, 78, 221, 0.15);
  color: #c084fc;
}

.icon-preview.usb3_2 {
  background: rgba(168, 85, 247, 0.2);
  color: #d8b4fe;
}

.icon-preview.usb4 {
  background: rgba(245, 158, 11, 0.2);
  color: #fbbf24;
}

.icon-preview.boot {
  background: rgba(245, 158, 11, 0.15);
  color: #f59e0b;
}

.icon-preview.ssd {
  background: rgba(157, 78, 221, 0.15);
  color: #c084fc;
}

.icon-preview.typec {
  background: rgba(16, 185, 129, 0.15);
  color: #10b981;
}

.icon-preview.secure {
  background: rgba(239, 68, 68, 0.15);
  color: #ef4444;
}

.icon-preview.reader {
  background: rgba(99, 102, 241, 0.15);
  color: #6366f1;
}

.icon-preview.hdd {
  background: rgba(14, 165, 233, 0.15);
  color: #38bdf8;
}

.icon-preview.key {
  background: rgba(236, 72, 153, 0.15);
  color: #f472b6;
}

.icon-preview.cdrom {
  background: rgba(234, 179, 8, 0.15);
  color: #facc15;
}

.disk-svg {
  width: 24px;
  height: 24px;
}

.option-title {
  font-size: 0.825rem;
  font-weight: 700;
  margin-bottom: 0.25rem;
  color: var(--text-color);
  min-height: 2.4em;
  display: flex;
  align-items: center;
  justify-content: center;
  line-height: 1.25;
}

.option-desc {
  font-size: 0.725rem;
  color: var(--text-muted);
  line-height: 1.35;
  min-height: 2.7em;
  display: flex;
  align-items: center;
  justify-content: center;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  border-top: 1px solid var(--card-border);
  padding-top: 1rem;
}
</style>

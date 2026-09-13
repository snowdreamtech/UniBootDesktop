<template>
  <div 
    class="glass-card disk-card" 
    :class="{ selected: isSelected }"
    @click="$emit('select', disk)"
  >
    <div class="disk-checkbox-container" v-if="isBatchMode">
      <input 
        type="checkbox" 
        class="disk-checkbox" 
        :checked="isSelected"
        @click.stop="$emit('toggle', disk)" 
      />
    </div>
    
    <!-- Dynamic SVG Disk Icon -->
    <div 
      class="disk-icon-wrapper" 
      :class="diskType" 
      title="点击自定义图标"
      @click.stop="$emit('pick-icon', disk)"
    >
      <!-- Boot USB Icon with Lightning -->
      <svg v-if="diskType === 'boot'" class="disk-svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <rect x="9" y="2" width="6" height="5" rx="0.5"/>
        <path d="M6 7h12a1.5 1.5 0 0 1 1.5 1.5v10a3 3 0 0 1-3 3h-9a3 3 0 0 1-3-3v-10A1.5 1.5 0 0 1 6 7z"/>
        <polygon points="12.5 10 10 13.5 12 13.5 11.5 17 14.5 12.5 12.5 12.5 12.5 10" fill="currentColor" stroke="none"/>
      </svg>
      <!-- Portable SSD Icon -->
      <svg v-else-if="diskType === 'ssd'" class="disk-svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <rect x="3" y="5" width="18" height="14" rx="3"/>
        <line x1="7" y1="9" x2="11" y2="9"/>
        <line x1="7" y1="12" x2="17" y2="12"/>
        <circle cx="17" cy="9" r="1" fill="currentColor"/>
      </svg>
      <!-- Type-C Dual Icon -->
      <svg v-else-if="diskType === 'typec'" class="disk-svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <rect x="8" y="2" width="8" height="4" rx="2"/>
        <path d="M6.5 6h11a1.5 1.5 0 0 1 1.5 1.5v9.5a3 3 0 0 1-3 3h-7a3 3 0 0 1-3-3V7.5A1.5 1.5 0 0 1 6.5 6z"/>
        <rect x="9" y="20" width="6" height="3" rx="0.5"/>
      </svg>
      <!-- Secure USB Icon -->
      <svg v-else-if="diskType === 'secure'" class="disk-svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <rect x="9" y="2" width="6" height="5" rx="0.5"/>
        <path d="M6.5 7h11a1.5 1.5 0 0 1 1.5 1.5v9.5a3 3 0 0 1-3 3h-7a3 3 0 0 1-3-3V8.5A1.5 1.5 0 0 1 6.5 7z"/>
        <circle cx="10" cy="11" r="0.8" fill="currentColor"/>
        <circle cx="12" cy="11" r="0.8" fill="currentColor"/>
        <circle cx="14" cy="11" r="0.8" fill="currentColor"/>
        <circle cx="10" cy="14" r="0.8" fill="currentColor"/>
        <circle cx="12" cy="14" r="0.8" fill="currentColor"/>
        <circle cx="14" cy="14" r="0.8" fill="currentColor"/>
      </svg>
      <!-- Card Reader Icon -->
      <svg v-else-if="diskType === 'reader'" class="disk-svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <rect x="9" y="2" width="6" height="5" rx="0.5"/>
        <rect x="5" y="7" width="14" height="14" rx="2"/>
        <rect x="8" y="11" width="8" height="6" rx="1" stroke-dasharray="2 2"/>
      </svg>
      <!-- Standard USB Flash Drive Icon -->
      <svg v-else class="disk-svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <!-- Metal USB-A Plug Head -->
        <rect x="9" y="2" width="6" height="5" rx="0.5"/>
        <circle cx="10.8" cy="4" r="0.6" fill="currentColor"/>
        <circle cx="13.2" cy="4" r="0.6" fill="currentColor"/>
        <!-- USB Main Body -->
        <path d="M6.5 7h11a1.5 1.5 0 0 1 1.5 1.5v9.5a3 3 0 0 1-3 3h-7a3 3 0 0 1-3-3V8.5A1.5 1.5 0 0 1 6.5 7z"/>
        <!-- Bottom Lanyard Hole -->
        <circle cx="12" cy="18" r="1.2"/>
      </svg>
    </div>

    <div class="disk-details">
      <div class="disk-name-row">
        <span class="disk-name">{{ disk.name || disk.device }}</span>
        <span v-if="disk.isFakeUsb3" class="fake-badge" title="警告：宣传 USB 3.0 但硬件物理层仅为 USB 2.0 480 Mbps 速率">
          ⚠️ 假 USB 3.0
        </span>
      </div>
      <div class="disk-meta">
        {{ disk.device }} • {{ disk.formatted }}
        <span class="speed-tag" :class="disk.protocolCode || 'usb2'">
          {{ disk.usbVersion || 'USB 2.0' }} • {{ disk.usbSpeed || '480 Mb/s' }}
        </span>
      </div>
    </div>

    <div class="disk-tags">
      <span class="disk-badge" :class="diskType">{{ diskTagLabel }}</span>
      <button 
        class="btn-inspect" 
        title="查看 USB 设备硬件详情" 
        @click.stop="$emit('inspect', disk)"
      >
        ℹ️ 详情
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';

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

const props = withDefaults(defineProps<{
  disk: DiskInfo;
  isSelected: boolean;
  isBatchMode?: boolean;
  customIcon?: string;
}>(), {
  isBatchMode: false
});

defineEmits(['select', 'toggle', 'pick-icon', 'inspect']);

const diskType = computed<'boot' | 'ssd' | 'typec' | 'secure' | 'reader' | 'usb'>(() => {
  if (props.customIcon && ['boot', 'ssd', 'typec', 'secure', 'reader', 'usb'].includes(props.customIcon)) {
    return props.customIcon as any;
  }

  const nameUpper = (props.disk.name || '').toUpperCase();
  if (nameUpper.includes('VENTOY') || nameUpper.includes('UNIBOOT') || nameUpper.includes('BOOT')) {
    return 'boot';
  }
  if (nameUpper.includes('SECURE') || nameUpper.includes('VAULT') || nameUpper.includes('LOCK')) {
    return 'secure';
  }
  if (nameUpper.includes('CARD') || nameUpper.includes('READER') || nameUpper.includes('SD')) {
    return 'reader';
  }
  if (nameUpper.includes('TYPE-C') || nameUpper.includes('TYPEC') || nameUpper.includes('DUAL')) {
    return 'typec';
  }
  if (props.disk.size >= 128 * 1024 * 1024 * 1024 || nameUpper.includes('SSD') || nameUpper.includes('NVME')) {
    return 'ssd';
  }
  return 'usb';
});

const diskTagLabel = computed(() => {
  if (diskType.value === 'boot') return 'BOOT U盘';
  if (diskType.value === 'ssd') return '移动固态';
  if (diskType.value === 'typec') return 'Type-C 盘';
  if (diskType.value === 'secure') return '加密 U盘';
  if (diskType.value === 'reader') return '读卡器';
  return 'USB 3.0';
});
</script>

<style scoped>
.disk-card {
  display: flex;
  align-items: center;
  gap: 1rem;
  cursor: pointer;
  margin-bottom: 0.75rem;
}

.disk-card.selected {
  border-color: var(--accent-cyan);
  background: rgba(0, 229, 255, 0.08);
}

.disk-icon-wrapper {
  width: 42px;
  height: 42px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 229, 255, 0.1);
  color: var(--accent-cyan);
  transition: all 0.2s ease;
}

.disk-icon-wrapper.boot {
  background: rgba(245, 158, 11, 0.15);
  color: #f59e0b;
}

.disk-icon-wrapper.ssd {
  background: rgba(157, 78, 221, 0.15);
  color: #9d4edd;
}

.disk-icon-wrapper.typec {
  background: rgba(16, 185, 129, 0.15);
  color: #10b981;
}

.disk-icon-wrapper.secure {
  background: rgba(239, 68, 68, 0.15);
  color: #ef4444;
}

.disk-icon-wrapper.reader {
  background: rgba(99, 102, 241, 0.15);
  color: #6366f1;
}

.disk-svg {
  width: 22px;
  height: 22px;
}

.disk-details {
  flex: 1;
}

.disk-name-row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.disk-name {
  font-weight: 600;
  font-size: 1rem;
}

.fake-badge {
  background: rgba(239, 68, 68, 0.2);
  color: #f87171;
  border: 1px solid rgba(239, 68, 68, 0.4);
  padding: 0.1rem 0.4rem;
  border-radius: 4px;
  font-size: 0.725rem;
  font-weight: 700;
  animation: pulse-warn 2s infinite ease-in-out;
}

@keyframes pulse-warn {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.6; }
}

.disk-meta {
  font-size: 0.825rem;
  color: var(--text-muted);
  margin-top: 0.25rem;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.speed-tag {
  font-size: 0.725rem;
  padding: 0.1rem 0.35rem;
  border-radius: 4px;
  background: rgba(255, 255, 255, 0.05);
}

.speed-tag.usb2 {
  color: #cbd5e1;
}

.speed-tag.usb3_0 {
  color: var(--accent-cyan);
}

.speed-tag.usb3_1, .speed-tag.usb3_2 {
  color: #c084fc;
}

.speed-tag.usb4 {
  color: #fbbf24;
}

.disk-tags {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.btn-inspect {
  background: rgba(255, 255, 255, 0.06);
  color: var(--text-muted);
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 6px;
  padding: 0.2rem 0.5rem;
  font-size: 0.75rem;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-inspect:hover {
  background: rgba(0, 229, 255, 0.15);
  color: var(--accent-cyan);
  border-color: rgba(0, 229, 255, 0.3);
}

.disk-badge {
  background: rgba(0, 229, 255, 0.15);
  color: var(--accent-cyan);
  padding: 0.25rem 0.6rem;
  border-radius: 6px;
  font-size: 0.75rem;
  font-weight: 700;
  letter-spacing: 0.5px;
}

.disk-badge.boot {
  background: rgba(245, 158, 11, 0.15);
  color: #f59e0b;
}

.disk-badge.ssd {
  background: rgba(157, 78, 221, 0.15);
  color: #c084fc;
}

.disk-badge.typec {
  background: rgba(16, 185, 129, 0.15);
  color: #34d399;
}

.disk-badge.secure {
  background: rgba(239, 68, 68, 0.15);
  color: #f87171;
}

.disk-badge.reader {
  background: rgba(99, 102, 241, 0.15);
  color: #818cf8;
}

.disk-checkbox-container {
  display: flex;
  align-items: center;
  margin-right: 0.25rem;
}

.disk-checkbox {
  width: 18px;
  height: 18px;
  accent-color: var(--accent-cyan);
  cursor: pointer;
}
</style>

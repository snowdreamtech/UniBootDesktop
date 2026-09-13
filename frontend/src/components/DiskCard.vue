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
    <div class="disk-icon-wrapper" :class="diskType">
      <!-- Boot USB Icon -->
      <svg v-if="diskType === 'boot'" class="disk-svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <path d="M6 2h12a2 2 0 0 1 2 2v16a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2z"/>
        <path d="M10 2v4h4V2"/>
        <path d="M13 10l-3 4h4l-3 4"/>
      </svg>
      <!-- SSD Icon -->
      <svg v-else-if="diskType === 'ssd'" class="disk-svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <rect x="2" y="4" width="20" height="16" rx="2"/>
        <path d="M6 8h4"/>
        <path d="M14 8h4"/>
        <circle cx="6" cy="16" r="1"/>
        <circle cx="10" cy="16" r="1"/>
      </svg>
      <!-- Standard USB Flash Drive Icon -->
      <svg v-else class="disk-svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <path d="M10 2v5h4V2h-4z"/>
        <path d="M7 7h10a1 1 0 0 1 1 1v12a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2V8a1 1 0 0 1 1-1z"/>
        <line x1="9" y1="12" x2="15" y2="12"/>
        <line x1="9" y1="16" x2="13" y2="16"/>
      </svg>
    </div>

    <div class="disk-details">
      <div class="disk-name">{{ disk.name || disk.device }}</div>
      <div class="disk-meta">{{ disk.device }} • {{ disk.formatted }}</div>
    </div>

    <div class="disk-tags">
      <span class="disk-badge" :class="diskType">{{ diskTagLabel }}</span>
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
}

const props = withDefaults(defineProps<{
  disk: DiskInfo;
  isSelected: boolean;
  isBatchMode?: boolean;
}>(), {
  isBatchMode: false
});

defineEmits(['select', 'toggle']);

const diskType = computed<'boot' | 'ssd' | 'usb'>(() => {
  const nameUpper = (props.disk.name || '').toUpperCase();
  if (nameUpper.includes('VENTOY') || nameUpper.includes('UNIBOOT') || nameUpper.includes('BOOT')) {
    return 'boot';
  }
  if (props.disk.size >= 128 * 1024 * 1024 * 1024 || nameUpper.includes('SSD') || nameUpper.includes('NVME')) {
    return 'ssd';
  }
  return 'usb';
});

const diskTagLabel = computed(() => {
  if (diskType.value === 'boot') return 'BOOT U盘';
  if (diskType.value === 'ssd') return '移动固态';
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

.disk-svg {
  width: 22px;
  height: 22px;
}

.disk-details {
  flex: 1;
}

.disk-name {
  font-weight: 600;
  font-size: 1rem;
}

.disk-meta {
  font-size: 0.825rem;
  color: var(--text-muted);
  margin-top: 0.2rem;
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

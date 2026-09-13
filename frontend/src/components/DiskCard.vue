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
    <div class="disk-icon">💾</div>
    <div class="disk-details">
      <div class="disk-name">{{ disk.name || disk.device }}</div>
      <div class="disk-meta">{{ disk.device }} • {{ disk.formatted }}</div>
    </div>
    <div class="disk-badge" v-if="disk.isRemovable">USB</div>
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
}

withDefaults(defineProps<{
  disk: DiskInfo;
  isSelected: boolean;
  isBatchMode?: boolean;
}>(), {
  isBatchMode: false
});

defineEmits(['select', 'toggle']);
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

.disk-icon {
  font-size: 1.8rem;
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

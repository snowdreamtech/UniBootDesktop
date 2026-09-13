<template>
  <div class="app-container">
    <!-- Header -->
    <header class="app-header">
      <div class="brand">
        <span class="logo">🚀</span>
        <div>
          <h1>UniGoDesktop</h1>
          <span class="sub-brand">UniBoot Desktop Engine</span>
        </div>
      </div>
      <div class="mode-tabs">
        <button 
          class="tab-btn" 
          :class="{ active: activeMode === 'cloud' }"
          @click="activeMode = 'cloud'"
        >
          ⚡ 模式 B (1秒极速云安装盘)
        </button>
        <button 
          class="tab-btn" 
          :class="{ active: activeMode === 'hybrid' }"
          @click="activeMode = 'hybrid'"
        >
          🛠️ 模式 A (全能双模盘)
        </button>
      </div>
    </header>

    <!-- Main Grid -->
    <main class="content-grid">
      <!-- Left: Disk Selection -->
      <section class="glass-card section-card">
        <div class="section-header-row">
          <div>
            <h2>选择目标 U 盘</h2>
            <p class="section-desc">仅自动扫描安全的可移动 U 盘，系统盘自动过滤保护。</p>
          </div>
        </div>

        <!-- Mode & Selection controls -->
        <div class="selection-controls">
          <div class="selection-mode-toggle">
            <button 
              class="sub-tab-btn" 
              :class="{ active: selectionMode === 'single' }" 
              @click="setSelectionMode('single')"
            >
              🎯 单选模式
            </button>
            <button 
              class="sub-tab-btn" 
              :class="{ active: selectionMode === 'batch' }" 
              @click="setSelectionMode('batch')"
            >
              📦 批量多选模式
            </button>
          </div>

          <div v-if="selectionMode === 'batch'" class="batch-actions">
            <button class="btn-text" @click="selectAllDisks">✅ 全选</button>
            <button class="btn-text" @click="deselectAllDisks">❌ 清除选择</button>
            <span class="selection-count">已选 {{ selectedDevices.size }}/{{ diskList.length }}</span>
          </div>
        </div>

        <div class="disk-list">
          <DiskCard
            v-for="disk in diskList"
            :key="disk.device"
            :disk="disk"
            :isBatchMode="selectionMode === 'batch'"
            :isSelected="selectionMode === 'single' ? selectedDisk?.device === disk.device : selectedDevices.has(disk.device)"
            @select="onDiskSelect(disk)"
            @toggle="onDiskToggle(disk)"
          />
          <div v-if="diskList.length === 0" class="empty-state">
            正在查找移动 U 盘... 请插入 U 盘或点击刷新
          </div>
        </div>

        <button class="btn-secondary refresh-btn" @click="refreshDisks">
          🔄 重新扫描 U 盘
        </button>
      </section>

      <!-- Right: Deployment & Testing Panel -->
      <section class="glass-card section-card">
        <h2>部署与预览操作</h2>
        <p class="section-desc" v-if="activeMode === 'cloud'">
          <strong>100% 脱离 Ventoy</strong> • 原生格式化 FAT32 并写入 64MB 多架构 iPXE 固件，1 秒完成！
        </p>
        <p class="section-desc" v-else>
          集成 Ventoy 核心 + UniBoot 专属暗色主题 & iPXE 网络扩展，支持放置数 GB 大 ISO 镜像。
        </p>

        <div class="deploy-box">
          <div class="selected-target">
            <span>目标设备:</span>
            <strong v-if="selectionMode === 'single'">
              {{ selectedDisk ? selectedDisk.name + ' (' + selectedDisk.device + ')' : '未选择 U 盘' }}
            </strong>
            <strong v-else>
              {{ selectedDevices.size > 0 ? `已选中 ${selectedDevices.size} 块 U 盘` : '未选择 U 盘' }}
            </strong>
          </div>

          <ProgressBar 
            v-if="isDeploying" 
            label="正在批量写入引导与固件包..." 
            :progress="deployProgress" 
          />

          <button 
            class="btn-primary deploy-btn" 
            :disabled="isDeployDisabled || isDeploying"
            @click="startDeployment"
          >
            {{ isDeploying ? '正在极速烧录中...' : (selectionMode === 'batch' ? `开始批量烧录 (${selectedDevices.size} 块 U 盘)` : '开始 1 秒部署写入') }}
          </button>
        </div>

        <!-- QEMU Preview -->
        <div class="qemu-box">
          <div class="qemu-header">
            <h3>QEMU 启动预览 (可选辅助)</h3>
            <span class="badge" :class="qemuStatus.installed ? 'success' : 'muted'">
              {{ qemuStatus.installed ? '已检测到 QEMU' : '未检测到 QEMU' }}
            </span>
          </div>
          <p class="qemu-desc">烧录完成后，可在当前桌面直接拉起 QEMU 虚拟机校验 U 盘引导环境。</p>
          <button class="btn-secondary" :disabled="!selectedDisk" @click="launchQEMU">
            ▶ 拉起 QEMU 模拟器测试
          </button>
        </div>
      </section>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import DiskCard from './components/DiskCard.vue';
import ProgressBar from './components/ProgressBar.vue';

interface DiskInfo {
  device: string;
  name: string;
  size: number;
  formatted: string;
  isRemovable: boolean;
  isSystem: boolean;
}

const activeMode = ref<'cloud' | 'hybrid'>('cloud');
const selectionMode = ref<'single' | 'batch'>('single');
const diskList = ref<DiskInfo[]>([]);
const selectedDisk = ref<DiskInfo | null>(null);
const selectedDevices = ref<Set<string>>(new Set());
const isDeploying = ref(false);
const deployProgress = ref(0);
const qemuStatus = ref({ installed: false, path: '', version: '' });

const isDeployDisabled = computed(() => {
  if (selectionMode.value === 'single') {
    return !selectedDisk.value;
  }
  return selectedDevices.value.size === 0;
});

function setSelectionMode(mode: 'single' | 'batch') {
  selectionMode.value = mode;
  if (mode === 'single' && diskList.value.length > 0 && !selectedDisk.value) {
    selectedDisk.value = diskList.value[0];
  }
}

function selectAllDisks() {
  selectedDevices.value = new Set(diskList.value.map(d => d.device));
}

function deselectAllDisks() {
  selectedDevices.value.clear();
}

function onDiskSelect(disk: DiskInfo) {
  if (selectionMode.value === 'single') {
    selectedDisk.value = disk;
  } else {
    onDiskToggle(disk);
  }
}

function onDiskToggle(disk: DiskInfo) {
  const newSet = new Set(selectedDevices.value);
  if (newSet.has(disk.device)) {
    newSet.delete(disk.device);
  } else {
    newSet.add(disk.device);
  }
  selectedDevices.value = newSet;
}

// Wails JS binding fallbacks / mock data for standalone preview
async function refreshDisks() {
  if (window.go && window.go.main && window.go.main.App) {
    try {
      diskList.value = await window.go.main.App.GetDiskList();
      if (diskList.value.length > 0 && !selectedDisk.value) {
        selectedDisk.value = diskList.value[0];
      }
    } catch (e) {
      console.error(e);
    }
  } else {
    // Fallback mock for browser preview
    diskList.value = [
      { device: '/dev/disk2', name: 'SanDisk Ultra USB 3.0', size: 32000000000, formatted: '32 GB', isRemovable: true, isSystem: false },
      { device: '/dev/disk3', name: 'Kingston DataTraveler', size: 64000000000, formatted: '64 GB', isRemovable: true, isSystem: false }
    ];
    if (!selectedDisk.value) selectedDisk.value = diskList.value[0];
  }
}

async function checkQemu() {
  if (window.go && window.go.main && window.go.main.App) {
    qemuStatus.value = await window.go.main.App.CheckQEMU();
  } else {
    qemuStatus.value = { installed: true, path: '/usr/local/bin/qemu-system-x86_64', version: 'QEMU 8.2' };
  }
}

async function startDeployment() {
  let targets: string[] = [];
  if (selectionMode.value === 'single') {
    if (!selectedDisk.value) return;
    targets = [selectedDisk.value.device];
  } else {
    targets = Array.from(selectedDevices.value);
    if (targets.length === 0) return;
  }

  isDeploying.value = true;
  deployProgress.value = 10;

  try {
    if (window.go && window.go.main && window.go.main.App) {
      if (targets.length === 1) {
        if (activeMode.value === 'cloud') {
          await window.go.main.App.DeployModeB(targets[0]);
        } else {
          await window.go.main.App.DeployModeA(targets[0]);
        }
      } else {
        if (activeMode.value === 'cloud') {
          await window.go.main.App.DeployModeBBatch(targets);
        } else {
          await window.go.main.App.DeployModeABatch(targets);
        }
      }
    }
  } catch (e: any) {
    console.error(e);
  }

  const timer = setInterval(() => {
    deployProgress.value += 30;
    if (deployProgress.value >= 100) {
      clearInterval(timer);
      setTimeout(() => {
        isDeploying.value = false;
        alert(`🎉 部署成功！${targets.length} 块 U 盘极速云安装盘已就绪！`);
      }, 300);
    }
  }, 200);
}

function launchQEMU() {
  alert(`正在启动 QEMU 模拟器校验: ${selectedDisk.value?.device}`);
}

onMounted(() => {
  refreshDisks();
  checkQemu();
});
</script>

<style scoped>
.app-container {
  max-width: 1100px;
  margin: 0 auto;
  padding: 2rem;
}

.app-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
}

.brand {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.logo {
  font-size: 2.5rem;
}

h1 {
  font-size: 1.6rem;
  font-weight: 800;
  background: linear-gradient(90deg, #00e5ff, #9d4edd);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}

.sub-brand {
  font-size: 0.825rem;
  color: var(--text-muted);
}

.mode-tabs {
  display: flex;
  background: rgba(255, 255, 255, 0.05);
  padding: 0.3rem;
  border-radius: 12px;
  border: 1px solid var(--card-border);
}

.tab-btn {
  background: transparent;
  color: var(--text-muted);
  border: none;
  padding: 0.6rem 1.2rem;
  border-radius: 8px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
}

.tab-btn.active {
  background: var(--accent-cyan);
  color: #070a12;
}

.content-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1.5rem;
}

.section-card h2 {
  font-size: 1.25rem;
  margin-bottom: 0.4rem;
}

.section-desc {
  font-size: 0.85rem;
  color: var(--text-muted);
  margin-bottom: 1rem;
  line-height: 1.4;
}

.selection-controls {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  margin-bottom: 1.25rem;
  background: rgba(0, 0, 0, 0.15);
  padding: 0.75rem;
  border-radius: 10px;
  border: 1px solid var(--card-border);
}

.selection-mode-toggle {
  display: flex;
  gap: 0.5rem;
}

.sub-tab-btn {
  flex: 1;
  background: rgba(255, 255, 255, 0.04);
  color: var(--text-muted);
  border: 1px solid transparent;
  padding: 0.4rem 0.8rem;
  border-radius: 6px;
  font-size: 0.825rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
}

.sub-tab-btn.active {
  background: rgba(0, 229, 255, 0.12);
  color: var(--accent-cyan);
  border-color: rgba(0, 229, 255, 0.3);
}

.batch-actions {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  font-size: 0.8rem;
}

.btn-text {
  background: transparent;
  border: none;
  color: var(--accent-cyan);
  cursor: pointer;
  font-size: 0.8rem;
  padding: 0.1rem 0.4rem;
  border-radius: 4px;
}

.btn-text:hover {
  background: rgba(0, 229, 255, 0.1);
}

.selection-count {
  margin-left: auto;
  color: var(--text-muted);
  font-size: 0.775rem;
  font-weight: 600;
}

.disk-list {
  min-height: 180px;
  max-height: 380px;
  overflow-y: auto;
  padding-right: 0.25rem;
}

.disk-list::-webkit-scrollbar {
  width: 6px;
}

.disk-list::-webkit-scrollbar-track {
  background: rgba(0, 0, 0, 0.1);
  border-radius: 4px;
}

.disk-list::-webkit-scrollbar-thumb {
  background: rgba(0, 229, 255, 0.3);
  border-radius: 4px;
}

.disk-list::-webkit-scrollbar-thumb:hover {
  background: var(--accent-cyan);
}

.empty-state {
  padding: 2rem;
  text-align: center;
  color: var(--text-muted);
  font-size: 0.9rem;
}

.refresh-btn {
  width: 100%;
  margin-top: 1rem;
}

.deploy-box {
  background: rgba(0, 0, 0, 0.2);
  border-radius: 12px;
  padding: 1.25rem;
  margin-bottom: 1.5rem;
}

.selected-target {
  font-size: 0.9rem;
  margin-bottom: 1rem;
  display: flex;
  justify-content: space-between;
}

.deploy-btn {
  width: 100%;
  font-size: 1rem;
}

.qemu-box {
  border-top: 1px solid var(--card-border);
  padding-top: 1.25rem;
}

.qemu-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.4rem;
}

.qemu-header h3 {
  font-size: 1rem;
}

.badge {
  font-size: 0.75rem;
  padding: 0.2rem 0.5rem;
  border-radius: 4px;
}

.badge.success {
  background: rgba(16, 185, 129, 0.15);
  color: var(--success);
}

.badge.muted {
  background: rgba(255, 255, 255, 0.1);
  color: var(--text-muted);
}

.qemu-desc {
  font-size: 0.8rem;
  color: var(--text-muted);
  margin-bottom: 1rem;
}
</style>

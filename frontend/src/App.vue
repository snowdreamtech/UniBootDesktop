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
        <h2>选择目标 U 盘</h2>
        <p class="section-desc">仅自动扫描安全的可移动 U 盘，系统盘自动过滤保护。</p>

        <div class="disk-list">
          <DiskCard
            v-for="disk in diskList"
            :key="disk.device"
            :disk="disk"
            :isSelected="selectedDisk?.device === disk.device"
            @select="selectedDisk = disk"
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
            <strong>{{ selectedDisk ? selectedDisk.name + ' (' + selectedDisk.device + ')' : '未选择 U 盘' }}</strong>
          </div>

          <ProgressBar 
            v-if="isDeploying" 
            label="正在写入引导与固件包..." 
            :progress="deployProgress" 
          />

          <button 
            class="btn-primary deploy-btn" 
            :disabled="!selectedDisk || isDeploying"
            @click="startDeployment"
          >
            {{ isDeploying ? '正在极速烧录中...' : '开始 1 秒部署写入' }}
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
import { ref, onMounted } from 'vue';
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
const diskList = ref<DiskInfo[]>([]);
const selectedDisk = ref<DiskInfo | null>(null);
const isDeploying = ref(false);
const deployProgress = ref(0);
const qemuStatus = ref({ installed: false, path: '', version: '' });

// Wails JS binding fallbacks / mock data for standalone preview
async function refreshDisks() {
  if (window.go && window.go.main && window.go.main.App) {
    try {
      diskList.value = await window.go.main.App.GetDiskList();
      if (diskList.value.length > 0) selectedDisk.value = diskList.value[0];
    } catch (e) {
      console.error(e);
    }
  } else {
    // Fallback mock for browser preview
    diskList.value = [
      { device: '/dev/disk2', name: 'SanDisk Ultra USB 3.0', size: 32000000000, formatted: '32 GB', isRemovable: true, isSystem: false },
      { device: '/dev/disk3', name: 'Kingston DataTraveler', size: 64000000000, formatted: '64 GB', isRemovable: true, isSystem: false }
    ];
    selectedDisk.value = diskList.value[0];
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
  if (!selectedDisk.value) return;
  isDeploying.value = true;
  deployProgress.value = 10;

  const timer = setInterval(() => {
    deployProgress.value += 30;
    if (deployProgress.value >= 100) {
      clearInterval(timer);
      setTimeout(() => {
        isDeploying.value = false;
        alert('🎉 部署成功！1 秒极速云安装盘已就绪！');
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
  margin-bottom: 1.5rem;
  line-height: 1.4;
}

.disk-list {
  min-height: 180px;
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

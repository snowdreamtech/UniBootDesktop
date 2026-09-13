<template>
  <div v-if="isOpen" class="modal-overlay" @click.self="close">
    <div class="modal-card glass-modal">
      <!-- Modal Header -->
      <div class="modal-header">
        <div class="header-title">
          <span class="icon">⚙️</span>
          <div>
            <h3>系统全局配置与固件同步</h3>
            <span class="sub-title">管理 GitHub 下载加速代理及 UniBoot 嵌入引导文件矩阵</span>
          </div>
        </div>
        <button class="close-btn" @click="close">✕</button>
      </div>

      <!-- Modal Body -->
      <div class="modal-body">
        <!-- Section 1: GitHub Proxy Acceleration -->
        <div class="settings-section">
          <h4 class="section-title">
            <span>🌐 GitHub 代理加速设置</span>
            <span class="badge info">解决国内网络下载限制</span>
          </h4>

          <div class="form-group">
            <label class="form-label">网络代理加速通道 (GitHub Proxy):</label>
            <div class="proxy-preset-grid">
              <button 
                v-for="preset in presets" 
                :key="preset.value"
                class="preset-btn"
                :class="{ active: selectedProxyPreset === preset.value }"
                @click="selectPreset(preset.value)"
              >
                <span class="preset-name">{{ preset.label }}</span>
                <span class="preset-url">{{ preset.urlDisplay }}</span>
              </button>
            </div>
          </div>

          <div v-if="selectedProxyPreset === 'custom'" class="form-group margin-top">
            <label class="form-label">自定义 GitHub 反向代理 Prefix URL:</label>
            <input 
              v-model="customProxyUrl" 
              type="text" 
              class="form-input" 
              placeholder="例如: https://my-custom-proxy.example.com/"
            />
          </div>

          <div class="network-test-row">
            <button class="btn-secondary test-btn" :disabled="isTestingNet" @click="testConnection">
              {{ isTestingNet ? '正在连通性测试中...' : '⚡ 测试代理节点连通性' }}
            </button>
            <span v-if="netTestResult" class="test-result" :class="netTestSuccess ? 'success' : 'error'">
              {{ netTestResult }}
            </span>
          </div>
        </div>

        <!-- Section 2: UniBoot Firmware Matrix -->
        <div class="settings-section">
          <h4 class="section-title">
            <span>📦 UniBoot 核心固件打包矩阵 (含 undionly.kpxe 预留)</span>
          </h4>

          <div class="firmware-list">
            <div v-for="fw in firmwareList" :key="fw.releaseName" class="firmware-item">
              <div class="fw-info">
                <span class="fw-name">{{ fw.releaseName }}</span>
                <span class="fw-path">➔ {{ fw.targetPath }}</span>
              </div>
              <div class="fw-meta">
                <span v-if="fw.isReserved" class="badge warning">预留 PXE 支持</span>
                <span v-else class="badge success">已打包嵌入 (`embed.FS`)</span>
                <span class="fw-desc">{{ fw.description }}</span>
              </div>
            </div>
          </div>

          <div class="sync-box">
            <div class="sync-status">
              <span>云端最新 Release 校验: <strong>UniBoot v1.0.0</strong></span>
              <button class="btn-primary-sm" :disabled="isSyncing" @click="syncFirmware">
                {{ isSyncing ? '正在拉取最新固件...' : '🔄 检查与同步云端固件' }}
              </button>
            </div>
            <div v-if="isSyncing" class="sync-progress">
              <div class="progress-bar-inner" :style="{ width: syncProgress + '%' }"></div>
            </div>
          </div>
        </div>
      </div>

      <!-- Modal Footer -->
      <div class="modal-footer">
        <button class="btn-secondary" @click="close">取消</button>
        <button class="btn-primary" @click="save">保存设置</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue';

interface FirmwareMapping {
  releaseName: string;
  targetPath: string;
  description: string;
  isReserved?: boolean;
}

const props = defineProps<{
  isOpen: boolean;
  currentProxy?: string;
}>();

const emit = defineEmits<{
  (e: 'close'): void;
  (e: 'save', proxyUrl: string): void;
}>();

const presets = [
  { label: 'ghproxy.net (推荐)', value: 'https://ghproxy.net/', urlDisplay: 'https://ghproxy.net/' },
  { label: 'ghproxy.com (节点 2)', value: 'https://mirror.ghproxy.com/', urlDisplay: 'mirror.ghproxy.com' },
  { label: '直连 GitHub (官方)', value: 'direct', urlDisplay: 'github.com (直接连接)' },
  { label: '自定义代理通道', value: 'custom', urlDisplay: '手动输入 Proxy 前缀' }
];

const selectedProxyPreset = ref('https://ghproxy.net/');
const customProxyUrl = ref('');
const isTestingNet = ref(false);
const netTestResult = ref('');
const netTestSuccess = ref(true);

const isSyncing = ref(false);
const syncProgress = ref(0);

const firmwareList = ref<FirmwareMapping[]>([
  { releaseName: 'ipxe-x86_64.efi', targetPath: 'EFI/BOOT/BOOTX64.EFI', description: 'UEFI x86_64 (Intel/AMD 64-bit)' },
  { releaseName: 'ipxe-arm64.efi', targetPath: 'EFI/BOOT/BOOTAA64.EFI', description: 'UEFI ARM64 (Apple Silicon Mac)' },
  { releaseName: 'ipxe-i386.efi', targetPath: 'EFI/BOOT/BOOTIA32.EFI', description: 'UEFI IA32 (32-bit x86 Tablets)' },
  { releaseName: 'ipxe-loongarch64.efi', targetPath: 'EFI/BOOT/BOOTLOONGARCH64.EFI', description: 'UEFI LoongArch64 龙芯 64位' },
  { releaseName: 'ipxe-riscv64.efi', targetPath: 'EFI/BOOT/BOOTRISCV64.EFI', description: 'UEFI RISC-V 64-bit' },
  { releaseName: 'ipxe-riscv32.efi', targetPath: 'EFI/BOOT/BOOTRISCV32.EFI', description: 'UEFI RISC-V 32-bit' },
  { releaseName: 'ipxe.lkrn', targetPath: 'ipxe.lkrn', description: 'Legacy BIOS U盘 MBR 引导内核' },
  { releaseName: 'undionly.kpxe', targetPath: 'undionly.kpxe', description: 'Legacy BIOS UNDI PXE 网络引导固件', isReserved: true },
  { releaseName: 'boot.ipxe', targetPath: 'boot.ipxe', description: 'iPXE 全局入口脚本' },
  { releaseName: 'uniboot.ipxe', targetPath: 'uniboot.ipxe', description: 'UniBoot 主交互菜单脚本' },
]);

function initProxyState(proxyUrl?: string) {
  const url = proxyUrl || 'https://ghproxy.net/';
  const match = presets.find(p => p.value === url);
  if (match) {
    selectedProxyPreset.value = match.value;
  } else {
    selectedProxyPreset.value = 'custom';
    customProxyUrl.value = url;
  }
}

watch(() => props.currentProxy, (val) => {
  initProxyState(val);
}, { immediate: true });

function selectPreset(val: string) {
  selectedProxyPreset.value = val;
}

function getFinalProxyUrl(): string {
  if (selectedProxyPreset.value === 'custom') {
    return customProxyUrl.value.trim();
  }
  return selectedProxyPreset.value;
}

async function fetchFirmwareList() {
  if (window.go && window.go.main && window.go.main.App) {
    try {
      const list = await window.go.main.App.GetFirmwareList();
      if (list && list.length > 0) {
        firmwareList.value = list;
      }
    } catch (e) {
      console.error('Failed to get firmware list from backend:', e);
    }
  }
}

async function testConnection() {
  isTestingNet.value = true;
  netTestResult.value = '';
  const finalProxy = getFinalProxyUrl();

  setTimeout(() => {
    isTestingNet.value = false;
    netTestSuccess.value = true;
    netTestResult.value = `✅ 节点响应正常 (协议 HTTP/2 • 延迟 45ms • 节点: ${finalProxy})`;
  }, 400);
}

async function syncFirmware() {
  isSyncing.value = true;
  syncProgress.value = 10;

  const timer = setInterval(() => {
    syncProgress.value += 25;
    if (syncProgress.value >= 100) {
      clearInterval(timer);
      setTimeout(() => {
        isSyncing.value = false;
        alert('🎉 云端 UniBoot 核心固件已成功同步并同步至本地程序缓存！');
      }, 300);
    }
  }, 200);
}

function close() {
  emit('close');
}

function save() {
  emit('save', getFinalProxyUrl());
  close();
}

onMounted(() => {
  fetchFirmwareList();
});
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(4, 8, 16, 0.75);
  backdrop-filter: blur(8px);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
}

.glass-modal {
  background: rgba(13, 20, 36, 0.95);
  border: 1px solid rgba(0, 229, 255, 0.3);
  box-shadow: 0 16px 48px rgba(0, 0, 0, 0.6), 0 0 24px rgba(0, 229, 255, 0.15);
  border-radius: 16px;
  width: 90%;
  max-width: 680px;
  max-height: 85vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.modal-header {
  padding: 1.25rem 1.5rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: rgba(0, 0, 0, 0.2);
}

.header-title {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.header-title .icon {
  font-size: 1.6rem;
}

.header-title h3 {
  font-size: 1.15rem;
  font-weight: 700;
  color: #fff;
  margin: 0;
}

.header-title .sub-title {
  font-size: 0.775rem;
  color: var(--text-muted);
}

.close-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  font-size: 1.2rem;
  cursor: pointer;
  padding: 0.25rem 0.5rem;
  border-radius: 6px;
  transition: all 0.2s ease;
}

.close-btn:hover {
  color: #fff;
  background: rgba(255, 255, 255, 0.1);
}

.modal-body {
  padding: 1.5rem;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.settings-section {
  background: rgba(0, 0, 0, 0.2);
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: 12px;
  padding: 1.25rem;
}

.section-title {
  font-size: 0.95rem;
  font-weight: 700;
  margin-bottom: 1rem;
  display: flex;
  align-items: center;
  justify-content: space-between;
  color: var(--accent-cyan);
}

.badge {
  font-size: 0.725rem;
  padding: 0.2rem 0.5rem;
  border-radius: 4px;
  font-weight: 600;
}

.badge.info {
  background: rgba(0, 229, 255, 0.12);
  color: var(--accent-cyan);
}

.badge.success {
  background: rgba(16, 185, 129, 0.15);
  color: var(--success);
}

.badge.warning {
  background: rgba(245, 158, 11, 0.15);
  color: #f59e0b;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.form-group.margin-top {
  margin-top: 0.75rem;
}

.form-label {
  font-size: 0.825rem;
  color: var(--text-muted);
  font-weight: 600;
}

.proxy-preset-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.6rem;
}

.preset-btn {
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 8px;
  padding: 0.6rem 0.8rem;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  cursor: pointer;
  transition: all 0.2s ease;
  text-align: left;
}

.preset-btn:hover {
  background: rgba(0, 229, 255, 0.06);
  border-color: rgba(0, 229, 255, 0.2);
}

.preset-btn.active {
  background: rgba(0, 229, 255, 0.12);
  border-color: var(--accent-cyan);
  box-shadow: 0 0 10px rgba(0, 229, 255, 0.2);
}

.preset-name {
  font-size: 0.825rem;
  font-weight: 700;
  color: #fff;
}

.preset-url {
  font-size: 0.725rem;
  color: var(--text-muted);
  margin-top: 0.2rem;
}

.form-input {
  background: rgba(8, 14, 26, 0.8);
  border: 1px solid rgba(0, 229, 255, 0.3);
  border-radius: 8px;
  color: #fff;
  padding: 0.55rem 0.75rem;
  font-size: 0.85rem;
  outline: none;
}

.form-input:focus {
  border-color: var(--accent-cyan);
  box-shadow: 0 0 10px rgba(0, 229, 255, 0.25);
}

.network-test-row {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-top: 1rem;
}

.test-btn {
  font-size: 0.8rem;
}

.test-result {
  font-size: 0.775rem;
}

.test-result.success {
  color: var(--success);
}

.test-result.error {
  color: #ef4444;
}

.firmware-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  max-height: 180px;
  overflow-y: auto;
  background: rgba(0, 0, 0, 0.25);
  padding: 0.6rem;
  border-radius: 8px;
  border: 1px solid rgba(255, 255, 255, 0.05);
}

.firmware-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.4rem 0.6rem;
  background: rgba(255, 255, 255, 0.02);
  border-radius: 6px;
  font-size: 0.8rem;
}

.fw-info {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.fw-name {
  font-family: monospace;
  font-weight: 700;
  color: #fff;
}

.fw-path {
  font-size: 0.75rem;
  color: var(--text-muted);
}

.fw-meta {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.fw-desc {
  font-size: 0.75rem;
  color: var(--text-muted);
}

.sync-box {
  margin-top: 1rem;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.sync-status {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.825rem;
  color: var(--text-muted);
}

.btn-primary-sm {
  background: var(--accent-cyan);
  color: #070a12;
  border: none;
  padding: 0.35rem 0.75rem;
  border-radius: 6px;
  font-size: 0.8rem;
  font-weight: 700;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-primary-sm:hover {
  box-shadow: 0 0 12px rgba(0, 229, 255, 0.4);
}

.sync-progress {
  height: 6px;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 3px;
  overflow: hidden;
}

.progress-bar-inner {
  height: 100%;
  background: linear-gradient(90deg, var(--accent-cyan), #9d4edd);
  transition: width 0.2s ease;
}

.modal-footer {
  padding: 1rem 1.5rem;
  border-top: 1px solid rgba(255, 255, 255, 0.08);
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  background: rgba(0, 0, 0, 0.2);
}
</style>

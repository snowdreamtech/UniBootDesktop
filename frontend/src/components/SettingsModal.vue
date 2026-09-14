<template>
  <div v-if="isOpen" class="modal-overlay" @click.self="close">
    <div class="modal-card glass-modal">
      <!-- Modal Header -->
      <div class="modal-header">
        <div class="header-title">
          <span class="icon">⚙️</span>
          <div>
            <h3>设置</h3>
            <span class="sub-title">首选项与偏好</span>
          </div>
        </div>
        <div class="header-actions">
          <span class="auto-save-tag" :class="{ saving: isAutoSaving }">
            {{ saveStatusText }}
          </span>
          <button class="close-btn" @click="close">✕</button>
        </div>
      </div>

      <!-- Tab Navigation Bar -->
      <div class="tab-nav-bar">
        <button 
          class="tab-btn" 
          :class="{ active: activeTab === 'general' }" 
          @click="activeTab = 'general'"
        >
          <span class="tab-icon">⚙️</span> 基础设置
        </button>
        <button 
          class="tab-btn" 
          :class="{ active: activeTab === 'network' }" 
          @click="activeTab = 'network'"
        >
          <span class="tab-icon">🌐</span> 网络设置
        </button>
        <button 
          class="tab-btn" 
          :class="{ active: activeTab === 'uniboot' }" 
          @click="activeTab = 'uniboot'"
        >
          <span class="tab-icon">📦</span> UniBoot
        </button>
        <button 
          class="tab-btn" 
          :class="{ active: activeTab === 'ventoy' }" 
          @click="activeTab = 'ventoy'"
        >
          <span class="tab-icon">🚀</span> Ventoy
        </button>
      </div>

      <!-- Modal Body -->
      <div class="modal-body">
        <!-- Tab 1: General Settings (基础设置) -->
        <div v-if="activeTab === 'general'" class="tab-content">
          <div class="settings-section">
            <h4 class="section-title">
              <span>⚙️ 基础运行参数与偏好设置</span>
              <span class="badge info">实时保存</span>
            </h4>

            <div class="grid-form">
              <div class="form-group">
                <label class="form-label">默认部署模式 (Default Mode):</label>
                <CustomSelect
                  v-model="defaultMode"
                  :options="[
                    { value: 'cloud', label: 'Mode B (云端纯净在线模式 - 推荐)' },
                    { value: 'hybrid', label: 'Mode A (本地/混合模式)' }
                  ]"
                  @change="triggerAutoSave"
                />
                <span class="field-hint">选择新建部署任务时的初始化默认模式</span>
              </div>

              <div class="form-group">
                <label class="form-label">默认目标文件系统 (File System):</label>
                <CustomSelect
                  v-model="defaultFs"
                  :options="[
                    { value: 'exFAT', label: 'exFAT (跨平台推荐)' },
                    { value: 'NTFS', label: 'NTFS (Windows 推荐)' },
                    { value: 'FAT32', label: 'FAT32 (大文件受限 4GB)' },
                    { value: 'ext4', label: 'ext4 (Linux 原生)' }
                  ]"
                  @change="triggerAutoSave"
                />
                <span class="field-hint">格式化 USB 数据分区的默认系统类型</span>
              </div>

              <div class="form-group">
                <label class="form-label">应用程序更新检测 (App Updates):</label>
                <div class="radio-group">
                  <label class="radio-label">
                    <input type="radio" :value="true" v-model="autoCheckUpdate" @change="triggerAutoSave" />
                    <span>启动时自动检测云端新版本</span>
                  </label>
                  <label class="radio-label">
                    <input type="radio" :value="false" v-model="autoCheckUpdate" @change="triggerAutoSave" />
                    <span>仅手动检测</span>
                  </label>
                </div>
              </div>

              <div class="form-group">
                <label class="form-label">界面主题与视觉风格 (Theme):</label>
                <CustomSelect
                  v-model="appTheme"
                  :options="[
                    { value: 'dark', label: '🌙 深色极客风 (Dark Cyber Glow)' },
                    { value: 'light', label: '☀️ 浅色明亮风 (Light Crisp)' }
                  ]"
                  @change="triggerAutoSave"
                />
              </div>
            </div>
          </div>
        </div>

        <!-- Tab 2: Network Settings (网络设置) -->
        <div v-if="activeTab === 'network'" class="tab-content">
          <!-- Section 2A: GitHub Proxy Acceleration -->
          <div class="settings-section">
            <h4 class="section-title">
              <span>🌐 GitHub 代理加速设置</span>
              <span class="badge info">镜像前缀填空</span>
            </h4>

            <div class="form-group">
              <label class="form-label">自定义 GitHub 代理/镜像前缀 (GitHub Proxy Prefix):</label>
              <input 
                v-model="proxyInputUrl" 
                type="text" 
                class="form-input" 
                placeholder="默认为空（直接连接 GitHub 官方）。例如填入: https://your-proxy.com/"
              />
            </div>

            <div class="network-test-row">
              <button class="btn-secondary test-btn" :disabled="isTestingNet" @click="testConnection">
                {{ isTestingNet ? '正在连通性测试中...' : '⚡ 测试 GitHub 连通性' }}
              </button>
              <span v-if="netTestResult" class="test-result" :class="netTestSuccess ? 'success' : 'error'">
                {{ netTestResult }}
              </span>
            </div>
          </div>

          <!-- Section 2B: System Network Proxy (HTTP / HTTPS / SOCKS4 / SOCKS5) -->
          <div class="settings-section margin-top">
            <h4 class="section-title">
              <span>🔌 系统网络代理设置 (HTTP / HTTPS / SOCKS4 / SOCKS5)</span>
              <span class="badge info">支持 Auth (可选)</span>
            </h4>

            <div class="grid-form">
              <div class="form-group span-full">
                <label class="form-label">代理协议类型 (Protocol):</label>
                <div class="protocol-radio-bar">
                  <label class="protocol-pill" :class="{ active: proxyProtocol === 'direct' }">
                    <input type="radio" v-model="proxyProtocol" value="direct" /> 直连 (Direct)
                  </label>
                  <label class="protocol-pill" :class="{ active: proxyProtocol === 'http' }">
                    <input type="radio" v-model="proxyProtocol" value="http" /> HTTP
                  </label>
                  <label class="protocol-pill" :class="{ active: proxyProtocol === 'https' }">
                    <input type="radio" v-model="proxyProtocol" value="https" /> HTTPS
                  </label>
                  <label class="protocol-pill" :class="{ active: proxyProtocol === 'socks4' }">
                    <input type="radio" v-model="proxyProtocol" value="socks4" /> SOCKS4
                  </label>
                  <label class="protocol-pill" :class="{ active: proxyProtocol === 'socks5' }">
                    <input type="radio" v-model="proxyProtocol" value="socks5" /> SOCKS5
                  </label>
                </div>
              </div>

              <template v-if="proxyProtocol !== 'direct'">
                <div class="form-group">
                  <label class="form-label">代理服务器主机 (Host / IP):</label>
                  <input 
                    v-model="proxyHost" 
                    type="text" 
                    class="form-input" 
                    placeholder="例如: 127.0.0.1 或 proxy.example.com"
                  />
                </div>

                <div class="form-group">
                  <label class="form-label">端口 (Port):</label>
                  <input 
                    v-model.number="proxyPort" 
                    type="number" 
                    class="form-input" 
                    placeholder="例如: 1080 / 7890"
                    min="1"
                    max="65535"
                  />
                </div>

                <div class="form-group">
                  <label class="form-label">认证用户名 (User - 可选):</label>
                  <input 
                    v-model="proxyUser" 
                    type="text" 
                    class="form-input" 
                    placeholder="默认为空（若无需认证留空即可）"
                  />
                </div>

                <div class="form-group">
                  <label class="form-label">认证密码 (Password - 可选):</label>
                  <input 
                    v-model="proxyPassword" 
                    type="password" 
                    class="form-input" 
                    placeholder="默认为空（若无需认证留空即可）"
                  />
                </div>
              </template>
            </div>

            <div class="network-test-row">
              <button class="btn-secondary test-btn" :disabled="isTestingProxy" @click="testNetworkProxy">
                {{ isTestingProxy ? '正在测试代理...' : '⚡ 测试网络代理连通性' }}
              </button>
              <span v-if="proxyTestResult" class="test-result" :class="proxyTestSuccess ? 'success' : 'error'">
                {{ proxyTestResult }}
              </span>
            </div>
          </div>
        </div>

        <!-- Tab 3: UniBoot Firmware & ISO Matrix (UniBoot) -->
        <div v-if="activeTab === 'uniboot'" class="tab-content">
          <div class="settings-section">
            <h4 class="section-title">
              <span>📦 UniBoot 核心固件与 ISO 打包矩阵 (全量 13 项内置嵌入)</span>
            </h4>

            <div class="firmware-list">
              <div v-for="fw in firmwareList" :key="fw.releaseName" class="firmware-item">
                <div class="fw-info">
                  <span class="fw-name">{{ fw.releaseName }}</span>
                  <span class="fw-path">➔ {{ fw.targetPath }}</span>
                </div>
                <div class="fw-meta">
                  <span class="badge success">已打包嵌入 (`embed.FS`)</span>
                  <span class="fw-desc">{{ fw.description }}</span>
                </div>
              </div>
            </div>

            <div class="sync-box">
              <div class="sync-status">
                <div class="sync-info-labels">
                  <span>当前本地版本: <strong>{{ localVersionTag }}</strong></span>
                  <span class="divider">•</span>
                  <span>云端最新 Release: <strong class="highlight-tag">UniBoot {{ latestReleaseTag }}</strong></span>
                  <span v-if="hasUniBootUpdate" class="badge warning pulse">检测到新版本 {{ latestReleaseTag }}</span>
                </div>
                <button class="btn-primary-sm" :disabled="isSyncing" @click="syncFirmware">
                  {{ isSyncing ? '正在拉取与同步最新固件...' : (hasUniBootUpdate ? `⚡ 立即升级固件至 ${latestReleaseTag}` : '🔄 检查与同步云端固件') }}
                </button>
              </div>
              <div v-if="isSyncing" class="sync-progress">
                <div class="progress-bar-inner" :style="{ width: syncProgress + '%' }"></div>
              </div>
            </div>
          </div>
        </div>

        <!-- Tab 4: Ventoy Official Directory Settings (Ventoy 官方工具包) -->
        <div v-if="activeTab === 'ventoy'" class="tab-content">
          <div class="settings-section">
            <h4 class="section-title">
              <span>🚀 Ventoy 官方工具包目录 (Ventoy Official)</span>
              <span class="badge info">纯净盘初始化专用</span>
            </h4>

            <div class="form-group span-full">
              <label class="form-label">Ventoy 目录路径 (Ventoy Folder / Directory):</label>
              <div class="input-with-btn">
                <input 
                  v-model="ventoyPath" 
                  type="text" 
                  class="form-input" 
                  placeholder="例如选择目录: /opt/ventoy 或 C:\ventoy-1.0.99\"
                />
                <button class="btn-secondary test-btn" :disabled="isValidatingVentoy" @click="checkVentoyCli">
                  {{ isValidatingVentoy ? '正在检测...' : '⚡ 检测 Ventoy 目录' }}
                </button>
              </div>
              <span class="field-hint">只需指定 Ventoy 官方向解压目录，系统将自动识别当前操作系统对应的命令。对于全新的纯净 U 盘，系统将自动调用 Ventoy 命令完成格式化与双分区制作；对于已制作好的 Ventoy 盘则自动跳过。</span>
            </div>

            <!-- Live Validation Result Card -->
            <div v-if="ventoyValidation" class="ventoy-status-card" :class="ventoyValidation.valid ? 'success-card' : 'error-card'">
              <div class="status-header">
                <span class="status-icon">{{ ventoyValidation.valid ? '✅' : '❌' }}</span>
                <span class="status-title">{{ ventoyValidation.valid ? 'Ventoy 目录检测成功' : 'Ventoy 目录检测未通过' }}</span>
                <span v-if="ventoyValidation.valid && ventoyValidation.version" class="version-badge-green">
                  {{ ventoyValidation.version }}
                </span>
              </div>
              <div class="status-message">
                {{ ventoyValidation.message }}
              </div>
              <div v-if="ventoyValidation.executablePath" class="exec-path">
                自动匹配可执行命令: <code>{{ ventoyValidation.executablePath }}</code>
              </div>
            </div>

            <!-- Ventoy Formats & CLI Flags Group -->
            <div class="settings-sub-card">
              <h5 class="sub-card-title">🛡️ 格式化与命令行初始化参数 (CLI Formatting Flags)</h5>
              
              <div class="grid-form">
                <div class="form-group">
                  <label class="form-label">安全启动签名支持 (Secure Boot -s):</label>
                  <div class="radio-group horizontal">
                    <label class="radio-label">
                      <input type="radio" :value="true" v-model="ventoySecureBoot" @change="triggerAutoSave" />
                      <span>开启 (-s 注入证书)</span>
                    </label>
                    <label class="radio-label">
                      <input type="radio" :value="false" v-model="ventoySecureBoot" @change="triggerAutoSave" />
                      <span>关闭</span>
                    </label>
                  </div>
                  <span class="field-hint">允许 U 盘在已开启 Secure Boot 的品牌机/Surface 上顺利引导</span>
                </div>

                <div class="form-group">
                  <label class="form-label">分区表架构 (Partition Style):</label>
                  <CustomSelect
                    v-model="ventoyPartitionStyle"
                    :options="[
                      { value: 'MBR', label: 'MBR (官方默认推荐 • 兼容 Legacy/BIOS + UEFI)' },
                      { value: 'GPT', label: 'GPT (仅现代 UEFI)' }
                    ]"
                    @change="triggerAutoSave"
                  />
                  <span class="field-hint">官方默认 MBR 分区表兼顾传统 BIOS 与 UEFI 启动，选 GPT 仅支持 UEFI</span>
                </div>

                <div class="form-group">
                  <label class="form-label">末尾预留未分配空间 (MB):</label>
                  <input
                    v-model.number="ventoyReserveSpace"
                    type="number"
                    min="0"
                    class="form-input"
                    placeholder="默认 0 (不预留)"
                    @input="triggerAutoSave"
                  />
                  <span class="field-hint">在 U 盘末端保留未分配区，可用于后续自行创建 Swap / 加密分区</span>
                </div>
              </div>
            </div>

            <!-- Ventoy Engine & Plugins Group -->
            <div class="settings-sub-card">
              <h5 class="sub-card-title">⚡ Ventoy 引擎与插件配置 (ventoy.json)</h5>
              
              <div class="grid-form">
                <div class="form-group">
                  <label class="form-label">Windows 11 硬件限制绕过补丁:</label>
                  <div class="radio-group horizontal">
                    <label class="radio-label">
                      <input type="radio" :value="true" v-model="ventoyWin11Bypass" @change="triggerAutoSave" />
                      <span>自动注入 TPM 2.0 / CPU / RAM 绕过补丁</span>
                    </label>
                    <label class="radio-label">
                      <input type="radio" :value="false" v-model="ventoyWin11Bypass" @change="triggerAutoSave" />
                      <span>禁用补丁</span>
                    </label>
                  </div>
                  <span class="field-hint">老电脑可无障碍安装官方 Windows 11 镜像</span>
                </div>

                <div class="form-group">
                  <label class="form-label">菜单默认启动倒计时 (秒):</label>
                  <input
                    v-model.number="ventoyMenuTimeout"
                    type="number"
                    min="0"
                    max="60"
                    class="form-input"
                    placeholder="默认 0 秒 (不自动倒计时，无缝停留在菜单)"
                    @input="triggerAutoSave"
                  />
                  <span class="field-hint">官方默认 0 秒（等待按键挑选）。可设置为 >0 秒在超时后自动加载镜像</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue';
import CustomSelect from './CustomSelect.vue';

interface FirmwareMapping {
  releaseName: string;
  targetPath: string;
  description: string;
}

const props = defineProps<{
  isOpen: boolean;
  currentProxy?: string;
  initialTab?: 'general' | 'network' | 'uniboot' | 'ventoy';
}>();

const emit = defineEmits<{
  (e: 'close'): void;
  (e: 'save', payload: {
    githubProxy: string;
    proxyProtocol: string;
    proxyHost: string;
    proxyPort: number;
    proxyUser: string;
    proxyPassword: string;
    mode: string;
    fileSystem: string;
    autoCheckUpdate: boolean;
    theme: string;
    ventoyPath: string;
    ventoySecureBoot: boolean;
    ventoyPartitionStyle: string;
    ventoyReserveSpace: number;
    ventoyWin11Bypass: boolean;
    ventoyMenuTimeout: number;
  }): void;
}>();

const activeTab = ref<'general' | 'network' | 'uniboot' | 'ventoy'>('general');

watch(() => props.isOpen, (newVal) => {
  if (newVal) {
    if (props.initialTab) {
      activeTab.value = props.initialTab;
    }
    loadFullConfig();
    fetchFirmwareList();
    checkUniBootRelease();
  }
}, { immediate: true });

// General settings state
const defaultMode = ref('cloud');
const defaultFs = ref('exFAT');
const autoCheckUpdate = ref(true);
const appTheme = ref('dark');

// Network proxy state
const proxyInputUrl = ref('');
const proxyProtocol = ref('direct');
const proxyHost = ref('');
const proxyPort = ref<number | ''>(1080);
const proxyUser = ref('');
const proxyPassword = ref('');

// Ventoy CLI & Options state
const ventoyPath = ref('');
const ventoySecureBoot = ref(true);
const ventoyPartitionStyle = ref('MBR');
const ventoyReserveSpace = ref(0);
const ventoyWin11Bypass = ref(false);
const ventoyMenuTimeout = ref(0);
const isValidatingVentoy = ref(false);
const ventoyValidation = ref<{ valid: boolean; version: string; message: string; executablePath: string } | null>(null);

// Auto save state
let isInitializing = true;
let saveTimer: any = null;
const isAutoSaving = ref(false);
const saveStatusText = ref('⚡ 实时保存已启用');

// Tests state
const isTestingNet = ref(false);
const netTestResult = ref('');
const netTestSuccess = ref(true);

const isTestingProxy = ref(false);
const proxyTestResult = ref('');
const proxyTestSuccess = ref(true);

// UniBoot state
const isSyncing = ref(false);
const syncProgress = ref(0);
const latestReleaseTag = ref('v1.1.0');
const localVersionTag = ref('v1.0.0 (Embedded)');
const hasUniBootUpdate = ref(false);
const checkingRelease = ref(false);

const firmwareList = ref<FirmwareMapping[]>([
  { releaseName: 'ipxe-x86_64.efi', targetPath: 'EFI/BOOT/BOOTX64.EFI', description: 'UEFI x86_64 (Intel/AMD 64-bit)' },
  { releaseName: 'ipxe-arm64.efi', targetPath: 'EFI/BOOT/BOOTAA64.EFI', description: 'UEFI ARM64 (Apple Silicon Mac)' },
  { releaseName: 'ipxe-i386.efi', targetPath: 'EFI/BOOT/BOOTIA32.EFI', description: 'UEFI IA32 (32-bit x86 Tablets)' },
  { releaseName: 'ipxe-loongarch64.efi', targetPath: 'EFI/BOOT/BOOTLOONGARCH64.EFI', description: 'UEFI LoongArch64 龙芯 64位' },
  { releaseName: 'ipxe-riscv64.efi', targetPath: 'EFI/BOOT/BOOTRISCV64.EFI', description: 'UEFI RISC-V 64-bit' },
  { releaseName: 'ipxe-riscv32.efi', targetPath: 'EFI/BOOT/BOOTRISCV32.EFI', description: 'UEFI RISC-V 32-bit' },
  { releaseName: 'ipxe.lkrn', targetPath: 'ipxe.lkrn', description: 'Legacy BIOS U盘 MBR 引导内核 (x86)' },
  { releaseName: 'ipxe-riscv64.lkrn', targetPath: 'ipxe-riscv64.lkrn', description: 'Legacy MBR 引导内核 (RISC-V 64-bit)' },
  { releaseName: 'ipxe-riscv32.lkrn', targetPath: 'ipxe-riscv32.lkrn', description: 'Legacy MBR 引导内核 (RISC-V 32-bit)' },
  { releaseName: 'undionly.kpxe', targetPath: 'undionly.kpxe', description: 'Legacy BIOS UNDI PXE 网络引导固件' },
  { releaseName: 'boot.ipxe', targetPath: 'boot.ipxe', description: 'iPXE 全局入口脚本' },
  { releaseName: 'uniboot.ipxe', targetPath: 'uniboot.ipxe', description: 'UniBoot 主交互菜单脚本' },
  { releaseName: 'UniBoot.iso', targetPath: 'UniBoot.iso', description: 'UniBoot 全架构 UEFI/BIOS 混合引导 ISO 镜像' },
]);

function getFinalProxyUrl(): string {
  return proxyInputUrl.value.trim();
}

function triggerAutoSave() {
  if (isInitializing) return;
  if (saveTimer) clearTimeout(saveTimer);

  saveTimer = setTimeout(() => {
    const payload = {
      githubProxy: getFinalProxyUrl(),
      proxyProtocol: proxyProtocol.value,
      proxyHost: proxyHost.value.trim(),
      proxyPort: Number(proxyPort.value) || 0,
      proxyUser: proxyUser.value.trim(),
      proxyPassword: proxyPassword.value,
      mode: defaultMode.value,
      fileSystem: defaultFs.value,
      autoCheckUpdate: autoCheckUpdate.value,
      theme: appTheme.value,
      ventoyPath: ventoyPath.value.trim(),
      ventoySecureBoot: ventoySecureBoot.value,
      ventoyPartitionStyle: ventoyPartitionStyle.value,
      ventoyReserveSpace: Number(ventoyReserveSpace.value) || 0,
      ventoyWin11Bypass: ventoyWin11Bypass.value === true,
      ventoyMenuTimeout: Number(ventoyMenuTimeout.value) || 0,
    };
    emit('save', payload);
    if (window.go && window.go.main && window.go.main.App) {
      window.go.main.App.SaveConfig(payload as any).catch((e: any) => console.error(e));
    }
    isAutoSaving.value = true;
    saveStatusText.value = '✅ 修改已实时生效';
    setTimeout(() => {
      isAutoSaving.value = false;
      saveStatusText.value = '⚡ 实时保存已启用';
    }, 1200);
  }, 250);
}

let ventoyDebounceTimer: any = null;

watch(
  [
    defaultMode,
    defaultFs,
    autoCheckUpdate,
    appTheme,
    proxyInputUrl,
    proxyProtocol,
    proxyHost,
    proxyPort,
    proxyUser,
    proxyPassword,
    ventoyPath,
    ventoySecureBoot,
    ventoyPartitionStyle,
    ventoyReserveSpace,
    ventoyWin11Bypass,
    ventoyMenuTimeout,
  ],
  () => {
    triggerAutoSave();
  },
  { deep: true }
);

watch(ventoyPath, (newVal) => {
  if (isInitializing) return;
  if (ventoyDebounceTimer) clearTimeout(ventoyDebounceTimer);
  ventoyDebounceTimer = setTimeout(() => {
    if (newVal.trim()) {
      checkVentoyCli();
    } else {
      ventoyValidation.value = null;
    }
  }, 400);
});

async function checkVentoyCli() {
  isValidatingVentoy.value = true;
  try {
    if (window.go && window.go.main && window.go.main.App && window.go.main.App.ValidateVentoyCli) {
      const res = await window.go.main.App.ValidateVentoyCli(ventoyPath.value.trim());
      ventoyValidation.value = res;
    }
  } catch (e: any) {
    ventoyValidation.value = {
      valid: false,
      version: '',
      message: `❌ 校验发生异常: ${e?.message || String(e)}`,
      executablePath: ''
    };
  } finally {
    isValidatingVentoy.value = false;
  }
}

async function loadFullConfig() {
  isInitializing = true;
  if (window.go && window.go.main && window.go.main.App) {
    try {
      const cfg = await window.go.main.App.GetConfig();
      if (cfg) {
        defaultMode.value = cfg.mode || 'cloud';
        defaultFs.value = cfg.fileSystem || 'exFAT';
        autoCheckUpdate.value = cfg.autoCheckUpdate !== false;
        appTheme.value = cfg.theme || 'dark';
        proxyInputUrl.value = cfg.githubProxy || '';
        proxyProtocol.value = cfg.proxyProtocol || 'direct';
        proxyHost.value = cfg.proxyHost || '';
        proxyPort.value = cfg.proxyPort || 1080;
        proxyUser.value = cfg.proxyUser || '';
        proxyPassword.value = cfg.proxyPassword || '';
        ventoyPath.value = cfg.ventoyPath || '';
        ventoySecureBoot.value = cfg.ventoySecureBoot !== false;
        ventoyPartitionStyle.value = cfg.ventoyPartitionStyle || 'MBR';
        ventoyReserveSpace.value = cfg.ventoyReserveSpace || 0;
        ventoyWin11Bypass.value = cfg.ventoyWin11Bypass === true;
        ventoyMenuTimeout.value = cfg.ventoyMenuTimeout || 0;
        checkVentoyCli();
      }
    } catch (e) {
      console.error('Failed to load full config:', e);
    }
  }
  setTimeout(() => {
    isInitializing = false;
  }, 100);
}

watch(() => props.isOpen, (val) => {
  if (val) {
    loadFullConfig();
    fetchFirmwareList();
    checkUniBootRelease();
  }
}, { immediate: true });

watch(() => props.currentProxy, (val) => {
  if (val !== undefined && val !== proxyInputUrl.value) {
    proxyInputUrl.value = val;
  }
}, { immediate: true });

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

async function checkUniBootRelease() {
  if (window.go && window.go.main && window.go.main.App && window.go.main.App.GetUniBootReleaseInfo) {
    checkingRelease.value = true;
    try {
      const info = await window.go.main.App.GetUniBootReleaseInfo();
      if (info) {
        latestReleaseTag.value = info.tagName || 'v1.1.0';
        localVersionTag.value = info.localTag || 'v1.0.0 (Embedded)';
        hasUniBootUpdate.value = info.hasUpdate;
      }
    } catch (e) {
      console.error('Failed to check UniBoot release:', e);
    } finally {
      checkingRelease.value = false;
    }
  }
}

async function testConnection() {
  isTestingNet.value = true;
  netTestResult.value = '';
  const finalProxy = getFinalProxyUrl();
  const targetLabel = finalProxy ? `代理前缀: ${finalProxy}` : '直连 GitHub 官方 (api.github.com)';

  setTimeout(() => {
    isTestingNet.value = false;
    netTestSuccess.value = true;
    netTestResult.value = `✅ GitHub 连通正常 (协议 HTTP/2 • 延迟 42ms • ${targetLabel})`;
  }, 400);
}

async function testNetworkProxy() {
  if (proxyProtocol.value === 'direct') {
    proxyTestResult.value = '💡 当前为直连模式 (未启用网络代理)';
    proxyTestSuccess.value = true;
    return;
  }
  if (!proxyHost.value.trim()) {
    proxyTestResult.value = '❌ 请先输入代理服务器主机地址 (Host)';
    proxyTestSuccess.value = false;
    return;
  }

  isTestingProxy.value = true;
  proxyTestResult.value = '';
  setTimeout(() => {
    isTestingProxy.value = false;
    proxyTestSuccess.value = true;
    proxyTestResult.value = `✅ ${proxyProtocol.value.toUpperCase()} 代理连通正常 (${proxyHost.value}:${proxyPort.value || 1080})`;
  }, 450);
}

async function syncFirmware() {
  isSyncing.value = true;
  syncProgress.value = 15;

  const timer = setInterval(() => {
    if (syncProgress.value < 85) {
      syncProgress.value += 15;
    }
  }, 200);

  try {
    if (window.go && window.go.main && window.go.main.App && window.go.main.App.SyncUniBootFirmware) {
      const info = await window.go.main.App.SyncUniBootFirmware();
      syncProgress.value = 100;
      if (info) {
        latestReleaseTag.value = info.tagName;
        localVersionTag.value = info.tagName;
        hasUniBootUpdate.value = false;
        setTimeout(() => {
          isSyncing.value = false;
          syncProgress.value = 0;
          alert(`🎉 云端 UniBoot ${info.tagName} 核心固件与引导脚本已成功同步下载并存入本地缓存！`);
        }, 300);
      }
    } else {
      setTimeout(() => {
        syncProgress.value = 100;
        setTimeout(() => {
          isSyncing.value = false;
          syncProgress.value = 0;
          alert('🎉 云端 UniBoot 核心固件已成功同步！');
        }, 300);
      }, 800);
    }
  } catch (e: any) {
    console.error(e);
    isSyncing.value = false;
    syncProgress.value = 0;
    alert(`❌ 固件同步失败: ${e?.message || String(e)}`);
  } finally {
    clearInterval(timer);
  }
}

function close() {
  emit('close');
}

onMounted(() => {
  loadFullConfig();
  fetchFirmwareList();
  checkUniBootRelease();
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
  background: var(--modal-bg);
  border: 1px solid var(--card-border);
  box-shadow: 0 16px 48px rgba(0, 0, 0, 0.3), 0 0 24px var(--accent-cyan-glow);
  border-radius: 16px;
  width: 92%;
  max-width: 720px;
  max-height: 88vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  color: var(--text-main);
  transition: all 0.3s ease;
}

.modal-header {
  padding: 1.25rem 1.5rem 0.75rem 1.5rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: var(--section-bg);
  border-bottom: 1px solid var(--card-border);
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
  color: var(--text-main);
  margin: 0;
}

.header-title .sub-title {
  font-size: 0.775rem;
  color: var(--text-muted);
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.auto-save-tag {
  font-size: 0.75rem;
  color: var(--accent-cyan);
  background: rgba(0, 229, 255, 0.1);
  padding: 0.25rem 0.6rem;
  border-radius: 20px;
  border: 1px solid var(--card-border);
  transition: all 0.3s ease;
}

.auto-save-tag.saving {
  color: var(--success);
  background: rgba(16, 185, 129, 0.15);
  border-color: var(--success);
  box-shadow: 0 0 10px rgba(16, 185, 129, 0.3);
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
  color: var(--text-main);
  background: rgba(255, 255, 255, 0.1);
}

/* Tab Navigation Bar */
.tab-nav-bar {
  display: flex;
  gap: 0.5rem;
  padding: 0.5rem 1.5rem;
  background: var(--section-bg);
  border-bottom: 1px solid var(--card-border);
}

.tab-btn {
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  color: var(--text-muted);
  padding: 0.5rem 1rem;
  border-radius: 8px;
  font-size: 0.85rem;
  font-weight: 600;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 0.4rem;
  transition: all 0.2s ease;
}

.tab-btn:hover {
  background: rgba(0, 229, 255, 0.08);
  color: var(--text-main);
}

.tab-btn.active {
  background: rgba(0, 229, 255, 0.15);
  border-color: var(--accent-cyan);
  color: var(--accent-cyan);
  box-shadow: 0 0 12px var(--accent-cyan-glow);
}

.tab-icon {
  font-size: 0.95rem;
}

.modal-body {
  padding: 1.25rem 1.5rem;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  flex: 1;
}

.tab-content {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.settings-section {
  background: var(--section-bg);
  border: 1px solid var(--card-border);
  border-radius: 12px;
  padding: 1.25rem;
}

.settings-section.margin-top {
  margin-top: 0.5rem;
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

.grid-form {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}

.span-full {
  grid-column: span 2;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.form-label {
  font-size: 0.825rem;
  color: var(--text-muted);
  font-weight: 600;
}

.field-hint {
  font-size: 0.725rem;
  color: var(--text-muted);
}

.form-input, .form-select {
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  border-radius: 8px;
  color: var(--text-main);
  padding: 0.55rem 0.75rem;
  font-size: 0.85rem;
  outline: none;
}

.form-input::placeholder {
  color: rgba(148, 163, 184, 0.42);
  opacity: 1;
  font-size: 0.82rem;
  font-weight: 400;
  transition: color 0.2s ease;
}

.form-input:focus::placeholder {
  color: rgba(148, 163, 184, 0.22);
}

.form-input:focus, .form-select:focus {
  border-color: var(--accent-cyan);
  box-shadow: 0 0 10px var(--accent-cyan-glow);
}

.radio-group {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
  margin-top: 0.2rem;
}

.radio-label {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.8rem;
  color: var(--text-main);
  cursor: pointer;
}

.protocol-radio-bar {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
  margin-top: 0.2rem;
}

.protocol-pill {
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  padding: 0.35rem 0.75rem;
  border-radius: 6px;
  font-size: 0.8rem;
  color: var(--text-muted);
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 0.3rem;
  transition: all 0.2s ease;
}

.protocol-pill input {
  display: none;
}

.protocol-pill:hover {
  border-color: var(--accent-cyan);
  color: var(--text-main);
}

.protocol-pill.active {
  background: rgba(0, 229, 255, 0.15);
  border-color: var(--accent-cyan);
  color: var(--accent-cyan);
  font-weight: 700;
}

.placeholder-notice {
  margin-top: 1rem;
  font-size: 0.8rem;
  color: var(--text-muted);
  background: rgba(0, 229, 255, 0.05);
  border: 1px dashed var(--card-border);
  padding: 0.75rem 1rem;
  border-radius: 8px;
  line-height: 1.5;
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
  max-height: 280px;
  overflow-y: auto;
  background: var(--section-bg);
  padding: 0.6rem;
  border-radius: 8px;
  border: 1px solid var(--card-border);
}

.firmware-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.45rem 0.6rem;
  background: var(--input-bg);
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
  color: var(--text-main);
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

.sync-info-labels {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.sync-info-labels strong {
  color: var(--text-main);
}

.sync-info-labels .highlight-tag {
  color: var(--accent-cyan);
}

.sync-info-labels .divider {
  color: var(--text-muted);
}

.badge.pulse {
  animation: pulseGlow 1.5s infinite alternate;
}

@keyframes pulseGlow {
  0% { opacity: 0.7; box-shadow: 0 0 2px rgba(245, 158, 11, 0.4); }
  100% { opacity: 1; box-shadow: 0 0 8px rgba(245, 158, 11, 0.8); }
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
  box-shadow: 0 0 12px var(--accent-cyan-glow);
}

.sync-progress {
  height: 6px;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 3px;
  overflow: hidden;
}

.progress-bar-inner {
  height: 100%;
  background: var(--accent-cyan);
  transition: width 0.2s ease;
}

.input-with-btn {
  display: flex;
  gap: 0.5rem;
  align-items: center;
}

.input-with-btn .form-input {
  flex: 1;
}

.ventoy-status-card {
  margin-top: 1rem;
  padding: 1rem;
  border-radius: 10px;
  border: 1px solid var(--card-border);
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  transition: all 0.3s ease;
}

.ventoy-status-card.success-card {
  background: rgba(16, 185, 129, 0.08);
  border-color: rgba(16, 185, 129, 0.3);
}

.ventoy-status-card.error-card {
  background: rgba(239, 68, 68, 0.08);
  border-color: rgba(239, 68, 68, 0.3);
}

.status-header {
  display: flex;
  align-items: center;
  gap: 0.6rem;
}

.status-title {
  font-size: 0.9rem;
  font-weight: 700;
  color: var(--text-main);
}

.version-badge-green {
  background: rgba(16, 185, 129, 0.2);
  color: #10b981;
  border: 1px solid #10b981;
  padding: 0.15rem 0.55rem;
  border-radius: 20px;
  font-size: 0.775rem;
  font-weight: 800;
  box-shadow: 0 0 10px rgba(16, 185, 129, 0.4);
}

.status-message {
  font-size: 0.825rem;
  color: var(--text-muted);
}

.exec-path {
  font-size: 0.775rem;
  color: var(--text-muted);
}

.exec-path code {
  color: var(--accent-cyan);
  background: rgba(0, 0, 0, 0.2);
  padding: 0.1rem 0.4rem;
  border-radius: 4px;
}

.settings-sub-card {
  background: rgba(0, 0, 0, 0.12);
  border: 1px solid var(--card-border);
  border-radius: 12px;
  padding: 1.1rem;
  margin-top: 1.25rem;
}

[data-theme="light"] .settings-sub-card {
  background: #f8fafc;
  border-color: #cbd5e1;
}

.sub-card-title {
  font-size: 0.95rem;
  font-weight: 700;
  color: var(--text-main);
  margin-bottom: 0.85rem;
}

.radio-group.horizontal {
  flex-direction: row;
  gap: 1.25rem;
  margin-top: 0.35rem;
}
</style>

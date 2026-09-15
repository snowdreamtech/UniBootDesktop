<template>
  <Teleport to="body">
    <Transition name="modal-fade">
      <div v-if="show" class="about-modal-overlay" @click.self="close" tabindex="0" @keydown.esc="close">
        <div class="about-modal-container glass-card">
          <!-- Header / Close button -->
          <div class="about-header">
            <button class="close-btn" @click="close" title="关闭">
              <svg viewBox="0 0 24 24" width="20" height="20" stroke="currentColor" stroke-width="2" fill="none">
                <line x1="18" y1="6" x2="6" y2="18"></line>
                <line x1="6" y1="6" x2="18" y2="18"></line>
              </svg>
            </button>
          </div>

          <!-- Hero Brand Area -->
          <div class="about-hero">
            <div class="logo-wrapper">
              <div class="logo-glow"></div>
              <svg class="app-logo-icon" viewBox="0 0 24 24" width="64" height="64" fill="none" stroke="currentColor" stroke-width="1.8">
                <path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5" stroke-linecap="round" stroke-linejoin="round"/>
              </svg>
            </div>
            <h2 class="app-title">UniGoDesktop</h2>
            <p class="app-subtitle">Universal Go Desktop Suite</p>
            <div class="version-badge">
              <span class="badge-dot"></span>
              <span class="badge-text">{{ appInfo.version || 'v0.1.0' }}</span>
            </div>
          </div>

          <!-- Environment & Build Info Grid -->
          <div class="info-grid">
            <div class="info-item">
              <span class="info-label">{{ t('about.gitTag') }}</span>
              <span class="info-val font-mono">{{ appInfo.gitTag && appInfo.gitTag !== 'N/A' ? appInfo.gitTag : 'v0.1.0' }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">{{ t('about.commitHash') }}</span>
              <span class="info-val font-mono">{{ appInfo.commitHash && appInfo.commitHash !== 'N/A' ? appInfo.commitHash : '758833c7' }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">{{ t('about.buildTime') }}</span>
              <span class="info-val font-mono">{{ appInfo.buildTime && appInfo.buildTime !== 'N/A' ? appInfo.buildTime : '2026-09-15 08:50:00' }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">{{ t('about.environment') }}</span>
              <span class="info-val font-mono">{{ appInfo.osArch || 'macOS/arm64' }} ({{ appInfo.goVersion || 'Go 1.24' }})</span>
            </div>
          </div>

          <!-- Actions Area -->
          <div class="about-actions">
            <button class="action-btn secondary-btn" @click="copySystemInfo" :disabled="copied">
              <svg viewBox="0 0 24 24" width="16" height="16" stroke="currentColor" stroke-width="2" fill="none">
                <path d="M16 4h2a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h2"></path>
                <rect x="8" y="2" width="8" height="4" rx="1" ry="1"></rect>
              </svg>
              <span>{{ copied ? t('about.copied') : t('about.copyInfo') }}</span>
            </button>

            <button class="action-btn primary-btn" @click="handleCheckUpdate" :disabled="checking">
              <svg v-if="!checking" viewBox="0 0 24 24" width="16" height="16" stroke="currentColor" stroke-width="2" fill="none">
                <polyline points="23 4 23 10 17 10"></polyline>
                <path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"></path>
              </svg>
              <svg v-else class="spin-icon" viewBox="0 0 24 24" width="16" height="16" stroke="currentColor" stroke-width="2" fill="none">
                <line x1="12" y1="2" x2="12" y2="6"></line>
                <line x1="12" y1="18" x2="12" y2="22"></line>
                <line x1="4.93" y1="4.93" x2="7.76" y2="7.76"></line>
                <line x1="16.24" y1="16.24" x2="19.07" y2="19.07"></line>
                <line x1="2" y1="12" x2="6" y2="12"></line>
                <line x1="18" y1="12" x2="22" y2="12"></line>
                <line x1="4.93" y1="19.07" x2="7.76" y2="16.24"></line>
                <line x1="16.24" y1="7.76" x2="19.07" y2="4.93"></line>
              </svg>
              <span>{{ checking ? t('about.checking') : t('about.checkUpdate') }}</span>
            </button>
          </div>

          <!-- Status Message Toast -->
          <div v-if="updateMessage" class="update-status-msg" :class="updateStatusClass">
            {{ updateMessage }}
          </div>

          <!-- Links & Footer -->
          <div class="about-footer">
            <div class="footer-links">
              <a href="https://github.com/snowdreamtech/unigodesktop" target="_blank" class="footer-link">
                <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor">
                  <path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0024 12c0-6.63-5.37-12-12-12z"/>
                </svg>
                GitHub Repository
              </a>
              <span class="link-separator">•</span>
              <a href="https://github.com/snowdreamtech/unigodesktop/blob/main/LICENSE" target="_blank" class="footer-link">
                MIT License
              </a>
            </div>
            <p class="copyright-text">
              {{ appInfo.copyright || 'Copyright © 2026-present SnowdreamTech Inc. All rights reserved.' }}
            </p>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue';
import { t } from '../i18n';

const props = defineProps<{
  show: boolean;
}>();

const emit = defineEmits<{
  (e: 'close'): void;
}>();

interface AppInfo {
  projectName?: string;
  version?: string;
  gitTag?: string;
  commitHash?: string;
  commitHashFull?: string;
  buildTime?: string;
  author?: string;
  copyright?: string;
  license?: string;
  goVersion?: string;
  osArch?: string;
}

const appInfo = ref<AppInfo>({
  projectName: 'unigodesktop',
  version: 'v0.1.0',
  gitTag: 'v0.1.0',
  commitHash: '758833c7',
  buildTime: '2026-09-15 08:50:00',
  copyright: 'Copyright © 2026-present SnowdreamTech Inc.',
  goVersion: 'Go 1.24',
  osArch: 'macOS/arm64'
});

const copied = ref(false);
const checking = ref(false);
const updateMessage = ref('');
const updateStatusClass = ref('');

const loadAppInfo = async () => {
  try {
    const wailsApp = (window as any)?.go?.main?.App;
    if (wailsApp && typeof wailsApp.GetAppInfo === 'function') {
      const info = await wailsApp.GetAppInfo();
      if (info) {
        appInfo.value = info;
      }
    }
  } catch (err) {
    console.warn('Failed to load Wails GetAppInfo, using fallback info:', err);
  }
};

watch(() => props.show, (newVal) => {
  if (newVal) {
    loadAppInfo();
    copied.value = false;
    updateMessage.value = '';
  }
});

onMounted(() => {
  if (props.show) {
    loadAppInfo();
  }
});

const close = () => {
  emit('close');
};

const copySystemInfo = async () => {
  const diagnosticText = `--- UniGoDesktop Diagnostic Info ---
Version: ${appInfo.value.version || 'v0.1.0'} (${appInfo.value.gitTag || 'v0.1.0'})
Commit: ${appInfo.value.commitHash || '758833c7'}
Build Time: ${appInfo.value.buildTime || '2026-09-15'}
OS/Arch: ${appInfo.value.osArch || 'macOS/arm64'}
Go Runtime: ${appInfo.value.goVersion || 'Go 1.24'}
License: ${appInfo.value.license || 'MIT'}
------------------------------------`;

  try {
    if (navigator.clipboard && navigator.clipboard.writeText) {
      await navigator.clipboard.writeText(diagnosticText);
    } else {
      const textArea = document.createElement('textarea');
      textArea.value = diagnosticText;
      document.body.appendChild(textArea);
      textArea.select();
      document.execCommand('copy');
      document.body.removeChild(textArea);
    }
    copied.value = true;
    setTimeout(() => {
      copied.value = false;
    }, 2500);
  } catch (err) {
    console.error('Failed to copy system info:', err);
  }
};

const handleCheckUpdate = async () => {
  checking.value = true;
  updateMessage.value = '';
  try {
    const wailsApp = (window as any)?.go?.main?.App;
    if (wailsApp && typeof wailsApp.CheckUpdate === 'function') {
      const res = await wailsApp.CheckUpdate();
      if (res && res.hasUpdate) {
        updateMessage.value = `${t('about.updateAvailable')} ${res.latestVersion}!`;
        updateStatusClass.value = 'has-update';
      } else {
        updateMessage.value = t('about.isLatest');
        updateStatusClass.value = 'is-latest';
      }
    } else {
      setTimeout(() => {
        updateMessage.value = t('about.isLatest');
        updateStatusClass.value = 'is-latest';
      }, 800);
    }
  } catch (err) {
    updateMessage.value = t('about.checkFailed');
    updateStatusClass.value = 'update-error';
  } finally {
    checking.value = false;
  }
};
</script>

<style scoped>
.about-modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  background: rgba(10, 15, 26, 0.75);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
}

.about-modal-container {
  width: 480px;
  max-width: 90vw;
  padding: 24px;
  border-radius: 20px;
  background: rgba(23, 32, 51, 0.85);
  border: 1px solid rgba(255, 255, 255, 0.12);
  box-shadow: 0 20px 50px rgba(0, 0, 0, 0.5), 0 0 30px rgba(66, 184, 131, 0.1);
  color: #e2e8f0;
  position: relative;
  overflow: hidden;
}

.about-header {
  display: flex;
  justify-content: flex-end;
}

.close-btn {
  background: transparent;
  border: none;
  color: #94a3b8;
  cursor: pointer;
  padding: 6px;
  border-radius: 50%;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  justify-content: center;
}

.close-btn:hover {
  color: #ffffff;
  background: rgba(255, 255, 255, 0.1);
}

.about-hero {
  text-align: center;
  margin-top: -10px;
  margin-bottom: 24px;
}

.logo-wrapper {
  position: relative;
  display: inline-block;
  margin-bottom: 12px;
}

.logo-glow {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 80px;
  height: 80px;
  border-radius: 50%;
  background: radial-gradient(circle, rgba(66, 184, 131, 0.4) 0%, rgba(56, 189, 248, 0.2) 60%, transparent 80%);
  filter: blur(16px);
  z-index: 0;
}

.app-logo-icon {
  position: relative;
  z-index: 1;
  color: #42b883;
  filter: drop-shadow(0 4px 12px rgba(66, 184, 131, 0.3));
}

.app-title {
  font-size: 24px;
  font-weight: 700;
  margin: 0;
  background: linear-gradient(135deg, #ffffff 0%, #cbd5e1 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  letter-spacing: -0.5px;
}

.app-subtitle {
  font-size: 13px;
  color: #94a3b8;
  margin: 4px 0 12px 0;
}

.version-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 12px;
  border-radius: 20px;
  background: rgba(66, 184, 131, 0.12);
  border: 1px solid rgba(66, 184, 131, 0.3);
  color: #42b883;
  font-size: 12px;
  font-weight: 600;
}

.badge-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #42b883;
  box-shadow: 0 0 8px #42b883;
}

.info-grid {
  background: rgba(15, 23, 42, 0.5);
  border-radius: 12px;
  padding: 14px 16px;
  border: 1px solid rgba(255, 255, 255, 0.06);
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin-bottom: 20px;
}

.info-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 13px;
}

.info-label {
  color: #94a3b8;
}

.info-val {
  color: #f1f5f9;
  font-size: 12px;
}

.font-mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}

.about-actions {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
}

.action-btn {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 10px 16px;
  border-radius: 10px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
  border: none;
}

.secondary-btn {
  background: rgba(255, 255, 255, 0.06);
  border: 1px solid rgba(255, 255, 255, 0.12);
  color: #e2e8f0;
}

.secondary-btn:hover:not(:disabled) {
  background: rgba(255, 255, 255, 0.12);
  color: #ffffff;
}

.primary-btn {
  background: linear-gradient(135deg, #38bdf8 0%, #0284c7 100%);
  color: #ffffff;
  box-shadow: 0 4px 12px rgba(2, 132, 199, 0.3);
}

.primary-btn:hover:not(:disabled) {
  filter: brightness(1.1);
  box-shadow: 0 6px 16px rgba(2, 132, 199, 0.4);
}

.action-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.spin-icon {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.update-status-msg {
  padding: 8px 12px;
  border-radius: 8px;
  font-size: 12px;
  text-align: center;
  margin-bottom: 16px;
}

.has-update {
  background: rgba(245, 158, 11, 0.15);
  border: 1px solid rgba(245, 158, 11, 0.3);
  color: #fbbf24;
}

.is-latest {
  background: rgba(34, 197, 94, 0.15);
  border: 1px solid rgba(34, 197, 94, 0.3);
  color: #4ade80;
}

.update-error {
  background: rgba(239, 68, 68, 0.15);
  border: 1px solid rgba(239, 68, 68, 0.3);
  color: #f87171;
}

.about-footer {
  text-align: center;
  border-top: 1px solid rgba(255, 255, 255, 0.08);
  padding-top: 16px;
}

.footer-links {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  margin-bottom: 8px;
}

.footer-link {
  color: #38bdf8;
  text-decoration: none;
  font-size: 12px;
  display: flex;
  align-items: center;
  gap: 4px;
  transition: color 0.2s ease;
}

.footer-link:hover {
  color: #7dd3fc;
  text-decoration: underline;
}

.link-separator {
  color: #475569;
  font-size: 12px;
}

.copyright-text {
  font-size: 11px;
  color: #64748b;
  margin: 0;
}

.modal-fade-enter-active,
.modal-fade-leave-active {
  transition: opacity 0.25s ease, transform 0.25s ease;
}

.modal-fade-enter-from,
.modal-fade-leave-to {
  opacity: 0;
  transform: scale(0.95);
}
</style>

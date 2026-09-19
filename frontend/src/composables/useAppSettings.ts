import { ref, watch, type Ref } from 'vue';
import { setLanguage } from '../i18n';
import { logUserAction } from '../utils/logger';
import {
  GetConfig,
  SaveConfig,
  ReloadAppMenu,
} from '../../wailsjs/go/main/App';

export interface UseAppSettingsOptions {
  selectedFsType: Ref<'exFAT' | 'NTFS' | 'FAT32' | 'ext4'>;
  activeMode: Ref<'cloud' | 'hybrid'>;
  autoEjectAfterDeploy: Ref<boolean>;
}

export function useAppSettings(options: UseAppSettingsOptions) {
  const { selectedFsType, activeMode, autoEjectAfterDeploy } = options;

  const settingsInitialTab = ref<'general' | 'network' | 'uniboot' | 'ventoy'>('general');
  const isSettingsOpen = ref(false);
  const isAboutOpen = ref(false);
  const currentGithubProxy = ref('');

  watch(isAboutOpen, (val) => {
    if (val) {
      logUserAction('INFO', 'User opened About modal');
    }
  });

  function openSettings(tab: 'general' | 'network' | 'uniboot' | 'ventoy' = 'general') {
    settingsInitialTab.value = tab;
    isSettingsOpen.value = true;
    logUserAction('INFO', 'User opened settings modal', tab);
  }

  function applyTheme(themeName?: string) {
    const theme = themeName === 'light' ? 'light' : 'dark';
    document.documentElement.setAttribute('data-theme', theme);
    try {
      localStorage.setItem('unigo_theme_cache', theme);
    } catch (e) {
      // localStorage unavailable, ignore
    }
  }

  async function saveLangToConfig(langVal: string) {
    try {
      const cfg = await GetConfig();
      if (cfg) {
        cfg.language = langVal;
        await SaveConfig(cfg);
      }
    } catch (e) {
      console.error('Failed to save language config:', e);
    }
  }

  function selectLanguage(langVal: string) {
    setLanguage(langVal);
    saveLangToConfig(langVal);
    ReloadAppMenu(langVal).catch((err: any) => {
      console.warn('Failed to reload app menu:', err);
    });
  }

  async function loadConfig() {
    if (window.go && window.go.main && window.go.main.App) {
      try {
        const cfg = await window.go.main.App.GetConfig();
        if (cfg) {
          if (cfg.githubProxy) currentGithubProxy.value = cfg.githubProxy;
          if (cfg.fileSystem) selectedFsType.value = cfg.fileSystem as any;
          if (cfg.language) {
            setLanguage(cfg.language);
            if (window.go.main.App.ReloadAppMenu) {
              window.go.main.App.ReloadAppMenu(cfg.language).catch(() => {});
            }
          }
          applyTheme(cfg.theme);
          autoEjectAfterDeploy.value = cfg.autoEjectAfterDeploy === true;
        }
      } catch (e) {
        console.error('Failed to load config:', e);
      }
    }
  }

  async function onSaveSettings(payload: any) {
    let proxyUrl = '';
    if (typeof payload === 'string') {
      proxyUrl = payload;
      currentGithubProxy.value = payload;
    } else if (payload && typeof payload === 'object') {
      proxyUrl = payload.githubProxy || '';
      currentGithubProxy.value = proxyUrl;
      if (payload.fileSystem) selectedFsType.value = payload.fileSystem as any;
      if (payload.mode) activeMode.value = payload.mode as any;
      if (payload.theme) applyTheme(payload.theme);
      if (typeof payload.autoEjectAfterDeploy === 'boolean') autoEjectAfterDeploy.value = payload.autoEjectAfterDeploy;
    }

    if (window.go && window.go.main && window.go.main.App) {
      try {
        const configObj = typeof payload === 'object' && payload !== null ? {
          mode: payload.mode || activeMode.value,
          autoCheckUpdate: payload.autoCheckUpdate !== false,
          theme: payload.theme || 'dark',
          language: payload.language || 'auto',
          githubProxy: proxyUrl,
          fileSystem: payload.fileSystem || selectedFsType.value,
          proxyProtocol: payload.proxyProtocol || 'direct',
          proxyHost: payload.proxyHost || '',
          proxyPort: Number(payload.proxyPort) || 0,
          proxyUser: payload.proxyUser || '',
          proxyPassword: payload.proxyPassword || '',
          ventoyPath: payload.ventoyPath || '',
          ventoySecureBoot: payload.ventoySecureBoot !== false,
          ventoyPartitionStyle: payload.ventoyPartitionStyle || 'MBR',
          ventoyReserveSpace: Number(payload.ventoyReserveSpace) || 0,
          ventoyWin11Bypass: payload.ventoyWin11Bypass === true,
          ventoyMenuTimeout: Number(payload.ventoyMenuTimeout) || 0,
          autoEjectAfterDeploy: payload.autoEjectAfterDeploy === true,
        } : {
          mode: activeMode.value,
          autoCheckUpdate: true,
          theme: 'dark',
          githubProxy: proxyUrl,
          fileSystem: selectedFsType.value,
        };

        await window.go.main.App.SaveConfig(configObj as any);
        if (configObj.language && window.go.main.App.ReloadAppMenu) {
          await window.go.main.App.ReloadAppMenu(configObj.language);
        }
      } catch (e) {
        console.error('Failed to save config:', e);
      }
    }
  }

  return {
    settingsInitialTab,
    isSettingsOpen,
    isAboutOpen,
    currentGithubProxy,
    openSettings,
    applyTheme,
    selectLanguage,
    loadConfig,
    onSaveSettings,
  };
}

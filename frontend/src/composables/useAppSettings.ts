import { ref, watch, type Ref } from "vue";
import { GetConfig, ReloadAppMenu, SaveConfig } from "../../wailsjs/go/main/App";
import { selectedLangSetting, setLanguage } from "../i18n";
import { logUserAction } from "../utils/logger";
import { useTheme } from "./useTheme";

export interface UseAppSettingsOptions {
  selectedFsType: Ref<"exFAT" | "NTFS" | "FAT32" | "ext4">;
  activeMode: Ref<"cloud" | "hybrid">;
  autoEjectAfterDeploy: Ref<boolean>;
}

export function useAppSettings(options: UseAppSettingsOptions) {
  const { selectedFsType, activeMode, autoEjectAfterDeploy } = options;

  const { currentTheme, effectiveTheme, applyTheme, getActiveTheme, toggleTheme } = useTheme();
  const settingsInitialTab = ref<"general" | "network" | "uniboot" | "ventoy">("general");
  const isSettingsOpen = ref(false);
  const isAboutOpen = ref(false);
  const currentGithubProxy = ref("");

  watch(isAboutOpen, (val) => {
    if (val) {
      logUserAction("INFO", "User opened About modal");
    }
  });

  function openSettings(tab: "general" | "network" | "uniboot" | "ventoy" = "general") {
    settingsInitialTab.value = tab;
    isSettingsOpen.value = true;
    logUserAction("INFO", "User opened settings modal", tab);
  }

  async function saveLangToConfig(langVal: string) {
    try {
      const cfg = await GetConfig();
      if (cfg) {
        cfg.language = langVal;
        await SaveConfig(cfg);
      }
    } catch (e) {
      console.error("Failed to save language config:", e);
    }
  }

  function selectLanguage(langVal: string) {
    setLanguage(langVal);
    saveLangToConfig(langVal);
    ReloadAppMenu(langVal).catch((err: any) => {
      console.warn("Failed to reload app menu:", err);
    });
  }

  async function loadConfig() {
    let cfg: any = null;

    try {
      if (window.go && window.go.main && window.go.main.App) {
        try {
          cfg = await window.go.main.App.GetConfig();
        } catch (e) {
          console.error("Failed to load config:", e);
        }
      }

      if (!cfg) {
        // Retry in case Wails IPC is still initializing when Vue mounts
        for (let i = 0; i < 15; i++) {
          await new Promise((resolve) => setTimeout(resolve, 60));
          if (window.go && window.go.main && window.go.main.App) {
            try {
              cfg = await window.go.main.App.GetConfig();
              if (cfg) break;
            } catch (e) {
              // Wait for next attempt
            }
          }
        }
      }

      if (cfg) {
        if (cfg.githubProxy) currentGithubProxy.value = cfg.githubProxy;
        if (cfg.fileSystem) selectedFsType.value = cfg.fileSystem as any;
        if (cfg.language) {
          await setLanguage(cfg.language);
          if (window.go?.main?.App?.ReloadAppMenu) {
            window.go.main.App.ReloadAppMenu(cfg.language).catch(() => {});
          }
        }
        if (cfg.theme) {
          applyTheme(cfg.theme);
        }
        autoEjectAfterDeploy.value = cfg.autoEjectAfterDeploy === true;
      }
    } finally {
      window.dispatchEvent(new Event("uniboot:config-ready"));
    }
  }

  async function onSaveSettings(payload: any) {
    let proxyUrl = "";
    if (typeof payload === "string") {
      proxyUrl = payload;
      currentGithubProxy.value = payload;
    } else if (payload && typeof payload === "object") {
      if (typeof payload.githubProxy === "string") {
        proxyUrl = payload.githubProxy;
        currentGithubProxy.value = proxyUrl;
      }
      if (payload.fileSystem) selectedFsType.value = payload.fileSystem as any;
      if (payload.mode) activeMode.value = payload.mode as any;
      if (payload.theme) applyTheme(payload.theme);
      if (payload.language) await setLanguage(payload.language);
      if (typeof payload.autoEjectAfterDeploy === "boolean") autoEjectAfterDeploy.value = payload.autoEjectAfterDeploy;
    }

    if (window.go && window.go.main && window.go.main.App) {
      try {
        let currentCfg: any = null;
        try {
          currentCfg = await window.go.main.App.GetConfig();
        } catch {
          // ignore
        }

        const isDirect =
          payload && payload.proxyProtocol
            ? payload.proxyProtocol === "direct"
            : (currentCfg?.proxyProtocol || "direct") === "direct";
        const host =
          payload && payload.proxyHost !== undefined ? payload.proxyHost.trim() : currentCfg?.proxyHost || "";
        const port = !isDirect && host ? Number(payload?.proxyPort ?? currentCfg?.proxyPort) || 0 : 0;

        const effectiveTheme =
          payload && payload.theme !== undefined
            ? payload.theme
            : currentCfg?.theme || currentTheme.value || "system";
        const effectiveLang =
          payload && payload.language !== undefined
            ? payload.language
            : currentCfg?.language || selectedLangSetting.value || "auto";

        const configObj = {
          ...(currentCfg || {}),
          mode: payload?.mode || currentCfg?.mode || activeMode.value,
          autoCheckUpdate:
            typeof payload?.autoCheckUpdate === "boolean"
              ? payload.autoCheckUpdate
              : currentCfg?.autoCheckUpdate !== false,
          theme: effectiveTheme,
          language: effectiveLang,
          githubProxy:
            typeof payload?.githubProxy === "string"
              ? payload.githubProxy
              : currentCfg?.githubProxy || currentGithubProxy.value || proxyUrl,
          fileSystem: payload?.fileSystem || currentCfg?.fileSystem || selectedFsType.value,
          proxyProtocol: payload?.proxyProtocol || currentCfg?.proxyProtocol || "direct",
          proxyHost: host,
          proxyPort: port,
          proxyUser: payload?.proxyUser !== undefined ? payload.proxyUser : currentCfg?.proxyUser || "",
          proxyPassword: payload?.proxyPassword !== undefined ? payload.proxyPassword : currentCfg?.proxyPassword || "",
          ventoyPath: payload?.ventoyPath !== undefined ? payload.ventoyPath : currentCfg?.ventoyPath || "",
          unibootPath: payload?.unibootPath !== undefined ? payload.unibootPath : currentCfg?.unibootPath || "",
          ventoySecureBoot:
            typeof payload?.ventoySecureBoot === "boolean"
              ? payload.ventoySecureBoot
              : currentCfg?.ventoySecureBoot !== false,
          ventoyPartitionStyle: payload?.ventoyPartitionStyle || currentCfg?.ventoyPartitionStyle || "MBR",
          ventoyReserveSpace:
            payload?.ventoyReserveSpace !== undefined
              ? Number(payload.ventoyReserveSpace) || 0
              : currentCfg?.ventoyReserveSpace || 0,
          ventoyWin11Bypass:
            typeof payload?.ventoyWin11Bypass === "boolean"
              ? payload.ventoyWin11Bypass
              : currentCfg?.ventoyWin11Bypass === true,
          ventoyMenuTimeout:
            payload?.ventoyMenuTimeout !== undefined
              ? Number(payload.ventoyMenuTimeout) || 0
              : currentCfg?.ventoyMenuTimeout || 0,
          ventoySecondaryMenu:
            typeof payload?.ventoySecondaryMenu === "boolean"
              ? payload.ventoySecondaryMenu
              : currentCfg?.ventoySecondaryMenu === true,
          autoEjectAfterDeploy:
            typeof payload?.autoEjectAfterDeploy === "boolean"
              ? payload.autoEjectAfterDeploy
              : currentCfg?.autoEjectAfterDeploy === true,
        };

        await window.go.main.App.SaveConfig(configObj as any);
        if (configObj.language && window.go.main.App.ReloadAppMenu) {
          await window.go.main.App.ReloadAppMenu(configObj.language);
        }
      } catch (e) {
        console.error("Failed to save config:", e);
      } finally {
        window.dispatchEvent(new Event("uniboot:config-ready"));
      }
    }
  }

  return {
    settingsInitialTab,
    isSettingsOpen,
    isAboutOpen,
    currentGithubProxy,
    currentTheme,
    effectiveTheme,
    applyTheme,
    getActiveTheme,
    toggleTheme,
    openSettings,
    selectLanguage,
    loadConfig,
    onSaveSettings,
  };
}

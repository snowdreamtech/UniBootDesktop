import { ref, watch, type Ref } from "vue";
import { selectedLangSetting, setLanguage, SUPPORTED_LANGUAGES } from "../i18n";
import { logUserAction } from "../utils/logger";
import { useTheme, type AppTheme } from "./useTheme";

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

  async function selectLanguage(langVal: string) {
    await setLanguage(langVal);
    await onSaveSettings({
      language: langVal,
      theme: currentTheme.value,
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

        const savedTheme = (typeof localStorage !== "undefined" && localStorage.getItem("uniboot_theme_cache")) as any;
        const hasConcreteCfgTheme = cfg.theme === "light" || cfg.theme === "dark";
        const hasConcreteSavedTheme = savedTheme === "light" || savedTheme === "dark";

        let resolvedTheme: AppTheme = "system";
        if (hasConcreteCfgTheme) {
          resolvedTheme = cfg.theme;
        } else if (hasConcreteSavedTheme) {
          resolvedTheme = savedTheme;
        } else if (cfg.theme === "system" || savedTheme === "system") {
          resolvedTheme = "system";
        }
        applyTheme(resolvedTheme);

        const savedLang = typeof localStorage !== "undefined" ? localStorage.getItem("uniboot_locale") : null;
        const hasConcreteCfgLang = Boolean(
          cfg.language && cfg.language !== "auto" && SUPPORTED_LANGUAGES.some((l) => l.code === cfg.language)
        );
        const hasConcreteSavedLang = Boolean(
          savedLang && savedLang !== "auto" && SUPPORTED_LANGUAGES.some((l) => l.code === savedLang)
        );

        let resolvedLang = "auto";
        if (hasConcreteCfgLang && cfg.language) {
          resolvedLang = cfg.language;
        } else if (hasConcreteSavedLang && savedLang) {
          resolvedLang = savedLang;
        } else if (cfg.language && cfg.language !== "") {
          resolvedLang = cfg.language;
        } else if (savedLang && savedLang !== "") {
          resolvedLang = savedLang;
        } else {
          resolvedLang = "zh-CN";
        }

        await setLanguage(resolvedLang);
        if (window.go?.main?.App?.ReloadAppMenu) {
          window.go.main.App.ReloadAppMenu(resolvedLang).catch(() => {});
        }

        // If backend config was missing or defaulted while user had a concrete choice saved in localStorage, sync it back to backend atomically
        if ((!hasConcreteCfgTheme && hasConcreteSavedTheme) || (!hasConcreteCfgLang && hasConcreteSavedLang)) {
          cfg.theme = resolvedTheme;
          cfg.language = resolvedLang;
          if (window.go?.main?.App?.SaveConfig) {
            window.go.main.App.SaveConfig(cfg).catch((e: any) => console.error("Initial config sync failed:", e));
          }
        }

        autoEjectAfterDeploy.value = cfg.autoEjectAfterDeploy === true;
      }
    } finally {
      if (typeof window !== "undefined" && typeof window.dispatchEvent === "function") {
        window.dispatchEvent(new Event("uniboot:config-ready"));
      }
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

        const savedTheme = (typeof localStorage !== "undefined" && localStorage.getItem("uniboot_theme_cache")) as any;
        const savedLang = typeof localStorage !== "undefined" ? localStorage.getItem("uniboot_locale") : null;

        const effectiveTheme =
          payload && payload.theme !== undefined
            ? payload.theme
            : currentTheme.value || savedTheme || currentCfg?.theme || "system";
        const effectiveLang =
          payload && payload.language !== undefined
            ? payload.language
            : selectedLangSetting.value || savedLang || currentCfg?.language || "auto";

        if (effectiveTheme && effectiveTheme !== currentTheme.value) {
          applyTheme(effectiveTheme);
        }
        if (effectiveLang && effectiveLang !== selectedLangSetting.value) {
          await setLanguage(effectiveLang);
        }

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
        if (typeof window !== "undefined" && typeof window.dispatchEvent === "function") {
          window.dispatchEvent(new Event("uniboot:config-ready"));
        }
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

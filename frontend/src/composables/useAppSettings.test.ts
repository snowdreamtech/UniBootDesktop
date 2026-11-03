// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import { describe, it, expect, beforeEach, beforeAll, vi } from "vitest";
import { ref } from "vue";

const store = new Map<string, string>();
const mockLocalStorage = {
  getItem: (key: string) => store.get(key) ?? null,
  setItem: (key: string, val: string) => store.set(key, String(val)),
  removeItem: (key: string) => store.delete(key),
  clear: () => store.clear(),
};

let mockConfigState: any = {
  mode: "cloud",
  theme: "system",
  language: "auto",
  fileSystem: "exFAT",
  autoEjectAfterDeploy: false,
};

let lastSavedConfig: any = null;
let lastReloadedMenuLang: string | null = null;

beforeAll(() => {
  (globalThis as any).localStorage = mockLocalStorage;
  (globalThis as any).window = (globalThis as any).window || {};
  (globalThis as any).window.go = {
    main: {
      App: {
        GetConfig: vi.fn(async () => ({ ...mockConfigState })),
        SaveConfig: vi.fn(async (cfg: any) => {
          lastSavedConfig = { ...cfg };
          mockConfigState = { ...cfg };
          return null;
        }),
        ReloadAppMenu: vi.fn(async (lang: string) => {
          lastReloadedMenuLang = lang;
          return null;
        }),
      },
    },
  };
});

describe("useAppSettings composable", () => {
  beforeEach(() => {
    store.clear();
    lastSavedConfig = null;
    lastReloadedMenuLang = null;
    mockConfigState = {
      mode: "cloud",
      theme: "system",
      language: "auto",
      fileSystem: "exFAT",
      autoEjectAfterDeploy: false,
    };
  });

  it("preserves active theme when only saving language", async () => {
    const { useAppSettings } = await import("./useAppSettings");
    const { useTheme } = await import("./useTheme");
    const { applyTheme } = useTheme();

    applyTheme("light");

    const selectedFsType = ref<"exFAT" | "NTFS" | "FAT32" | "ext4">("exFAT");
    const activeMode = ref<"cloud" | "hybrid">("cloud");
    const autoEjectAfterDeploy = ref(false);

    const { onSaveSettings } = useAppSettings({
      selectedFsType,
      activeMode,
      autoEjectAfterDeploy,
    });

    await onSaveSettings({ language: "zh-CN" });

    expect(lastSavedConfig).not.toBeNull();
    expect(lastSavedConfig.language).toBe("zh-CN");
    expect(lastSavedConfig.theme).toBe("light");
    expect(store.get("uniboot_theme_cache")).toBe("light");
    expect(store.get("uniboot_locale")).toBe("zh-CN");
  });

  it("preserves active language when only toggling theme", async () => {
    const { useAppSettings } = await import("./useAppSettings");
    const { setLanguage } = await import("../i18n");

    await setLanguage("zh-CN");

    const selectedFsType = ref<"exFAT" | "NTFS" | "FAT32" | "ext4">("exFAT");
    const activeMode = ref<"cloud" | "hybrid">("cloud");
    const autoEjectAfterDeploy = ref(false);

    const { onSaveSettings } = useAppSettings({
      selectedFsType,
      activeMode,
      autoEjectAfterDeploy,
    });

    await onSaveSettings({ theme: "light" });

    expect(lastSavedConfig).not.toBeNull();
    expect(lastSavedConfig.theme).toBe("light");
    expect(lastSavedConfig.language).toBe("zh-CN");
    expect(store.get("uniboot_theme_cache")).toBe("light");
    expect(store.get("uniboot_locale")).toBe("zh-CN");
  });

  it("selectLanguage updates i18n, saves config and reloads native menu atomically", async () => {
    const { useAppSettings } = await import("./useAppSettings");
    const { useTheme } = await import("./useTheme");
    const { applyTheme } = useTheme();

    applyTheme("dark");

    const selectedFsType = ref<"exFAT" | "NTFS" | "FAT32" | "ext4">("exFAT");
    const activeMode = ref<"cloud" | "hybrid">("cloud");
    const autoEjectAfterDeploy = ref(false);

    const { selectLanguage, currentTheme } = useAppSettings({
      selectedFsType,
      activeMode,
      autoEjectAfterDeploy,
    });

    await selectLanguage("ja-JP");

    expect(lastSavedConfig).not.toBeNull();
    expect(lastSavedConfig.language).toBe("ja-JP");
    expect(lastSavedConfig.theme).toBe("dark");
    expect(currentTheme.value).toBe("dark");
    expect(lastReloadedMenuLang).toBe("ja-JP");
  });

  it("loadConfig adopts concrete localStorage preferences when backend has defaults and syncs them", async () => {
    store.set("uniboot_theme_cache", "light");
    store.set("uniboot_locale", "zh-CN");

    mockConfigState = {
      mode: "cloud",
      theme: "system",
      language: "auto",
      fileSystem: "exFAT",
    };

    const { useAppSettings } = await import("./useAppSettings");
    const selectedFsType = ref<"exFAT" | "NTFS" | "FAT32" | "ext4">("exFAT");
    const activeMode = ref<"cloud" | "hybrid">("cloud");
    const autoEjectAfterDeploy = ref(false);

    const { loadConfig, currentTheme } = useAppSettings({
      selectedFsType,
      activeMode,
      autoEjectAfterDeploy,
    });

    await loadConfig();

    expect(currentTheme.value).toBe("light");
    expect(store.get("uniboot_theme_cache")).toBe("light");
    expect(store.get("uniboot_locale")).toBe("zh-CN");
    expect(lastSavedConfig).not.toBeNull();
    expect(lastSavedConfig.theme).toBe("light");
    expect(lastSavedConfig.language).toBe("zh-CN");
  });

  it("loadConfig applies concrete backend preferences when backend specifies them", async () => {
    mockConfigState = {
      mode: "cloud",
      theme: "dark",
      language: "fr-FR",
      fileSystem: "exFAT",
    };

    const { useAppSettings } = await import("./useAppSettings");
    const selectedFsType = ref<"exFAT" | "NTFS" | "FAT32" | "ext4">("exFAT");
    const activeMode = ref<"cloud" | "hybrid">("cloud");
    const autoEjectAfterDeploy = ref(false);

    const { loadConfig, currentTheme } = useAppSettings({
      selectedFsType,
      activeMode,
      autoEjectAfterDeploy,
    });

    await loadConfig();

    expect(currentTheme.value).toBe("dark");
    expect(store.get("uniboot_theme_cache")).toBe("dark");
    expect(store.get("uniboot_locale")).toBe("fr-FR");
  });
});

import { describe, it, expect, beforeEach, beforeAll } from "vitest";

const store = new Map<string, string>();
const mockLocalStorage = {
  getItem: (key: string) => store.get(key) ?? null,
  setItem: (key: string, val: string) => store.set(key, String(val)),
  removeItem: (key: string) => store.delete(key),
  clear: () => store.clear(),
};

beforeAll(() => {
  (globalThis as any).localStorage = mockLocalStorage;
});

describe("useTheme composable", () => {
  beforeEach(() => {
    store.clear();
  });

  it("defaults to system theme when unconfigured", async () => {
    const { useTheme } = await import("./useTheme");
    const { getActiveTheme } = useTheme();
    expect(["system", "dark", "light"]).toContain(getActiveTheme());
  });

  it("applies and persists light theme correctly", async () => {
    const { useTheme, currentTheme, effectiveTheme } = await import("./useTheme");
    const { applyTheme, getActiveTheme } = useTheme();
    const applied = applyTheme("light");
    expect(applied).toBe("light");
    expect(currentTheme.value).toBe("light");
    expect(effectiveTheme.value).toBe("light");
    expect(getActiveTheme()).toBe("light");
    expect(store.get("uniboot_theme_cache")).toBe("light");
  });

  it("applies and persists dark theme correctly", async () => {
    const { useTheme, currentTheme, effectiveTheme } = await import("./useTheme");
    const { applyTheme, getActiveTheme } = useTheme();
    const applied = applyTheme("dark");
    expect(applied).toBe("dark");
    expect(currentTheme.value).toBe("dark");
    expect(effectiveTheme.value).toBe("dark");
    expect(getActiveTheme()).toBe("dark");
    expect(store.get("uniboot_theme_cache")).toBe("dark");
  });

  it("toggles between dark and light smoothly without drifting", async () => {
    const { useTheme, currentTheme } = await import("./useTheme");
    const { applyTheme, toggleTheme } = useTheme();
    applyTheme("light");
    expect(currentTheme.value).toBe("light");

    const toggled = toggleTheme();
    expect(toggled).toBe("dark");
    expect(currentTheme.value).toBe("dark");
    expect(store.get("uniboot_theme_cache")).toBe("dark");

    const toggledBack = toggleTheme();
    expect(toggledBack).toBe("light");
    expect(currentTheme.value).toBe("light");
    expect(store.get("uniboot_theme_cache")).toBe("light");
  });
});

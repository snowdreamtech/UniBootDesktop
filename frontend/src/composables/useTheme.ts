// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import { ref, computed } from "vue";
import { SetTheme } from "../../wailsjs/go/main/App";

export type AppTheme = "dark" | "light" | "system";

function getSystemPreferredTheme(): "dark" | "light" {
  if (typeof window !== "undefined" && window.matchMedia) {
    return window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light";
  }
  return "dark";
}

function resolveEffectiveTheme(theme: AppTheme): "dark" | "light" {
  if (theme === "system") {
    return getSystemPreferredTheme();
  }
  return theme === "light" ? "light" : "dark";
}

function getInitialTheme(): AppTheme {
  try {
    const cached = localStorage.getItem("uniboot_theme_cache");
    if (cached === "light" || cached === "dark" || cached === "system") {
      return cached as AppTheme;
    }
  } catch (e) {
    // localStorage may be unavailable
  }

  return "system";
}

const currentTheme = ref<AppTheme>(getInitialTheme());
const effectiveTheme = computed<"dark" | "light">(() => resolveEffectiveTheme(currentTheme.value));

function updateDomAndWindow(resolved: "dark" | "light") {
  if (typeof document !== "undefined") {
    document.documentElement.setAttribute("data-theme", resolved);
  }

  if (typeof window !== "undefined") {
    if ((window as any).runtime) {
      try {
        if (
          currentTheme.value === "system" &&
          typeof (window as any).runtime.WindowSetSystemDefaultTheme === "function"
        ) {
          (window as any).runtime.WindowSetSystemDefaultTheme();
        } else if (resolved === "dark") {
          if (typeof (window as any).runtime.WindowSetDarkTheme === "function") {
            (window as any).runtime.WindowSetDarkTheme();
          }
        } else {
          if (typeof (window as any).runtime.WindowSetLightTheme === "function") {
            (window as any).runtime.WindowSetLightTheme();
          }
        }

        if (typeof (window as any).runtime.WindowSetBackgroundColour === "function") {
          if (resolved === "dark") {
            (window as any).runtime.WindowSetBackgroundColour(11, 15, 25, 255);
          } else {
            (window as any).runtime.WindowSetBackgroundColour(248, 250, 252, 255);
          }
        }
      } catch (e) {
        // Ignore errors in non-wails environment
      }
    }

    // Trigger backend SetTheme to force native Windows/DWM frame recalculation and title bar update
    try {
      if (typeof SetTheme === "function") {
        SetTheme(currentTheme.value).catch(() => {});
      } else if (typeof (window as any).go?.main?.App?.SetTheme === "function") {
        (window as any).go.main.App.SetTheme(currentTheme.value);
      }
    } catch (e) {
      // Ignore errors in non-wails environment
    }
  }
}

export function syncWindowTheme() {
  updateDomAndWindow(resolveEffectiveTheme(currentTheme.value));
}

// Ensure DOM and window theme are synchronized immediately upon script evaluation
if (typeof document !== "undefined") {
  updateDomAndWindow(resolveEffectiveTheme(currentTheme.value));
}

// Global listener for system theme changes (e.g. macOS appearance toggle or sunset/sunrise)
if (typeof window !== "undefined" && window.matchMedia) {
  const mql = window.matchMedia("(prefers-color-scheme: dark)");
  const handleSystemThemeChange = (e: MediaQueryListEvent | MediaQueryList) => {
    if (currentTheme.value === "system") {
      const nextResolved = e.matches ? "dark" : "light";
      updateDomAndWindow(nextResolved);
    }
  };

  if (typeof mql.addEventListener === "function") {
    mql.addEventListener("change", handleSystemThemeChange);
  } else if (typeof (mql as any).addListener === "function") {
    (mql as any).addListener(handleSystemThemeChange);
  }
}

export function useTheme() {
  function applyTheme(themeName?: string): AppTheme {
    let theme: AppTheme = "system";
    if (themeName === "light" || themeName === "dark" || themeName === "system") {
      theme = themeName;
    }

    currentTheme.value = theme;
    try {
      localStorage.setItem("uniboot_theme_cache", theme);
    } catch (e) {
      // localStorage may be unavailable
    }

    const resolved = resolveEffectiveTheme(theme);
    updateDomAndWindow(resolved);
    return theme;
  }

  function getActiveTheme(): AppTheme {
    return currentTheme.value;
  }

  function toggleTheme(): AppTheme {
    const active = resolveEffectiveTheme(currentTheme.value);
    const next: AppTheme = active === "dark" ? "light" : "dark";
    applyTheme(next);
    return next;
  }

  return {
    currentTheme,
    effectiveTheme,
    applyTheme,
    getActiveTheme,
    toggleTheme,
  };
}

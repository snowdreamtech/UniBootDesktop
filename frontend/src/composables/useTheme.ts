import { ref } from "vue";

export type AppTheme = "dark" | "light";

function getInitialTheme(): AppTheme {
  if (typeof document !== "undefined") {
    const domTheme = document.documentElement.getAttribute("data-theme");
    if (domTheme === "light" || domTheme === "dark") {
      return domTheme;
    }
  }

  try {
    const cached = localStorage.getItem("uniboot_theme_cache");
    if (cached === "light" || cached === "dark") {
      return cached;
    }
  } catch (e) {
    // localStorage may be unavailable
  }

  if (typeof window !== "undefined" && window.matchMedia("(prefers-color-scheme: light)").matches) {
    return "light";
  }

  return "dark";
}

const currentTheme = ref<AppTheme>(getInitialTheme());

export function useTheme() {
  function applyTheme(themeName?: string): AppTheme {
    const theme: AppTheme = themeName === "light" ? "light" : "dark";
    currentTheme.value = theme;
    if (typeof document !== "undefined") {
      document.documentElement.setAttribute("data-theme", theme);
    }
    try {
      localStorage.setItem("uniboot_theme_cache", theme);
    } catch (e) {
      // localStorage may be unavailable
    }

    if (typeof window !== "undefined" && (window as any).runtime) {
      try {
        if (theme === "dark") {
          if (typeof (window as any).runtime.WindowSetDarkTheme === "function") {
            (window as any).runtime.WindowSetDarkTheme();
          }
          if (typeof (window as any).runtime.WindowSetBackgroundColour === "function") {
            (window as any).runtime.WindowSetBackgroundColour(11, 15, 25, 255);
          }
        } else {
          if (typeof (window as any).runtime.WindowSetLightTheme === "function") {
            (window as any).runtime.WindowSetLightTheme();
          }
          if (typeof (window as any).runtime.WindowSetBackgroundColour === "function") {
            (window as any).runtime.WindowSetBackgroundColour(248, 250, 252, 255);
          }
        }
      } catch (e) {
        // Ignore errors in non-wails environment
      }
    }

    return theme;
  }

  function getActiveTheme(): AppTheme {
    if (typeof document !== "undefined") {
      const domTheme = document.documentElement.getAttribute("data-theme");
      if (domTheme === "light" || domTheme === "dark") {
        currentTheme.value = domTheme;
        return domTheme;
      }
    }
    return currentTheme.value;
  }

  return {
    currentTheme,
    applyTheme,
    getActiveTheme,
  };
}

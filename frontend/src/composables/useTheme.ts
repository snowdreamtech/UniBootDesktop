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
    const cached = localStorage.getItem("unigo_theme_cache");
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
      localStorage.setItem("unigo_theme_cache", theme);
    } catch (e) {
      // localStorage may be unavailable
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

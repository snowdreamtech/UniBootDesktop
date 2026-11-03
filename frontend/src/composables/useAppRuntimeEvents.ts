import { onMounted, onUnmounted } from "vue";
import { isWailsRuntime } from "../utils/wails";
import { syncWindowTheme } from "./useTheme";

export interface UseAppRuntimeEventsOptions {
  openAbout: () => void;
  openSettings: () => void;
  onThemeChanged?: (theme: string) => void;
}

export function useAppRuntimeEvents(options: UseAppRuntimeEventsOptions) {
  const { openAbout, openSettings, onThemeChanged } = options;

  function handleKeyDown(e: KeyboardEvent) {
    // Cmd+, or Ctrl+, to open settings
    if ((e.metaKey || e.ctrlKey) && e.key === ",") {
      e.preventDefault();
      openSettings();
    }
  }

  onMounted(() => {
    syncWindowTheme();
    window.addEventListener("keydown", handleKeyDown);

    if (isWailsRuntime() && window.runtime && typeof window.runtime.EventsOn === "function") {
      window.runtime.EventsOn("open-about-modal", () => {
        openAbout();
      });
      window.runtime.EventsOn("open-settings-modal", () => {
        openSettings();
      });
      if (onThemeChanged) {
        window.runtime.EventsOn("theme-changed", (theme: string) => {
          onThemeChanged(theme);
        });
      }
    }
  });

  onUnmounted(() => {
    window.removeEventListener("keydown", handleKeyDown);
    if (isWailsRuntime() && window.runtime && typeof window.runtime.EventsOff === "function") {
      window.runtime.EventsOff("open-about-modal");
      window.runtime.EventsOff("open-settings-modal");
      if (onThemeChanged) {
        window.runtime.EventsOff("theme-changed");
      }
    }
  });
}

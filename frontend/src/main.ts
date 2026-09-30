import { createApp } from "vue";
import App from "./App.vue";
import "./styles/theme.css";

const STARTUP_LOADER_MIN_DURATION_MS = 650;
const startupStartedAt = performance.now();

function setStartupProgress(progress: number) {
  document.documentElement.style.setProperty("--startup-progress", `${progress}%`);
}

function dismissStartupLoader() {
  const elapsed = performance.now() - startupStartedAt;
  const remaining = Math.max(0, STARTUP_LOADER_MIN_DURATION_MS - elapsed);
  setStartupProgress(92);

  window.setTimeout(() => {
    setStartupProgress(100);
    document.documentElement.classList.add("app-ready");
    window.setTimeout(() => {
      document.getElementById("startup-loader")?.remove();
    }, 260);
  }, remaining);
}

window.addEventListener("uniboot:config-ready", dismissStartupLoader, { once: true });

// Fallback safety timeout: ensure startup loader is dismissed if event doesn't fire
window.setTimeout(() => {
  if (!document.documentElement.classList.contains("app-ready")) {
    dismissStartupLoader();
  }
}, 2500);

// Prevent accidental external file drop from navigating away
window.addEventListener("dragover", (e: DragEvent) => e.preventDefault(), false);
window.addEventListener("drop", (e: DragEvent) => e.preventDefault(), false);

// Native Window Focus / Blur State Adaptation
window.addEventListener("focus", () => {
  document.documentElement.classList.remove("window-inactive");
});
window.addEventListener("blur", () => {
  document.documentElement.classList.add("window-inactive");
});

setStartupProgress(12);
createApp(App).mount("#app");
setStartupProgress(36);

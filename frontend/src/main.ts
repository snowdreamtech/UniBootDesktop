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

// Prevent default browser context menu except on text inputs and textareas
window.addEventListener("contextmenu", (e: MouseEvent) => {
  const target = e.target as HTMLElement | null;
  if (target && (target.tagName === "INPUT" || target.tagName === "TEXTAREA" || target.isContentEditable)) {
    return;
  }
  e.preventDefault();
});

// Prevent browser zoom shortcuts (Cmd/Ctrl + +/-/0) and pinch/wheel zoom
window.addEventListener(
  "wheel",
  (e: WheelEvent) => {
    if (e.ctrlKey || e.metaKey) {
      e.preventDefault();
    }
  },
  { passive: false }
);

window.addEventListener("keydown", (e: KeyboardEvent) => {
  if ((e.ctrlKey || e.metaKey) && (e.key === "=" || e.key === "-" || e.key === "0" || e.key === "+")) {
    e.preventDefault();
  }
});

// Native Window Focus / Blur State Adaptation
window.addEventListener("focus", () => {
  document.documentElement.classList.remove("window-inactive");
});
window.addEventListener("blur", () => {
  document.documentElement.classList.add("window-inactive");
});

import { isClickOnScrollbar, triggerNativeDrag } from "./utils/windowDrag";

// Enable native window dragging when dragging on background blank areas
window.addEventListener("mousedown", (e: MouseEvent) => {
  // Only trigger on primary mouse button single clicks
  if (e.buttons !== 1 || e.detail > 1) {
    return;
  }

  // 1. Strictly prevent dragging when clicking any scrollbar (viewport or container level)
  if (isClickOnScrollbar(e)) {
    return;
  }

  const target = e.target as HTMLElement | null;
  if (!target) return;

  // 2. Do not drag if interacting with buttons, inputs, links, list items, terminals, modals, etc.
  const interactiveSelector = [
    "button",
    "input",
    "textarea",
    "select",
    "option",
    "a",
    "pre",
    "code",
    "label",
    "dialog",
    ".modal-overlay",
    ".modal-card",
    ".modal-body",
    ".modal-container",
    ".glass-modal",
    ".disk-item",
    ".iso-file-item",
    ".mode-tabs",
    ".lang-selector-header",
    ".lang-dropdown-menu",
    ".terminal-body",
    ".terminal-container",
    ".log-line",
    ".no-drag",
    "[contenteditable='true']",
    "[role='button']",
    "[role='checkbox']",
    "[role='radio']",
  ].join(",");

  if (target.closest(interactiveSelector)) {
    return;
  }

  // 3. Do not drag if user is selecting text
  const selection = window.getSelection();
  if (selection && selection.toString().length > 0 && selection.containsNode(target, true)) {
    return;
  }

  triggerNativeDrag();
});

setStartupProgress(12);
createApp(App).mount("#app");
setStartupProgress(36);

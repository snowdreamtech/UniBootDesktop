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

window.addEventListener("unigo:config-ready", dismissStartupLoader, { once: true });

// Fallback safety timeout: ensure startup loader is dismissed if event doesn't fire
window.setTimeout(() => {
  if (!document.documentElement.classList.contains("app-ready")) {
    dismissStartupLoader();
  }
}, 2500);

// 1. Prevent default browser context menu on non-editable UI elements
window.addEventListener("contextmenu", (e: MouseEvent) => {
  const target = e.target as HTMLElement | null;
  if (target && (target.tagName === "INPUT" || target.tagName === "TEXTAREA" || target.isContentEditable)) {
    return;
  }
  e.preventDefault();
});

// 2. Prevent accidental browser reload, find, print, and save shortcuts
window.addEventListener("keydown", (e: KeyboardEvent) => {
  const key = e.key.toLowerCase();
  const isCmdOrCtrl = e.metaKey || e.ctrlKey;

  // Prevent reload: F5, Ctrl+R, Cmd+R
  if (e.key === "F5" || (isCmdOrCtrl && key === "r")) {
    e.preventDefault();
  }
  // Prevent browser in-page search: Ctrl+F, Cmd+F
  if (isCmdOrCtrl && key === "f") {
    e.preventDefault();
  }
  // Prevent browser print: Ctrl+P, Cmd+P
  if (isCmdOrCtrl && key === "p") {
    e.preventDefault();
  }
  // Prevent browser save webpage: Ctrl+S, Cmd+S
  if (isCmdOrCtrl && key === "s") {
    e.preventDefault();
  }
  // Prevent browser history navigation: Alt+Left/Right, Cmd+[/]
  if (
    (e.altKey && (e.key === "ArrowLeft" || e.key === "ArrowRight")) ||
    (e.metaKey && (e.key === "[" || e.key === "]"))
  ) {
    e.preventDefault();
  }
});

// 3. Prevent mouse navigation side-buttons (Back/Forward) from navigating webview
window.addEventListener("mousedown", (e: MouseEvent) => {
  if (e.button === 3 || e.button === 4) {
    e.preventDefault();
    e.stopPropagation();
  }
});
window.addEventListener("mouseup", (e: MouseEvent) => {
  if (e.button === 3 || e.button === 4) {
    e.preventDefault();
    e.stopPropagation();
  }
});

// 4. Prevent pinch-to-zoom and Ctrl/Cmd + wheel zooming
window.addEventListener(
  "wheel",
  (e: WheelEvent) => {
    if (e.ctrlKey || e.metaKey) {
      e.preventDefault();
    }
  },
  { passive: false }
);

window.addEventListener("gesturestart", (e: Event) => e.preventDefault());
window.addEventListener("gesturechange", (e: Event) => e.preventDefault());
window.addEventListener("gestureend", (e: Event) => e.preventDefault());

// 5. Prevent accidental external file drop from navigating away
window.addEventListener("dragover", (e: DragEvent) => e.preventDefault(), false);
window.addEventListener("drop", (e: DragEvent) => e.preventDefault(), false);

// 6. Native Window Focus / Blur State Adaptation
window.addEventListener("focus", () => {
  document.documentElement.classList.remove("window-inactive");
});
window.addEventListener("blur", () => {
  document.documentElement.classList.add("window-inactive");
});

import { isClickOnScrollbar, triggerNativeDrag } from "./utils/windowDrag";

// 7. Enable native window dragging when dragging on background blank areas
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
    ".lang-selector-header",
    ".lang-dropdown-menu",
    ".terminal-body",
    ".log-line",
    ".modal-overlay",
    ".modal-card",
    ".modal-body",
    ".settings-modal-card",
    ".about-modal-card",
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

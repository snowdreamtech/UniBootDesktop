import { createApp } from "vue";
import App from "./App.vue";
import "./styles/theme.css";

// 1. Prevent default browser context menu on non-editable UI elements
window.addEventListener("contextmenu", (e: MouseEvent) => {
  const target = e.target as HTMLElement | null;
  if (
    target &&
    (target.tagName === "INPUT" ||
      target.tagName === "TEXTAREA" ||
      target.isContentEditable)
  ) {
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
});

// 3. Prevent pinch-to-zoom and Ctrl/Cmd + wheel zooming
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

// 4. Prevent accidental external file drop from navigating away
window.addEventListener("dragover", (e: DragEvent) => e.preventDefault(), false);
window.addEventListener("drop", (e: DragEvent) => e.preventDefault(), false);

createApp(App).mount("#app");


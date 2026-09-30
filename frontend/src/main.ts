import { createApp } from "vue";
import App from "./App.vue";
import "./styles/theme.css";

// Prevent default browser context menu on non-editable UI elements
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

createApp(App).mount("#app");

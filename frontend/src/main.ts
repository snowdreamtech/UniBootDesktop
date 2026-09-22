import { createApp } from "vue";
import App from "./App.vue";
import "./styles/theme.css";

const STARTUP_LOADER_MIN_DURATION_MS = 650;
const startupStartedAt = performance.now();

function dismissStartupLoader() {
  const elapsed = performance.now() - startupStartedAt;
  const remaining = Math.max(0, STARTUP_LOADER_MIN_DURATION_MS - elapsed);

  window.setTimeout(() => {
    document.documentElement.classList.add("app-ready");
    window.setTimeout(() => {
      document.getElementById("startup-loader")?.remove();
    }, 260);
  }, remaining);
}

window.addEventListener("unigo:config-ready", dismissStartupLoader, { once: true });
createApp(App).mount("#app");

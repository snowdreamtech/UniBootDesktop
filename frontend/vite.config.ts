import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import { fileURLToPath, URL } from "node:url";
import fs from "node:fs";

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    {
      name: "preserve-dist-gitkeep",
      closeBundle() {
        const gitkeepPath = fileURLToPath(new URL("./dist/.gitkeep", import.meta.url));
        fs.writeFileSync(gitkeepPath, "");
      },
    },
  ],
  root: fileURLToPath(new URL(".", import.meta.url)),
  build: {
    outDir: fileURLToPath(new URL("./dist", import.meta.url)),
    emptyOutDir: true,
  },
  server: {
    port: 5180,
    strictPort: false,
  },
});

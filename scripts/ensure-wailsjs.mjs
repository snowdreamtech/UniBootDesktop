// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { execSync } from "node:child_process";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const rootDir = path.resolve(__dirname, "..");
const runtimeFile = path.join(rootDir, "frontend", "wailsjs", "runtime", "runtime.js");
const appFile = path.join(rootDir, "frontend", "wailsjs", "go", "main", "App.js");

if (!fs.existsSync(runtimeFile) || !fs.existsSync(appFile)) {
  console.log("⚡ frontend/wailsjs is missing. Generating Wails bindings...");
  let succeeded = false;

  try {
    execSync("unirtm exec -- wails generate module", {
      cwd: rootDir,
      stdio: "inherit",
    });
    succeeded = true;
  } catch {
    try {
      execSync("wails generate module", {
        cwd: rootDir,
        stdio: "inherit",
      });
      succeeded = true;
    } catch (err) {
      console.warn("⚠️  Failed to generate Wails bindings automatically:", err.message);
    }
  }

  if (succeeded) {
    console.log("✓ Wails JS/TS bindings generated successfully.");
  }
}

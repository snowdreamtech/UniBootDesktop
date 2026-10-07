// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import fs from "node:fs";
import path from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";
import { execSync } from "node:child_process";

const __filename = fileURLToPath(import.meta.url);
const scriptsDir = path.dirname(__filename);
const projectRoot = path.resolve(scriptsDir, "..");
const frontendDir = path.join(projectRoot, "frontend");
const nodeModulesDir = path.join(frontendDir, "node_modules");
const binDir = path.join(nodeModulesDir, ".bin");

/**
 * Ensure frontend dependencies and platform-specific binary wrappers are present and healthy.
 */
export function ensureFrontendDeps() {
  const isWindows = process.platform === "win32";
  const requiredBins = isWindows ? ["vue-tsc.cmd", "vite.cmd"] : ["vue-tsc", "vite"];

  let needsInstall = false;

  if (!fs.existsSync(nodeModulesDir) || !fs.existsSync(binDir)) {
    needsInstall = true;
  } else {
    for (const bin of requiredBins) {
      const binPath = path.join(binDir, bin);
      if (!fs.existsSync(binPath)) {
        needsInstall = true;
        break;
      }
      if (!isWindows) {
        try {
          fs.accessSync(binPath, fs.constants.X_OK);
        } catch {
          try {
            fs.chmodSync(binPath, 0o755);
            fs.accessSync(binPath, fs.constants.X_OK);
          } catch {
            needsInstall = true;
            break;
          }
        }
      }
    }

    // Auto-heal all POSIX execute bits in node_modules/.bin
    if (!needsInstall && !isWindows) {
      try {
        const files = fs.readdirSync(binDir);
        for (const file of files) {
          if (!file.endsWith(".cmd") && !file.endsWith(".ps1")) {
            const filePath = path.join(binDir, file);
            try {
              const stat = fs.statSync(filePath);
              if (stat.isFile() && (stat.mode & 0o111) === 0) {
                fs.chmodSync(filePath, stat.mode | 0o755);
              }
            } catch {
              // ignore individual stat/chmod errors
            }
          }
        }
      } catch {
        // ignore readdir errors
      }

      // Check platform native rollup binding when rollup is present
      const rollupDir = path.join(nodeModulesDir, "rollup");
      if (fs.existsSync(rollupDir)) {
        const targetRollupDir = path.join(nodeModulesDir, "@rollup", `rollup-${process.platform}-${process.arch}`);
        if (!fs.existsSync(targetRollupDir)) {
          console.warn(
            `\x1b[33m[Preflight] Platform native rollup module missing (@rollup/rollup-${process.platform}-${process.arch}).\x1b[0m`
          );
          needsInstall = true;
        }
      }
    }
  }

  if (needsInstall) {
    console.log(
      `\x1b[33m[Preflight] Frontend dependencies or platform wrappers missing for ${process.platform}. Running npm ci...\x1b[0m`
    );

    // Invalidate Wails cached package.json.md5 so Wails won't assume old state
    const md5File = path.join(frontendDir, "package.json.md5");
    if (fs.existsSync(md5File)) {
      try {
        fs.unlinkSync(md5File);
      } catch {
        // ignore
      }
    }

    let succeeded = false;
    try {
      execSync("unirtm exec -- npm ci", {
        cwd: frontendDir,
        stdio: "inherit",
      });
      succeeded = true;
    } catch {
      try {
        execSync("npm ci", {
          cwd: frontendDir,
          stdio: "inherit",
        });
        succeeded = true;
      } catch (err) {
        console.error(`\x1b[31m[Preflight] Failed to install frontend dependencies: ${err.message}\x1b[0m`);
        throw err;
      }
    }

    if (succeeded) {
      console.log("\x1b[32m[Preflight] Frontend dependencies and platform wrappers verified.\x1b[0m");
    }
  }
}

/**
 * Ensure Wails JavaScript/TypeScript bindings exist.
 */
export function ensureWailsBindings() {
  const runtimeFile = path.join(frontendDir, "wailsjs", "runtime", "runtime.js");
  const appFile = path.join(frontendDir, "wailsjs", "go", "main", "App.js");

  if (!fs.existsSync(runtimeFile) || !fs.existsSync(appFile)) {
    console.log("\x1b[33m[Preflight] frontend/wailsjs bindings missing. Generating...\x1b[0m");
    let succeeded = false;

    try {
      execSync("unirtm exec -- wails generate module", {
        cwd: projectRoot,
        stdio: "inherit",
      });
      succeeded = true;
    } catch {
      try {
        execSync("wails generate module", {
          cwd: projectRoot,
          stdio: "inherit",
        });
        succeeded = true;
      } catch (err) {
        console.warn(
          `\x1b[33m[Preflight] Warning: Failed to generate Wails bindings automatically: ${err.message}\x1b[0m`
        );
      }
    }

    if (succeeded) {
      console.log("\x1b[32m[Preflight] Wails bindings generated successfully.\x1b[0m");
    }
  }
}

/**
 * Ensure application icon and platform bundle assets are prepared.
 */
export async function ensureAppIcon() {
  const iconScript = path.join(scriptsDir, "prepare-app-icon.mjs");
  if (fs.existsSync(iconScript)) {
    await import(pathToFileURL(iconScript).href);
  }
}

/**
 * Run full preflight verification suite.
 */
export async function preflight() {
  ensureFrontendDeps();
  await ensureAppIcon();
  ensureWailsBindings();
}

// Only execute when run directly from command line
const isMain = () => {
  if (!process.argv[1]) return false;
  const currentPath = path.resolve(fileURLToPath(import.meta.url));
  const entryPath = path.resolve(process.argv[1]);
  return process.platform === "win32"
    ? currentPath.toLowerCase() === entryPath.toLowerCase()
    : currentPath === entryPath;
};

if (isMain()) {
  preflight().catch((err) => {
    console.error(`\x1b[31m[Preflight] Verification failed: ${err.message}\x1b[0m`);
    process.exit(1);
  });
}

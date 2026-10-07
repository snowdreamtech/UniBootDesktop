// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { spawn, execSync } from "node:child_process";
import { preflight } from "./preflight.mjs";

const __filename = fileURLToPath(import.meta.url);
const scriptsDir = path.dirname(__filename);
const projectRoot = path.resolve(scriptsDir, "..");

async function main() {
  await preflight();

  // Reset and prepare Go cache directories
  const cacheDir = path.join(projectRoot, ".cache");
  const goCacheDir = path.join(cacheDir, "go");
  const goTmpDir = path.join(cacheDir, "go-tmp");

  try {
    fs.rmSync(cacheDir, { recursive: true, force: true });
  } catch {
    // ignore
  }
  fs.mkdirSync(goCacheDir, { recursive: true });
  fs.mkdirSync(goTmpDir, { recursive: true });

  const getGitOutput = (cmd, fallback) => {
    try {
      return execSync(cmd, { cwd: projectRoot, stdio: ["ignore", "pipe", "ignore"] })
        .toString()
        .trim();
    } catch {
      return fallback;
    }
  };

  const tag = getGitOutput("git describe --tags --abbrev=0", "dev");
  const commit = getGitOutput("git rev-parse --short HEAD", "unknown");
  const commitFull = getGitOutput("git rev-parse HEAD", "unknown");
  const date = new Date().toISOString();

  let modulePath = "github.com/snowdreamtech/unibootdesktop";
  try {
    modulePath = execSync("go list -m", { cwd: projectRoot, stdio: ["ignore", "pipe", "ignore"] })
      .toString()
      .trim();
  } catch {
    // fallback
  }

  const pkg = `${modulePath}/internal/env`;
  const projectName = modulePath.split("/").pop() || "unibootdesktop";

  const userArgs = process.argv.slice(2);
  const autoTags = [];
  if (process.platform === "linux") {
    const hasTags = userArgs.some((arg) => arg === "-tags" || arg.startsWith("-tags="));
    if (!hasTags) {
      try {
        let has40 = false;
        try {
          execSync("pkg-config --exists webkit2gtk-4.0", { stdio: "ignore" });
          has40 = true;
        } catch {
          // ignore
        }
        if (!has40) {
          try {
            execSync("pkg-config --exists webkit2gtk-4.1", { stdio: "ignore" });
            autoTags.push("-tags", "webkit2_41");
          } catch {
            // ignore
          }
        }
      } catch {
        // ignore
      }
    }
  }

  const ldflags = [
    "-s",
    "-w",
    `-X '${pkg}.GitTag=${tag}'`,
    `-X '${pkg}.CommitHash=${commit}'`,
    `-X '${pkg}.CommitHashFull=${commitFull}'`,
    `-X '${pkg}.BuildTime=${date}'`,
    `-X '${pkg}.ProjectName=${projectName}'`,
  ].join(" ");

  const args = ["build", "-clean", "-trimpath", "-m", "-nosyncgomod", ...autoTags, "-ldflags", ldflags, ...userArgs];

  const childEnv = {
    ...process.env,
    GOCACHE: goCacheDir,
    GOTMPDIR: goTmpDir,
  };

  const isWindows = process.platform === "win32";
  const wailsCmd = isWindows ? "wails.exe" : "wails";

  const child = spawn(wailsCmd, args, {
    cwd: projectRoot,
    env: childEnv,
    stdio: "inherit",
    shell: isWindows,
  });

  child.on("exit", (code, signal) => {
    if (signal) {
      process.kill(process.pid, signal);
    } else {
      process.exit(code ?? 0);
    }
  });

  child.on("error", (err) => {
    console.error(`\x1b[31m[Build] Failed to spawn Wails process: ${err.message}\x1b[0m`);
    process.exit(1);
  });
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
  main().catch((err) => {
    console.error(`\x1b[31m[Build] Build failed: ${err.message}\x1b[0m`);
    process.exit(1);
  });
}

// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { spawn, execSync } from "node:child_process";
import { preflight } from "./preflight.mjs";
import { cleanDevArtifacts, resolveWailsCommand } from "./run-dev.mjs";

const __filename = fileURLToPath(import.meta.url);
const scriptsDir = path.dirname(__filename);
const projectRoot = path.resolve(scriptsDir, "..");

async function main() {
  await preflight();
  cleanDevArtifacts(projectRoot);

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

  let modulePath = "github.com/snowdreamtech/unigodesktop";
  try {
    modulePath = execSync("go list -m", { cwd: projectRoot, stdio: ["ignore", "pipe", "ignore"] })
      .toString()
      .trim();
  } catch {
    // fallback
  }

  const pkg = `${modulePath}/internal/env`;
  const projectName = modulePath.split("/").pop() || "unigodesktop";

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

  const isWindows = process.platform === "win32";
  const wailsCmd = resolveWailsCommand();

  // Normalize PATH and Path for Windows cross-shell compatibility (MSYS2 / Git Bash / PowerShell)
  const effectivePath = process.env.PATH || process.env.Path || "";
  const extraPaths = [];

  // 1. Priority 1: The Node.js and npm directory running this script (direct binary, bypassing shims)
  const nodeDir = path.dirname(process.execPath);
  if (fs.existsSync(nodeDir)) {
    extraPaths.push(nodeDir);
  }

  // 2. Priority 2: Direct Wails executable directory
  const wailsDir = path.dirname(wailsCmd);
  if (fs.existsSync(wailsDir) && !extraPaths.includes(wailsDir)) {
    extraPaths.push(wailsDir);
  }

  // 3. Priority 3: Resolve active tools via unirtm which
  const userBin = path.join(process.env.USERPROFILE || process.env.HOME || "", "bin");
  const unirtmCandidates = ["unirtm", path.join(userBin, isWindows ? "unirtm.exe" : "unirtm")];
  for (const cmd of ["node", "npm", "wails"]) {
    for (const u of unirtmCandidates) {
      try {
        const out = execSync(`"${u}" which ${cmd}`, { stdio: ["ignore", "pipe", "ignore"] })
          .toString()
          .trim();
        if (out && fs.existsSync(out)) {
          const binDir = path.dirname(out);
          if (!extraPaths.includes(binDir)) {
            extraPaths.push(binDir);
          }
          break;
        }
      } catch {
        // ignore
      }
    }
  }

  // 4. Priority 4: UniRTM shims directory as fallback
  const localAppData = process.env.LOCALAPPDATA;
  if (isWindows && localAppData) {
    const shimDir = path.join(localAppData, "unirtm", "shims");
    if (fs.existsSync(shimDir) && !extraPaths.includes(shimDir)) {
      extraPaths.push(shimDir);
    }
  }
  const homeDir = process.env.HOME || process.env.USERPROFILE;
  if (homeDir) {
    const unixShimDir = path.join(homeDir, ".local", "share", "unirtm", "shims");
    if (fs.existsSync(unixShimDir) && !extraPaths.includes(unixShimDir)) {
      extraPaths.push(unixShimDir);
    }
  }
  if (fs.existsSync(userBin) && !extraPaths.includes(userBin)) {
    extraPaths.push(userBin);
  }

  // Construct deduplicated PATH with extraPaths prepended
  const pathParts = [];
  for (const p of extraPaths) {
    if (p && !pathParts.includes(p)) {
      pathParts.push(p);
    }
  }
  for (const p of effectivePath.split(path.delimiter)) {
    if (p && !pathParts.includes(p)) {
      pathParts.push(p);
    }
  }
  const fullPath = pathParts.join(path.delimiter);

  // Guarantee PATHEXT on Windows for MSYS2/Git Bash/CMD (so npm resolves to npm.cmd)
  const defaultPathExt = ".COM;.EXE;.BAT;.CMD;.VBS;.VBE;.JS;.JSE;.WSF;.WSH;.MSC";
  const effectivePathExt = process.env.PATHEXT || process.env.PathExt || defaultPathExt;

  const childEnv = { ...process.env };
  // Purge any case collisions of PATH / PATHEXT
  for (const key of Object.keys(childEnv)) {
    const lower = key.toLowerCase();
    if (lower === "path") {
      delete childEnv[key];
    } else if (isWindows && lower === "pathext") {
      delete childEnv[key];
    }
  }

  childEnv.PATH = fullPath;
  if (isWindows) {
    childEnv.Path = fullPath;
    childEnv.PATHEXT = effectivePathExt;
    childEnv.PathExt = effectivePathExt;
  }
  childEnv.GOCACHE = goCacheDir;
  childEnv.GOTMPDIR = goTmpDir;

  const child = spawn(wailsCmd, args, {
    cwd: projectRoot,
    env: childEnv,
    stdio: "inherit",
    shell: isWindows,
  });

  child.on("exit", (code, signal) => {
    cleanDevArtifacts(projectRoot);
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

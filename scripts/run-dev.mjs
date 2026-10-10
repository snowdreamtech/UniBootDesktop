// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import fs from "node:fs";
import path from "node:path";
import { createServer } from "node:net";
import { spawn, execSync } from "node:child_process";
import { fileURLToPath } from "node:url";
import { preflight } from "./preflight.mjs";

/**
 * Check if a TCP port is currently free on the specified host.
 * @param {number} port
 * @param {string} host
 * @returns {Promise<boolean>}
 */
export function isPortFree(port, host = "127.0.0.1") {
  return new Promise((resolve) => {
    const server = createServer();
    server.once("error", () => resolve(false));
    server.once("listening", () => {
      server.close(() => resolve(true));
    });
    server.listen(port, host);
  });
}

/**
 * Find the next available TCP port starting from startPort.
 * @param {number} startPort
 * @param {string} host
 * @param {number} maxAttempts
 * @returns {Promise<{ port: number, switched: boolean }>}
 */
export async function findNextFreePort(startPort, host = "127.0.0.1", maxAttempts = 100) {
  for (let offset = 0; offset < maxAttempts; offset++) {
    const port = startPort + offset;
    if (await isPortFree(port, host)) {
      return { port, switched: offset > 0 };
    }
  }
  throw new Error(`Unable to find an available port in range [${startPort}, ${startPort + maxAttempts})`);
}

/**
 * Determine if this script is being executed directly.
 * @returns {boolean}
 */
export function isMainModule() {
  if (!process.argv[1]) return false;
  try {
    const scriptPath = fileURLToPath(import.meta.url);
    const entryPath = path.resolve(process.argv[1]);
    return (
      scriptPath === entryPath || (process.platform === "win32" && scriptPath.toLowerCase() === entryPath.toLowerCase())
    );
  } catch {
    return false;
  }
}

/**
 * Resolve the Wails executable binary path reliably across all environments.
 * @returns {string}
 */
export function resolveWailsCommand() {
  const isWindows = process.platform === "win32";
  const defaultCmd = isWindows ? "wails.exe" : "wails";
  const userBin = path.join(process.env.USERPROFILE || process.env.HOME || "", "bin");
  const unirtmCandidates = ["unirtm", path.join(userBin, isWindows ? "unirtm.exe" : "unirtm")];

  // 1. Try resolving via `unirtm which wails`
  for (const u of unirtmCandidates) {
    try {
      const whichOut = execSync(`"${u}" which wails`, { stdio: ["ignore", "pipe", "ignore"] })
        .toString()
        .trim();
      if (whichOut && fs.existsSync(whichOut)) {
        return whichOut;
      }
    } catch {
      // ignore
    }
  }

  // 2. Check unirtm installs directory for wails
  const localAppData = process.env.LOCALAPPDATA;
  if (isWindows && localAppData) {
    const installsDir = path.join(localAppData, "unirtm", "installs");
    if (fs.existsSync(installsDir)) {
      try {
        const entries = fs.readdirSync(installsDir);
        for (const entry of entries) {
          if (entry.includes("wails")) {
            const candidate = path.join(installsDir, entry);
            for (const sub of [defaultCmd, path.join("bin", defaultCmd)]) {
              const fullCandidate = path.join(candidate, sub);
              if (fs.existsSync(fullCandidate)) {
                return fullCandidate;
              }
            }
          }
        }
      } catch {
        // ignore
      }
    }
  }

  // 3. Check unirtm shims
  if (isWindows && localAppData) {
    const shimPath = path.join(localAppData, "unirtm", "shims", defaultCmd);
    if (fs.existsSync(shimPath)) {
      return shimPath;
    }
  }
  const homeDir = process.env.HOME || process.env.USERPROFILE;
  if (homeDir) {
    const unixShimPath = path.join(homeDir, ".local", "share", "unirtm", "shims", defaultCmd);
    if (fs.existsSync(unixShimPath)) {
      return unixShimPath;
    }
  }

  return defaultCmd;
}

/**
 * Remove any leftover development binaries (*-dev, *-dev.exe) from build/bin.
 * @param {string} root
 */
export function cleanDevArtifacts(root) {
  const binDir = path.join(root, "build", "bin");
  if (!fs.existsSync(binDir)) return;
  try {
    const entries = fs.readdirSync(binDir);
    for (const entry of entries) {
      if (entry.includes("-dev")) {
        const fullPath = path.join(binDir, entry);
        try {
          fs.rmSync(fullPath, { force: true });
        } catch {
          // ignore locked files
        }
      }
    }
  } catch {
    // ignore
  }
}

async function main() {
  await preflight();

  const rootDir = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
  cleanDevArtifacts(rootDir);

  const preferredBackendPort = 34115;
  const preferredFrontendPort = 5173;

  const backend = await findNextFreePort(preferredBackendPort);
  if (backend.switched) {
    console.log(
      `\x1b[33m[SmartPort] Backend port ${preferredBackendPort} is in use; auto-switched to free port ${backend.port}\x1b[0m`
    );
  } else {
    console.log(`\x1b[32m[SmartPort] Backend port ${backend.port} is available\x1b[0m`);
  }

  const frontend = await findNextFreePort(preferredFrontendPort);
  if (frontend.switched) {
    console.log(
      `\x1b[33m[SmartPort] Frontend port ${preferredFrontendPort} is in use; auto-switched to free port ${frontend.port}\x1b[0m`
    );
  } else {
    console.log(`\x1b[32m[SmartPort] Frontend port ${frontend.port} is available\x1b[0m`);
  }

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
          // webkit2gtk-4.0 not available
        }
        if (!has40) {
          try {
            execSync("pkg-config --exists webkit2gtk-4.1", { stdio: "ignore" });
            autoTags.push("-tags", "webkit2_41");
          } catch {
            // neither available
          }
        }
      } catch {
        // pkg-config not available
      }
    }
  }

  const args = ["dev", "-devserver", `127.0.0.1:${backend.port}`, "-viteservertimeout", "60", ...autoTags, ...userArgs];

  console.log(
    `\x1b[36m[SmartPort] Launching Wails Dev (Backend: 127.0.0.1:${backend.port}, Frontend: 127.0.0.1:${frontend.port})...\x1b[0m\n`
  );

  const cacheGo = path.join(rootDir, ".cache", "go");
  const cacheGoTmp = path.join(rootDir, ".cache", "go-tmp");
  fs.mkdirSync(cacheGo, { recursive: true });
  fs.mkdirSync(cacheGoTmp, { recursive: true });

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
  childEnv.GOCACHE = process.env.GOCACHE || cacheGo;
  childEnv.GOTMPDIR = process.env.GOTMPDIR || cacheGoTmp;
  childEnv.PORT = String(frontend.port);
  childEnv.VITE_PORT = String(frontend.port);

  const child = spawn(wailsCmd, args, {
    env: childEnv,
    stdio: "inherit",
    shell: isWindows,
  });

  const forwardSignal = (sig) => {
    if (child && !child.killed) {
      try {
        child.kill(sig);
      } catch {
        // process may have already exited
      }
    }
  };

  process.on("SIGINT", () => forwardSignal("SIGINT"));
  process.on("SIGTERM", () => forwardSignal("SIGTERM"));

  child.on("exit", (code, signal) => {
    cleanDevArtifacts(rootDir);
    if (signal) {
      process.kill(process.pid, signal);
    } else {
      process.exit(code ?? 0);
    }
  });

  child.on("error", (err) => {
    console.error(`\x1b[31m[SmartPort] Failed to spawn Wails process: ${err.message}\x1b[0m`);
    process.exit(1);
  });
}

// Only execute when run directly from command line
if (isMainModule()) {
  main().catch((err) => {
    console.error(`\x1b[31m[SmartPort] Error during smart port resolution: ${err.message}\x1b[0m`);
    process.exit(1);
  });
}

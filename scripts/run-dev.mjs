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

async function main() {
  await preflight();

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

  const args = ["dev", "-devserver", `localhost:${backend.port}`, ...autoTags, ...userArgs];

  console.log(
    `\x1b[36m[SmartPort] Launching Wails Dev (Backend: localhost:${backend.port}, Frontend: localhost:${frontend.port})...\x1b[0m\n`
  );

  const rootDir = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
  const cacheGo = path.join(rootDir, ".cache", "go");
  const cacheGoTmp = path.join(rootDir, ".cache", "go-tmp");
  fs.mkdirSync(cacheGo, { recursive: true });
  fs.mkdirSync(cacheGoTmp, { recursive: true });

  const childEnv = {
    ...process.env,
    GOCACHE: process.env.GOCACHE || cacheGo,
    GOTMPDIR: process.env.GOTMPDIR || cacheGoTmp,
    PORT: String(frontend.port),
    VITE_PORT: String(frontend.port),
  };

  const isWindows = process.platform === "win32";
  const wailsCmd = isWindows ? "wails.exe" : "wails";

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

// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import { createServer } from "node:net";
import { spawn } from "node:child_process";

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

async function main() {
  const preferredBackendPort = 34115;
  const preferredFrontendPort = 5173;

  const backend = await findNextFreePort(preferredBackendPort);
  if (backend.switched) {
    console.log(`\x1b[33m[SmartPort] Backend port ${preferredBackendPort} is in use; auto-switched to free port ${backend.port}\x1b[0m`);
  } else {
    console.log(`\x1b[32m[SmartPort] Backend port ${backend.port} is available\x1b[0m`);
  }

  const frontend = await findNextFreePort(preferredFrontendPort);
  if (frontend.switched) {
    console.log(`\x1b[33m[SmartPort] Frontend port ${preferredFrontendPort} is in use; auto-switched to free port ${frontend.port}\x1b[0m`);
  } else {
    console.log(`\x1b[32m[SmartPort] Frontend port ${frontend.port} is available\x1b[0m`);
  }

  const args = [
    "dev",
    "-devserver",
    `localhost:${backend.port}`,
    ...process.argv.slice(2),
  ];

  console.log(`\x1b[36m[SmartPort] Launching Wails Dev (Backend: localhost:${backend.port}, Frontend: localhost:${frontend.port})...\x1b[0m\n`);

  const childEnv = {
    ...process.env,
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
if (import.meta.url === `file://${process.argv[1]}`) {
  main().catch((err) => {
    console.error(`\x1b[31m[SmartPort] Error during smart port resolution: ${err.message}\x1b[0m`);
    process.exit(1);
  });
}

// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import fs from "node:fs";
import path from "node:path";
import crypto from "node:crypto";
import { fileURLToPath } from "node:url";
import https from "node:https";
import http from "node:http";

const __filename = fileURLToPath(import.meta.url);
const scriptsDir = path.dirname(__filename);
const projectRoot = path.resolve(scriptsDir, "..");
const assetsDir = path.join(projectRoot, "pkg", "firmware", "assets");
const versionFile = path.join(assetsDir, "version.json");
const checksumsFile = path.join(assetsDir, "checksums.sha256");

function calculateSha256(filePath) {
  return new Promise((resolve, reject) => {
    const hash = crypto.createHash("sha256");
    const stream = fs.createReadStream(filePath);
    stream.on("data", (chunk) => hash.update(chunk));
    stream.on("end", () => resolve(hash.digest("hex").toLowerCase()));
    stream.on("error", (err) => reject(err));
  });
}

async function downloadToFile(url, destPath) {
  const tmpPath = `${destPath}.tmp-${Date.now()}`;
  try {
    const res = await fetch(url, {
      redirect: "follow",
      headers: {
        "User-Agent": "UniBootDesktop-Build/1.0",
        Accept: "*/*",
      },
    });

    if (!res.ok) {
      throw new Error(`HTTP status ${res.status} for ${url}`);
    }

    const { pipeline } = await import("node:stream/promises");
    const { Readable } = await import("node:stream");
    const fileStream = fs.createWriteStream(tmpPath);
    await pipeline(Readable.fromWeb(res.body), fileStream);
    fs.renameSync(tmpPath, destPath);
  } finally {
    fs.rmSync(tmpPath, { force: true });
  }
}

export async function fetchFirmwareAssets(options = {}) {
  const { silent = false } = options;

  if (!fs.existsSync(versionFile) || !fs.existsSync(checksumsFile)) {
    throw new Error("Missing firmware metadata (version.json or checksums.sha256)");
  }

  const versionConfig = JSON.parse(fs.readFileSync(versionFile, "utf-8"));
  const tagName = versionConfig.tagName || `v${versionConfig.version}`;

  const checksumLines = fs
    .readFileSync(checksumsFile, "utf-8")
    .split("\n")
    .map((l) => l.trim())
    .filter((l) => l.length > 0);

  const manifest = new Map();
  for (const line of checksumLines) {
    const parts = line.split(/\s+/);
    if (parts.length >= 2) {
      manifest.set(parts[1], parts[0].toLowerCase());
    }
  }

  // Repository files tracked directly in git are excluded from downloads
  const gitTrackedAssets = new Set(["background.png", "boot.ipxe", "uniboot.ipxe", "version.json", "checksums.sha256"]);

  const proxy = process.env.GITHUB_PROXY || process.env.ENABLE_GITHUB_PROXY === "1" ? process.env.GITHUB_PROXY : "";
  const proxyPrefix = proxy && proxy !== "0" && proxy !== "direct" ? (proxy.endsWith("/") ? proxy : `${proxy}/`) : "";

  let downloadedCount = 0;
  let verifiedCount = 0;

  for (const [filename, expectedHash] of manifest.entries()) {
    if (gitTrackedAssets.has(filename)) {
      continue;
    }

    const filePath = path.join(assetsDir, filename);

    // Verify existing file integrity
    if (fs.existsSync(filePath)) {
      const existingHash = await calculateSha256(filePath);
      if (existingHash === expectedHash) {
        verifiedCount++;
        continue;
      }
      if (!silent) {
        console.warn(`[Firmware] Checksum mismatch for ${filename}, re-downloading...`);
      }
      fs.rmSync(filePath, { force: true });
    }

    const rawUrl = `https://github.com/snowdreamtech/UniBoot/releases/download/${tagName}/${filename}`;
    const candidateUrls = [];
    if (proxyPrefix) {
      candidateUrls.push(`${proxyPrefix}${rawUrl}`);
    }
    candidateUrls.push(rawUrl);

    let success = false;
    let lastError = null;

    for (const url of candidateUrls) {
      for (let attempt = 1; attempt <= 3; attempt++) {
        try {
          if (!silent) {
            console.log(`[Firmware] Downloading ${filename} from ${url} (attempt ${attempt}/3)...`);
          }
          await downloadToFile(url, filePath);
          const downloadedHash = await calculateSha256(filePath);
          if (downloadedHash !== expectedHash) {
            fs.rmSync(filePath, { force: true });
            throw new Error(`Checksum mismatch: expected ${expectedHash}, got ${downloadedHash}`);
          }
          downloadedCount++;
          success = true;
          break;
        } catch (err) {
          lastError = err;
          fs.rmSync(filePath, { force: true });
          await new Promise((r) => setTimeout(r, 1000 * attempt));
        }
      }
      if (success) break;
    }

    if (!success) {
      throw new Error(
        `Failed to download firmware asset ${filename}: ${lastError ? lastError.message : "unknown error"}`
      );
    }
  }

  if (!silent) {
    console.log(`[Firmware] Assets verification complete: ${verifiedCount} cached, ${downloadedCount} downloaded.`);
  }

  return { verifiedCount, downloadedCount };
}

// Only execute directly when invoked as a script
const isMain = () => {
  if (!process.argv[1]) return false;
  const currentPath = path.resolve(fileURLToPath(import.meta.url));
  const entryPath = path.resolve(process.argv[1]);
  return process.platform === "win32"
    ? currentPath.toLowerCase() === entryPath.toLowerCase()
    : currentPath === entryPath;
};

if (isMain()) {
  fetchFirmwareAssets().catch((err) => {
    console.error(`[Firmware] Error: ${err.message}`);
    process.exit(1);
  });
}

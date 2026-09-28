import { readFile, writeFile } from "node:fs/promises";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import { execSync } from "node:child_process";

const projectRoot = resolve(dirname(fileURLToPath(import.meta.url)), "..");

let tag = process.env.TAG || "";
if (!tag) {
  try {
    tag = execSync("git describe --tags --exact-match 2>/dev/null", { encoding: "utf-8" }).trim();
  } catch {
    tag = "";
  }
}

if (!tag) {
  console.log("No git tag detected, skipping build metadata version sync.");
  process.exit(0);
}

const cleanVer = tag.replace(/^v/i, "").trim();
const semverMatch = cleanVer.match(/^(\d+)\.(\d+)\.(\d+)/);
if (!semverMatch) {
  console.log(`Tag '${tag}' is not standard semver, keeping existing metadata.`);
  process.exit(0);
}

const major = parseInt(semverMatch[1], 10);
const minor = parseInt(semverMatch[2], 10);
const patch = parseInt(semverMatch[3], 10);
const build = 0;

// 1. Update wails.json
const wailsPath = resolve(projectRoot, "wails.json");
try {
  const wailsContent = JSON.parse(await readFile(wailsPath, "utf-8"));
  if (wailsContent.info) {
    wailsContent.info.productVersion = cleanVer;
    await writeFile(wailsPath, JSON.stringify(wailsContent, null, 2) + "\n", "utf-8");
    console.log(`Updated wails.json productVersion to ${cleanVer}`);
  }
} catch (e) {
  console.warn("Failed to update wails.json:", e);
}

// 2. Update build/windows/info.json
const infoPath = resolve(projectRoot, "build/windows/info.json");
try {
  const infoContent = JSON.parse(await readFile(infoPath, "utf-8"));
  if (infoContent.FixedFileInfo) {
    infoContent.FixedFileInfo.FileVersion = { Major: major, Minor: minor, Patch: patch, Build: build };
    infoContent.FixedFileInfo.ProductVersion = { Major: major, Minor: minor, Patch: patch, Build: build };
  }
  if (infoContent.StringFileInfo) {
    infoContent.StringFileInfo.FileVersion = cleanVer;
    infoContent.StringFileInfo.ProductVersion = cleanVer;
  }
  await writeFile(infoPath, JSON.stringify(infoContent, null, 2) + "\n", "utf-8");
  console.log(`Updated build/windows/info.json versions to ${cleanVer}`);
} catch (e) {
  console.warn("Failed to update build/windows/info.json:", e);
}

// 3. Update build/windows/wails.exe.manifest
const manifestPath = resolve(projectRoot, "build/windows/wails.exe.manifest");
try {
  let manifest = await readFile(manifestPath, "utf-8");
  const fullVer = `${major}.${minor}.${patch}.${build}`;
  manifest = manifest.replace(/version="\d+\.\d+\.\d+\.\d+"/, `version="${fullVer}"`);
  await writeFile(manifestPath, manifest, "utf-8");
  console.log(`Updated build/windows/wails.exe.manifest version to ${fullVer}`);
} catch (e) {
  console.warn("Failed to update build/windows/wails.exe.manifest:", e);
}

// 4. Update frontend/package.json
const pkgPath = resolve(projectRoot, "frontend/package.json");
try {
  const pkgContent = JSON.parse(await readFile(pkgPath, "utf-8"));
  pkgContent.version = cleanVer;
  await writeFile(pkgPath, JSON.stringify(pkgContent, null, 2) + "\n", "utf-8");
  console.log(`Updated frontend/package.json version to ${cleanVer}`);
} catch (e) {
  console.warn("Failed to update frontend/package.json:", e);
}

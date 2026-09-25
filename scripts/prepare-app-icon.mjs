import { copyFile, mkdir, readFile, stat } from "node:fs/promises";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const projectRoot = resolve(dirname(fileURLToPath(import.meta.url)), "..");
const sourcePath = resolve(projectRoot, "assets/branding/app-icon.png");
const targetPath = resolve(projectRoot, "build/appicon.png");

const sourceStats = await stat(sourcePath);
if (sourceStats.size === 0) {
  throw new Error(`App icon source is empty: ${sourcePath}`);
}

await mkdir(dirname(targetPath), { recursive: true });
await copyFile(sourcePath, targetPath);
const signature = (await readFile(targetPath)).subarray(0, 8).toString("hex");
if (signature !== "89504e470d0a1a0a") {
  throw new Error(`App icon is not a valid PNG: ${targetPath}`);
}

console.log(`Prepared Wails App Icon: ${targetPath}`);

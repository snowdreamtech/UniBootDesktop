import { execSync } from "node:child_process";
import { readdir, readFile, writeFile } from "node:fs/promises";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const projectRoot = resolve(dirname(fileURLToPath(import.meta.url)), "..");
const localesDir = resolve(projectRoot, "frontend/src/i18n/locales");
const isCheckMode = process.argv.includes("--check");
const isFormatOnly = process.argv.includes("--format");

/**
 * Converts a locale code like "zh-CN" or "sr-Latn" to camelCase identifier "zhCn" or "srLatn".
 * @param {string} code
 * @returns {string}
 */
function toCamelCase(code) {
  return code
    .split("-")
    .map((part, index) =>
      index === 0 ? part.toLowerCase() : part.charAt(0).toUpperCase() + part.slice(1).toLowerCase()
    )
    .join("");
}

/**
 * Extracts key-value pairs from a locale TypeScript file.
 * Supports single/double quotes, Prettier line-wrapped keys, and trailing comments.
 * @param {string} content
 * @returns {Map<string, string>}
 */
function parseLocaleEntries(content) {
  const map = new Map();
  const regex =
    /^\s*(?:"([^"\\]*(?:\\.[^"\\]*)*)"|'([^'\\]*(?:\\.[^'\\]*)*)')\s*:\s*(?:"((?:[^"\\]|\\.)*)"|'((?:[^'\\]|\\.)*)')\s*,?\s*(?:\/\/.*)?$/gm;
  let match;
  while ((match = regex.exec(content)) !== null) {
    const key = match[1] ?? match[2];
    const value = match[3] ?? match[4];
    map.set(key, value);
  }
  return map;
}

/**
 * Formats given file paths using Prettier to maintain 100% style consistency with unirtm verify.
 * @param {string[]} filePaths
 */
function formatFilesWithPrettier(filePaths) {
  if (!filePaths || filePaths.length === 0) return;
  const args = filePaths.map((f) => `"${f}"`).join(" ");
  try {
    execSync(`prettier --write ${args}`, { stdio: "inherit" });
  } catch {
    try {
      execSync(`unirtm exec -- prettier --write ${args}`, { stdio: "inherit" });
    } catch {
      try {
        execSync(`npx prettier --write ${args}`, { stdio: "inherit" });
      } catch (err) {
        console.warn("⚠️  Warning: could not run prettier on updated locale files:", err.message);
      }
    }
  }
}

/**
 * Generates formatted TypeScript file content for a locale.
 * @param {string} varName
 * @param {Map<string, string>} entries
 * @returns {string}
 */
function generateLocaleContent(varName, entries) {
  const lines = [
    'import type { TranslationDict } from "../types";',
    "",
    `export const ${varName}: TranslationDict = {`,
  ];

  for (const [key, value] of entries) {
    const raw = value.replace(/\\"/g, '"');
    const escapedValue = raw.replace(/"/g, '\\"');
    lines.push(`  "${key}": "${escapedValue}",`);
  }

  lines.push("};", "");
  return lines.join("\n");
}

async function main() {
  if (isFormatOnly) {
    console.log(`Formatting all locale files in ${localesDir} with Prettier...`);
    formatFilesWithPrettier([resolve(localesDir, "*.ts")]);
    process.exit(0);
  }

  const enUsPath = resolve(localesDir, "en-US.ts");
  const enUsContent = await readFile(enUsPath, "utf-8");
  const enUsEntries = parseLocaleEntries(enUsContent);
  const canonicalKeys = Array.from(enUsEntries.keys());

  const allFiles = (await readdir(localesDir)).filter((file) => file.endsWith(".ts"));
  let hasMissing = false;
  const updatedFiles = [];

  console.log(
    `Checking ${allFiles.length} locale files against en-US canonical schema (${canonicalKeys.length} keys)...`
  );

  for (const file of allFiles) {
    const localeCode = file.replace(/\.ts$/, "");
    const filePath = resolve(localesDir, file);
    const content = await readFile(filePath, "utf-8");
    const currentEntries = parseLocaleEntries(content);

    const missingKeys = canonicalKeys.filter((key) => !currentEntries.has(key));
    const obsoleteKeys = Array.from(currentEntries.keys()).filter((key) => !canonicalKeys.includes(key));

    if (missingKeys.length > 0 || obsoleteKeys.length > 0) {
      hasMissing = true;
      if (missingKeys.length > 0) {
        console.warn(`⚠️  [${localeCode}] missing ${missingKeys.length} keys: ${missingKeys.join(", ")}`);
      }
      if (obsoleteKeys.length > 0) {
        console.warn(`🗑️  [${localeCode}] removing ${obsoleteKeys.length} obsolete keys: ${obsoleteKeys.join(", ")}`);
      }

      if (!isCheckMode) {
        // Backfill missing keys in canonical order using en-US fallback
        const mergedEntries = new Map();
        for (const key of canonicalKeys) {
          if (currentEntries.has(key)) {
            mergedEntries.set(key, currentEntries.get(key));
          } else {
            mergedEntries.set(key, enUsEntries.get(key) || "");
          }
        }

        const varName = toCamelCase(localeCode);
        const newContent = generateLocaleContent(varName, mergedEntries);
        await writeFile(filePath, newContent, "utf-8");
        updatedFiles.push(filePath);
        console.log(`✅ [${localeCode}] synchronized and updated.`);
      }
    }
  }

  if (isCheckMode) {
    if (hasMissing) {
      console.error("\n❌ Locale check failed: missing translation keys detected.");
      process.exit(1);
    } else {
      console.log(`\n🎉 All ${allFiles.length} locale files are complete and synchronized!`);
      process.exit(0);
    }
  } else {
    if (updatedFiles.length > 0) {
      if (!updatedFiles.includes(enUsPath)) {
        updatedFiles.push(enUsPath);
      }
      console.log(`\n🎨 Formatting ${updatedFiles.length} updated files with Prettier...`);
      formatFilesWithPrettier(updatedFiles);
      console.log(`\n✨ Successfully updated and formatted ${updatedFiles.length} locale files.`);
    } else {
      console.log(`\n🎉 All ${allFiles.length} locale files already up to date!`);
    }
  }
}

main().catch((err) => {
  console.error("Fatal error:", err);
  process.exit(1);
});

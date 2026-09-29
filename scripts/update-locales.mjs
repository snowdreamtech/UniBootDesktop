import { readdir, readFile, writeFile } from "node:fs/promises";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const projectRoot = resolve(dirname(fileURLToPath(import.meta.url)), "..");
const localesDir = resolve(projectRoot, "frontend/src/i18n/locales");
const isCheckMode = process.argv.includes("--check");

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
 * @param {string} content
 * @returns {Map<string, string>}
 */
function parseLocaleEntries(content) {
  const map = new Map();
  const regex = /^\s*"([^"\\]*(?:\\.[^"\\]*)*)"\s*:\s*"((?:[^"\\]|\\.)*)"\s*,?\s*$/gm;
  let match;
  while ((match = regex.exec(content)) !== null) {
    map.set(match[1], match[2]);
  }
  return map;
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
    const escapedValue = value.replace(/"/g, '\\"');
    lines.push(`  "${key}": "${escapedValue}",`);
  }

  lines.push("};", "");
  return lines.join("\n");
}

async function main() {
  const enUsPath = resolve(localesDir, "en-US.ts");
  const enUsContent = await readFile(enUsPath, "utf-8");
  const enUsEntries = parseLocaleEntries(enUsContent);
  const canonicalKeys = Array.from(enUsEntries.keys());

  const allFiles = (await readdir(localesDir)).filter((file) => file.endsWith(".ts"));
  let hasMissing = false;
  let updatedCount = 0;

  console.log(
    `Checking ${allFiles.length} locale files against en-US canonical schema (${canonicalKeys.length} keys)...`
  );

  for (const file of allFiles) {
    const localeCode = file.replace(/\.ts$/, "");
    const filePath = resolve(localesDir, file);
    const content = await readFile(filePath, "utf-8");
    const currentEntries = parseLocaleEntries(content);

    const missingKeys = canonicalKeys.filter((key) => !currentEntries.has(key));

    if (missingKeys.length > 0) {
      hasMissing = true;
      console.warn(`⚠️  [${localeCode}] missing ${missingKeys.length} keys: ${missingKeys.join(", ")}`);

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
        updatedCount++;
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
    if (updatedCount > 0) {
      console.log(`\n✨ Successfully updated ${updatedCount} locale files.`);
    } else {
      console.log(`\n🎉 All ${allFiles.length} locale files already up to date!`);
    }
  }
}

main().catch((err) => {
  console.error("Fatal error:", err);
  process.exit(1);
});

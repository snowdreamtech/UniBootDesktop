#!/usr/bin/env node

/**
 * i18n Key Alignment & Type Integrity Checker
 * Validates interface definitions in types.ts against locales and code usage.
 */

const fs = require("fs");
const path = require("path");

const workspaceRoot = path.resolve(__dirname, "..");
const frontendDir = path.join(workspaceRoot, "frontend");
const typesFile = path.join(frontendDir, "src/i18n/types.ts");
const localesDir = path.join(frontendDir, "src/i18n/locales");
const srcDir = path.join(frontendDir, "src");

console.log("🔍 [i18n Check] Running i18n Key & Type Integrity Validation...");

let hasErrors = false;

// 1. Parse TranslationDict keys in types.ts
if (!fs.existsSync(typesFile)) {
  console.error(`❌ Cannot find types file: ${typesFile}`);
  process.exit(1);
}

const typesContent = fs.readFileSync(typesFile, "utf8");
const declaredKeys = new Set();
const typeKeyRegex = /"([^"]+)":\s*string;/g;
let match;
while ((match = typeKeyRegex.exec(typesContent)) !== null) {
  declaredKeys.add(match[1]);
}

console.log(`ℹ️  Found ${declaredKeys.size} declared keys in TranslationDict (types.ts).`);

// 2. Parse locale keys
function getLocaleKeys(localeFileName) {
  const filePath = path.join(localesDir, localeFileName);
  if (!fs.existsSync(filePath)) return new Set();
  const content = fs.readFileSync(filePath, "utf8");
  const keys = new Set();
  const keyRegex = /"([^"]+)":\s*"/g;
  let m;
  while ((m = keyRegex.exec(content)) !== null) {
    keys.add(m[1]);
  }
  return keys;
}

function getLocaleEntries(localeFileName) {
  const filePath = path.join(localesDir, localeFileName);
  if (!fs.existsSync(filePath)) return new Map();
  const content = fs.readFileSync(filePath, "utf8");
  const entries = new Map();
  const entryRegex = /"([^"\n]+)":\s*"((?:\\.|[^"\\])*)"/g;
  let m;
  while ((m = entryRegex.exec(content)) !== null) {
    entries.set(m[1], m[2]);
  }
  return entries;
}

function getPlaceholders(value) {
  return [...value.matchAll(/\{[^}]+\}/g)].map((match) => match[0]).sort();
}

const zhCnKeys = getLocaleKeys("zh-CN.ts");
const enUsKeys = getLocaleKeys("en-US.ts");
const enUsEntries = getLocaleEntries("en-US.ts");

const qualityKeys = [
  "iso.preflight_title",
  "iso.conflict_title",
  "iso.conflict_desc",
  "iso.conflict_action",
  "iso.conflict_keep_both",
  "iso.conflict_replace",
  "iso.conflict_skip",
  "iso.conflict_apply_all",
  "iso.conflict_cancel",
  "iso.conflict_confirm",
  "iso.preflight_failed",
];

console.log(`ℹ️  Found ${zhCnKeys.size} keys in zh-CN.ts, ${enUsKeys.size} keys in en-US.ts.`);

// Check for keys in zh-CN missing in types.ts
const missingInTypesFromZh = [...zhCnKeys].filter((k) => !declaredKeys.has(k));
if (missingInTypesFromZh.length > 0) {
  hasErrors = true;
  console.error(
    `\n❌ Found ${missingInTypesFromZh.length} key(s) in zh-CN.ts NOT declared in TranslationDict (types.ts):`
  );
  missingInTypesFromZh.forEach((k) => console.error(`   - "${k}"`));
}

// Check for keys in types.ts missing in zh-CN.ts
const missingInZhFromTypes = [...declaredKeys].filter((k) => !zhCnKeys.has(k));
if (missingInZhFromTypes.length > 0) {
  hasErrors = true;
  console.error(`\n❌ Found ${missingInZhFromTypes.length} key(s) in TranslationDict NOT present in zh-CN.ts:`);
  missingInZhFromTypes.forEach((k) => console.error(`   - "${k}"`));
}

// Check for keys in types.ts missing in en-US.ts
const missingInEnFromTypes = [...declaredKeys].filter((k) => !enUsKeys.has(k));
if (missingInEnFromTypes.length > 0) {
  hasErrors = true;
  console.error(`\n❌ Found ${missingInEnFromTypes.length} key(s) in TranslationDict NOT present in en-US.ts:`);
  missingInEnFromTypes.forEach((k) => console.error(`   - "${k}"`));
}

// Check every locale to prevent non-default languages from silently falling back to English.
const localeFiles = fs.readdirSync(localesDir).filter((fileName) => fileName.endsWith(".ts"));
for (const localeFile of localeFiles) {
  const localeKeys = getLocaleKeys(localeFile);
  const localeEntries = getLocaleEntries(localeFile);
  const missingKeys = [...declaredKeys].filter((key) => !localeKeys.has(key));
  const extraKeys = [...localeKeys].filter((key) => !declaredKeys.has(key));

  if (missingKeys.length > 0) {
    hasErrors = true;
    console.error(`\n❌ ${localeFile} is missing ${missingKeys.length} declared key(s):`);
    missingKeys.forEach((key) => console.error(`   - "${key}"`));
  }

  if (extraKeys.length > 0) {
    hasErrors = true;
    console.error(`\n❌ ${localeFile} contains ${extraKeys.length} undeclared key(s):`);
    extraKeys.forEach((key) => console.error(`   - "${key}"`));
  }

  for (const key of qualityKeys) {
    const sourceValue = enUsEntries.get(key);
    const localeValue = localeEntries.get(key);
    if (sourceValue && localeValue && JSON.stringify(getPlaceholders(sourceValue)) !== JSON.stringify(getPlaceholders(localeValue))) {
      hasErrors = true;
      console.error(`\n❌ ${localeFile} has placeholder mismatch for "${key}":`);
      console.error(`   - English: ${getPlaceholders(sourceValue).join(", ") || "(none)"}`);
      console.error(`   - ${localeFile}: ${getPlaceholders(localeValue).join(", ") || "(none)"}`);
    }
  }

  if (localeFile !== "en-US.ts") {
    for (const key of qualityKeys) {
      const sourceValue = enUsEntries.get(key);
      const localeValue = localeEntries.get(key);
      const sameTextAllowed = new Set(["iso.conflict_action"]);
      if (sourceValue && localeValue && localeValue === sourceValue && !sameTextAllowed.has(key)) {
        hasErrors = true;
        console.error(`\n❌ ${localeFile} reuses English copy for localized key "${key}"`);
      }
    }
  }
}

// 3. Scan src/ for t('...') usages in code
function scanDirectory(dir) {
  let codeUsed = [];
  const entries = fs.readdirSync(dir, { withFileTypes: true });
  for (const entry of entries) {
    const fullPath = path.join(dir, entry.name);
    if (entry.isDirectory()) {
      if (entry.name !== "locales") {
        codeUsed = codeUsed.concat(scanDirectory(fullPath));
      }
    } else if (
      entry.isFile() &&
      (entry.name.endsWith(".vue") || entry.name.endsWith(".ts") || entry.name.endsWith(".js"))
    ) {
      if (fullPath.includes("types.ts")) continue;
      const content = fs.readFileSync(fullPath, "utf8");
      const lines = content.split("\n");
      const tRegex = /\b(?:t|\$t)\(\s*['"]([a-zA-Z0-9_.-]+)['"]/g;
      lines.forEach((lineText, idx) => {
        let m;
        while ((m = tRegex.exec(lineText)) !== null) {
          const key = m[1];
          if (key.includes(".")) {
            codeUsed.push({ key, file: path.relative(frontendDir, fullPath), line: idx + 1 });
          }
        }
      });
    }
  }
  return codeUsed;
}

const usedInCode = scanDirectory(srcDir);
const undeclaredUsedInCode = usedInCode.filter((item) => !declaredKeys.has(item.key));

if (undeclaredUsedInCode.length > 0) {
  hasErrors = true;
  console.error(
    `\n❌ Found ${undeclaredUsedInCode.length} i18n key(s) used in frontend code but NOT declared in TranslationDict:`
  );
  undeclaredUsedInCode.forEach((item) => console.error(`   - "${item.key}" at ${item.file}:${item.line}`));
}

if (hasErrors) {
  console.error("\n💥 [i18n Check Failed] Please fix the above key misalignments!");
  process.exit(1);
} else {
  console.log("\n✅ [i18n Check Passed] All i18n keys are 100% declared, defined, and aligned!");
  process.exit(0);
}

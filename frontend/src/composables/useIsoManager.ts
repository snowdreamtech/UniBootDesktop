// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import { ref } from "vue";
import { currentLang, t } from "../i18n";
import { SelectIsoFiles } from "../../wailsjs/go/main/App";
import { logUserAction } from "../utils/logger";

export interface IsoFileItem {
  name: string;
  path: string;
}

export const SUPPORTED_IMAGE_EXTS = [
  ".iso",
  ".img",
  ".wim",
  ".vhd",
  ".vhdx",
  ".vti",
  ".efi",
  ".bin",
  ".xz",
  ".gz",
  ".raw",
];

export function normalizeIsoSourcePath(filePath: string): string {
  return (filePath || "").trim().replace(/\\/g, "/").replace(/\/+/g, "/");
}

export function useIsoManager(showToast: (msg: string, type: "info" | "warning" | "error" | "success") => void) {
  const selectedIsoFiles = ref<IsoFileItem[]>([]);
  const isoCopyStatus = ref<string>("");

  function isSupportedImage(filePath: string): boolean {
    const lower = (filePath || "").toLowerCase();
    return SUPPORTED_IMAGE_EXTS.some((ext) => lower.endsWith(ext));
  }

  function addIsoFilesByPaths(paths: string[]): number {
    if (!paths || paths.length === 0) return 0;
    let added = 0;
    let duplicateCount = 0;
    let invalidCount = 0;

    for (const p of paths) {
      if (!p) continue;
      if (!isSupportedImage(p)) {
        invalidCount++;
        continue;
      }
      const name = p.split(/[/\\]/).pop() || p;
      const normalizedPath = normalizeIsoSourcePath(p);
      const existingIndex = selectedIsoFiles.value.findIndex((f) => normalizeIsoSourcePath(f.path) === normalizedPath);

      if (existingIndex >= 0) {
        // If the existing entry only has the filename, upgrade it to the full absolute path
        if (
          (p.includes("/") || p.includes("\\")) &&
          !selectedIsoFiles.value[existingIndex].path.includes("/") &&
          !selectedIsoFiles.value[existingIndex].path.includes("\\")
        ) {
          selectedIsoFiles.value[existingIndex].path = p;
        }
        duplicateCount++;
        continue;
      }

      selectedIsoFiles.value.push({ name, path: p });
      added++;
      logUserAction("INFO", "Added image source file", `${name} (${p})`);
    }

    if (added > 0) {
      showToast(t("deploy.toast_added_iso", { count: added }), "success");
    } else if (duplicateCount > 0) {
      showToast(
        currentLang.value.startsWith("zh")
          ? "所选镜像文件已在列表中，已自动忽略重复项"
          : "Selected image file is already in the list, duplicate ignored",
        "info"
      );
    } else if (invalidCount > 0) {
      showToast(t("iso.drag_unsupported"), "warning");
    }

    return added;
  }

  function triggerHtmlFileInput() {
    if (typeof document === "undefined") return;
    const input = document.createElement("input");
    input.type = "file";
    input.multiple = true;
    input.accept = SUPPORTED_IMAGE_EXTS.join(",");
    input.style.display = "none";

    input.onchange = () => {
      if (!input.files || input.files.length === 0) return;
      const paths: string[] = [];
      for (let i = 0; i < input.files.length; i++) {
        const file = input.files[i];
        // In WebKit/Electron or desktop webview, file.path or webkitRelativePath might exist
        const resolvedPath = (file as any).path || file.name;
        if (resolvedPath) {
          paths.push(resolvedPath);
        }
      }
      if (paths.length > 0) {
        addIsoFilesByPaths(paths);
      }
      document.body.removeChild(input);
    };

    input.oncancel = () => {
      try {
        document.body.removeChild(input);
      } catch {
        // ignore
      }
    };

    document.body.appendChild(input);
    input.click();
  }

  async function handleSelectIsoFiles() {
    const hasWails =
      typeof window !== "undefined" &&
      typeof (window as any)?.go?.main?.App?.SelectIsoFiles === "function";

    if (hasWails) {
      try {
        const paths: string[] = await SelectIsoFiles(
          t("dialog.selectIsoTitle"),
          t("dialog.ventoyFilter"),
          t("dialog.allFilesFilter")
        );
        if (paths && paths.length > 0) {
          addIsoFilesByPaths(paths);
        }
      } catch (err: any) {
        console.error("SelectIsoFiles error:", err);
        showToast(err?.message || "打开原生文件选择器失败，已切换至备用文件选择", "warning");
        triggerHtmlFileInput();
      }
    } else {
      triggerHtmlFileInput();
    }
  }

  function removeIsoFile(index: number) {
    const item = selectedIsoFiles.value[index];
    const name = item ? (typeof item === "string" ? item : item.name || item.path) : "";
    selectedIsoFiles.value.splice(index, 1);
    if (name) {
      logUserAction("INFO", "User removed ISO source file from selection list", name);
    }
  }

  function clearIsoFiles() {
    selectedIsoFiles.value = [];
    logUserAction("INFO", "User cleared all ISO source files from selection list");
  }

  return {
    selectedIsoFiles,
    isoCopyStatus,
    isSupportedImage,
    addIsoFilesByPaths,
    handleSelectIsoFiles,
    removeIsoFile,
    clearIsoFiles,
  };
}

import { onMounted, onUnmounted, type Ref } from "vue";
import { GetRecentLogs } from "../../wailsjs/go/main/App";
import { syncWindowTheme } from "./useTheme";

export interface UseAppRuntimeEventsOptions {
  t: (key: any, named?: Record<string, any>) => string;
  loadConfig: () => Promise<void>;
  refreshDisks: () => Promise<void>;
  checkQemu: () => Promise<void>;
  checkVentoyStatus: () => Promise<void>;
  setInitialLogs: (logs: any[]) => void;
  appendLogEntry: (entry: any) => void;
  isoCopyStatus: Ref<string>;
  deployProgress?: Ref<number>;
  deploySpeedMBps: Ref<number>;
  deployElapsedSec: Ref<number>;
  deployEtaSec: Ref<number>;
  isDeploying: Ref<boolean>;
  isAboutOpen: Ref<boolean>;
  addIsoFilesByPaths: (paths: string[]) => void;
  batchDeployInfo?: Ref<any>;
  showToast?: (message: string, type?: any) => void;
}

export function useAppRuntimeEvents(options: UseAppRuntimeEventsOptions) {
  const {
    t,
    loadConfig,
    refreshDisks,
    checkQemu,
    checkVentoyStatus,
    setInitialLogs,
    appendLogEntry,
    isoCopyStatus,
    deploySpeedMBps,
    deployElapsedSec,
    deployEtaSec,
    isDeploying,
    isAboutOpen,
    addIsoFilesByPaths,
    batchDeployInfo,
  } = options;

  let onConfigReady: (() => void) | null = null;

  onMounted(() => {
    syncWindowTheme();
    loadConfig();
    refreshDisks();
    checkQemu();
    checkVentoyStatus();

    onConfigReady = () => {
      checkVentoyStatus();
    };
    window.addEventListener("uniboot:config-ready", onConfigReady);

    GetRecentLogs()
      .then((logs: any[]) => {
        setInitialLogs(logs);
      })
      .catch(() => {
        setInitialLogs([]);
      });

    if (window.runtime && window.runtime.EventsOn) {
      window.runtime.EventsOn("ventoy-status-changed", () => {
        checkVentoyStatus();
      });

      window.runtime.EventsOn("log:entry", (entry: any) => {
        appendLogEntry(entry);
      });

      window.runtime.EventsOn("iso-copy-progress", (data: any) => {
        if (data) {
          isoCopyStatus.value = t("disk.writingImageProgress", {
            fileIndex: data.fileIndex,
            totalFiles: data.totalFiles,
            currentFile: data.currentFile,
            progress: data.progress.toFixed(1),
          });
          if (batchDeployInfo?.value) {
            batchDeployInfo.value.currentStage = `${t("deploy.stage_iso_copy")} (${data.currentFile} ${data.progress.toFixed(0)}%)`;
          }
          if (data.speedMBps !== undefined) deploySpeedMBps.value = data.speedMBps;
          if (data.elapsedSec !== undefined) deployElapsedSec.value = data.elapsedSec;
          if (data.etaSec !== undefined) deployEtaSec.value = data.etaSec;
        }
      });

      let diskChangedDebounceTimer: ReturnType<typeof setTimeout> | null = null;

      window.runtime.EventsOn("disk-list-changed", () => {
        console.log("[RuntimeEvents] Removable storage change detected, debouncing drive list refresh");
        appendLogEntry({
          id: Date.now(),
          timestamp: new Date().toISOString(),
          level: "INFO",
          message: "Removable storage change detected, refreshing drive list",
        });
        if (!isDeploying.value) {
          if (diskChangedDebounceTimer) {
            clearTimeout(diskChangedDebounceTimer);
          }
          diskChangedDebounceTimer = setTimeout(() => {
            diskChangedDebounceTimer = null;
            refreshDisks();
          }, 250);
        }
      });

      window.runtime.EventsOn("vm-session-ended", (data: any) => {
        if (!isDeploying.value) {
          refreshDisks();
        }
        if (options.showToast) {
          if (data && data.error) {
            options.showToast(t("vm.session_ended_error", { error: data.error }), "warning");
          } else {
            options.showToast(t("vm.session_ended_success"), "success");
          }
        }
      });

      window.runtime.EventsOn("open-about-modal", () => {
        isAboutOpen.value = true;
      });

      window.runtime.EventsOn("open-log-modal", () => {
        const el = document.querySelector(".log-section-card");
        if (el) {
          el.scrollIntoView({ behavior: "smooth" });
        }
      });

      if (typeof window.runtime.OnFileDrop === "function") {
        window.runtime.OnFileDrop((_x: number, _y: number, paths: string[]) => {
          if (paths && paths.length > 0) {
            addIsoFilesByPaths(paths);
          }
        }, false);
      }
    }
  });

  onUnmounted(() => {
    if (onConfigReady) {
      window.removeEventListener("uniboot:config-ready", onConfigReady);
      onConfigReady = null;
    }
    if (window.runtime && typeof window.runtime.OnFileDropOff === "function") {
      window.runtime.OnFileDropOff();
    }
  });
}

import { onMounted, onUnmounted, type Ref } from 'vue';
import { GetRecentLogs } from '../../wailsjs/go/main/App';

export interface UseAppRuntimeEventsOptions {
  t: (key: any, named?: Record<string, any>) => string;
  loadConfig: () => Promise<void>;
  refreshDisks: () => Promise<void>;
  checkQemu: () => Promise<void>;
  checkVentoyStatus: () => Promise<void>;
  setInitialLogs: (logs: any[]) => void;
  appendLogEntry: (entry: any) => void;
  isoCopyStatus: Ref<string>;
  deployProgress: Ref<number>;
  deploySpeedMBps: Ref<number>;
  deployElapsedSec: Ref<number>;
  deployEtaSec: Ref<number>;
  isDeploying: Ref<boolean>;
  isAboutOpen: Ref<boolean>;
  addIsoFilesByPaths: (paths: string[]) => void;
  batchDeployInfo?: Ref<any>;
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
    deployProgress,
    deploySpeedMBps,
    deployElapsedSec,
    deployEtaSec,
    isDeploying,
    isAboutOpen,
    addIsoFilesByPaths,
    batchDeployInfo,
  } = options;

  onMounted(() => {
    loadConfig();
    refreshDisks();
    checkQemu();
    checkVentoyStatus();

    GetRecentLogs().then((logs: any[]) => {
      setInitialLogs(logs);
    }).catch(() => {
      setInitialLogs([]);
    });

    if (window.runtime && window.runtime.EventsOn) {
      window.runtime.EventsOn("log:entry", (entry: any) => {
        appendLogEntry(entry);
      });

      window.runtime.EventsOn("iso-copy-progress", (data: any) => {
        if (data) {
          isoCopyStatus.value = t('disk.writingImageProgress', {
            fileIndex: data.fileIndex,
            totalFiles: data.totalFiles,
            currentFile: data.currentFile,
            progress: data.progress.toFixed(1)
          });
          // Only allow single-disk ISO progress to control deployProgress if not in batch deployment mode
          if (!batchDeployInfo?.value) {
            deployProgress.value = Math.min(99, Math.max(50, Math.floor(50 + data.progress / 2)));
          }
          if (data.speedMBps !== undefined) deploySpeedMBps.value = data.speedMBps;
          if (data.elapsedSec !== undefined) deployElapsedSec.value = data.elapsedSec;
          if (data.etaSec !== undefined) deployEtaSec.value = data.etaSec;
        }
      });

      window.runtime.EventsOn("disk-list-changed", () => {
        if (!isDeploying.value) {
          refreshDisks();
        }
      });

      window.runtime.EventsOn("open-about-modal", () => {
        isAboutOpen.value = true;
      });

      window.runtime.EventsOn("open-log-modal", () => {
        const el = document.querySelector('.log-section-card');
        if (el) {
          el.scrollIntoView({ behavior: 'smooth' });
        }
      });

      if (typeof window.runtime.OnFileDrop === 'function') {
        window.runtime.OnFileDrop((_x: number, _y: number, paths: string[]) => {
          if (paths && paths.length > 0) {
            addIsoFilesByPaths(paths);
          }
        }, false);
      }
    }
  });

  onUnmounted(() => {
    if (window.runtime && typeof window.runtime.OnFileDropOff === 'function') {
      window.runtime.OnFileDropOff();
    }
  });
}

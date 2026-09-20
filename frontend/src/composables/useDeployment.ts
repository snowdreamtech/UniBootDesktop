import { computed, ref, watch, type Ref } from 'vue';
import type { InstallDiagnosticsData } from '../components/DiagnosticsModal.vue';
import type { DiskInfo } from '../components/DiskPanel.vue';
import { logUserAction } from '../utils/logger';

export interface DeployBannerState {
  visible: boolean;
  msg: string;
  targets: string[];
  autoEjected?: boolean;
  mode?: 'cloud' | 'hybrid';
  dismissed?: boolean;
}

export interface BatchDeployProgress {
  totalDisks: number;
  currentDiskIndex: number;
  currentDisk: string;
  currentStage: string;
  diskProgress: number;
  overallProgress: number;
  speedMBps: number;
  elapsedSec: number;
  etaSec: number;
}

export interface UseDeploymentOptions {
  t: (key: any, named?: Record<string, any>) => string;
  showToast: (msg: string, type?: 'info' | 'success' | 'warning' | 'error') => void;
  selectionMode: Ref<'single' | 'batch'>;
  selectedDisk: Ref<DiskInfo | null>;
  selectedDevices: Ref<Set<string>>;
  diskList: Ref<DiskInfo[]>;
  selectedIsoFiles: Ref<Array<{ path: string }>>;
  refreshDisks: () => Promise<void>;
  openSettings?: (tab?: 'general' | 'network' | 'uniboot' | 'ventoy') => void;
}

export function useDeployment(options: UseDeploymentOptions) {
  const {
    t,
    showToast,
    selectionMode,
    selectedDisk,
    selectedDevices,
    diskList,
    selectedIsoFiles,
    refreshDisks,
    openSettings,
  } = options;

  const activeMode = ref<'cloud' | 'hybrid'>('cloud');
  const selectedFsType = ref<'exFAT' | 'NTFS' | 'FAT32' | 'ext4'>('exFAT');
  const autoEjectAfterDeploy = ref(false);

  const isDeploying = ref(false);
  const deployProgress = ref(0);
  const deploySpeedMBps = ref(0);
  const deployElapsedSec = ref(0);
  const deployEtaSec = ref(0);

  // Batch deployment state
  const batchDeployInfo = ref<BatchDeployProgress | null>(null);

  const isDeployConfirmOpen = ref(false);
  const pendingTargets = ref<string[]>([]);
  const pendingTargetSnapshots = ref<DiskInfo[]>([]);

  const isVentoyAlertOpen = ref(false);
  const ventoyAlertTitle = ref('');
  const ventoyAlertMessage = ref('');
  const ventoyAlertAction = ref<'open_settings' | 'switch_b'>('open_settings');
  const ventoyStatus = ref({ valid: true, version: '', message: '', executablePath: '' });

  const isDiagnosticsOpen = ref(false);
  const currentDiagnostics = ref<InstallDiagnosticsData | null>(null);
  const currentDiagErrorMsg = ref('');

  const deploySuccessBanner = ref<DeployBannerState>({
    visible: false,
    msg: '',
    targets: [],
    autoEjected: false,
    dismissed: false,
  });

  const isMacOs = computed(() => navigator.userAgent.includes('Mac') || navigator.platform.includes('Mac'));

  watch(activeMode, (newMode) => {
    logUserAction('INFO', 'User switched deployment mode', newMode);
  });

  watch(selectedFsType, (newFs) => {
    logUserAction('INFO', 'User selected target file system', newFs);
  });

  const showDeploySuccessBanner = computed(() => {
    const b = deploySuccessBanner.value;
    return Boolean(
      !isDeploying.value &&
      b.visible &&
      !b.dismissed &&
      b.mode === activeMode.value
    );
  });

  function dismissDeploySuccessBanner() {
    deploySuccessBanner.value.dismissed = true;
    deploySuccessBanner.value.visible = false;
    logUserAction('DEBUG', 'User dismissed deployment success banner');
  }

  function checkDiskCanUpdateNonDestructively(d: DiskInfo, mode: 'cloud' | 'hybrid'): boolean {
    if (!d) return false;
    if (mode === 'cloud') {
      // Cloud boot mode: both existing cloud boot and Ventoy hybrid disks can be updated non-destructively without wiping user partitions
      return Boolean(d.isCloudMode || d.isRealVentoy);
    } else {
      // Hybrid mode: only genuine Ventoy disks support non-destructive upgrade (ventoy -u) preserving files.
      // Pure cloud boot disks lack Ventoy MBR and dual-partition layouts and MUST be fully formatted to install Ventoy.
      return Boolean(d.isRealVentoy);
    }
  }

  const isSelectedVentoyDisk = computed(() => {
    if (selectionMode.value === 'single' && selectedDisk.value) {
      return checkDiskCanUpdateNonDestructively(selectedDisk.value, activeMode.value);
    }
    return false;
  });

  const isNonDestructive = computed(() => {
    if (selectionMode.value === 'single') {
      return isSelectedVentoyDisk.value;
    }
    if (selectedDevices.value.size === 0) return false;
    return Array.from(selectedDevices.value).every((dev: string) => {
      const d = diskList.value.find((disk: DiskInfo) => disk.device === dev);
      return d ? checkDiskCanUpdateNonDestructively(d, activeMode.value) : false;
    });
  });

  const ventoyCountInBatch = computed(() => {
    if (selectionMode.value !== 'batch' || selectedDevices.value.size === 0) return 0;
    let count = 0;
    selectedDevices.value.forEach((dev: string) => {
      const d = diskList.value.find((disk: DiskInfo) => disk.device === dev);
      if (d && checkDiskCanUpdateNonDestructively(d, activeMode.value)) {
        count++;
      }
    });
    return count;
  });

  const deployBtnText = computed(() => {
    if (isDeploying.value) return t('deploy.writing');

    if (selectionMode.value === 'single') {
      if (isNonDestructive.value) {
        return t('deploy.start_update');
      }
      return activeMode.value === 'cloud' ? t('deploy.start_cloud_create') : t('deploy.start_create');
    }

    const total = selectedDevices.value.size;
    const bootCount = ventoyCountInBatch.value;
    const blankCount = total - bootCount;

    if (bootCount === total && total > 0) {
      return t('deploy.batch_update', { count: total });
    } else if (blankCount === total && total > 0) {
      return t('deploy.batch_create', { count: total });
    } else {
      return t('deploy.batch_mixed', { count: total });
    }
  });

  const deployDisabledReason = computed(() => {
    if (isDeploying.value) return t('deploy.tip_writing');
    if (selectionMode.value === 'single' && !selectedDisk.value) return t('deploy.tip_select_single');
    if (selectionMode.value === 'batch' && selectedDevices.value.size === 0) return t('deploy.tip_select_batch');
    if (activeMode.value === 'hybrid' && !isNonDestructive.value && !ventoyStatus.value.valid) {
      if (isMacOs.value) {
        return t('deploy.tip_macos_unsupported');
      }
      return ventoyStatus.value.message || t('deploy.tip_need_ventoy');
    }
    if (selectionMode.value === 'batch') {
      const total = selectedDevices.value.size;
      const bootCount = ventoyCountInBatch.value;
      const blankCount = total - bootCount;
      if (bootCount > 0 && blankCount > 0) {
        return t('deploy.tip_batch_mixed', { bootCount, blankCount });
      }
      if (bootCount === total && total > 0) {
        return t('deploy.tip_batch_update_all', { count: total });
      }
    }
    return '';
  });

  const activeVmTargetDevice = computed(() => {
    if (selectionMode.value === 'single') {
      return selectedDisk.value?.device || '';
    }
    if (selectedDevices.value.size > 0) {
      return Array.from(selectedDevices.value)[0];
    }
    return '';
  });

  const activeVmTargetName = computed(() => {
    if (selectionMode.value === 'single') {
      return selectedDisk.value?.name || selectedDisk.value?.device || '';
    }
    if (selectedDevices.value.size > 0) {
      const firstDev = Array.from(selectedDevices.value)[0];
      const found = diskList.value.find(d => d.device === firstDev);
      return found?.name || firstDev;
    }
    return '';
  });

  async function checkVentoyStatus() {
    if (window.go && window.go.main && window.go.main.App && window.go.main.App.ValidateVentoyCli) {
      try {
        const res = await window.go.main.App.ValidateVentoyCli('');
        if (res) {
          ventoyStatus.value = res;
        }
      } catch (e) {
        console.error('Failed to validate Ventoy status:', e);
      }
    }
  }

  async function selectMode(mode: 'cloud' | 'hybrid') {
    activeMode.value = mode;
    if (mode === 'hybrid') {
      await checkVentoyStatus();
      if (!isNonDestructive.value && !ventoyStatus.value.valid) {
        if (isMacOs.value) {
          showToast(t('deploy.tip_macos_unsupported'), 'warning');
        } else {
          showToast(t('deploy.tip_need_ventoy'), 'warning');
        }
      }
    }
  }

  function openVentoyAlert(title: string, message: string, action: 'open_settings' | 'switch_b') {
    ventoyAlertTitle.value = title;
    ventoyAlertMessage.value = message;
    ventoyAlertAction.value = action;
    isVentoyAlertOpen.value = true;
  }

  function handleVentoyAlertAction() {
    isVentoyAlertOpen.value = false;
    if (openSettings) {
      openSettings('ventoy');
    }
  }

  function handleVentoyAlertSwitchB() {
    isVentoyAlertOpen.value = false;
    activeMode.value = 'cloud';
    showToast(t('deploy.toast_switched_b'), 'success');
  }

  function openDiagnosticsModal(diag: InstallDiagnosticsData | null, msg: string) {
    currentDiagnostics.value = diag;
    currentDiagErrorMsg.value = msg || t('deploy.alert_fail', { msg: '' });
    isDiagnosticsOpen.value = true;
  }

  function handleCopyReport() {
    logUserAction('INFO', 'User copied diagnostics report to clipboard');
    showToast(t('diag.toast_copied'), 'info');
  }

  function handleRetryDeploy() {
    logUserAction('INFO', 'User clicked retry deployment from diagnostics modal');
    isDiagnosticsOpen.value = false;
    openDeployConfirm();
  }

  function openDeployConfirm() {
    dismissDeploySuccessBanner();

    let targets: string[] = [];
    if (selectionMode.value === 'single') {
      if (!selectedDisk.value) return;
      targets = [selectedDisk.value.device];
    } else {
      targets = Array.from(selectedDevices.value);
      if (targets.length === 0) return;
    }

    const refreshedSnapshots = targets.map((device) => diskList.value.find((disk) => disk.device === device));
    if (refreshedSnapshots.some((disk) => !disk)) {
      showToast(t('deploy.toast_target_changed'), 'error');
      return;
    }

    pendingTargets.value = targets;
    pendingTargetSnapshots.value = refreshedSnapshots as DiskInfo[];
    isDeployConfirmOpen.value = true;
  }

  async function handleDeployBtnClick() {
    if (isDeploying.value) return;

    dismissDeploySuccessBanner();

    if (selectionMode.value === 'single' && !selectedDisk.value) {
      showToast(t('deploy.toast_select_target'), 'warning');
      return;
    }
    if (selectionMode.value === 'batch' && selectedDevices.value.size === 0) {
      showToast(t('deploy.toast_select_batch'), 'warning');
      return;
    }

    if (activeMode.value === 'hybrid' && !isNonDestructive.value) {
      if (!ventoyStatus.value.valid) {
        if (isMacOs.value) {
          openVentoyAlert(
            t('deploy.macos_alert_title'),
            t('deploy.macos_alert_desc'),
            'switch_b'
          );
        } else {
          openVentoyAlert(
            t('deploy.no_ventoy_title'),
            ventoyStatus.value.message || t('deploy.no_ventoy_desc'),
            'open_settings'
          );
        }
        return;
      }
      checkVentoyStatus().catch(() => {});
    }

    openDeployConfirm();
  }

  async function startDeployment() {
    if (isDeploying.value) return;

    dismissDeploySuccessBanner();

    let targets: string[] = [];
    if (selectionMode.value === 'single') {
      if (!selectedDisk.value) return;
      targets = [selectedDisk.value.device];
    } else {
      targets = Array.from(selectedDevices.value);
      if (targets.length === 0) return;
    }

    isDeploying.value = true;
    deployProgress.value = 5;
    batchDeployInfo.value = null;

    let unsubCloudProgress: (() => void) | null = null;
    let unsubBatchProgress: (() => void) | null = null;

    if (activeMode.value === 'cloud' && window.runtime && window.runtime.EventsOn) {
      // Listen for batch deployment progress (multi-disk)
      window.runtime.EventsOn('cloud-deploy-batch-progress', (progress: BatchDeployProgress) => {
        deployProgress.value = progress.overallProgress;
        batchDeployInfo.value = progress;
        // Update speed and time information from batch progress
        deploySpeedMBps.value = progress.speedMBps;
        deployElapsedSec.value = progress.elapsedSec;
        deployEtaSec.value = progress.etaSec;
      });
      unsubBatchProgress = () => {
        if (window.runtime && window.runtime.EventsOff) {
          window.runtime.EventsOff('cloud-deploy-batch-progress');
        }
      };

      // Listen for single disk progress (for backward compatibility)
      window.runtime.EventsOn('cloud-deploy-progress', (progress: number) => {
        if (!batchDeployInfo.value) {
          deployProgress.value = progress;
        }
      });
      unsubCloudProgress = () => {
        if (window.runtime && window.runtime.EventsOff) {
          window.runtime.EventsOff('cloud-deploy-progress');
        }
      };
    }

    let success = true;
    let resultMsg = '';
    let latestDiagnostics: any = null;

    try {
      if (window.go && window.go.main && window.go.main.App) {
        const isoPaths = selectedIsoFiles.value.map(f => f.path);
        if (targets.length === 1) {
          const expected = pendingTargetSnapshots.value[0];
          if (!expected) throw new Error(t('deploy.toast_target_changed'));
          let res: any;
          if (activeMode.value === 'cloud') {
            res = await window.go.main.App.DeployCloudMode(targets[0], selectedFsType.value, expected);
          } else {
            res = await window.go.main.App.DeployHybridMode(targets[0], selectedFsType.value, isoPaths, expected);
          }
          if (res) {
            success = res.success;
            resultMsg = res.message || '';
            if (res.diagnostics) {
              latestDiagnostics = res.diagnostics;
            }
          }
        } else {
          if (pendingTargetSnapshots.value.length !== targets.length) {
            throw new Error(t('deploy.toast_target_changed'));
          }
          let resList: any[];
          if (activeMode.value === 'cloud') {
            resList = await window.go.main.App.DeployCloudModeBatch(targets, selectedFsType.value, pendingTargetSnapshots.value);
          } else {
            resList = await window.go.main.App.DeployHybridModeBatch(targets, selectedFsType.value, isoPaths, pendingTargetSnapshots.value);
          }
          if (resList && resList.length > 0) {
            const failed = resList.filter(r => !r.success);
            if (failed.length > 0) {
              success = false;
              resultMsg = failed.map(f => `${f.target}: ${f.message}`).join('\n');
              if (failed[0].diagnostics) {
                latestDiagnostics = failed[0].diagnostics;
              }
            } else {
              resultMsg = t('deploy.result_batch_success', { count: resList.length });
            }
          }
        }
      } else {
        // Mock execution for browser demo
        await new Promise(r => setTimeout(r, 800));
        const modeLabel = activeMode.value === 'cloud' ? 'B' : 'A';
        resultMsg = t('deploy.result_success', { mode: modeLabel, targets: targets.join(', ') });
      }
    } catch (e: any) {
      console.error(e);
      success = false;
      resultMsg = e?.message || String(e);
    } finally {
      if (unsubCloudProgress) {
        unsubCloudProgress();
        unsubCloudProgress = null;
      }
      if (unsubBatchProgress) {
        unsubBatchProgress();
        unsubBatchProgress = null;
      }
    }

    if (success) {
      deployProgress.value = 100;
      setTimeout(async () => {
        isDeploying.value = false;
        deployProgress.value = 0;
        batchDeployInfo.value = null;

        let autoEjectedCount = 0;
        if (autoEjectAfterDeploy.value) {
          for (const dev of targets) {
            try {
              if (window.go && window.go.main && window.go.main.App && window.go.main.App.EjectDisk) {
                await window.go.main.App.EjectDisk(dev);
                autoEjectedCount++;
              }
            } catch (ejectErr) {
              console.warn(`Auto eject failed for ${dev}:`, ejectErr);
            }
          }
        }

        await refreshDisks();

        deploySuccessBanner.value = {
          visible: true,
          msg: resultMsg,
          targets: [...targets],
          autoEjected: autoEjectedCount > 0,
          mode: activeMode.value,
          dismissed: false,
        };

        if (autoEjectedCount > 0) {
          showToast(t('deploy.toast_auto_ejected', { count: autoEjectedCount }), 'success');
        }
      }, 200);
    } else {
      isDeploying.value = false;
      deployProgress.value = 0;
      batchDeployInfo.value = null;
      deploySpeedMBps.value = 0;
      deployElapsedSec.value = 0;
      deployEtaSec.value = 0;
      openDiagnosticsModal(latestDiagnostics, resultMsg);
    }
  }

  async function handleCancelDeploy() {
    logUserAction('WARN', 'User clicked cancel deployment button');
    const app = (window as any)?.go?.main?.App;
    if (app && typeof app.CancelDeployment === 'function') {
      try {
        const cancelled = await app.CancelDeployment();
        if (cancelled) {
          showToast(t('deploy.toast_cancelled'), 'info');
          logUserAction('INFO', 'Deployment task cancelled successfully');

          // Clean up deployment state
          isDeploying.value = false;
          deployProgress.value = 0;
          batchDeployInfo.value = null;
          deploySpeedMBps.value = 0;
          deployElapsedSec.value = 0;
          deployEtaSec.value = 0;
        }
      } catch (e: any) {
        console.error('Failed to cancel deployment:', e);
      }
    }
  }

  return {
    activeMode,
    selectedFsType,
    autoEjectAfterDeploy,
    isDeploying,
    deployProgress,
    batchDeployInfo,
    deploySpeedMBps,
    deployElapsedSec,
    deployEtaSec,
    isDeployConfirmOpen,
    pendingTargets,
    pendingTargetSnapshots,
    isVentoyAlertOpen,
    ventoyAlertTitle,
    ventoyAlertMessage,
    ventoyAlertAction,
    ventoyStatus,
    isDiagnosticsOpen,
    currentDiagnostics,
    currentDiagErrorMsg,
    deploySuccessBanner,
    isMacOs,
    isNonDestructive,
    deployBtnText,
    deployDisabledReason,
    activeVmTargetDevice,
    activeVmTargetName,
    showDeploySuccessBanner,
    checkVentoyStatus,
    selectMode,
    openVentoyAlert,
    handleVentoyAlertAction,
    handleVentoyAlertSwitchB,
    openDiagnosticsModal,
    handleCopyReport,
    handleRetryDeploy,
    openDeployConfirm,
    handleDeployBtnClick,
    startDeployment,
    handleCancelDeploy,
    dismissDeploySuccessBanner,
  };
}

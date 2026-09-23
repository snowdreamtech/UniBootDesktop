import { computed, ref, watch, type Ref } from "vue";
import type { InstallDiagnosticsData } from "../components/DiagnosticsModal.vue";
import type { DiskInfo } from "../components/DiskPanel.vue";
import { logUserAction } from "../utils/logger";
import { getVentoyValidationMessage, type VentoyValidation } from "../utils/ventoyValidation";

export interface IsoCopyPlanEntry {
  sourcePath: string;
  targetName: string;
  action: "replace" | "skip" | "rename";
}

export interface IsoCopyDiskPlan {
  targetDisk: string;
  entries: IsoCopyPlanEntry[];
}

export interface IsoCopyConflict {
  targetDisk: string;
  sourcePath: string;
  fileName: string;
  suggestedName: string;
  conflictType?: "source_duplicate" | "target_exists" | "source_duplicate_target_exists";
}

export interface DeployBannerState {
  visible: boolean;
  msg: string;
  targets: string[];
  autoEjected?: boolean;
  mode?: "cloud" | "hybrid";
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

function getDeployResultMessage(result: { code?: string; message?: string }): string {
  if (result.code) {
    return getVentoyValidationMessage({
      valid: false,
      version: "",
      code: result.code,
      executablePath: "",
    });
  }
  return result.message || "";
}

export interface UseDeploymentOptions {
  t: (key: any, named?: Record<string, any>) => string;
  showToast: (msg: string, type?: "info" | "success" | "warning" | "error") => void;
  selectionMode: Ref<"single" | "batch">;
  selectedDisk: Ref<DiskInfo | null>;
  selectedDevices: Ref<Set<string>>;
  diskList: Ref<DiskInfo[]>;
  selectedIsoFiles: Ref<Array<{ path: string }>>;
  refreshDisks: () => Promise<void>;
  openSettings?: (tab?: "general" | "network" | "uniboot" | "ventoy") => void;
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

  const activeMode = ref<"cloud" | "hybrid">("cloud");
  const selectedFsType = ref<"exFAT" | "NTFS" | "FAT32" | "ext4">("exFAT");
  const autoEjectAfterDeploy = ref(false);

  const isDeploying = ref(false);
  const deployProgress = ref(0);
  const deploySpeedMBps = ref(0);
  const deployElapsedSec = ref(0);
  const deployEtaSec = ref(0);

  // Batch deployment state
  const batchDeployInfo = ref<BatchDeployProgress | null>(null);

  const isDeployConfirmOpen = ref(false);
  const isIsoConflictOpen = ref(false);
  const isoConflicts = ref<IsoCopyConflict[]>([]);
  const pendingIsoPlans = ref<IsoCopyDiskPlan[]>([]);
  const isPreflight = ref(false);
  const pendingTargets = ref<string[]>([]);
  const pendingTargetSnapshots = ref<DiskInfo[]>([]);
  // Lock the ISO file list at deploy-initiation time alongside pendingTargets.
  // selectedIsoFiles is reactive and can change while conflict/confirm dialogs
  // are open, which would cause isoPaths sent to the backend to diverge from
  // the pendingIsoPlans entries that were built at T1.
  const pendingIsoFiles = ref<typeof selectedIsoFiles.value>([]);

  const isVentoyAlertOpen = ref(false);
  const ventoyAlertTitle = ref("");
  const ventoyAlertMessage = ref("");
  const ventoyAlertAction = ref<"open_settings" | "switch_b">("open_settings");
  const ventoyStatus = ref<VentoyValidation>({
    valid: true,
    version: "",
    code: "validated",
    message: "",
    executablePath: "",
  });

  const isDiagnosticsOpen = ref(false);
  const currentDiagnostics = ref<InstallDiagnosticsData | null>(null);
  const currentDiagErrorMsg = ref("");

  const deploySuccessBanner = ref<DeployBannerState>({
    visible: false,
    msg: "",
    targets: [],
    autoEjected: false,
    dismissed: false,
  });

  const isMacOs = computed(() => navigator.userAgent.includes("Mac") || navigator.platform.includes("Mac"));

  watch(activeMode, (newMode) => {
    logUserAction("INFO", "User switched deployment mode", newMode);
  });

  watch(selectedFsType, (newFs) => {
    logUserAction("INFO", "User selected target file system", newFs);
  });

  const showDeploySuccessBanner = computed(() => {
    const b = deploySuccessBanner.value;
    return Boolean(!isDeploying.value && b.visible && !b.dismissed && b.mode === activeMode.value);
  });

  function dismissDeploySuccessBanner() {
    deploySuccessBanner.value.dismissed = true;
    deploySuccessBanner.value.visible = false;
    logUserAction("DEBUG", "User dismissed deployment success banner");
  }

  function checkDiskCanUpdateNonDestructively(d: DiskInfo, mode: "cloud" | "hybrid"): boolean {
    if (!d) return false;
    if (mode === "cloud") {
      // Cloud boot mode: both existing cloud boot and Ventoy hybrid disks can be updated non-destructively without wiping user partitions
      return Boolean(d.isCloudMode || d.isRealVentoy);
    } else {
      // Hybrid mode: only genuine Ventoy disks support non-destructive upgrade (ventoy -u) preserving files.
      // Pure cloud boot disks lack Ventoy MBR and dual-partition layouts and MUST be fully formatted to install Ventoy.
      return Boolean(d.isRealVentoy);
    }
  }

  const isSelectedVentoyDisk = computed(() => {
    if (selectionMode.value === "single" && selectedDisk.value) {
      return checkDiskCanUpdateNonDestructively(selectedDisk.value, activeMode.value);
    }
    return false;
  });

  const isNonDestructive = computed(() => {
    if (selectionMode.value === "single") {
      return isSelectedVentoyDisk.value;
    }
    if (selectedDevices.value.size === 0) return false;
    return Array.from(selectedDevices.value).every((dev: string) => {
      const d = diskList.value.find((disk: DiskInfo) => disk.device === dev);
      return d ? checkDiskCanUpdateNonDestructively(d, activeMode.value) : false;
    });
  });

  const ventoyCountInBatch = computed(() => {
    if (selectionMode.value !== "batch" || selectedDevices.value.size === 0) return 0;
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
    if (isDeploying.value) return t("deploy.writing");

    if (selectionMode.value === "single") {
      if (isNonDestructive.value) {
        return t("deploy.start_update");
      }
      return activeMode.value === "cloud" ? t("deploy.start_cloud_create") : t("deploy.start_create");
    }

    const total = selectedDevices.value.size;
    const bootCount = ventoyCountInBatch.value;
    const blankCount = total - bootCount;

    if (bootCount === total && total > 0) {
      return t("deploy.batch_update", { count: total });
    } else if (blankCount === total && total > 0) {
      return t("deploy.batch_create", { count: total });
    } else {
      return t("deploy.batch_mixed", { count: total });
    }
  });

  const deployDisabledReason = computed(() => {
    if (isDeploying.value) return t("deploy.tip_writing");
    if (selectionMode.value === "single" && !selectedDisk.value) return t("deploy.tip_select_single");
    if (selectionMode.value === "batch" && selectedDevices.value.size === 0) return t("deploy.tip_select_batch");
    if (activeMode.value === "hybrid" && !isNonDestructive.value && !ventoyStatus.value.valid) {
      if (isMacOs.value) {
        return t("deploy.tip_macos_unsupported");
      }
      return getVentoyValidationMessage(ventoyStatus.value) || t("deploy.tip_need_ventoy");
    }
    if (selectionMode.value === "batch") {
      const total = selectedDevices.value.size;
      const bootCount = ventoyCountInBatch.value;
      const blankCount = total - bootCount;
      if (bootCount > 0 && blankCount > 0) {
        return t("deploy.tip_batch_mixed", { bootCount, blankCount });
      }
      if (bootCount === total && total > 0) {
        return t("deploy.tip_batch_update_all", { count: total });
      }
    }
    return "";
  });

  const activeVmTargetDevice = computed(() => {
    if (selectionMode.value === "single") {
      return selectedDisk.value?.device || "";
    }
    if (selectedDevices.value.size > 0) {
      return Array.from(selectedDevices.value)[0];
    }
    return "";
  });

  const activeVmTargetName = computed(() => {
    if (selectionMode.value === "single") {
      return selectedDisk.value?.name || selectedDisk.value?.device || "";
    }
    if (selectedDevices.value.size > 0) {
      const firstDev = Array.from(selectedDevices.value)[0];
      const found = diskList.value.find((d) => d.device === firstDev);
      return found?.name || firstDev;
    }
    return "";
  });

  async function checkVentoyStatus() {
    if (window.go && window.go.main && window.go.main.App && window.go.main.App.ValidateVentoyCli) {
      try {
        const res = await window.go.main.App.ValidateVentoyCli("");
        if (res) {
          ventoyStatus.value = res;
        }
      } catch (e) {
        console.error("Failed to validate Ventoy status:", e);
      }
    }
  }

  async function selectMode(mode: "cloud" | "hybrid") {
    activeMode.value = mode;
    if (mode === "hybrid") {
      await checkVentoyStatus();
      if (!isNonDestructive.value && !ventoyStatus.value.valid) {
        if (isMacOs.value) {
          showToast(t("deploy.tip_macos_unsupported"), "warning");
        } else {
          showToast(t("deploy.tip_need_ventoy"), "warning");
        }
      }
    }
  }

  function openVentoyAlert(title: string, message: string, action: "open_settings" | "switch_b") {
    ventoyAlertTitle.value = title;
    ventoyAlertMessage.value = message;
    ventoyAlertAction.value = action;
    isVentoyAlertOpen.value = true;
  }

  function handleVentoyAlertAction() {
    isVentoyAlertOpen.value = false;
    if (openSettings) {
      openSettings("ventoy");
    }
  }

  function handleVentoyAlertSwitchB() {
    isVentoyAlertOpen.value = false;
    activeMode.value = "cloud";
    showToast(t("deploy.toast_switched_b"), "success");
  }

  function openDiagnosticsModal(diag: InstallDiagnosticsData | null, msg: string) {
    currentDiagnostics.value = diag;
    currentDiagErrorMsg.value = msg || t("deploy.alert_fail", { msg: "" });
    isDiagnosticsOpen.value = true;
  }

  function handleCopyReport() {
    logUserAction("INFO", "User copied diagnostics report to clipboard");
    showToast(t("diag.toast_copied"), "info");
  }

  function handleRetryDeploy() {
    logUserAction("INFO", "User clicked retry deployment from diagnostics modal");
    isDiagnosticsOpen.value = false;
    openDeployConfirm();
  }

  function openDeployConfirm() {
    dismissDeploySuccessBanner();

    // pendingTargets must be locked before this function is called.
    // We intentionally do NOT re-derive targets from the current UI selection:
    // - For the no-conflict deploy path, targets are locked by handleDeployBtnClick
    //   before the async preflight call, so they always match pendingIsoPlans.
    // - For the retry path, we reuse the original targets from the failed deploy,
    //   ensuring the ISO plans (built for those targets) remain consistent.
    // This function's only job is to take a fresh disk-state snapshot of the
    // already-decided targets and open the confirmation dialog.
    const targets = pendingTargets.value;
    if (targets.length === 0) {
      showToast(t("deploy.toast_target_changed"), "error");
      return;
    }

    const refreshedSnapshots = targets.map((device) => diskList.value.find((disk) => disk.device === device));
    if (refreshedSnapshots.some((disk) => !disk)) {
      showToast(t("deploy.toast_target_changed"), "error");
      return;
    }

    pendingTargetSnapshots.value = refreshedSnapshots as DiskInfo[];
    isDeployConfirmOpen.value = true;
  }

  function buildDefaultIsoPlans(targets: string[]) {
    return targets.map((targetDisk) => ({
      targetDisk,
      entries: selectedIsoFiles.value.map((file) => ({
        sourcePath: file.path,
        targetName: "",
        action: "rename" as const,
      })),
    }));
  }

  async function preflightIsoCopies(targets: string[]): Promise<boolean> {
    pendingIsoPlans.value = buildDefaultIsoPlans(targets);
    if (activeMode.value !== "hybrid" || selectedIsoFiles.value.length === 0) return true;

    const app = window.go?.main?.App;
    if (!app?.PreflightIsoCopy) return true;

    try {
      const conflicts = ((await app.PreflightIsoCopy(
        targets,
        selectedIsoFiles.value.map((file) => file.path),
      )) ?? []) as IsoCopyConflict[];
      if (conflicts.length > 0) {
        // Snapshot disk state before showing the conflict dialog so the entire
        // preflight → conflict → confirm → deploy chain uses a single consistent
        // disk state captured at this exact moment.
        const snapshots = targets.map((device) => diskList.value.find((disk) => disk.device === device));
        if (snapshots.some((d) => !d)) {
          // At least one target disk disappeared during preflight — abort.
          // Opening the conflict dialog with a stale/missing snapshot would
          // risk deploying to the wrong disk after the user resolves conflicts.
          showToast(t("deploy.toast_target_changed"), "error");
          return false;
        }
        // pendingTargets was already locked by handleDeployBtnClick before
        // the async preflight call, so we only need the snapshot here.
        pendingTargetSnapshots.value = snapshots as DiskInfo[];
        isoConflicts.value = conflicts;
        isIsoConflictOpen.value = true;
        return false;
      }
      return true;
    } catch (err: any) {
      showToast(err?.message || t("iso.preflight_failed"), "error");
      return false;
    }
  }

  function cancelIsoConflictPreflight() {
    isIsoConflictOpen.value = false;
    isoConflicts.value = [];
    pendingIsoPlans.value = [];
    // Clear the snapshot captured during preflight so stale data cannot
    // accidentally be reused if the next deploy attempt hits the silent
    // snapshot-failure branch before overwriting these refs.
    pendingTargets.value = [];
    pendingTargetSnapshots.value = [];
    pendingIsoFiles.value = [];
  }

  function confirmIsoConflictPreflight(
    decisions: Array<IsoCopyPlanEntry & { targetDisk: string }>,
  ) {
    const decisionByKey = new Map(
      decisions.map((decision) => [`${decision.targetDisk}\n${decision.sourcePath}`, decision]),
    );
    pendingIsoPlans.value = pendingIsoPlans.value.map((plan) => ({
      ...plan,
      entries: plan.entries.map((entry) => {
        const decision = decisionByKey.get(`${plan.targetDisk}\n${entry.sourcePath}`);
        return decision ? { ...entry, ...decision } : entry;
      }),
    }));
    isIsoConflictOpen.value = false;
    isoConflicts.value = [];
    // Do NOT call openDeployConfirm() here — it would re-fetch disk snapshots
    // from diskList, which may have changed since the user started deciding.
    // pendingTargets and pendingTargetSnapshots were already captured in
    // preflightIsoCopies when the conflicts were first discovered.
    if (pendingTargets.value.length === 0) {
      // Should not normally happen, but guard against the silent snapshot-failure
      // edge case where pendingTargets was never populated.
      showToast(t("deploy.toast_target_changed"), "error");
      return;
    }
    dismissDeploySuccessBanner();
    isDeployConfirmOpen.value = true;
  }

  async function handleDeployBtnClick() {
    // Guard against concurrent deploy flows. isPreflight covers the async Go
    // preflight call, but once preflightIsoCopies returns the flag is cleared
    // even if a dialog is still open. isIsoConflictOpen / isDeployConfirmOpen
    // guard the remaining window so a second click cannot corrupt pendingTargets
    // while the user is resolving a conflict or reviewing the confirmation.
    if (isDeploying.value || isPreflight.value || isIsoConflictOpen.value || isDeployConfirmOpen.value) return;

    dismissDeploySuccessBanner();

    if (selectionMode.value === "single" && !selectedDisk.value) {
      showToast(t("deploy.toast_select_target"), "warning");
      return;
    }
    if (selectionMode.value === "batch" && selectedDevices.value.size === 0) {
      showToast(t("deploy.toast_select_batch"), "warning");
      return;
    }

    if (activeMode.value === "hybrid" && !isNonDestructive.value) {
      if (!ventoyStatus.value.valid) {
        if (isMacOs.value) {
          openVentoyAlert(t("deploy.macos_alert_title"), t("deploy.macos_alert_desc"), "switch_b");
        } else {
          openVentoyAlert(
            t("deploy.no_ventoy_title"),
            getVentoyValidationMessage(ventoyStatus.value) || t("deploy.no_ventoy_desc"),
            "open_settings"
          );
        }
        return;
      }
      checkVentoyStatus().catch(() => {});
    }

    let targets: string[] = [];
    if (selectionMode.value === "single") {
      if (!selectedDisk.value) return;
      // Verify the selected disk still exists in the current disk list
      const currentDisk = diskList.value.find((d) => d.device === selectedDisk.value!.device);
      if (!currentDisk) {
        showToast(t("deploy.toast_target_changed"), "error");
        return;
      }
      targets = [currentDisk.device];
    } else {
      targets = Array.from(selectedDevices.value);
      if (targets.length === 0) return;
    }

    // Lock pendingTargets and pendingIsoFiles NOW, before the async preflight call.
    // buildDefaultIsoPlans (inside preflightIsoCopies) stamps each plan with
    // plan.targetDisk from this exact targets array. If we waited until
    // openDeployConfirm to set pendingTargets, a selection change during the
    // async Go call would cause pendingIsoPlans[i].targetDisk (T1) to diverge
    // from pendingTargets[i] (T2), routing ISO plans to the wrong disks.
    // Similarly, selectedIsoFiles is locked so that the isoPaths sent to the
    // backend in startDeployment always match the entries in pendingIsoPlans.
    pendingTargets.value = targets;
    pendingTargetSnapshots.value = [];
    pendingIsoFiles.value = [...selectedIsoFiles.value];

    isPreflight.value = true;
    try {
      if (!(await preflightIsoCopies(targets))) return;
      openDeployConfirm();
    } finally {
      isPreflight.value = false;
    }
  }

  async function startDeployment() {
    if (isDeploying.value) return;
    // Close the confirm dialog immediately so that the isDeployConfirmOpen guard
    // in handleDeployBtnClick does not permanently block subsequent deploys.
    // (The dialog emits @confirm which calls this function, but does not emit
    // @close, so isDeployConfirmOpen stays true unless we clear it here.)
    isDeployConfirmOpen.value = false;
    dismissDeploySuccessBanner();

    const hasNativeRuntime = Boolean(window.go && window.go.main && window.go.main.App);
    if (!hasNativeRuntime) {
      deployProgress.value = 0;
      batchDeployInfo.value = null;
      showToast("Desktop runtime is unavailable in browser preview mode", "warning");
      return;
    }

    // Use pendingTargets as the single authoritative source of truth.
    // These were locked in either by openDeployConfirm() or by preflightIsoCopies()
    // (when conflicts were found) at the moment the user initiated the deploy flow.
    // Re-deriving targets from the current selection here would allow a changed
    // selection to silently mismatch the pendingIsoPlans that were built for the
    // original targets, potentially writing ISO plans to the wrong disks.
    const targets = pendingTargets.value;
    if (targets.length === 0) {
      showToast(t("deploy.toast_target_changed"), "error");
      return;
    }

    isDeploying.value = true;
    deployProgress.value = 0;
    batchDeployInfo.value = {
      totalDisks: targets.length,
      currentDiskIndex: 1,
      currentDisk: targets[0],
      currentStage: t("deploy.stage_preparing"),
      diskProgress: 0,
      overallProgress: 0,
      speedMBps: 0,
      elapsedSec: 0,
      etaSec: 0,
    };

    let unsubCloudProgress: (() => void) | null = null;
    let unsubBatchProgress: (() => void) | null = null;

    if (window.runtime && window.runtime.EventsOn) {
      // Listen for unified batch deployment progress (unified across modes)
      const onBatchProgress = (progress: BatchDeployProgress) => {
        deployProgress.value = progress.overallProgress;
        batchDeployInfo.value = progress;
        // Update speed and time information from batch progress
        deploySpeedMBps.value = progress.speedMBps;
        deployElapsedSec.value = progress.elapsedSec;
        deployEtaSec.value = progress.etaSec;
      };

      window.runtime.EventsOn("deploy-batch-progress", onBatchProgress);
      window.runtime.EventsOn("cloud-deploy-batch-progress", onBatchProgress);

      unsubBatchProgress = () => {
        if (window.runtime && window.runtime.EventsOff) {
          window.runtime.EventsOff("deploy-batch-progress");
          window.runtime.EventsOff("cloud-deploy-batch-progress");
        }
      };

      // Listen for single disk progress fallback
      window.runtime.EventsOn("cloud-deploy-progress", (progress: number) => {
        deployProgress.value = progress;
      });
      unsubCloudProgress = () => {
        if (window.runtime && window.runtime.EventsOff) {
          window.runtime.EventsOff("cloud-deploy-progress");
        }
      };
    }

    let success = true;
    let resultMsg = "";
    let latestDiagnostics: any = null;

    try {
      const app = window.go && window.go.main && window.go.main.App;
      if (!app) {
        throw new Error("Desktop runtime is unavailable in browser preview mode");
      }

      const isoPaths = pendingIsoFiles.value.map((f) => f.path);
      if (targets.length === 1) {
        const expected = pendingTargetSnapshots.value[0];
        if (!expected) throw new Error(t("deploy.toast_target_changed"));
        let resList: any[];
        if (activeMode.value === "cloud") {
          resList = await app.DeployCloudModeBatch(targets, selectedFsType.value, [expected]);
        } else if (app.DeployHybridModeBatchWithPlans) {
          resList = await app.DeployHybridModeBatchWithPlans(
            targets,
            selectedFsType.value,
            isoPaths,
            pendingIsoPlans.value,
            [expected],
          );
        } else {
          resList = await app.DeployHybridModeBatch(targets, selectedFsType.value, isoPaths, [expected]);
        }
        if (resList && resList.length > 0) {
          const res = resList[0];
          success = res.success;
          resultMsg = getDeployResultMessage(res);
          if (res.diagnostics) {
            latestDiagnostics = res.diagnostics;
          }
        } else {
          // Backend returned null or an empty result list — treat as failure
          // so the user sees a diagnostics modal instead of a false success banner.
          success = false;
          resultMsg = t("deploy.alert_fail", { msg: "no result returned" });
        }
      } else {
        if (pendingTargetSnapshots.value.length !== targets.length) {
          throw new Error(t("deploy.toast_target_changed"));
        }
        let resList: any[];
        if (activeMode.value === "cloud") {
          resList = await app.DeployCloudModeBatch(targets, selectedFsType.value, pendingTargetSnapshots.value);
        } else if (app.DeployHybridModeBatchWithPlans) {
          resList = await app.DeployHybridModeBatchWithPlans(
            targets,
            selectedFsType.value,
            isoPaths,
            pendingIsoPlans.value,
            pendingTargetSnapshots.value,
          );
        } else {
          resList = await app.DeployHybridModeBatch(
            targets,
            selectedFsType.value,
            isoPaths,
            pendingTargetSnapshots.value
          );
        }
        if (resList && resList.length > 0) {
          const failed = resList.filter((r) => !r.success);
          if (failed.length > 0) {
            success = false;
            resultMsg = failed.map((f) => `${f.target}: ${getDeployResultMessage(f)}`).join("\n");
            if (failed[0].diagnostics) {
              latestDiagnostics = failed[0].diagnostics;
            }
          } else {
            resultMsg = t("deploy.result_batch_success", { count: resList.length });
          }
        } else {
          success = false;
          resultMsg = t("deploy.alert_fail", { msg: "no result returned" });
        }
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

        deploySuccessBanner.value = {
          visible: true,
          msg: resultMsg,
          targets: [...targets],
          autoEjected: autoEjectedCount > 0,
          mode: activeMode.value,
          dismissed: false,
        };

        if (autoEjectedCount > 0) {
          showToast(t("deploy.toast_auto_ejected", { count: autoEjectedCount }), "success");
        }

        try {
          await refreshDisks();
        } catch (refreshErr) {
          console.warn("Refresh disks after deploy failed:", refreshErr);
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
    logUserAction("WARN", "User clicked cancel deployment button");
    const app = (window as any)?.go?.main?.App;
    if (app && typeof app.CancelDeployment === "function") {
      try {
        const cancelled = await app.CancelDeployment();
        if (cancelled) {
          showToast(t("deploy.toast_cancelled"), "info");
          logUserAction("INFO", "Deployment task cancelled successfully");

          // Clean up deployment state
          isDeploying.value = false;
          deployProgress.value = 0;
          batchDeployInfo.value = null;
          deploySpeedMBps.value = 0;
          deployElapsedSec.value = 0;
          deployEtaSec.value = 0;
        }
      } catch (e: any) {
        console.error("Failed to cancel deployment:", e);
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
    isIsoConflictOpen,
    isoConflicts,
    pendingIsoPlans,
    isPreflight,
    pendingTargets,
    pendingTargetSnapshots,
    pendingIsoFiles,
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
    cancelIsoConflictPreflight,
    confirmIsoConflictPreflight,
    handleDeployBtnClick,
    startDeployment,
    handleCancelDeploy,
    dismissDeploySuccessBanner,
  };
}

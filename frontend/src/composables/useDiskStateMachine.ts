// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import { ref, computed, type Ref, type ComputedRef } from 'vue';
import type { DiskInfo } from '../components/DiskPanel.vue';

/**
 * Approved finite disk states:
 * - UNDEPLOYED: Blank / unconfigured disk (no bootloader)
 * - DEPLOYED: Ready bootable system disk (UniBoot / Ventoy installed)
 * - DEPLOYING: Writing / installing / updating in progress
 * - VERIFYING: Checksum calculating / verifying in progress
 * - TESTING: VM preview simulation running
 * - READONLY: Write-protected / read-only permissions
 * - EJECTED: Safely ejected
 * - ERROR: Partition corrupted / write failure
 */
export type DiskState =
  | 'UNDEPLOYED'
  | 'DEPLOYED'
  | 'DEPLOYING'
  | 'VERIFYING'
  | 'TESTING'
  | 'READONLY'
  | 'EJECTED'
  | 'ERROR';

import type { VentoyValidation } from '../utils/ventoyValidation';

export interface DiskStateMachineOptions {
  t: (key: string, params?: Record<string, any>) => string;
  activeMode: Ref<'cloud' | 'hybrid'>;
  selectionMode: Ref<'single' | 'batch'>;
  selectedDisk: Ref<DiskInfo | null>;
  selectedDevices: Ref<Set<string>>;
  diskList: Ref<DiskInfo[]>;
  ventoyStatus: Ref<VentoyValidation>;
  isMacOs: Ref<boolean>;
  isDeploying: Ref<boolean>;
  isPreflight: Ref<boolean>;
  isVerifying?: Ref<boolean>;
  hypervisorCount?: Ref<number>;
  isLaunchingQemu?: Ref<boolean>;
  checkDiskCanUpdateNonDestructively: (disk: DiskInfo, mode: 'cloud' | 'hybrid') => boolean;
  getVentoyValidationMessage: (status: VentoyValidation) => string;
}

export interface DiskStateMachine {
  // Evaluated state of current disk selection
  currentDiskState: ComputedRef<DiskState>;
  getDiskState: (disk: DiskInfo | null) => DiskState;

  // Unified controlled UI variables (1:1 with existing UI props)
  isNonDestructive: ComputedRef<boolean>;
  deployBtnText: ComputedRef<string>;
  isDeployDisabled: ComputedRef<boolean>;
  deployDisabledReason: ComputedRef<string>;
  isVmDisabled: ComputedRef<boolean>;
  vmDisabledReason: ComputedRef<string>;
  isVmRunning: ComputedRef<boolean>;
  activeVmTargetDevice: ComputedRef<string>;
  activeVmTargetName: ComputedRef<string>;

  // State mutation actions
  setRunningVmTarget: (device: string, name: string) => void;
  clearRunningVmTarget: () => void;
  setVerifyingDevice: (device: string, isVerifying: boolean) => void;
  setDeviceError: (device: string, error: string) => void;
  clearDeviceError: (device: string) => void;
  setHypervisorCount: (count: number) => void;
  setIsLaunchingQemu: (launching: boolean) => void;
}

export function useDiskStateMachine(options: DiskStateMachineOptions): DiskStateMachine {
  const {
    t,
    activeMode,
    selectionMode,
    selectedDisk,
    selectedDevices,
    diskList,
    ventoyStatus,
    isMacOs,
    isDeploying,
    checkDiskCanUpdateNonDestructively,
    getVentoyValidationMessage,
  } = options;

  const localHypervisorCount = ref(options.hypervisorCount ? options.hypervisorCount.value : 1);
  const localIsLaunchingQemu = ref(options.isLaunchingQemu ? options.isLaunchingQemu.value : false);

  function setHypervisorCount(count: number) {
    localHypervisorCount.value = count;
  }

  function setIsLaunchingQemu(launching: boolean) {
    localIsLaunchingQemu.value = launching;
  }

  // Track active operations per physical device
  const runningVmTarget = ref<{ device: string; name: string } | null>(null);
  const verifyingDevices = ref<Set<string>>(new Set());
  const errorDevices = ref<Map<string, string>>(new Map());

  function setRunningVmTarget(device: string, name: string) {
    runningVmTarget.value = { device, name };
  }

  function clearRunningVmTarget() {
    runningVmTarget.value = null;
  }

  function setVerifyingDevice(device: string, isVerifying: boolean) {
    if (isVerifying) {
      verifyingDevices.value.add(device);
    } else {
      verifyingDevices.value.delete(device);
    }
  }

  function setDeviceError(device: string, error: string) {
    errorDevices.value.set(device, error);
  }

  function clearDeviceError(device: string) {
    errorDevices.value.delete(device);
  }

  // Pure function evaluating the finite state of a single disk
  function getDiskState(disk: DiskInfo | null): DiskState {
    if (!disk) return 'UNDEPLOYED';

    if (runningVmTarget.value?.device === disk.device) {
      return 'TESTING';
    }

    if (isDeploying.value) {
      if (selectionMode.value === 'single' && selectedDisk.value?.device === disk.device) {
        return 'DEPLOYING';
      }
      if (selectionMode.value === 'batch' && selectedDevices.value.has(disk.device)) {
        return 'DEPLOYING';
      }
    }

    if (verifyingDevices.value.has(disk.device) || (options.isVerifying && options.isVerifying.value)) {
      return 'VERIFYING';
    }

    if ((disk as any).is_ejected) {
      return 'EJECTED';
    }

    if ((disk as any).is_readonly) {
      return 'READONLY';
    }

    if (errorDevices.value.has(disk.device) || (disk as any).is_error) {
      return 'ERROR';
    }

    if (checkDiskCanUpdateNonDestructively(disk, activeMode.value)) {
      return 'DEPLOYED';
    }

    return 'UNDEPLOYED';
  }

  // Active target device resolution (preserves device during VM run when unmounted)
  const activeVmTargetDevice = computed(() => {
    if (runningVmTarget.value && runningVmTarget.value.device) {
      return runningVmTarget.value.device;
    }
    if (selectionMode.value === 'single') {
      return selectedDisk.value?.device || '';
    }
    if (selectedDevices.value.size > 0) {
      return Array.from(selectedDevices.value)[0];
    }
    return '';
  });

  // Active target name resolution
  const activeVmTargetName = computed(() => {
    if (runningVmTarget.value && runningVmTarget.value.name) {
      return runningVmTarget.value.name;
    }
    if (selectionMode.value === 'single') {
      return selectedDisk.value?.name || selectedDisk.value?.device || '';
    }
    if (selectedDevices.value.size > 0) {
      const firstDev = Array.from(selectedDevices.value)[0];
      const found = diskList.value.find((d) => d.device === firstDev);
      return found?.name || firstDev;
    }
    return '';
  });

  // Consolidated state for current active selection context
  const currentDiskState = computed<DiskState>(() => {
    if (runningVmTarget.value) {
      return 'TESTING';
    }

    if (selectionMode.value === 'single') {
      return getDiskState(selectedDisk.value);
    }

    // Batch mode: evaluate collective state across selected disks
    if (selectedDevices.value.size === 0) {
      return 'UNDEPLOYED';
    }

    if (isDeploying.value) {
      return 'DEPLOYING';
    }

    let allDeployed = true;
    let anyVerifying = false;

    selectedDevices.value.forEach((dev) => {
      const d = diskList.value.find((disk) => disk.device === dev);
      const st = getDiskState(d || null);
      if (st !== 'DEPLOYED') {
        allDeployed = false;
      }
      if (st === 'VERIFYING') {
        anyVerifying = true;
      }
    });

    if (anyVerifying) return 'VERIFYING';
    if (allDeployed) return 'DEPLOYED';
    return 'UNDEPLOYED';
  });

  // 1. isNonDestructive
  const isNonDestructive = computed(() => {
    if (selectionMode.value === 'single') {
      return currentDiskState.value === 'DEPLOYED';
    }
    if (selectedDevices.value.size === 0) return false;
    return Array.from(selectedDevices.value).every((dev) => {
      const d = diskList.value.find((disk) => disk.device === dev);
      return d ? checkDiskCanUpdateNonDestructively(d, activeMode.value) : false;
    });
  });

  const ventoyCountInBatch = computed(() => {
    if (selectionMode.value !== 'batch' || selectedDevices.value.size === 0) return 0;
    let count = 0;
    selectedDevices.value.forEach((dev) => {
      const d = diskList.value.find((disk) => disk.device === dev);
      if (d && checkDiskCanUpdateNonDestructively(d, activeMode.value)) {
        count++;
      }
    });
    return count;
  });

  // 2. deployBtnText
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

  // 3. isDeployDisabled & deployDisabledReason
  const isDeployDisabled = computed(() => {
    if (isDeploying.value) return true;
    if (currentDiskState.value === 'TESTING') return true;
    if (currentDiskState.value === 'VERIFYING') return true;
    if (currentDiskState.value === 'READONLY') return true;
    if (currentDiskState.value === 'EJECTED') return true;

    if (selectionMode.value === 'single' && !selectedDisk.value) return true;
    if (selectionMode.value === 'batch' && selectedDevices.value.size === 0) return true;

    // Guard: Hybrid mode Ventoy requirement on fresh drive
    if (activeMode.value === 'hybrid' && !isNonDestructive.value && !ventoyStatus.value.valid) {
      return true;
    }

    return false;
  });

  const deployDisabledReason = computed(() => {
    if (isDeploying.value) return t('deploy.tip_writing');
    if (currentDiskState.value === 'TESTING') return t('vm.tip_running');
    if (currentDiskState.value === 'VERIFYING') return t('checksum.calculating');

    if (selectionMode.value === 'single' && !selectedDisk.value) return t('deploy.tip_select_single');
    if (selectionMode.value === 'batch' && selectedDevices.value.size === 0) return t('deploy.tip_select_batch');

    // Guard: Hybrid mode Ventoy requirement on fresh drive
    if (activeMode.value === 'hybrid' && !isNonDestructive.value && !ventoyStatus.value.valid) {
      if (isMacOs.value) {
        return t('deploy.tip_macos_unsupported');
      }
      return getVentoyValidationMessage(ventoyStatus.value) || t('deploy.tip_need_ventoy');
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

  // 4. isVmRunning
  const isVmRunning = computed(() => {
    return currentDiskState.value === 'TESTING';
  });

  // 5. isVmDisabled
  const isVmDisabled = computed(() => {
    if (localIsLaunchingQemu.value) return true;
    if (isDeploying.value) return true;
    if (localHypervisorCount.value === 0) return true;
    if (!activeVmTargetDevice.value) return true;

    // Blank disks (UNDEPLOYED) cannot be tested in VM
    if (currentDiskState.value === 'UNDEPLOYED') return true;
    if (currentDiskState.value === 'EJECTED') return true;
    if (currentDiskState.value === 'VERIFYING') return true;

    // When testing is active, allow clicking stop VM
    if (currentDiskState.value === 'TESTING') return false;

    return false;
  });

  // 6. vmDisabledReason
  const vmDisabledReason = computed(() => {
    if (localIsLaunchingQemu.value) {
      return t('vm.tip_launching');
    }
    if (isDeploying.value) {
      return t('vm.tip_deploying');
    }
    if (localHypervisorCount.value === 0) {
      return t('vm.tip_not_installed');
    }
    if (!activeVmTargetDevice.value) {
      return t('vm.tip_select_target');
    }
    if (currentDiskState.value === 'UNDEPLOYED') {
      return t('vm.tip_select_target');
    }
    if (currentDiskState.value === 'TESTING') {
      return t('vm.tip_running');
    }
    return t('vm.tip_ready');
  });

  return {
    currentDiskState,
    getDiskState,
    isNonDestructive,
    deployBtnText,
    isDeployDisabled,
    deployDisabledReason,
    isVmDisabled,
    vmDisabledReason,
    isVmRunning,
    activeVmTargetDevice,
    activeVmTargetName,
    setRunningVmTarget,
    clearRunningVmTarget,
    setVerifyingDevice,
    setDeviceError,
    clearDeviceError,
    setHypervisorCount,
    setIsLaunchingQemu,
  };
}

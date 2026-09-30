// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import { ref, computed, watch, type Ref, type ComputedRef } from 'vue';
import { t } from '../i18n';
import { logUserAction } from '../utils/logger';
import type { DiskStateMachine } from './useDiskStateMachine';

export interface VMStatus {
  type: string;
  name: string;
  installed: boolean;
  path?: string;
  version?: string;
  priority?: number;
  canBootRaw?: boolean;
}

export interface UseVirtualMachineOptions {
  activeVmTargetDevice: ComputedRef<string>;
  activeVmTargetName?: ComputedRef<string>;
  diskList: Ref<DiskInfo[]>;
  isDeploying?: Ref<boolean>;
  showToast: (msg: string, type: 'info' | 'warning' | 'error' | 'success') => void;
  refreshDisks?: () => Promise<void>;
  onVmSessionStarted?: (targetDevice: string, targetName: string) => void;
  onVmSessionEnded?: () => void;
  fsm?: DiskStateMachine;
}

export function useVirtualMachine(options: UseVirtualMachineOptions) {
  const { activeVmTargetDevice, diskList, showToast, fsm } = options;
  const isDeploying = fsm ? fsm.isDeploying : (options.isDeploying || ref<boolean>(false));

  const hypervisorList = ref<VMStatus[]>([]);
  const selectedBootMode = ref<string>('auto');
  const selectedVMType = ref<string>('qemu');
  const vmCpuCores = ref<number>(2);
  const vmMemoryMB = ref<number>(2048);
  const vmDisplayAccel = ref<boolean>(true);
  const isLaunchingQemu = fsm ? fsm.isLaunchingVm : ref<boolean>(false);
  const qemuStatus = ref<{ installed: boolean; path: string; version: string }>({
    installed: false,
    path: '',
    version: ''
  });

  function handleVmEnded() {
    isLaunchingQemu.value = false;
    if (fsm) {
      fsm.stopVm();
    }
    if (options.onVmSessionEnded) {
      options.onVmSessionEnded();
    }
    if (options.refreshDisks) {
      options.refreshDisks();
    }
  }

  if (typeof window !== 'undefined' && (window as any).runtime && (window as any).runtime.EventsOn) {
    (window as any).runtime.EventsOn('vm-session-ended', () => {
      handleVmEnded();
    });
  }

  const isVmRunning = computed(() => {
    return fsm ? fsm.isVmRunning.value : false;
  });

  const isVmDisabled = computed(() => {
    return fsm ? fsm.isVmDisabled.value : true;
  });

  const vmDisabledReason = computed(() => {
    return fsm ? fsm.vmDisabledReason.value : '';
  });

  const runningVmTargetDevice = computed(() => {
    return fsm ? fsm.activeVmTargetDevice.value : '';
  });

  const runningVmTargetName = computed(() => {
    return fsm ? fsm.activeVmTargetName.value : '';
  });

  watch(
    () => hypervisorList.value.length,
    (count) => {
      if (fsm) {
        fsm.setHypervisorCount(count);
      }
    },
    { immediate: true }
  );

  watch(
    () => isLaunchingQemu.value,
    (launching) => {
      if (fsm) {
        fsm.setIsLaunchingQemu(launching);
      }
    }
  );

  async function checkQemu() {
    if (window.go && window.go.main && window.go.main.App) {
      if (typeof window.go.main.App.DetectHypervisors === 'function') {
        try {
          const list = await window.go.main.App.DetectHypervisors();
          if (Array.isArray(list)) {
            const installed = list.filter((h: any) => h.installed).sort((a: any, b: any) => a.priority - b.priority);
            hypervisorList.value = installed;
            if (installed.length > 0) {
              selectedVMType.value = installed[0].type;
              qemuStatus.value = { installed: true, path: installed[0].path, version: installed[0].version };
            } else {
              qemuStatus.value = { installed: false, path: '', version: '' };
            }
            return;
          }
        } catch (e) {
          console.error('Failed to detect hypervisors:', e);
        }
      }
      const fallback = await window.go.main.App.CheckQEMU();
      if (fallback && fallback.installed) {
        hypervisorList.value = [{ type: 'qemu', name: 'QEMU', installed: true, path: fallback.path, version: fallback.version, priority: 1, canBootRaw: true }];
        selectedVMType.value = 'qemu';
        qemuStatus.value = fallback;
      } else {
        hypervisorList.value = [];
        qemuStatus.value = { installed: false, path: '', version: '' };
      }
    } else {
      // Browser demo mode: Mock installed hypervisors (QEMU, UTM, VirtualBox)
      hypervisorList.value = [
        { type: 'qemu', name: 'QEMU', installed: true, path: '/usr/local/bin/qemu-system-x86_64', version: 'QEMU 8.2', priority: 1, canBootRaw: true },
        { type: 'utm', name: 'UTM', installed: true, path: '/Applications/UTM.app', version: 'UTM 4.4', priority: 2, canBootRaw: true },
        { type: 'virtualbox', name: 'Oracle VM VirtualBox', installed: true, path: '/usr/local/bin/VBoxManage', version: 'VirtualBox 7.0', priority: 7, canBootRaw: true }
      ];
      selectedVMType.value = 'qemu';
      qemuStatus.value = { installed: true, path: '/usr/local/bin/qemu-system-x86_64', version: 'QEMU 8.2' };
    }
  }

  async function launchVM() {
    if (fsm ? !fsm.canLaunchVm.value : (isLaunchingQemu.value || isDeploying.value || isVmDisabled.value)) {
      return;
    }
    isLaunchingQemu.value = true;

    try {
      if (!diskList.value || diskList.value.length === 0) {
        showToast(t('deploy.toast_no_disks'), 'warning');
        return;
      }

      const targetDevice = activeVmTargetDevice.value;

      if (!targetDevice) {
        showToast(t('vm.toast_select_first'), 'warning');
        return;
      }

      if (hypervisorList.value.length === 0) {
        showToast(t('vm.toast_not_installed'), 'error');
        return;
      }

      const currentVM = hypervisorList.value.find(h => h.type === selectedVMType.value) || hypervisorList.value[0];
      const vmName = currentVM ? currentVM.name : 'QEMU';
      const vmConfig = {
        cpuCores: vmCpuCores.value,
        memoryMB: vmMemoryMB.value,
        bootMode: selectedBootMode.value,
        displayAccel: vmDisplayAccel.value,
        secureBoot: false,
      };

      if (window.go && window.go.main && window.go.main.App) {
        const app = window.go.main.App as any;
        const launchedDev = targetDevice;
        const launchedName = options.activeVmTargetName?.value || targetDevice;

        if (typeof app.LaunchVMWithConfig === 'function') {
          await app.LaunchVMWithConfig(targetDevice, selectedVMType.value, vmConfig);
          if (fsm) {
            fsm.startVm(launchedDev, launchedName);
          }
          if (options.onVmSessionStarted) {
            options.onVmSessionStarted(launchedDev, launchedName);
          }
          logUserAction('INFO', 'User launched hypervisor simulation test with VMConfig', `${vmName} (${selectedVMType.value}, ${selectedBootMode.value}, ${vmCpuCores.value} cores, ${vmMemoryMB.value}MB) on ${targetDevice}`);
          showToast(t('vm.startSuccess_vm', { name: vmName }), 'success');
        } else if (typeof app.LaunchVM === 'function') {
          await app.LaunchVM(targetDevice, selectedVMType.value, selectedBootMode.value);
          if (fsm) {
            fsm.startVm(launchedDev, launchedName);
          }
          if (options.onVmSessionStarted) {
            options.onVmSessionStarted(launchedDev, launchedName);
          }
          logUserAction('INFO', 'User launched hypervisor simulation test', `${vmName} (${selectedVMType.value}, ${selectedBootMode.value}) on ${targetDevice}`);
          showToast(t('vm.startSuccess_vm', { name: vmName }), 'success');
        } else if (typeof app.LaunchQEMU === 'function') {
          await app.LaunchQEMU(targetDevice);
          if (fsm) {
            fsm.startVm(launchedDev, launchedName);
          }
          if (options.onVmSessionStarted) {
            options.onVmSessionStarted(launchedDev, launchedName);
          }
          logUserAction('INFO', 'User launched hypervisor simulation test', `QEMU on ${targetDevice}`);
          showToast(t('vm.startSuccess', { name: 'QEMU' }), 'success');
        } else {
          showToast(t('vm.backendNotReady'), 'warning');
          return;
        }

        if (options.refreshDisks) {
          await options.refreshDisks();
        }
      } else {
        await new Promise(r => setTimeout(r, 600));
        const demoName = options.activeVmTargetName?.value || targetDevice;
        if (fsm) {
          fsm.startVm(targetDevice, demoName);
        }
        if (options.onVmSessionStarted) {
          options.onVmSessionStarted(targetDevice, demoName);
        }
        logUserAction('INFO', 'User launched hypervisor simulation test (demo mode)', `${vmName} on ${targetDevice}`);
        if (options.refreshDisks) {
          await options.refreshDisks();
        }
      }
    } catch (e: any) {
      console.error('[UniBoot] LaunchVM error:', e);
      showToast(t('vm.startFailed', { error: e?.message || String(e) }), 'error');
    } finally {
      isLaunchingQemu.value = false;
    }
  }

  async function stopVM() {
    try {
      if (window.go && window.go.main && window.go.main.App && typeof (window.go.main.App as any).StopVM === 'function') {
        await (window.go.main.App as any).StopVM();
      }
      handleVmEnded();
    } catch (e: any) {
      console.warn('Failed to stop VM:', e);
    }
  }

  return {
    hypervisorList,
    selectedBootMode,
    selectedVMType,
    vmCpuCores,
    vmMemoryMB,
    vmDisplayAccel,
    isLaunchingQemu,
    isVmRunning,
    runningVmTargetDevice,
    runningVmTargetName,
    qemuStatus,
    isVmDisabled,
    vmDisabledReason,
    checkQemu,
    launchVM,
    stopVM,
  };
}

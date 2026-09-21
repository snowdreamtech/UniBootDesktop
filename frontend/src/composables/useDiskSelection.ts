import { ref, watch } from 'vue';
import type { DiskInfo } from '../components/DiskPanel.vue';
import type { DiskIconType } from '../components/IconPickerModal.vue';
import { logUserAction } from '../utils/logger';

const CUSTOM_ICONS_KEY = 'unigo_custom_icons_v1';

export function loadCustomIcons(): Record<string, DiskIconType> {
  try {
    const raw = localStorage.getItem(CUSTOM_ICONS_KEY);
    if (raw) return JSON.parse(raw);
  } catch (e) {
    console.error('Failed to load custom icons from localStorage:', e);
  }
  return {};
}

export function saveCustomIcons(icons: Record<string, DiskIconType>) {
  try {
    localStorage.setItem(CUSTOM_ICONS_KEY, JSON.stringify(icons));
  } catch (e) {
    console.error('Failed to save custom icons to localStorage:', e);
  }
}

export function getDiskFingerprint(disk: DiskInfo): string {
  if (disk.serialNumber && disk.serialNumber.trim() !== '') {
    return `sn:${disk.serialNumber.trim()}`;
  }
  return `dev:${disk.name}_${disk.size}`;
}

export interface UseDiskSelectionOptions {
  t: (key: any, named?: Record<string, any>) => string;
  showToast: (msg: string, type?: 'info' | 'success' | 'warning' | 'error') => void;
  getTargetsToEject?: () => string[];
  onEjectSuccess?: (device: string) => void;
  onSafelyEjectSuccess?: () => void;
}

export function useDiskSelection(options: UseDiskSelectionOptions) {
  const { t, showToast, getTargetsToEject, onEjectSuccess, onSafelyEjectSuccess } = options;

  const selectionMode = ref<'single' | 'batch'>('single');
  const diskList = ref<DiskInfo[]>([]);
  const selectedDisk = ref<DiskInfo | null>(null);
  const selectedDevices = ref<Set<string>>(new Set());
  const customIcons = ref<Record<string, DiskIconType>>(loadCustomIcons());
  const isPickerOpen = ref(false);
  const targetPickerDisk = ref<DiskInfo | null>(null);
  const isInspectorOpen = ref(false);
  const targetInspectorDisk = ref<DiskInfo | null>(null);
  const isScanningDisks = ref(false);

  watch(selectionMode, (newMode) => {
    logUserAction('INFO', 'User switched disk selection mode', newMode);
  });

  watch(selectedDisk, (disk, oldDisk) => {
    if (disk) {
      logUserAction('INFO', 'User selected target disk drive', `${disk.name || disk.device} (${disk.formatted})`);
    } else if (oldDisk) {
      logUserAction('INFO', 'User deselected target disk drive', `${oldDisk.name || oldDisk.device}`);
    }
  });

  function setSelectionMode(mode: 'single' | 'batch') {
    selectionMode.value = mode;
  }

  function selectAllDisks() {
    selectedDevices.value = new Set(diskList.value.map(d => d.device));
  }

  function deselectAllDisks() {
    selectedDevices.value.clear();
  }

  function onDiskSelect(disk: DiskInfo) {
    if (selectionMode.value === 'single') {
      if (selectedDisk.value && selectedDisk.value.device === disk.device) {
        selectedDisk.value = null;
      } else {
        selectedDisk.value = disk;
      }
    } else {
      onDiskToggle(disk);
    }
  }

  function onDiskToggle(disk: DiskInfo) {
    const newSet = new Set(selectedDevices.value);
    const wasSelected = newSet.has(disk.device);
    if (wasSelected) {
      newSet.delete(disk.device);
      logUserAction('INFO', 'User unchecked disk drive in batch selection mode', `${disk.name || disk.device}`);
    } else {
      newSet.add(disk.device);
      logUserAction('INFO', 'User checked disk drive in batch selection mode', `${disk.name || disk.device} (${disk.formatted})`);
    }
    selectedDevices.value = newSet;
  }

  function openInspector(disk: DiskInfo) {
    targetInspectorDisk.value = disk;
    isInspectorOpen.value = true;
    logUserAction('INFO', 'User opened USB hardware inspector modal', `${disk.name || disk.device} (${disk.formatted})`);
  }

  function openIconPicker(disk: DiskInfo) {
    targetPickerDisk.value = disk;
    isPickerOpen.value = true;
    logUserAction('INFO', 'User opened custom icon picker modal', disk.name || disk.device);
  }

  function onIconSelected(type: DiskIconType) {
    if (targetPickerDisk.value) {
      const fp = getDiskFingerprint(targetPickerDisk.value);
      customIcons.value[fp] = type;
      saveCustomIcons(customIcons.value);
      logUserAction('INFO', 'User updated custom disk icon', `${targetPickerDisk.value.name || targetPickerDisk.value.device} -> ${type}`);
    }
  }

  function onIconReset() {
    if (targetPickerDisk.value) {
      const fp = getDiskFingerprint(targetPickerDisk.value);
      delete customIcons.value[fp];
      saveCustomIcons(customIcons.value);
      logUserAction('INFO', 'User reset custom disk icon to default', targetPickerDisk.value.name || targetPickerDisk.value.device);
    }
  }

  async function handleEjectDisk(disk: DiskInfo) {
    try {
      if (window.go && window.go.main && window.go.main.App && window.go.main.App.EjectDisk) {
        await window.go.main.App.EjectDisk(disk.device);
      }
      if (onEjectSuccess) {
        onEjectSuccess(disk.device);
      }
      showToast(t('disk.toast_ejected_success', { device: disk.device, name: disk.name || disk.device }), 'success');
      await refreshDisks();
    } catch (err: any) {
      showToast(t('disk.toast_ejected_failed', { device: disk.device, error: err?.toString() || 'Unknown error' }), 'error');
    }
  }

  async function handleSafelyEjectAfterDeploy() {
    const targets = getTargetsToEject ? getTargetsToEject() : [];
    if (onSafelyEjectSuccess) {
      onSafelyEjectSuccess();
    }
    let ejectedCount = 0;
    for (const dev of targets) {
      try {
        if (window.go && window.go.main && window.go.main.App && window.go.main.App.EjectDisk) {
          await window.go.main.App.EjectDisk(dev);
          ejectedCount++;
        }
      } catch (err) {
        console.warn(`Eject failed for ${dev}:`, err);
      }
    }
    await refreshDisks();
    if (ejectedCount > 0) {
      showToast(t('deploy.toast_auto_ejected', { count: ejectedCount }), 'success');
    }
  }

  async function handleBatchEjectDisks() {
    const targets = Array.from(selectedDevices.value);
    if (targets.length === 0) return;

    let successCount = 0;
    let failCount = 0;

    for (const device of targets) {
      try {
        if (window.go && window.go.main && window.go.main.App && window.go.main.App.EjectDisk) {
          await window.go.main.App.EjectDisk(device);
        }
        successCount++;
        selectedDevices.value.delete(device);
      } catch (err) {
        failCount++;
      }
    }

    if (failCount === 0) {
      showToast(t('disk.toast_batch_eject_success', { count: successCount }), 'success');
    } else {
      showToast(t('disk.toast_batch_eject_partial', { successCount, failCount }), 'warning');
    }

    await refreshDisks();
  }

  // Wails JS binding fallbacks / mock data for standalone preview
  async function refreshDisks() {
    if (isScanningDisks.value) return;
    isScanningDisks.value = true;

    const previousSelectedDevice = selectedDisk.value?.device;
    const previousSelectedDevices = new Set(selectedDevices.value);

    try {
      if (window.go && window.go.main && window.go.main.App) {
        try {
          const fetched = ((await window.go.main.App.GetDiskList()) || []).filter((d: DiskInfo) => {
            if (!d || d.isSystem) return false;
            if (d.writable === false) return false;
            const nameLower = (d.name || '').toLowerCase();
            const devLower = (d.device || '').toLowerCase();
            if (
              nameLower.includes('disk image') ||
              nameLower.includes('.dmg') ||
              nameLower.includes('virtual') ||
              devLower.includes('loop')
            ) {
              return false;
            }
            return true;
          });
          diskList.value = fetched;

          // Preserve single selection if the disk is still connected
          if (previousSelectedDevice) {
            const stillExists = diskList.value.find(d => d.device === previousSelectedDevice);
            selectedDisk.value = stillExists || (diskList.value.length > 0 ? diskList.value[0] : null);
          } else if (diskList.value.length > 0) {
            selectedDisk.value = diskList.value[0];
          } else {
            selectedDisk.value = null;
          }

          // Preserve batch selections, filtering out only disks that were legitimately ejected/unplugged
          const newSelectedDevices = new Set<string>();
          for (const dev of previousSelectedDevices) {
            if (diskList.value.some(d => d.device === dev)) {
              newSelectedDevices.add(dev);
            }
          }
          selectedDevices.value = newSelectedDevices;
        } catch (e) {
          console.error('Error refreshing disk list:', e);
        }
      } else {
        // Fallback mock for browser preview demonstrating genuine vs fake USB 3.0
        await new Promise(resolve => setTimeout(resolve, 450));
        diskList.value = [
          {
            device: '/dev/disk2',
            name: 'SanDisk Ultra USB 3.0 Flash Drive',
            size: 32000000000,
            formatted: '32 GB',
            isRemovable: true,
            isSystem: false,
            usbVersion: 'USB 2.0',
            usbSpeed: '480 Mb/s',
            vendor: 'SanDisk (Suspected Fake)',
            isFakeUsb3: true,
            protocolCode: 'usb2',
            freeSpace: 0,
            freeFormatted: '0 B',
            fileSystem: 'exFAT',
            partitionScheme: 'GPT',
            writable: true,
            serialNumber: '',
            vendorId: '',
            productId: '',
            smartStatus: '',
            busPower: '',
            busPowerUsed: '',
            sectorSize: '',
            transportProtocol: '',
            bootStatus: '',
            controllerVendor: '',
            isRealVentoy: false,
            isCloudMode: false,
            isGenericBoot: false,
            mountPoint: ''
          },
          {
            device: '/dev/disk3',
            name: 'Kingston DataTraveler 3.0',
            size: 64000000000,
            formatted: '64 GB',
            isRemovable: true,
            isSystem: false,
            usbVersion: 'USB 3.0',
            usbSpeed: '5 Gb/s',
            vendor: 'Kingston Technology',
            isFakeUsb3: false,
            protocolCode: 'usb3_0',
            freeSpace: 0,
            freeFormatted: '0 B',
            fileSystem: 'exFAT',
            partitionScheme: 'GPT',
            writable: true,
            serialNumber: '',
            vendorId: '',
            productId: '',
            smartStatus: '',
            busPower: '',
            busPowerUsed: '',
            sectorSize: '',
            transportProtocol: '',
            bootStatus: '',
            controllerVendor: '',
            isRealVentoy: false,
            isCloudMode: false,
            isGenericBoot: false,
            mountPoint: ''
          },
          {
            device: '/dev/disk4',
            name: 'Samsung Type-C Duo 3.1',
            size: 128000000000,
            formatted: '128 GB',
            isRemovable: true,
            isSystem: false,
            usbVersion: 'USB 3.1 Gen 2',
            usbSpeed: '10 Gb/s',
            vendor: 'Samsung Electronics',
            isFakeUsb3: false,
            protocolCode: 'usb3_1',
            freeSpace: 0,
            freeFormatted: '0 B',
            fileSystem: 'exFAT',
            partitionScheme: 'GPT',
            writable: true,
            serialNumber: '',
            vendorId: '',
            productId: '',
            smartStatus: '',
            busPower: '',
            busPowerUsed: '',
            sectorSize: '',
            transportProtocol: '',
            bootStatus: '',
            controllerVendor: '',
            isRealVentoy: false,
            isCloudMode: false,
            isGenericBoot: false,
            mountPoint: ''
          }
        ];
        if (previousSelectedDevice) {
          const stillExists = diskList.value.find(d => d.device === previousSelectedDevice);
          selectedDisk.value = stillExists || (diskList.value.length > 0 ? diskList.value[0] : null);
        } else if (diskList.value.length > 0) {
          selectedDisk.value = diskList.value[0];
        }
      }
    } finally {
      isScanningDisks.value = false;
    }
  }

  return {
    selectionMode,
    diskList,
    selectedDisk,
    selectedDevices,
    isScanningDisks,
    customIcons,
    isPickerOpen,
    targetPickerDisk,
    isInspectorOpen,
    targetInspectorDisk,
    setSelectionMode,
    selectAllDisks,
    deselectAllDisks,
    onDiskSelect,
    onDiskToggle,
    openInspector,
    openIconPicker,
    onIconSelected,
    onIconReset,
    handleEjectDisk,
    handleBatchEjectDisks,
    handleSafelyEjectAfterDeploy,
    refreshDisks,
    getDiskFingerprint,
  };
}

import { ref, watch } from "vue";
import type { DiskInfo } from "../components/DiskPanel.vue";
import type { DiskIconType } from "../components/IconPickerModal.vue";
import { logUserAction } from "../utils/logger";

const CUSTOM_ICONS_KEY = "uniboot_custom_icons_v1";

export function loadCustomIcons(): Record<string, DiskIconType> {
  try {
    const raw = localStorage.getItem(CUSTOM_ICONS_KEY);
    if (raw) return JSON.parse(raw);
  } catch (e) {
    console.error("Failed to load custom icons from localStorage:", e);
  }
  return {};
}

export function saveCustomIcons(icons: Record<string, DiskIconType>) {
  try {
    localStorage.setItem(CUSTOM_ICONS_KEY, JSON.stringify(icons));
  } catch (e) {
    console.error("Failed to save custom icons to localStorage:", e);
  }
}

export function getDiskFingerprint(disk: DiskInfo): string {
  if (disk.serialNumber && disk.serialNumber.trim() !== "") {
    return `sn:${disk.serialNumber.trim()}`;
  }
  return `dev:${disk.name}_${disk.size}`;
}

export interface UseDiskSelectionOptions {
  t: (key: any, named?: Record<string, any>) => string;
  showToast: (msg: string, type?: "info" | "success" | "warning" | "error") => void;
  getTargetsToEject?: () => string[];
  onEjectSuccess?: (device: string) => void;
  onSafelyEjectSuccess?: () => void;
  isLocked?: () => boolean;
  lockReason?: () => string;
}

export function useDiskSelection(options: UseDiskSelectionOptions) {
  const { t, showToast, getTargetsToEject, onEjectSuccess, onSafelyEjectSuccess } = options;

  function checkLocked(): boolean {
    if (options.isLocked && options.isLocked()) {
      const reason = options.lockReason ? options.lockReason() : "";
      if (reason) {
        showToast(reason, "warning");
      }
      return true;
    }
    return false;
  }

  const selectionMode = ref<"single" | "batch">("single");
  const diskList = ref<DiskInfo[]>([]);
  const selectedDisk = ref<DiskInfo | null>(null);
  const selectedDevices = ref<Set<string>>(new Set());
  const ejectingDevices = ref<Set<string>>(new Set());
  const recentlyEjectedDevices = ref<Map<string, number>>(new Map());
  const customIcons = ref<Record<string, DiskIconType>>(loadCustomIcons());
  const isPickerOpen = ref(false);
  const targetPickerDisk = ref<DiskInfo | null>(null);
  const isInspectorOpen = ref(false);
  const targetInspectorDisk = ref<DiskInfo | null>(null);
  const isScanningDisks = ref(false);
  const pendingRestoreDevice = ref<string>("");

  watch(selectionMode, (newMode) => {
    logUserAction("INFO", "User switched disk selection mode", newMode);
  });

  watch(selectedDisk, (disk, oldDisk) => {
    if (disk) {
      logUserAction("INFO", "User selected target disk drive", `${disk.name || disk.device} (${disk.formatted})`);
    } else if (oldDisk) {
      logUserAction("INFO", "User deselected target disk drive", `${oldDisk.name || oldDisk.device}`);
    }
  });

  function setSelectionMode(mode: "single" | "batch") {
    if (checkLocked()) return;
    selectionMode.value = mode;
  }

  function selectAllDisks() {
    if (checkLocked()) return;
    selectedDevices.value = new Set(diskList.value.map((d) => d.device));
  }

  function deselectAllDisks() {
    if (checkLocked()) return;
    selectedDevices.value.clear();
  }

  function onDiskSelect(disk: DiskInfo) {
    if (checkLocked()) return;
    if (selectionMode.value === "single") {
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
    if (checkLocked()) return;
    const newSet = new Set(selectedDevices.value);
    const wasSelected = newSet.has(disk.device);
    if (wasSelected) {
      newSet.delete(disk.device);
      logUserAction("INFO", "User unchecked disk drive in batch selection mode", `${disk.name || disk.device}`);
    } else {
      newSet.add(disk.device);
      logUserAction(
        "INFO",
        "User checked disk drive in batch selection mode",
        `${disk.name || disk.device} (${disk.formatted})`
      );
    }
    selectedDevices.value = newSet;
  }

  function openInspector(disk: DiskInfo) {
    targetInspectorDisk.value = disk;
    isInspectorOpen.value = true;
    logUserAction(
      "INFO",
      "User opened USB hardware inspector modal",
      `${disk.name || disk.device} (${disk.formatted})`
    );
  }

  function openIconPicker(disk: DiskInfo) {
    if (checkLocked()) return;
    targetPickerDisk.value = disk;
    isPickerOpen.value = true;
    logUserAction("INFO", "User opened custom icon picker modal", disk.name || disk.device);
  }

  function onIconSelected(type: DiskIconType) {
    if (targetPickerDisk.value) {
      const fp = getDiskFingerprint(targetPickerDisk.value);
      customIcons.value[fp] = type;
      saveCustomIcons(customIcons.value);
      logUserAction(
        "INFO",
        "User updated custom disk icon",
        `${targetPickerDisk.value.name || targetPickerDisk.value.device} -> ${type}`
      );
    }
  }

  function onIconReset() {
    if (targetPickerDisk.value) {
      const fp = getDiskFingerprint(targetPickerDisk.value);
      delete customIcons.value[fp];
      saveCustomIcons(customIcons.value);
      logUserAction(
        "INFO",
        "User reset custom disk icon to default",
        targetPickerDisk.value.name || targetPickerDisk.value.device
      );
    }
  }

  function markEjecting(device: string) {
    const next = new Set(ejectingDevices.value);
    next.add(device);
    ejectingDevices.value = next;
  }

  function unmarkEjecting(device: string) {
    const next = new Set(ejectingDevices.value);
    next.delete(device);
    ejectingDevices.value = next;
  }

  function removeDiskFromList(device: string) {
    const nextList = diskList.value.filter((d) => d.device !== device);
    diskList.value = nextList;

    if (selectedDisk.value && selectedDisk.value.device === device) {
      selectedDisk.value = nextList[0] ?? null;
    }

    if (selectedDevices.value.has(device)) {
      const nextSelected = new Set(selectedDevices.value);
      nextSelected.delete(device);
      selectedDevices.value = nextSelected;
    }
  }

  async function handleEjectDisk(disk: DiskInfo) {
    if (checkLocked()) return;
    if (ejectingDevices.value.has(disk.device)) return;
    markEjecting(disk.device);

    try {
      if (window.go && window.go.main && window.go.main.App && window.go.main.App.EjectDisk) {
        await window.go.main.App.EjectDisk(disk.device);
      }
      recentlyEjectedDevices.value.set(disk.device, Date.now());
      removeDiskFromList(disk.device);
      if (onEjectSuccess) {
        onEjectSuccess(disk.device);
      }
      showToast(t("disk.toast_ejected_success", { device: disk.device, name: disk.name || disk.device }), "success");
      await refreshDisks();
    } catch (err: any) {
      showToast(
        t("disk.toast_ejected_failed", { device: disk.device, error: err?.toString() || "Unknown error" }),
        "error"
      );
    } finally {
      unmarkEjecting(disk.device);
    }
  }

  async function handleSafelyEjectAfterDeploy() {
    const targets = getTargetsToEject ? getTargetsToEject() : [];
    if (onSafelyEjectSuccess) {
      onSafelyEjectSuccess();
    }
    let ejectedCount = 0;
    for (const dev of targets) {
      markEjecting(dev);
    }
    try {
      if (window.go && window.go.main && window.go.main.App && window.go.main.App.BatchEjectDisks) {
        const res = await window.go.main.App.BatchEjectDisks(targets);
        ejectedCount = res.success ? res.success.length : 0;
        if (res.success) {
          const now = Date.now();
          for (const dev of res.success) {
            recentlyEjectedDevices.value.set(dev, now);
            removeDiskFromList(dev);
          }
        }
      } else {
        await Promise.all(
          targets.map(async (dev) => {
            try {
              if (window.go && window.go.main && window.go.main.App && window.go.main.App.EjectDisk) {
                await window.go.main.App.EjectDisk(dev);
                recentlyEjectedDevices.value.set(dev, Date.now());
                removeDiskFromList(dev);
                ejectedCount++;
              }
            } catch (err) {
              console.warn(`Eject failed for ${dev}:`, err);
            }
          })
        );
      }
    } finally {
      await refreshDisks();
      for (const dev of targets) {
        unmarkEjecting(dev);
      }
    }
    if (ejectedCount > 0) {
      showToast(t("deploy.toast_auto_ejected", { count: ejectedCount }), "success");
    }
  }

  async function handleBatchEjectDisks() {
    if (checkLocked()) return;
    const targets = Array.from(selectedDevices.value);
    if (targets.length === 0) return;

    for (const dev of targets) {
      markEjecting(dev);
    }

    let successCount = 0;
    let failCount = 0;

    try {
      if (window.go && window.go.main && window.go.main.App && window.go.main.App.BatchEjectDisks) {
        const res = await window.go.main.App.BatchEjectDisks(targets);
        successCount = res.success ? res.success.length : 0;
        failCount = res.failed ? Object.keys(res.failed).length : 0;
        if (res.success) {
          const now = Date.now();
          for (const dev of res.success) {
            recentlyEjectedDevices.value.set(dev, now);
            removeDiskFromList(dev);
          }
        }
      } else {
        await Promise.all(
          targets.map(async (device) => {
            try {
              if (window.go && window.go.main && window.go.main.App && window.go.main.App.EjectDisk) {
                await window.go.main.App.EjectDisk(device);
              }
              recentlyEjectedDevices.value.set(device, Date.now());
              successCount++;
              removeDiskFromList(device);
            } catch (err) {
              failCount++;
            }
          })
        );
      }

      if (failCount === 0) {
        showToast(t("disk.toast_batch_eject_success", { count: successCount }), "success");
      } else {
        showToast(t("disk.toast_batch_eject_partial", { successCount, failCount }), "warning");
      }
    } finally {
      await refreshDisks();
      for (const device of targets) {
        unmarkEjecting(device);
      }
    }
  }

  // Wails JS binding fallbacks / mock data for standalone preview
  async function refreshDisks() {
    if (isScanningDisks.value) return;
    if (options.isLocked && options.isLocked()) return;
    isScanningDisks.value = true;

    const previousSelectedDevice = selectedDisk.value?.device;
    const previousSelectedDevices = new Set(selectedDevices.value);

    try {
      if (window.go && window.go.main && window.go.main.App) {
        try {
          const now = Date.now();
          for (const [dev, ts] of recentlyEjectedDevices.value.entries()) {
            if (now - ts > 6000) {
              recentlyEjectedDevices.value.delete(dev);
            }
          }

          const fetched = ((await window.go.main.App.GetDiskList()) || []).filter((d: DiskInfo) => {
            if (!d || d.isSystem) return false;
            if (d.writable === false) return false;
            if (!d.mountPoint || d.mountPoint.trim() === "") return false;
            if (ejectingDevices.value.has(d.device)) return false;
            if (recentlyEjectedDevices.value.has(d.device)) return false;
            const nameLower = (d.name || "").toLowerCase();
            const devLower = (d.device || "").toLowerCase();
            if (
              nameLower.includes("disk image") ||
              nameLower.includes(".dmg") ||
              nameLower.includes("virtual") ||
              devLower.includes("loop")
            ) {
              return false;
            }
            return true;
          });

          // Smoothly update existing disks in-place to avoid layout flashes
          const currentMap = new Map(diskList.value.map((d) => [d.device, d]));
          const updatedList: DiskInfo[] = [];

          for (const newDisk of fetched) {
            const oldDisk = currentMap.get(newDisk.device);
            if (oldDisk) {
              // Mutate in-place so Vue reactivity smoothly transitions badges and fields without DOM teardown
              Object.assign(oldDisk, newDisk);
              updatedList.push(oldDisk);
            } else {
              updatedList.push(newDisk);
            }
          }
          diskList.value = updatedList;

          // Preserve single selection if the disk is still connected, or restore after VM remount
          const targetToPreserve = pendingRestoreDevice.value || previousSelectedDevice;
          if (targetToPreserve) {
            const stillExists = diskList.value.find((d) => d.device === targetToPreserve);
            if (stillExists) {
              selectedDisk.value = stillExists;
              pendingRestoreDevice.value = "";
            } else {
              selectedDisk.value = null;
            }
          } else if (diskList.value.length > 0 && !selectedDisk.value) {
            selectedDisk.value = diskList.value[0];
          } else {
            selectedDisk.value = null;
          }

          // Preserve batch selections, filtering out only disks that were legitimately ejected/unplugged
          const newSelectedDevices = new Set<string>();
          for (const dev of previousSelectedDevices) {
            if (diskList.value.some((d) => d.device === dev)) {
              newSelectedDevices.add(dev);
            }
          }
          selectedDevices.value = newSelectedDevices;
        } catch (e) {
          console.error("Error refreshing disk list:", e);
        }
      } else {
        // Fallback mock for browser preview demonstrating genuine vs fake USB 3.0
        await new Promise((resolve) => setTimeout(resolve, 450));
        const isWin = typeof navigator !== "undefined" && /Win/i.test(navigator.platform || navigator.userAgent);
        diskList.value = [
          {
            device: isWin ? "\\\\.\\PhysicalDrive1" : "/dev/disk2",
            name: "SanDisk Ultra USB 3.0 Flash Drive",
            size: 32000000000,
            formatted: "32 GB",
            isRemovable: true,
            isSystem: false,
            usbVersion: "USB 2.0",
            usbSpeed: "480 Mb/s",
            vendor: "SanDisk (Suspected Fake)",
            isFakeUsb3: true,
            protocolCode: "usb2",
            freeSpace: 0,
            freeFormatted: "0 B",
            fileSystem: "exFAT",
            partitionScheme: "GPT",
            writable: true,
            serialNumber: "",
            vendorId: "",
            productId: "",
            smartStatus: "",
            busPower: "",
            busPowerUsed: "",
            sectorSize: "",
            transportProtocol: "",
            bootStatus: "",
            bootStatusCode: "",
            controllerVendor: "",
            isRealVentoy: false,
            isCloudMode: false,
            isGenericBoot: false,
            thirdPartyBootType: "",
            thirdPartyBootCode: "",
            mountPoint: isWin ? "E:\\" : "/Volumes/SANDISK",
          },
          {
            device: isWin ? "\\\\.\\PhysicalDrive2" : "/dev/disk3",
            name: "Kingston DataTraveler 3.0",
            size: 64000000000,
            formatted: "64 GB",
            isRemovable: true,
            isSystem: false,
            usbVersion: "USB 3.0",
            usbSpeed: "5 Gb/s",
            vendor: "Kingston Technology",
            isFakeUsb3: false,
            protocolCode: "usb3_0",
            freeSpace: 0,
            freeFormatted: "0 B",
            fileSystem: "exFAT",
            partitionScheme: "GPT",
            writable: true,
            serialNumber: "",
            vendorId: "",
            productId: "",
            smartStatus: "",
            busPower: "",
            busPowerUsed: "",
            sectorSize: "",
            transportProtocol: "",
            bootStatus: "",
            bootStatusCode: "",
            controllerVendor: "",
            isRealVentoy: false,
            isCloudMode: false,
            isGenericBoot: false,
            thirdPartyBootType: "",
            thirdPartyBootCode: "",
            mountPoint: isWin ? "F:\\" : "/Volumes/KINGSTON",
          },
          {
            device: isWin ? "\\\\.\\PhysicalDrive3" : "/dev/disk4",
            name: "Samsung Type-C Duo 3.1",
            size: 128000000000,
            formatted: "128 GB",
            isRemovable: true,
            isSystem: false,
            usbVersion: "USB 3.1 Gen 2",
            usbSpeed: "10 Gb/s",
            vendor: "Samsung Electronics",
            isFakeUsb3: false,
            protocolCode: "usb3_1",
            freeSpace: 0,
            freeFormatted: "0 B",
            fileSystem: "exFAT",
            partitionScheme: "GPT",
            writable: true,
            serialNumber: "",
            vendorId: "",
            productId: "",
            smartStatus: "",
            busPower: "",
            busPowerUsed: "",
            sectorSize: "",
            transportProtocol: "",
            bootStatus: "",
            bootStatusCode: "",
            controllerVendor: "",
            isRealVentoy: false,
            isCloudMode: false,
            isGenericBoot: false,
            thirdPartyBootType: "",
            thirdPartyBootCode: "",
            mountPoint: isWin ? "G:\\" : "/Volumes/SAMSUNG",
          },
        ];
        if (previousSelectedDevice) {
          const stillExists = diskList.value.find((d) => d.device === previousSelectedDevice);
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
    pendingRestoreDevice,
    setPendingRestoreDevice: (dev: string) => {
      pendingRestoreDevice.value = dev;
    },
    ejectingDevices,
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

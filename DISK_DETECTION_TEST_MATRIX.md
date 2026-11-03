# Disk Detection Test Matrix

## Architectural Principles

1. **Zero Volume Label Reliance**: Disk labels (volume names) are strictly ignored during disk state determination. They can be renamed, duplicated, or empty.
2. **MBR Hardware Signature**: Ventoy detection checks Sector 0 MBR bytes directly (`CheckVentoyMbrSignature`).
3. **UniBoot Official Manifest**: UniBoot identity is determined via `ipxe/uniboot.json` containing `UNIBOOT_DISK` magic, with fallback to legacy `ipxe` script files for backwards compatibility.

---

## Test Scenarios

### Scenario 1: Hybrid Mode Disk (Ventoy + UniBoot Cloud Boot)

**Physical Layout:**

- MBR Sector 0: Ventoy bootloader signature
- Partition 1 (Data): `/ventoy/` engine files, `/ipxe/uniboot.json` (`mode: "hybrid"`), `/ventoy/themes/uniboot/`
- Partition 2 (ESP): Ventoy EFI bootloader files, iPXE chainloader

**Expected Detection:**

- `CheckVentoyMbrSignature()` → `true`
- `HasVentoyEngineFiles(p1)` → `true`
- `HasUniBootCloudFiles(p1)` → `true` (valid `ipxe/uniboot.json`)
- `IsCloudModeDisk()` → `false` (has Ventoy MBR and engine)
- `IsVentoyDisk()` → `true` (physical MBR signature & engine files)
- `IsRealVentoyDisk()` → `true`

**Frontend:**

- `disk.isCloudMode` = `false`
- `disk.isRealVentoy` = `true`
- Status: "Ventoy / UniBoot (混合模式)"
- Hybrid mode upgrade → "Smart Upgrade" ✅

---

### Scenario 2: Pure Ventoy Disk (Official Ventoy without UniBoot)

**Physical Layout:**

- MBR Sector 0: Ventoy bootloader signature
- Partition 1 (Data): `/ventoy/` directory with official engine files (NO `ipxe/uniboot.json`)
- Partition 2 (ESP): Ventoy EFI bootloader files (`ventoy.disk.img`, `BOOTX64.EFI`)

**Expected Detection:**

- `CheckVentoyMbrSignature()` → `true`
- `HasVentoyEngineFiles(p1)` → `true`
- `HasUniBootCloudFiles(p1)` → `false`
- `IsCloudModeDisk()` → `false` (has Ventoy MBR)
- `IsVentoyDisk()` → `true` (physical MBR signature & engine files)
- `IsRealVentoyDisk()` → `true`

**Frontend:**

- `disk.isCloudMode` = `false`
- `disk.isRealVentoy` = `true`
- Status: "Ventoy / UniBoot (混合模式)" (Can be upgraded non-destructively)
- Hybrid mode upgrade → "Smart Upgrade" (Preserves ISOs) ✅
- Cloud mode upgrade → "Smart Upgrade" (Converts EFI partition) ✅

---

### Scenario 3: Pure Cloud Boot Disk (UniBoot Cloud Mode)

**Physical Layout:**

- MBR Sector 0: Standard Protective MBR / UEFI GPT (NO Ventoy signature)
- Partition 1 (Data): Clean user storage (NO `/ventoy/` directory)
- Partition 2 (ESP): `/ipxe/uniboot.json` (`mode: "cloud"`), `ipxe/ipxe.efi`, `ipxe/uniboot.ipxe`, `EFI/BOOT/BOOTX64.EFI`

**Expected Detection:**

- `CheckVentoyMbrSignature()` → `false`
- `HasVentoyEngineFiles(p1)` → `false`
- `HasUniBootCloudFiles(p2)` → `true` (valid `ipxe/uniboot.json` in ESP)
- `IsCloudModeDisk()` → `true`
- `IsVentoyDisk()` → `false`
- `IsRealVentoyDisk()` → `false`

**Frontend:**

- `disk.isCloudMode` = `true`
- `disk.isRealVentoy` = `false`
- Status: "UniBoot (1秒极速云引导盘)"
- Cloud mode upgrade → "Smart Upgrade" (Instant re-flash of ESP) ✅
- Hybrid mode upgrade → "Full Format" (with warning prompt) ✅

---

### Scenario 4: Blank or Standard Storage Disk

**Physical Layout:**

- MBR Sector 0: Empty or standard OS MBR
- Partitions: Unformatted or standard exFAT/FAT32/NTFS without boot files

**Expected Detection:**

- `CheckVentoyMbrSignature()` → `false`
- `HasVentoyEngineFiles()` → `false`
- `HasUniBootCloudFiles()` → `false`
- `IsCloudModeDisk()` → `false`
- `IsVentoyDisk()` → `false`
- `IsRealVentoyDisk()` → `false`
- `IsGenericBootDisk()` → `false`

**Frontend:**

- `disk.isCloudMode` = `false`
- `disk.isRealVentoy` = `false`
- Status: "数据存储盘 (未检测到引导包)"
- Any mode → "Full Format" ✅

---

### Scenario 5: Third-party Boot Disks (Fine-Grained Classification)

**Physical Layout:**

- MBR Sector 0: Third-party bootloader (NO Ventoy signature)
- Partitions: Specific vendor/distro system files

**Fingerprint Specifications & Detection:**

| Target Type | Path Fingerprints | `thirdPartyBootType` | `bootStatus` |
| :--- | :--- | :--- | :--- |
| **Rufus Disk** | `rufus.efi`, `EFI/rufus/`, or `autounattend.xml` + `autorun.ico` | `"Rufus 制作盘"` | `"第三方引导: Rufus 制作盘"` |
| **微PE (WePE)** | `WEPE/` directory, `wepe.efi` | `"微PE (WePE) 维护盘"` | `"第三方引导: 微PE (WePE) 维护盘"` |
| **优启通 (EasyU)** | `EASYU/`, `USBDATA/`, `SKY/`, or `ITSKY/` | `"优启通 (EasyU) 维护盘"` | `"第三方引导: 优启通 (EasyU) 维护盘"` |
| **YUMI Multiboot** | `multiboot/menu/yumi.cfg`, or `multiboot/` | `"YUMI 多系统引导盘"` | `"第三方引导: YUMI 多系统引导盘"` |
| **OpenCore Hackintosh** | `EFI/OC/OpenCore.efi` or `EFI/OC/config.plist` | `"OpenCore 黑苹果引导盘"` | `"第三方引导: OpenCore 黑苹果引导盘"` |
| **Clover Hackintosh** | `EFI/CLOVER/CloverX64.efi` or `EFI/CLOVER/config.plist` | `"Clover 黑苹果引导盘"` | `"第三方引导: Clover 黑苹果引导盘"` |
| **Windows Official Installer** | `sources/install.wim`, `sources/install.esd`, or `sources/install.swm` | `"Windows 官方安装介质"` | `"第三方引导: Windows 官方安装介质"` |
| **Generic WinPE Disk** | `PETOOLS/`, `winpe.ini`, `pe.cfg`, or standalone `sources/boot.wim` | `"通用 WinPE 维护盘"` | `"第三方引导: 通用 WinPE 维护盘"` |
| **Linux Live USB** | `casper/`, `LiveOS/`, `arch/boot/`, `isolinux/`, or `boot/grub/grub.cfg` | `"Linux Live 安装盘"` | `"第三方引导: Linux Live 安装盘"` |
| **Generic UEFI USB** | `EFI/BOOT/BOOTX64.EFI`, `BOOTAA64.EFI`, `bootmgr`, `boot/bcd` | `"通用 UEFI 引导盘"` | `"第三方引导: 通用 UEFI 引导盘"` |

**Frontend:**

- `disk.isCloudMode` = `false`
- `disk.isRealVentoy` = `false`
- `disk.isGenericBoot` = `true`
- `disk.thirdPartyBootType` = Specific classified type
- Status: `"第三方引导: <Type>"`
- User Protection: Prominent warning prompt before format/deployment to avoid accidental destruction of maintenance tools or installer media. ✅

---

## Verification Status

All scenarios adhere strictly to:

1. No reliance on volume labels (`VolumeName`, `LABEL`).
2. Exact device partition routing (`NormalizeDarwinDiskNode`).
3. Single source of truth via MBR signature and `ipxe/uniboot.json`.

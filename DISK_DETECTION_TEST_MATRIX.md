# Disk Detection Test Matrix

## Test Scenarios

### Scenario 1: Hybrid Mode Disk (Ventoy + Cloud Boot)
**Physical Layout:**
- MBR: Ventoy bootloader signature
- Partition 1: `/ventoy/` directory with engine files
- Partition 2 (ESP): `VTOYEFI` label, `boot.ipxe`, `BOOTX64.EFI`, `grubx64.efi`

**Expected Detection:**
- `CheckVentoyMbrSignature()` → `true`
- `IsCloudModeDisk()` → `false` (has Ventoy MBR)
- `IsVentoyDisk()` → `true` (has VTOYEFI label)
- `IsRealVentoyDisk()` → `true` (not cloud, is Ventoy)

**Frontend:**
- `disk.isCloudMode` = `false`
- `disk.isRealVentoy` = `true`
- Hybrid mode upgrade → "Smart Upgrade" ✅

---

### Scenario 2: Pure Ventoy Disk (No Cloud Boot)
**Physical Layout:**
- MBR: Ventoy bootloader signature
- Partition 1: `/ventoy/` directory with engine files
- Partition 2 (ESP): `VTOYEFI` label, `BOOTX64.EFI`, `grubx64.efi` (NO ipxe)

**Expected Detection:**
- `CheckVentoyMbrSignature()` → `true`
- `IsCloudModeDisk()` → `false` (has Ventoy MBR)
- `IsVentoyDisk()` → `true` (has VTOYEFI label)
- `IsRealVentoyDisk()` → `true`

**Frontend:**
- `disk.isCloudMode` = `false`
- `disk.isRealVentoy` = `true`
- Hybrid mode upgrade → "Smart Upgrade" ✅
- Cloud mode upgrade → "Smart Upgrade" ✅

---

### Scenario 3: Pure Cloud Boot Disk (No Ventoy)
**Physical Layout:**
- MBR: Standard UEFI (NO Ventoy signature)
- Partition 1: Empty or data
- Partition 2 (ESP): `boot.ipxe`, `ipxe/uniboot.ipxe` (NO Ventoy files)

**Expected Detection:**
- `CheckVentoyMbrSignature()` → `false`
- `IsVentoyDisk()` → `false` (no VTOYEFI label)
- `IsCloudModeDisk()` → `true` (has cloud files, no Ventoy)
- `IsRealVentoyDisk()` → `false` (is cloud mode)

**Frontend:**
- `disk.isCloudMode` = `true`
- `disk.isRealVentoy` = `false`
- Hybrid mode upgrade → "Full Format" with warning banner ✅
- Cloud mode upgrade → "Smart Upgrade" ✅

---

### Scenario 4: Blank Disk
**Physical Layout:**
- MBR: Empty or standard MBR
- Partitions: Empty or unformatted

**Expected Detection:**
- `CheckVentoyMbrSignature()` → `false`
- `IsVentoyDisk()` → `false`
- `IsCloudModeDisk()` → `false` (no cloud files)
- `IsRealVentoyDisk()` → `false`

**Frontend:**
- `disk.isCloudMode` = `false`
- `disk.isRealVentoy` = `false`
- Any mode → "Full Format" ✅

---

### Scenario 5: Third-party Boot Disk (Rufus, BalenaEtcher, etc.)
**Physical Layout:**
- MBR: Third-party bootloader
- ESP: Generic BOOTX64.EFI, no Ventoy/Cloud files

**Expected Detection:**
- `CheckVentoyMbrSignature()` → `false`
- `IsVentoyDisk()` → `false`
- `IsCloudModeDisk()` → `false`
- `IsRealVentoyDisk()` → `false`
- `IsGenericBootDisk()` → `true`

**Frontend:**
- `disk.isCloudMode` = `false`
- `disk.isRealVentoy` = `false`
- `disk.isGenericBoot` = `true`
- Any mode → "Full Format" ✅

---

## Verification Status

All scenarios pass logical verification ✅

## Notes

1. **Performance**: `IsCloudModeDisk` is called twice in macOS (once in IsRealVentoyDisk, once directly), but results are consistent due to MBR signature fast-path.

2. **Platform Differences**:
   - **macOS**: Full multi-partition detection with MBR check
   - **Linux**: Device path + lsblk for mount point discovery  
   - **Windows**: Limited to single drive letter, file-based detection only

3. **Known Limitation**: Unmounted Linux disks may not be accurately detected (requires mounted partitions for file checks). Not a problem in practice as UI only shows mounted disks.

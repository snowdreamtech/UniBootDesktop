// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package firmware

// FirmwareMapping defines the mapping between a UniBoot release asset name and its UEFI/BIOS standard target path.
type FirmwareMapping struct {
	ReleaseName string `json:"releaseName"` // Original Release asset filename (e.g. ipxe-x86_64.efi, undionly.kpxe)
	TargetPath  string `json:"targetPath"`  // Standard target path in UNIBOOTEFI / U-disk (e.g. EFI/BOOT/BOOTX64.EFI, undionly.kpxe)
	Description string `json:"description"` // Architecture / target description
	IsReserved  bool   `json:"isReserved"`  // Whether this is a reserved / optional firmware module (e.g. undionly.kpxe)
}

// StandardFirmwareMappings defines the full matrix of UniBoot firmware files to be deployed.
var StandardFirmwareMappings = []FirmwareMapping{
	// UEFI Architectures
	{ReleaseName: "ipxe-x86_64.efi", TargetPath: "EFI/BOOT/BOOTX64.EFI", Description: "UEFI x86_64 (Intel/AMD 64-bit)", IsReserved: false},
	{ReleaseName: "ipxe-arm64.efi", TargetPath: "EFI/BOOT/BOOTAA64.EFI", Description: "UEFI ARM64 (Apple Silicon Mac / ARM Server)", IsReserved: false},
	{ReleaseName: "ipxe-i386.efi", TargetPath: "EFI/BOOT/BOOTIA32.EFI", Description: "UEFI IA32 (32-bit x86 Tablets/Board)", IsReserved: false},
	{ReleaseName: "ipxe-loongarch64.efi", TargetPath: "EFI/BOOT/BOOTLOONGARCH64.EFI", Description: "UEFI LoongArch64 (Loongson 64-bit)", IsReserved: false},
	{ReleaseName: "ipxe-riscv64.efi", TargetPath: "EFI/BOOT/BOOTRISCV64.EFI", Description: "UEFI RISC-V 64-bit", IsReserved: false},
	{ReleaseName: "ipxe-riscv32.efi", TargetPath: "EFI/BOOT/BOOTRISCV32.EFI", Description: "UEFI RISC-V 32-bit", IsReserved: false},

	// Legacy BIOS Boot Images
	{ReleaseName: "ipxe.lkrn", TargetPath: "ipxe.lkrn", Description: "Legacy BIOS U-disk MBR Kernel Boot Image", IsReserved: false},
	{ReleaseName: "undionly.kpxe", TargetPath: "undionly.kpxe", Description: "Legacy BIOS UNDI PXE Network Boot Firmware", IsReserved: true},

	// Entry Scripts
	{ReleaseName: "boot.ipxe", TargetPath: "boot.ipxe", Description: "iPXE Global Entry Script", IsReserved: false},
	{ReleaseName: "uniboot.ipxe", TargetPath: "uniboot.ipxe", Description: "UniBoot Main Interactive Menu Script", IsReserved: false},
}

// GetFirmwareMappings returns a copy of all standard firmware mappings.
func GetFirmwareMappings() []FirmwareMapping {
	mappings := make([]FirmwareMapping, len(StandardFirmwareMappings))
	copy(mappings, StandardFirmwareMappings)
	return mappings
}

// GetMappingByReleaseName looks up a firmware mapping by its release asset name.
func GetMappingByReleaseName(name string) (FirmwareMapping, bool) {
	for _, m := range StandardFirmwareMappings {
		if m.ReleaseName == name {
			return m, true
		}
	}
	return FirmwareMapping{}, false
}

// TargetPathForReleaseAsset returns the target path for a given release asset filename,
// or empty string if not recognized.
func TargetPathForReleaseAsset(name string) string {
	if m, ok := GetMappingByReleaseName(name); ok {
		return m.TargetPath
	}
	return ""
}

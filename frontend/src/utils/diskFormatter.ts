// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

/**
 * Format a disk object into a user-friendly display name.
 * e.g., on Windows:
 * - with mount point "E:\" and label "UNIBOOT" -> "E: (UNIBOOT)"
 * - with mount point "E:\" and no label -> "E:"
 * - without mount point -> "UNIBOOT (\\.\PhysicalDrive1)" or "\\.\PhysicalDrive1"
 * On macOS/Linux:
 * - with name "UNIBOOT" and device "/dev/disk2" -> "UNIBOOT (/dev/disk2)"
 */
export function formatDiskFriendlyName(disk?: {
  mountPoint?: string;
  name?: string;
  device?: string;
}): string {
  if (!disk) return "";
  const mp = (disk.mountPoint || "").trim();
  const name = (disk.name || "").trim();
  const dev = (disk.device || "").trim();

  // 1. Check if Windows drive letter like "E:\" or "E:"
  const driveLetterMatch = mp.match(/^([A-Za-z]:)[\\/]?$/);
  if (driveLetterMatch) {
    const letter = driveLetterMatch[1].toUpperCase();
    if (name) {
      // Strip leading redundant drive letter like "E: " or "E: (UNIBOOT)" -> "UNIBOOT"
      const cleanName = name.replace(/^[A-Za-z]:\s*\(?/, "").replace(/\)$/, "").trim();
      if (cleanName) {
        return `${letter} (${cleanName})`;
      }
    }
    return letter;
  }

  // 2. Check if name is already defined and distinct from device
  if (name && name !== dev) {
    if (dev) {
      return `${name} (${dev})`;
    }
    return name;
  }

  return dev || name || "";
}

/**
 * Format a device path, preferring Windows drive letter if available.
 * e.g. "E:" instead of "\\.\PhysicalDrive1"
 */
export function formatDiskDriveLetter(disk?: {
  mountPoint?: string;
  device?: string;
}): string {
  if (!disk) return "";
  const mp = (disk.mountPoint || "").trim();
  const driveLetterMatch = mp.match(/^([A-Za-z]:)[\\/]?$/);
  if (driveLetterMatch) {
    return driveLetterMatch[1].toUpperCase();
  }
  return disk.device || "";
}

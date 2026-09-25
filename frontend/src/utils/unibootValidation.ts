// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import type { DiskInfo } from '../components/DiskPanel.vue';

/**
 * UniBoot engine status (cross-platform, native pure Go implementation).
 * Free of external path, CLI, or operating system incompatibilities.
 */
export interface UniBootStatus {
  /** Whether the native UniBoot engine is ready to deploy on the current host. */
  ready: boolean;
  /** Current built-in/embedded UniBoot firmware version (e.g. "1.0.0"). */
  version: string;
  /** Lightweight status code: 'ready' | 'updating' | 'offline'. */
  code: 'ready' | 'updating' | 'offline';
  /** Optional human-readable detail or diagnostics message. */
  message?: string;
}

/**
 * UniBoot deployment status for an individual disk target.
 */
export interface UniBootDiskStatus {
  /** Whether this disk contains a verified UniBoot bootloader environment. */
  installed: boolean;
  /** Version of UniBoot installed on this disk (e.g. "1.0.0"), or empty if not installed. */
  version: string;
  /** Deployment mode recorded in manifest ("cloud" | "hybrid" | ""). */
  mode: string;
  /** Whether the disk can be upgraded non-destructively to the latest built-in version. */
  upgradeable: boolean;
}

/**
 * Extracts UniBoot status from a given DiskInfo.
 */
export function getDiskUniBootStatus(disk: DiskInfo | null, engineVersion = '1.0.0'): UniBootDiskStatus {
  if (!disk) {
    return {
      installed: false,
      version: '',
      mode: '',
      upgradeable: false,
    };
  }

  const anyDisk = disk as any;
  const installed = Boolean(
    disk.unibootVersion ||
    disk.unibootMode ||
    anyDisk.isCloudMode ||
    anyDisk.isRealVentoy ||
    anyDisk.ventoy_version
  );

  const version = disk.unibootVersion || anyDisk.ventoy_version || '';
  const mode = disk.unibootMode || (anyDisk.isCloudMode ? 'cloud' : (anyDisk.isRealVentoy ? 'hybrid' : ''));
  const upgradeable = installed && version !== '' && version !== engineVersion;

  return {
    installed,
    version,
    mode,
    upgradeable,
  };
}

// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import { describe, it, expect } from "vitest";
import { formatDiskFriendlyName, formatDiskDriveLetter } from "./diskFormatter";

describe("diskFormatter", () => {
  describe("formatDiskFriendlyName", () => {
    it("formats Windows drive with label as 'E: (UNIBOOT)'", () => {
      const disk = {
        device: "\\\\.\\PhysicalDrive1",
        mountPoint: "E:\\",
        name: "UNIBOOT",
      };
      expect(formatDiskFriendlyName(disk)).toBe("E: (UNIBOOT)");
    });

    it("formats Windows drive without trailing backslash", () => {
      const disk = {
        device: "\\\\.\\PhysicalDrive1",
        mountPoint: "E:",
        name: "UNIBOOT",
      };
      expect(formatDiskFriendlyName(disk)).toBe("E: (UNIBOOT)");
    });

    it("formats Windows drive with lowercase drive letter properly uppercase", () => {
      const disk = {
        device: "\\\\.\\PhysicalDrive1",
        mountPoint: "e:\\",
        name: "Kingston",
      };
      expect(formatDiskFriendlyName(disk)).toBe("E: (Kingston)");
    });

    it("formats Windows drive with empty label as drive letter only", () => {
      const disk = {
        device: "\\\\.\\PhysicalDrive1",
        mountPoint: "E:\\",
        name: "",
      };
      expect(formatDiskFriendlyName(disk)).toBe("E:");
    });

    it("formats Windows drive where label already starts with drive letter without redundancy", () => {
      const disk = {
        device: "\\\\.\\PhysicalDrive1",
        mountPoint: "E:\\",
        name: "E: (UNIBOOT)",
      };
      expect(formatDiskFriendlyName(disk)).toBe("E: (UNIBOOT)");
    });

    it("formats unpartitioned/raw disk with name as 'Name (device)'", () => {
      const disk = {
        device: "\\\\.\\PhysicalDrive1",
        mountPoint: "",
        name: "Kingston DataTraveler",
      };
      expect(formatDiskFriendlyName(disk)).toBe("Kingston DataTraveler (\\\\.\\PhysicalDrive1)");
    });

    it("formats macOS / Linux disk with name as 'Name (device)'", () => {
      const disk = {
        device: "/dev/disk2",
        mountPoint: "/Volumes/UNIBOOT",
        name: "UNIBOOT",
      };
      expect(formatDiskFriendlyName(disk)).toBe("UNIBOOT (/dev/disk2)");
    });

    it("falls back to device when no name or mount point is present", () => {
      const disk = {
        device: "\\\\.\\PhysicalDrive1",
        mountPoint: "",
        name: "",
      };
      expect(formatDiskFriendlyName(disk)).toBe("\\\\.\\PhysicalDrive1");
    });

    it("returns empty string for undefined disk", () => {
      expect(formatDiskFriendlyName(undefined)).toBe("");
    });
  });

  describe("formatDiskDriveLetter", () => {
    it("returns drive letter for Windows disk", () => {
      const disk = {
        device: "\\\\.\\PhysicalDrive1",
        mountPoint: "E:\\",
      };
      expect(formatDiskDriveLetter(disk)).toBe("E:");
    });

    it("falls back to device for non-Windows disk", () => {
      const disk = {
        device: "/dev/disk2",
        mountPoint: "/Volumes/UNIBOOT",
      };
      expect(formatDiskDriveLetter(disk)).toBe("/dev/disk2");
    });
  });
});

# UniBoot Desktop (UniGoDesktop)

[![CI Pipeline](https://img.shields.io/github/actions/workflow/status/snowdreamtech/UniGoDesktop/ci.yml?branch=main&label=CI%20Pipeline)](https://github.com/snowdreamtech/UniGoDesktop/actions/workflows/ci.yml)
[![CD Pipeline](https://img.shields.io/github/actions/workflow/status/snowdreamtech/UniGoDesktop/cd.yml?branch=main&label=CD%20Pipeline)](https://github.com/snowdreamtech/UniGoDesktop/actions/workflows/cd.yml)
[![Multi-OS Verified](https://img.shields.io/badge/Verified-Linux%20%7C%20macOS%20%7C%20Windows-blue)](https://github.com/snowdreamtech/UniGoDesktop/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/license/MIT)
[![Release](https://img.shields.io/github/v/release/snowdreamtech/UniGoDesktop?logo=github&sort=semver)](https://github.com/snowdreamtech/UniGoDesktop/releases/latest)

[English](README.md) | [简体中文](README_zh-CN.md)

**UniBoot Desktop (UniGoDesktop)** is a fast, multi-architecture, dual-engine cross-platform bootable disk creator and diagnostic suite powered by **Go + Wails + Vue 3**. It natively supports **macOS (Apple Silicon M1~M4 native / Intel), Windows, and Linux**.

---

## 🌟 Key Features

- **Dual Boot Engines**:
  - **Hybrid Mode (Ventoy MultiBoot Pro)**: Powered by Ventoy core protocol. Non-destructive in-place upgrades preserving user space with unlimited ISO/WIM/VHD/IMG placement.
  - **Cloud Mode (1-Sec Cloud Disk)**: macOS-native friendly iPXE cloud network boot with all-architecture firmware support (x86_64, UEFI, Legacy MBR, ARM64, RISC-V 64).
- **Live Transfer Speed (MB/s), ETA & Cancellation Context**:
  - Dynamic live transfer speed counter (MB/s) and estimated time remaining (ETA). Supports in-flight deployment cancellation (`Context.WithCancel`) at any time.
- **ISO SHA256 Checksum Verification & Physical Disk Guard**:
  - Automatic pre/post ISO SHA256 integrity verification. Hardened system disk whitelist and target disk exclusive locking to prevent accidental erasure.
- **QEMU Simulator VM Test & Advanced Tuning**:
  - Embedded QEMU simulator test module in both GUI and CLI. Customize **CPU Cores (1~8), RAM (1GB~8GB), Hardware Acceleration (`⚡ HVF/KVM/WHPX`)**, and **SecureBoot simulation**.
- **Inverted Modern Log Viewer & Smart Batch Selector**:
  - Inverted log console displaying newest logs on top. Smart batch button state machine for Select All / Deselect All actions.
- **51 Native Locales (100% Ventoy Parity)**:
  - 100% translated across 423 UI keys with ZERO English fallbacks. Includes RTL (Right-to-Left) auto-layout flipping for Arabic, Hebrew, Persian, and Urdu.
- **Single & Batch Multi-Disk Parallel Deployment**:
  - Powerful CLI supporting single disk deployment, concurrent multi-disk batch deployment (`--disks`), and auto-all disk deployment (`--all-usb`).

---

## 📖 CLI Usage & Script Automation

UniBoot features a **Dual-Mode Engine** where 100% of GUI features are accessible via the Cobra CLI:

```bash
# 1. Inspect disks and hardware specs (supports --json)
unigodesktop df --usb

# 2. Deploy Hybrid Mode (Ventoy) to a single disk with ISO copy
unigodesktop deploy --disk /dev/disk2 --mode Hybrid --fs exfat -i ~/Downloads/Ubuntu.iso -y

# 3. High-concurrency batch parallel deployment for multiple disks
unigodesktop deploy --disks /dev/disk2,/dev/disk3,/dev/disk4 --mode Cloud -y

# 4. Automatically deploy to ALL detected removable disks
unigodesktop deploy --all-usb --mode Hybrid -y

# 5. Launch QEMU simulator to test target disk from CLI (custom RAM/CPU)
unigodesktop qemu --disk /dev/disk2 -m 4096

# 6. Configure GitHub cloud mirror speed acceleration
unigodesktop config set github_proxy "https://ghproxy.net/"
unigodesktop config get github_proxy

# 7. Manage 13 embedded firmware cache items
unigodesktop cache list
unigodesktop cache purge

# 8. Run system environment health diagnostic
unigodesktop doctor

# 9. Generate shell completion scripts
unigodesktop completion zsh > ~/.zsh/completion/_unigodesktop
```

---

## 🍏 macOS Security Notice (Gatekeeper Tip)

If opening the application triggers the macOS Gatekeeper message *"UniGoDesktop cannot be opened because the developer cannot be verified"*:

1. **Right-Click Open**: Right-click `UniGoDesktop.app` in Finder, choose **Open**, and click **Open** in the confirmation dialog.
2. **Terminal Quarantine Removal**:
   ```bash
   sudo xattr -cr /Applications/UniGoDesktop.app
   ```

---

## 🛠️ Build & Packaging

### 1. Build Desktop GUI App (Wails)

```bash
# Build local desktop app
wails build

# Build macOS Universal Binary (.app / .dmg)
wails build -platform darwin/universal -package
```

### 2. Multi-Platform Packaging Pipeline (GoReleaser)

```bash
# Trigger GoReleaser pipeline
goreleaser release --snapshot --clean
```

---

## 📄 License

This project is licensed under the [MIT License](LICENSE).

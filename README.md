# UniBoot Desktop (UniGoDesktop)

[![CI Pipeline](https://img.shields.io/github/actions/workflow/status/snowdreamtech/UniGoDesktop/ci.yml?branch=main&label=CI%20Pipeline)](https://github.com/snowdreamtech/UniGoDesktop/actions/workflows/ci.yml)
[![CD Pipeline](https://img.shields.io/github/actions/workflow/status/snowdreamtech/UniGoDesktop/cd.yml?branch=main&label=CD%20Pipeline)](https://github.com/snowdreamtech/UniGoDesktop/actions/workflows/cd.yml)
[![Multi-OS Verified](https://img.shields.io/badge/Verified-Linux%20%7C%20macOS%20%7C%20Windows-blue)](https://github.com/snowdreamtech/UniGoDesktop/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/license/MIT)
[![Release](https://img.shields.io/github/v/release/snowdreamtech/UniGoDesktop?logo=github&sort=semver)](https://github.com/snowdreamtech/UniGoDesktop/releases/latest)

[English](README.md) | [简体中文](README_zh-CN.md)

**UniBoot Desktop (UniGoDesktop)** is a fast, multi-architecture, dual-engine cross-platform bootable USB creator and diagnostic suite powered by **Go + Wails + Vue 3**. It natively supports **macOS (Apple Silicon M1~M4 native / Intel), Windows, and Linux**.

---

## 🌟 Key Features

- **Dual Boot Engines**:
  - **Mode A (Ventoy MultiBoot Hybrid Pro)**: Powered by Ventoy core protocol. Non-destructive in-place upgrades preserving user space with unlimited ISO/WIM/VHD/IMG placement.
  - **Mode B (1-Sec Cloud Disk)**: macOS-native friendly iPXE cloud network boot with all-architecture firmware support (x86_64, UEFI, Legacy MBR, ARM64, RISC-V 64).
- **QEMU Simulator VM Test**:
  - Embedded QEMU simulator test module in both GUI and CLI. Verify bootable USB drives instantly without rebooting your computer.
- **USB Hardware Inspector**:
  - Displays 480 Mb/s physical bus speed, SMART health, partition scheme (GPT/MBR), and filesystem type.
- **53 Native Locales (100% Ventoy Parity)**:
  - 100% translated across 256 UI keys with ZERO English fallbacks. Includes RTL (Right-to-Left) auto-layout flipping for Arabic, Hebrew, Persian, and Urdu.
- **Single & Batch Multi-USB Parallel Deployment**:
  - Powerful CLI supporting single USB deployment, concurrent multi-USB batch deployment (`--disks`), and auto-all USB deployment (`--all-usb`).
- **13 Embedded Firmware Matrix**:
  - Statically embedded Go `embed.FS` firmware matrix with automated GitHub cloud mirror synchronization.

---

## 📖 CLI Usage & Script Automation

UniBoot features a **Dual-Mode Engine** where 100% of GUI features are accessible via the Cobra CLI:

```bash
# 1. Inspect USB drives and hardware specs (supports --json)
unigodesktop df --usb

# 2. Deploy Mode A (Ventoy) to a single USB drive with ISO copy
unigodesktop deploy --disk /dev/disk2 --mode A --fs exfat -i ~/Downloads/Ubuntu.iso -y

# 3. High-concurrency batch parallel deployment for multiple USB drives
unigodesktop deploy --disks /dev/disk2,/dev/disk3,/dev/disk4 --mode B -y

# 4. Automatically deploy to ALL detected removable USB drives
unigodesktop deploy --all-usb --mode A -y

# 5. Launch QEMU simulator to test target USB drive from CLI
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

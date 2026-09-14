# UniBoot 桌面端 (UniGoDesktop)

[![CI 流水线](https://img.shields.io/github/actions/workflow/status/snowdreamtech/UniGoDesktop/ci.yml?branch=main&label=CI%20Pipeline)](https://github.com/snowdreamtech/UniGoDesktop/actions/workflows/ci.yml)
[![CD 自动化发布](https://img.shields.io/github/actions/workflow/status/snowdreamtech/UniGoDesktop/cd.yml?branch=main&label=CD%20Pipeline)](https://github.com/snowdreamtech/UniGoDesktop/actions/workflows/cd.yml)
[![跨平台验证](https://img.shields.io/badge/Verified-Linux%20%7C%20macOS%20%7C%20Windows-blue)](https://github.com/snowdreamtech/UniGoDesktop/actions/workflows/ci.yml)
[![许可证: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/license/MIT)
[![最新发布](https://img.shields.io/github/v/release/snowdreamtech/UniGoDesktop?logo=github&sort=semver)](https://github.com/snowdreamtech/UniGoDesktop/releases/latest)

[English](README.md) | [简体中文](README_zh-CN.md)

**UniBoot 桌面端 (UniGoDesktop)** 是一款基于 **Go + Wails + Vue 3** 架构打造的全架构、双引导引擎跨平台启动盘制作与运维测试桌面软件。完美适配 **macOS (Apple Silicon M1~M4 原生/Intel)、Windows 及 Linux** 桌面环境。

---

## 🌟 核心特性

- **双引导制作引擎 (Dual Boot Engines)**：
  - **Mode A (Ventoy MultiBoot 混合模式)**：基于 Ventoy 核心协议，零损伤保留 U 盘剩余空间，支持无限放入 ISO/WIM/VHD/IMG 镜像。
  - **Mode B (1-Sec Cloud 极速一秒云盘模式)**：macOS 友好原生方案，纯 iPXE 网络引导全架构固件（支持 x86_64, UEFI, Legacy MBR, ARM64, RISC-V 64）。
- **QEMU 可视化虚拟机测试 (VM Simulator)**：
  - 界面与命令行内置 QEMU 模拟验证模块，可在桌面端一键无需重启电脑直接测试启动盘装载效果。
- **USB 硬件深层检测 (Hardware Inspector)**：
  - 具备 480 Mb/s 物理总线速率检测、SMART 健康度、分区表架构 (GPT/MBR) 与文件系统类型展示。
- **53 种语言全量母语原生国际化 (53 Native Locales)**：
  - 100% 齐平 Ventoy 官方 53 种语言，每一个语言包的 256 个 UI 键值全量母语原生翻译（0 英文残留），支持阿拉伯语/希伯来语/波斯语/乌尔都语 **RTL (右至左) 自动布局翻转**。
- **单盘 & 多 U 盘高并发批量部署 (Batch Multi-USB Flashing)**：
  - CLI 命令行支持单盘部署、高并发多 U 盘批量并行刷盘 (`--disks`) 及全设备自动刷盘 (`--all-usb`)。
- **13 项嵌入式固件与 ISO 矩阵 (Embedded Firmware Matrix)**：
  - Go `embed.FS` 静态嵌入 13 项全局引导固件，支持 GitHub 镜像加速增量云端自动同步。

---

## 📖 CLI 命令行与脚本自动化指南

UniBoot 具备**桌面 GUI + Cobra CLI 100% 功能完全对齐**的双模运行能力：

### 常用 CLI 子命令

```bash
# 1. 打印 U 盘硬件规格与存储巡检 (支持 --json)
unigodesktop df --usb

# 2. 部署 Mode A (Ventoy) 到单个 U 盘并自动拷贝 ISO
unigodesktop deploy --disk /dev/disk2 --mode A --fs exfat -i ~/Downloads/Ubuntu.iso -y

# 3. 高并发批量并行部署多块 U 盘 (Batch Mode)
unigodesktop deploy --disks /dev/disk2,/dev/disk3,/dev/disk4 --mode B -y

# 4. 自动全量扫描并部署所有接入的 USB 盘
unigodesktop deploy --all-usb --mode A -y

# 5. 命令行启动 QEMU 虚拟机测试目标 U 盘
unigodesktop qemu --disk /dev/disk2 -m 4096

# 6. 检查与设置全局 GitHub 镜像加速代理
unigodesktop config set github_proxy "https://ghproxy.net/"
unigodesktop config get github_proxy

# 7. 管理 13 项嵌入式固件缓存
unigodesktop cache list
unigodesktop cache purge

# 8. 系统环境与运行健康诊断
unigodesktop doctor

# 9. 生成 Zsh / Bash 命令行自动补全
unigodesktop completion zsh > ~/.zsh/completion/_unigodesktop
```

---

## 🛠️ 打包与构建

### 1. 本地打包桌面应用 (Wails)

```bash
# 构建本地 GUI 桌面应用
wails build

# 构建 macOS 通用双架构 (.app / .dmg)
wails build -platform darwin/universal -package
```

### 2. 全平台全量自动构建 (GoReleaser)

```bash
# 触发 GoReleaser 自动化流水线
goreleaser release --snapshot --clean
```

---

## 📄 许可证 (License)

本项目遵循 [MIT License](LICENSE) 许可协议。

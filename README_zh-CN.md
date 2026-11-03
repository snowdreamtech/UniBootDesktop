# UniGoDesktop

[![CI 流水线](https://img.shields.io/github/actions/workflow/status/snowdreamtech/UniGoDesktop/wails.yml?branch=main&label=CI%20Pipeline)](https://github.com/snowdreamtech/UniGoDesktop/actions/workflows/wails.yml)
[![多平台验证](https://img.shields.io/badge/Verified-Linux%20%7C%20macOS%20%7C%20Windows-blue)](https://github.com/snowdreamtech/UniGoDesktop/actions/workflows/wails.yml)
[![开源协议: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/license/MIT)
[![最新发布](https://img.shields.io/github/v/release/snowdreamtech/UniGoDesktop?logo=github&sort=semver)](https://github.com/snowdreamtech/UniGoDesktop/releases/latest)

[English](README.md) | [简体中文](README_zh-CN.md)

**UniGoDesktop** 是一套现代化、高颜值、企业级通用的跨平台桌面应用模版，基于 **Go + Wails v2 + Vue 3 + TypeScript** 架构构建。原生支持 **macOS (Apple Silicon M系列及 Intel 原生)、Windows 以及 Linux**。

---

## 🌟 核心特性

- **现代跨平台桌面架构**：
  - 高性能 Go 后端与轻量级 WebKit/Chromium 前端无缝融合，内存占用远低于 Electron。
  - 单一可执行原生二进制分发，启动迅速。
- **毛玻璃极客美学与主题自适应**：
  - 现代化毛玻璃设计语言与流体微动画。
  - 完整支持**深色主题 (Dark Mode)**、**浅色主题 (Light Mode)** 及**跟随系统设置**。
- **53 种母语本地化与 RTL 布局**：
  - 支持 53 种自然语言本地化，自动识别系统首选语言，支持动态热切换。
  - 针对阿拉伯语、希伯来语、波斯语、乌尔都语提供自动双向文字排版 (RTL) 支持。
- **通用网络代理与 GitHub 镜像加速**：
  - 完整支持 Direct、HTTP、HTTPS、SOCKS4、SOCKS5 代理协议及认证。
  - 支持自定义 GitHub 加速镜像配置。
  - 内置实时网络延迟测速与联通性诊断。
- **CLI + GUI 双模运行能力**：
  - 完整 Cobra 命令行体系与交互式桌面 GUI 协同运行。
- **全平台自动化打包与 CI/CD**：
  - 配置 GitHub Actions 跨平台发布工作流，自动化构建生成 macOS (Universal DMG/ZIP)、Windows (NSIS 安装包/ZIP) 以及 Linux (AppImage/DEB/RPM)。

---

## 🚀 快速上手

### 环境要求

- **Go**: `1.24+` (推荐 `1.25+`)
- **Node.js**: `20+` 或 `22+`
- **Wails CLI**: `v2.10+` (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)

### 开发模式

```bash
# 启动热重载开发模式
wails dev

# 仅调试前端
npm --prefix frontend run dev
```

### 生产构建

```bash
# 构建桌面原生应用
wails build

# 仅构建前端产物
npm --prefix frontend run build
```

---

## 📖 CLI 命令行说明

UniGoDesktop 提供了基于 Cobra 的完整命令行接口：

```bash
# 1. 交互式 Hello 问候演示
unigodesktop hello

# 2. 系统诊断与运行环境检查
unigodesktop doctor

# 3. 检查数据目录空间占用
unigodesktop df

# 4. 检查版本更新
unigodesktop update --check

# 5. 启动桌面 GUI 界面
unigodesktop --desktop
```

---

## 📄 开源许可

本项目采用 [MIT 许可证](LICENSE)。

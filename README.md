# UniGoDesktop

[![CI Pipeline](https://img.shields.io/github/actions/workflow/status/snowdreamtech/UniGoDesktop/wails.yml?branch=main&label=CI%20Pipeline)](https://github.com/snowdreamtech/UniGoDesktop/actions/workflows/wails.yml)
[![Multi-OS Verified](https://img.shields.io/badge/Verified-Linux%20%7C%20macOS%20%7C%20Windows-blue)](https://github.com/snowdreamtech/UniGoDesktop/actions/workflows/wails.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/license/MIT)
[![Release](https://img.shields.io/github/v/release/snowdreamtech/UniGoDesktop?logo=github&sort=semver)](https://github.com/snowdreamtech/UniGoDesktop/releases/latest)

[English](README.md) | [简体中文](README_zh-CN.md)

**UniGoDesktop** is an enterprise-ready, cross-platform desktop application template powered by **Go + Wails v2 + Vue 3 + TypeScript**. It natively supports **macOS (Apple Silicon & Intel), Windows, and Linux**.

---

## 🌟 Key Features

- **Modern Desktop Architecture**:
  - High performance Go backend integrated with lightweight WebKit/Chromium frontend via Wails v2.
  - Zero-bloat native binary distribution with minimal memory footprint compared to Electron.
- **Glassmorphic UI & Theming**:
  - Modern, responsive glassmorphic design system.
  - Seamless support for **Dark Mode**, **Light Mode**, and **System Default** appearance.
- **53 Mother Tongue Locales & RTL**:
  - Comprehensive internationalization covering 53 native languages.
  - Automatic system language detection and dynamic runtime switching.
  - Bidirectional text layout with automatic RTL (Right-to-Left) support for Arabic, Hebrew, Persian, and Urdu.
- **Network Proxy & Mirror Acceleration**:
  - Full proxy routing support for `Direct`, `HTTP`, `HTTPS`, `SOCKS4`, and `SOCKS5` with credentials.
  - Configurable custom GitHub acceleration mirror.
  - Built-in real-time network latency diagnostic and connectivity speed test.
- **Dual Runtime Capabilities (CLI + GUI)**:
  - Full Cobra CLI command structure paired with interactive GUI.
- **Cross-Platform CI/CD Packaging**:
  - Automated GitHub Actions build workflows for macOS (Universal DMG/ZIP), Windows (NSIS EXE/ZIP), and Linux (AppImage/DEB/RPM).

---

## 🚀 Quick Start

### Prerequisites

- **Go**: `1.24+` (recommended `1.25+`)
- **Node.js**: `20+` or `22+`
- **Wails CLI**: `v2.10+` (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)

### Development Mode

```bash
# Run in development mode with hot-reloading
wails dev

# Run frontend alone
npm --prefix frontend run dev
```

### Production Build

```bash
# Build desktop native application
wails build

# Build production frontend bundle
npm --prefix frontend run build
```

---

## 📖 CLI Usage

UniGoDesktop provides powerful CLI commands via Cobra:

```bash
# 1. Show interactive greeting
unigodesktop hello

# 2. Check system diagnostics and environment specs
unigodesktop doctor

# 3. Check data directory space usage
unigodesktop df

# 4. Check for application releases and updates
unigodesktop update --check

# 5. Launch desktop GUI
unigodesktop --desktop
```

---

## 📄 License

This project is licensed under the [MIT License](LICENSE).

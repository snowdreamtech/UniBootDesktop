# UniGoDesktop

[![CI Pipeline](https://img.shields.io/github/actions/workflow/status/snowdreamtech/UniGoDesktop/ci.yml?branch=main&label=CI%20Pipeline)](https://github.com/snowdreamtech/UniGoDesktop/actions/workflows/ci.yml)
[![CD Pipeline](https://img.shields.io/github/actions/workflow/status/snowdreamtech/UniGoDesktop/cd.yml?branch=main&label=CD%20Pipeline)](https://github.com/snowdreamtech/UniGoDesktop/actions/workflows/cd.yml)
[![GitHub Pages](https://img.shields.io/github/actions/workflow/status/snowdreamtech/UniGoDesktop/pages.yml?branch=main&label=Docs&logo=github)](https://github.com/snowdreamtech/UniGoDesktop/actions/workflows/pages.yml)
[![CodeQL](https://img.shields.io/github/actions/workflow/status/snowdreamtech/UniGoDesktop/codeql.yml?branch=main&label=CodeQL&logo=github)](https://github.com/snowdreamtech/UniGoDesktop/actions/workflows/codeql.yml)
[![Multi-OS Verified](https://img.shields.io/badge/Verified-Linux%20%7C%20macOS%20%7C%20Windows-blue)](https://github.com/snowdreamtech/UniGoDesktop/actions/workflows/ci.yml)
[![Security Audit](https://img.shields.io/badge/Security-Zizmor%20%7C%20Trivy%20%7C%20Gitleaks-brightgreen)](https://github.com/snowdreamtech/UniGoDesktop/actions/workflows/ci.yml)
[![SBOM Available](https://img.shields.io/badge/SBOM-Available-success)](https://github.com/snowdreamtech/UniGoDesktop/releases/latest)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/license/MIT)
[![Release](https://img.shields.io/github/v/release/snowdreamtech/UniGoDesktop?logo=github&sort=semver)](https://github.com/snowdreamtech/UniGoDesktop/releases/latest)
[![Dependabot Enabled](https://img.shields.io/badge/Dependabot-Enabled-brightgreen?logo=dependabot)](https://github.com/snowdreamtech/UniGoDesktop/blob/main/.github/dependabot.yml)
[![pre-commit](https://img.shields.io/badge/pre--commit-enabled-brightgreen?logo=pre-commit)](https://github.com/pre-commit/pre-commit)
[![GitHub Stars](https://img.shields.io/github/stars/snowdreamtech/UniGoDesktop?style=social)](https://github.com/snowdreamtech/UniGoDesktop)
[![GitHub Issues](https://img.shields.io/github/issues/snowdreamtech/UniGoDesktop)](https://github.com/snowdreamtech/UniGoDesktop/issues)
[![Code Size](https://img.shields.io/github/languages/code-size/snowdreamtech/UniGoDesktop)](https://github.com/snowdreamtech/UniGoDesktop)

[English](README.md) | [简体中文](README_zh-CN.md)

UniGoDesktop is a fast, flexible, and enterprise-grade Golang Desktop Application Template inspired by UniRTM and helloworld. It provides a robust foundation for building modern cross-platform desktop applications in Go with an embedded web-based user interface, system tray support, and a CLI fallback mode.

## 🌟 Features

- **Golang Desktop Engine**: Structured application lifecycle (`internal/desktop`) managing windows, system tray, and local background services.
- **Embedded Web UI**: Serves embedded HTML/CSS/JS frontend assets (`embed.FS`) over a lightweight local web server bridge.
- **System Tray Integration**: Native system tray support with dynamic menus (Show Window, Status, Exit).
- **Dual-Mode Execution**: Run as an interactive desktop GUI (`unigodesktop desktop` or default) or as a headless CLI tool.
- **Modern CLI Architecture**: Built with Go 1.27+ and Cobra, providing structured command-line functionality.
- **Structured Logging & Cache**: Standard `log/slog` structured logging and local SQLite database storage.
- **Cross-Platform Ready**: Seamless cross-platform support for macOS, Linux, and Windows.

## 🏗️ Architecture & Design

### Overview

UniGoDesktop is engineered to solve the boilerplate problem when starting new Go CLI projects. It standardizes the development environment, architectural patterns, and automation pipelines out of the box.

### Core Components

- **CLI Framework**: Utilizes `spf13/cobra` and `spf13/viper` for powerful command parsing and configuration management.
- **UI & Logging**: Integrated `pterm` for UI components and `slog` for structured event logging.
- **Database & Caching**: SQLite-backed (via `modernc.org/sqlite`) local caching and transactional data storage.
- **Task Orchestration**: Configured with `UniRTM` for managing local development workflows (lint, test, verify).

## 📖 Usage Guide

### Prerequisites

- **Runtime**: Go (>= 1.24).
- **Git**: Global git installation required.
- **UniRTM**: Required for task execution and orchestration.

### Installation

**Via NPM**:

```sh-session
npm install -g @snowdreamtech/unigodesktop
```

**Via PyPI**:

```sh-session
pip install snowdreamtech-unigo
```

### Quick Start

1. **Install UniRTM**: Ensure `unirtm` is installed on your system.
2. **Initialize**: `unirtm run setup` (bootstraps core dependencies and hooks).
3. **Install**: `unirtm run install` (installs project dependencies).
4. **Verify**: `unirtm run verify` (ensures everything is green).
5. **Build**: `go build -o unigo main.go`

### Available Commands

- `unigo version`: Print the version number
- `unigo doctor`: Check system health and diagnose issues
- `unigo self-update`: Update to the latest version
- `unigo cache`: Manage local cache
- `unigo config`: Manage configuration
- `unigo license`: Manage copyright license headers

## 🛠️ Operations Guide

### Pre-deployment Checklist

1. Run `unirtm run verify` to ensure all quality gates are green (runs formatters, linters, tests, and security audits).
2. Ensure `CHANGELOG.md` is updated.

### Troubleshooting

- **Problem**: `unirtm run verify` fails with test errors.
  - **Solution**: Ensure your code passes all Go unit tests. UniGoDesktop enforces strict coverage and race condition checks.
- **Problem**: Pre-commit hooks fail.
  - **Solution**: The hooks often auto-fix issues (like formatting). Stage the modified files and commit again.

## 🔒 Security Considerations

### Security Model

- **Audit Logging**: All critical operations and security scans are executed during the `verify` task using tools like `trivy`, `gitleaks`, and `govulncheck`.
- **Dependency Management**: Powered by Dependabot to ensure all dependencies are kept up-to-date and secure.

## 🧑‍💻 Development Guide

### Local Development Setup

```bash
git clone https://github.com/snowdreamtech/UniGoDesktop.git
cd UniGoDesktop
unirtm run setup
unirtm run install
```

### 🚀 Proxy Usage Scenarios

The `GITHUB_PROXY` (default: `https://gh-proxy.sn0wdr1am.com/`) is optimized for specific network acceleration scenarios. Misusing it for unsupported protocols (like Git) will result in errors.

| Scenario              | Supported? | Example / Note                                         |
| :-------------------- | :--------- | :----------------------------------------------------- |
| **Release Files**     | ✅ Yes     | `.../releases/download/v1.0/tool.zip`                  |
| **Source Archives**   | ✅ Yes     | `.../archive/master.zip` or `.tar.gz`                  |
| **Direct File Links** | ✅ Yes     | `.../blob/master/filename`                             |
| **Git Clone**         | ❌ **No**  | Do **not** use for `git clone` or `insteadOf` configs. |
| **Project Folders**   | ❌ **No**  | Browsing/cloning via proxy is not supported.           |

> [!IMPORTANT]
> To prevent breaking toolchains (like `unirtm`), this template explicitly disables Git redirection via this proxy. Use it only for direct HTTP downloads in scripts.

## 📄 License

This project is licensed under the **MIT License**.
Copyright (c) 2026-present [SnowdreamTech Inc.](https://github.com/snowdreamtech)
See the [LICENSE](./LICENSE) file for the full license text.

## Star History

[![Star History Chart](https://api.star-history.com/image?repos=snowdreamtech/UniGoDesktop&type=date&legend=top-left)](https://www.star-history.com/?repos=snowdreamtech%2FUniGoDesktop&type=date&legend=top-left)

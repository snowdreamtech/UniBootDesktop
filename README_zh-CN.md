# UniGoDesktop

[![CI 流水线](https://img.shields.io/github/actions/workflow/status/snowdreamtech/UniGoDesktop/ci.yml?branch=main&label=CI%20Pipeline)](https://github.com/snowdreamtech/UniGoDesktop/actions/workflows/ci.yml)
[![CD 自动化发布](https://img.shields.io/github/actions/workflow/status/snowdreamtech/UniGoDesktop/cd.yml?branch=main&label=CD%20Pipeline)](https://github.com/snowdreamtech/UniGoDesktop/actions/workflows/cd.yml)
[![文档站点](https://img.shields.io/github/actions/workflow/status/snowdreamtech/UniGoDesktop/pages.yml?branch=main&label=%E6%96%87%E6%A1%A3&logo=github)](https://github.com/snowdreamtech/UniGoDesktop/actions/workflows/pages.yml)
[![CodeQL 审计](https://img.shields.io/github/actions/workflow/status/snowdreamtech/UniGoDesktop/codeql.yml?branch=main&label=CodeQL&logo=github)](https://github.com/snowdreamtech/UniGoDesktop/actions/workflows/codeql.yml)
[![跨平台验证](https://img.shields.io/badge/Verified-Linux%20%7C%20macOS%20%7C%20Windows-blue)](https://github.com/snowdreamtech/UniGoDesktop/actions/workflows/ci.yml)
[![安全审计](https://img.shields.io/badge/Security-Zizmor%20%7C%20Trivy%20%7C%20Gitleaks-brightgreen)](https://github.com/snowdreamtech/UniGoDesktop/actions/workflows/ci.yml)
[![SBOM 背书](https://img.shields.io/badge/SBOM-Available-success)](https://github.com/snowdreamtech/UniGoDesktop/releases/latest)
[![许可证: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/license/MIT)
[![最新发布](https://img.shields.io/github/v/release/snowdreamtech/UniGoDesktop?logo=github&sort=semver)](https://github.com/snowdreamtech/UniGoDesktop/releases/latest)
[![Dependabot 已启用](https://img.shields.io/badge/Dependabot-Enabled-brightgreen?logo=dependabot)](https://github.com/snowdreamtech/UniGoDesktop/blob/main/.github/dependabot.yml)
[![pre-commit 已启用](https://img.shields.io/badge/pre--commit-enabled-brightgreen?logo=pre-commit)](https://github.com/pre-commit/pre-commit)
[![GitHub Stars](https://img.shields.io/github/stars/snowdreamtech/UniGoDesktop?style=social)](https://github.com/snowdreamtech/UniGoDesktop)
[![GitHub Issues](https://img.shields.io/github/issues/snowdreamtech/UniGoDesktop)](https://github.com/snowdreamtech/UniGoDesktop/issues)
[![代码规模](https://img.shields.io/github/languages/code-size/snowdreamtech/UniGoDesktop)](https://github.com/snowdreamtech/UniGoDesktop)

[English](README.md) | [简体中文](README_zh-CN.md)

UniGoDesktop 是一个快速、灵活且企业级的 Golang 桌面端应用模版项目 (Golang Desktop Template)，深受 UniRTM 和 helloworld 的启发。它为构建现代跨平台 Go 桌面应用提供了坚实的基础，内置嵌入式 Web 界面、系统托盘 (System Tray) 支持以及 CLI/Headless 命令行模式。

## 🌟 特性

- **Golang 桌面引擎**：结构化的桌面应用生命周期管理 (`internal/desktop`)，管理主窗口、系统托盘与本地后台服务。
- **嵌入式 Web UI**：通过轻量级本地 Web Bridge 服务器提供嵌入式前端静态资源 (HTML/CSS/JS)。
- **系统托盘集成**：原生系统托盘菜单（显示主窗口、应用状态、退出）。
- **双模运行能力**：支持以交互式桌面 GUI 模式 (`unigodesktop desktop` 或默认) 或命令行 CLI 工具模式运行。
- **现代 CLI 架构**：基于 Go 1.27+ 和 Cobra 构建，提供完整的结构化命令行功能。
- **结构化日志与数据库缓存**：内置 Go `log/slog` 高性能结构化日志与 SQLite 本地持久化存储。
- **跨平台就绪**：在 macOS、Linux 和 Windows 上均可无缝构建与运行。

## 🏗️ 架构与设计

### 概览

UniGoDesktop 旨在解决每次启动新的 Go CLI 项目时遇到的模板化和重复配置问题。它提供了开箱即用的标准开发环境、架构模式和自动化流水线。

### 核心组件

- **CLI 框架**：使用 `spf13/cobra` 和 `spf13/viper` 进行强大的命令解析和配置管理。
- **UI 与 日志**：集成了 `pterm` 作为 UI 组件，并使用 `slog` 进行结构化的事件记录。
- **数据库与缓存**：由 SQLite 驱动 (通过 `modernc.org/sqlite`)，支持本地缓存和事务性数据存储。
- **任务编排**：内置了 `UniRTM` 配置，用于管理本地开发工作流（lint、test、verify）。

## 📖 使用指南

### 前置条件

- **运行时**: Go (>= 1.24)。
- **Git**: 需要全局安装 git。
- **UniRTM**: 必须安装以执行任务编排。

### 安装

**通过 NPM 安装**:

```sh-session
npm install -g @snowdreamtech/unigodesktop
```

**通过 PyPI 安装**:

```sh-session
pip install snowdreamtech-unigo
```

### 快速开始

1. **安装 UniRTM**：请确保系统已安装 `unirtm`。
2. **初始化**：`unirtm run setup`（引导安装核心工具与钩子）。
3. **安装依赖**：`unirtm run install`（安装项目依赖）。
4. **验证**：`unirtm run verify`（确保所有代码检查通过）。
5. **构建**：`go build -o unigo main.go`

### 可用命令

- `unigo version`: 打印版本号
- `unigo doctor`: 检查系统健康状况并诊断问题
- `unigo self-update`: 更新到最新版本
- `unigo cache`: 管理本地缓存
- `unigo config`: 管理配置
- `unigo license`: 管理源代码文件中的版权许可头

## 🛠️ 运维指南

### 部署前检查清单

1. 运行 `unirtm run verify` 确保所有质量门禁均为绿色（执行格式化、Lint、测试及安全审计）。
2. 确保 `CHANGELOG.md` 已更新。

### 故障排除

- **问题**: `unirtm run verify` 报测试失败。
  - **解决方案**: 确保代码通过所有 Go 单元测试。UniGoDesktop 会执行严格的代码覆盖率和竞态条件 (race) 检查。
- **问题**: Pre-commit 钩子报错。
  - **解决方案**: 许多钩子（如代码格式化）会自动修复问题。将修改后的文件重新 `git add` 并再次提交即可。

## 🔒 安全注意事项

### 安全模型

- **审计日志**: 所有的关键操作与安全扫描，均在使用 `unirtm run verify` 时，由 `trivy`、`gitleaks` 和 `govulncheck` 等工具自动执行。
- **依赖管理**: 由 Dependabot 支持，以确保所有项目依赖项始终保持最新且安全的状态。

## 🧑‍💻 开发者指南

### 本地开发设置

```bash
git clone https://github.com/snowdreamtech/UniGoDesktop.git
cd UniGoDesktop
unirtm run setup
unirtm run install
```

### 🚀 代理使用场景

`GITHUB_PROXY` (默认: `https://gh-proxy.sn0wdr1am.com/`) 针对特定的网络加速场景进行了优化。在不支持的协议（如 Git）上误用它会导致错误。

| 场景                   | 是否支持      | 示例 / 说明                                    |
| :--------------------- | :------------ | :--------------------------------------------- |
| **Release 文件**       | ✅ 支持       | `.../releases/download/v1.0/tool.zip`          |
| **源码归档 (Archive)** | ✅ 支持       | `.../archive/master.zip` 或 `.tar.gz`          |
| **文件直接链接**       | ✅ 支持       | `.../blob/master/filename`                     |
| **Git Clone**          | ❌ **不支持** | **请勿**用于 `git clone` 或 `insteadOf` 配置。 |
| **项目文件夹**         | ❌ **不支持** | 不支持通过代理进行项目文件夹的浏览或克隆。     |

> [!IMPORTANT]
> 为了防止破坏工具链（如 `unirtm`），本模板显式禁用了通过此代理进行的 Git 重定向。请仅在脚本中进行直接 HTTP 下载时使用它。

## 📄 许可证

本项目采用 **MIT 许可证** 授权。
版权所有 (c) 2026-现在 [SnowdreamTech Inc.](https://github.com/snowdreamtech)
详见 [LICENSE](./LICENSE) 文件。

## Star History

[![Star History Chart](https://api.star-history.com/image?repos=snowdreamtech/UniGoDesktop&type=date&legend=top-left)](https://www.star-history.com/?repos=snowdreamtech%2FUniGoDesktop&type=date&legend=top-left)

# ADR 0005: Dual-Mode GUI/CLI Architecture, Pure-Go SQLite Storage, and Lightweight Split i18n Engine

## Status

Accepted

## Context

UniGoDesktop is a desktop application suite combining an interactive graphical interface with headless/command-line automation capabilities across macOS, Windows, and Linux. When designing the application foundation, four key architectural challenges emerged:

1. **GUI and CLI Binary Coexistence:** The project needed to support both an interactive GUI desktop app and headless/command-line operational modes without duplicating business logic, database migrations, or configuration schemas across separate repositories.
2. **Cross-Compilation Simplicity:** Standard Go SQLite drivers (such as `mattn/go-sqlite3`) require `CGO_ENABLED=1` and platform-specific C cross-compilers (`gcc`, `mingw-w64`, `musl-gcc`). In GitHub Actions and automated release matrices (x86_64, aarch64, Windows, macOS, Linux), CGO introduces high build complexity, linker fragility, and cross-compilation bottlenecks.
3. **Desktop Framework Selection:** The GUI framework needed to produce compact, native-feeling desktop applications with fast startup, native menus/dialogs, low RAM consumption, and deep operating system integration.
4. **Comprehensive Multi-Language Support:** The application targets global users across 53 languages (including RTL languages such as Arabic, Hebrew, Urdu, and Persian). Heavy internationalization libraries (like `vue-i18n`) add substantial runtime bundle overhead and complex message format abstractions that complicate simple key-value lookups and parameter interpolation.

## Decision

We have established the following core design choices:

### 1. GUI + CLI Coexistence via Build Tags and Shared Core

- Business logic, database repositories, updater services, and configuration models reside in `pkg/` and `internal/` without any GUI dependencies.
- The GUI runner is isolated in `wails_gui.go` (and the `cmd/` package), configured with appropriate build constraints so that non-GUI tools, CLI workflows, and testing suites can compile with `CGO_ENABLED=0` without linking WebKitGTK or WebView2.
- Wails v2 is chosen as the GUI bridge: it provides native OS webview hosting (WebKit on macOS/Linux, WebView2 on Windows) without bundling Chromium, resulting in tiny binary footprints (< 30MB) compared to Electron (> 150MB).

### 2. Pure-Go Embedded Storage with `modernc.org/sqlite`

- The database layer (`internal/database`, `internal/repository/sqlite`) standardizes on `modernc.org/sqlite` (pure Go transpiled from SQLite C source).
- `CGO_ENABLED=0` is enforced across all test runners, CI lint jobs, and standard builds.
- This enables cross-compilation from any host OS to any target OS/architecture without requiring C cross-compilers or sysroot headers, ensuring 100% reproducible and portable builds.

### 3. Lightweight Custom i18n Engine with Route/Module Splitting

- Rather than importing third-party i18n frameworks, `frontend/src/i18n/` implements a type-safe, lightweight internationalization engine tailored to desktop requirements:
  - English (`en-US`) is statically bundled as the default fallback.
  - Remaining 52 locales are asynchronously lazy-loaded on demand via dynamic `import(...)` chunks, keeping the initial bundle under 135 KB.
  - Parameter interpolation supports standard tokens (`{param}` or `{ param }`) with safe regex handling.
  - Native RTL direction detection (`dir="rtl"`) automatically adjusts UI layout for Arabic, Persian, Hebrew, and Urdu.

## Consequences

### Positive

- **Trivial Cross-Compilation:** Pure-Go SQLite allows fast, deterministic compilation for all target OS/arch combinations without CGO dependencies.
- **Resource Efficiency:** Small binary sizes, low RAM usage, and instant startup times compared to Electron.
- **Clean Separation of Concerns:** Core domain logic is decoupled from presentation layers, allowing easy automated unit testing and reuse across CLI and GUI.
- **Fast Initial Load:** Dynamic chunk loading ensures only the active language bundle is fetched and parsed.

### Negative

- `modernc.org/sqlite` can exhibit slightly lower peak query throughput than native CGO SQLite under heavy concurrent write loads; however, for a single-user desktop client application, this difference is negligible and far outweighed by portability benefits.

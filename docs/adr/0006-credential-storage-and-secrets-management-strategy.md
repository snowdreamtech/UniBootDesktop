# ADR 0006: Credential Storage and Secrets Management Strategy

## Status

Accepted

## Context

UniGoDesktop supports optional network proxy authentication (HTTP/HTTPS/SOCKS4/SOCKS5) requiring username and password fields configured by users through the graphical preferences modal or command-line parameters.

Storing credentials raises security and platform-compatibility challenges across desktop operating systems:

1. **Local File Persistence:** Configuration files written to disk (`config.toml`) historically contained the `ProxyPassword` field in plaintext.
2. **Access Control:** Desktop operating systems employ varying native security credential stores:
   - **macOS:** Apple Keychain Services (`Security.framework`).
   - **Windows:** Windows Credential Manager (`wincred`).
   - **Linux:** Freedesktop Secret Service API via D-Bus (`libsecret` / `gnome-keyring` / `ksecretservice`).
3. **Headless & Containerized Execution:** When running as a CLI tool or inside containerized CI environments, native OS credential daemons or D-Bus session brokers may not be present or initialized, necessitating a deterministic fallback mechanism.

## Decision

We adopt a two-phase credential storage and protection architecture:

### Phase 1: Baseline Security & Filesystem Hardening (Current Implementation)

1. **Strict File Permissions:** Configuration files are explicitly written with `0600` permissions (`-rw-------`), ensuring access is restricted strictly to the current operating system user.
2. **Atomic Configuration Persistence:** All configuration updates use temporary file generation with atomic rename semantics (as established in ADR 0004), preventing partial writes and race conditions.
3. **Data Sanitization in Logs and Memory:** Credentials such as `ProxyPassword` are excluded from structured loggers (`zerolog`), debug outputs, and crash reports.
4. **Masked UI Input:** The desktop interface masks password entry fields (`type="password"`) and eliminates persistent DOM storage of cleartext tokens.

### Phase 2: Platform-Native Keyring Integration (Evolutionary Roadmap)

1. **Abstract Secret Storage Interface:** Introduce a lightweight storage contract `pkg/secret.Store` exposing `Get(key string)`, `Set(key string, val []byte)`, and `Delete(key string)`.
2. **OS Keychain Providers:**
   - Under macOS, delegate storage to macOS Keychain via native APIs.
   - Under Windows, delegate to Windows Credential Manager.
   - Under Linux desktop environments, communicate with the Secret Service D-Bus interface.
3. **Graceful Fallback:** If the system credential service is unavailable (e.g. headless Linux servers, Docker containers, or minimal CI runners), the application gracefully falls back to encrypted local storage derived from user-scoped machine keys or environment variables (`UNIGO_PROXY_PASSWORD`).

## Consequences

### Positive

- **Transparent Defense-in-Depth:** Filesystem permissions (`0600`) and sanitized log pipelines protect credentials against multi-user inspection on the same host.
- **Portability:** The application runs reliably in headless environments, CI pipelines, and all major desktop platforms without mandatory external C library dependencies.
- **Clear Evolution Path:** Developers and security auditors have an unambiguous architectural blueprint for native keyring integration.

### Negative

- Native OS keyring integrations require platform-specific API shims or pure-Go D-Bus client bindings, which will be implemented incrementally in Phase 2 to preserve pure-Go cross-compilation guarantees.

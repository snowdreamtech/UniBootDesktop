# Remaining Security Risk Audit

## Scope

This document records the remaining high-priority security risks after the shutdown safety hardening and user-eject validation work. The focus is on risk areas that still need a second round of enforcement before the project can be considered fully hardened.

## Audit Summary

### Priority 1 — raw device access and external command execution

These are the highest-risk areas because they can reach system resources directly and may trigger destructive or privilege-escalating operations.

1. Raw device path execution without a strict allowlist
   - Files involved:
     - [pkg/privilege/privilege.go](../../pkg/privilege/privilege.go)
     - [pkg/qemu/qemu.go](../../pkg/qemu/qemu.go)
     - [pkg/installer/installer.go](../../pkg/installer/installer.go)
   - Risk:
     - direct reads/writes to raw physical disks
     - misuse of device paths from user input or partially trusted config
     - possible writes to the wrong volume if validation is bypassed
   - Required mitigation:
     - require device path normalization and allowlisting
     - reject non-removable/system disks before raw I/O
     - keep raw access behind a single guarded helper rather than scattered direct calls

2. Unstructured OS command execution through multiple call sites
   - Files involved:
     - [pkg/installer/partition.go](../../pkg/installer/partition.go)
     - [pkg/installer/ventoy_cli.go](../../pkg/installer/ventoy_cli.go)
     - [pkg/hypervisor/driver_qemu.go](../../pkg/hypervisor/driver_qemu.go)
     - [pkg/hypervisor/driver_vbox.go](../../pkg/hypervisor/driver_vbox.go)
     - [pkg/hypervisor/driver_vmware.go](../../pkg/hypervisor/driver_vmware.go)
     - [pkg/hypervisor/hypervisor.go](../../pkg/hypervisor/hypervisor.go)
   - Risk:
     - command injection through untrusted path or argument strings
     - inconsistent escaping and validation patterns across OS-specific implementations
     - hard-to-audit shell execution pathways
   - Required mitigation:
     - implement one safety wrapper for all OS command execution
     - validate command names, arguments, and target paths before execution
     - reject shell metacharacters and unsafe file paths

3. Unsafe privilege escalation flow
   - File involved:
     - [pkg/privilege/privilege.go](../../pkg/privilege/privilege.go)
   - Risk:
     - even with current validation, the system still executes commands via shell as the final step
     - privilege boundaries are necessary, but they should be hardened and centralized
   - Required mitigation:
     - keep the current validation, but add a single policy-driven “safe elevated executor” abstraction
     - separate command construction from execution
     - log every elevated operation with the target and sanitized arguments

### Priority 2 — file write integrity and rollback

1. Final-write operations still need stricter atomicity and rollback across more paths
   - Files involved:
     - [pkg/config/config.go](../../pkg/config/config.go)
     - [pkg/firmware/firmware.go](../../pkg/firmware/firmware.go)
     - [pkg/installer/ventoy.go](../../pkg/installer/ventoy.go)
   - Risk:
     - partial writes or corrupted outputs if a write fails mid-operation
     - restore logic is only as strong as the upstream validation and path rules
   - Required mitigation:
     - ensure all destination files are created via temp-file + rename
     - validate content after write and restore backup on mismatch
     - enforce destination directory allowlists

2. Output path trust boundary
   - Risk:
     - if a user-provided path points outside a known safe directory, a write may end up in an unexpected place
   - Required mitigation:
     - canonicalize paths
     - reject parent traversal and non-expected directory roots
     - keep app-owned output paths restricted to the managed working directories

### Priority 3 — auditability and detection

1. External execution and raw-disk mutation are not yet fully logged
   - Risk:
     - when an issue occurs, the system may not expose enough evidence to determine which device or command caused it
   - Required mitigation:
     - add structured audit logs for all elevated actions, raw disk reads, and system command invocations
     - include target device, destination path, command fragment, and timestamp

2. CI enforcement gap
   - Risk:
     - a future regression could reintroduce risky patterns even if local tests pass
   - Required mitigation:
     - add `gosec` / `govulncheck` / dependency audit to CI
     - enforce high-severity blocking rules in pull requests

## Recommended Execution Plan

### Phase 1: document and prioritize

This document completes the first phase. It defines the remaining work as a staged security plan instead of a blanket rewrite.

### Phase 2: unify and harden execution paths

The next step is to implement a single guarded execution layer for all OS-command invocation and raw-device operations. This should include:

- normalized path validation
- device allowlist checks
- command allowlist and argument filtering
- explicit per-operation logging
- hard timeout and cancellation

### Phase 3: enforce in CI and regression tests

- add security-focused tests around bad device paths, unsafe command arguments, and write rollback behavior
- keep the test suite as a gate before merge

## Conclusion

The project has moved from a state with obvious shutdown and eject issues into a state with strong partial security controls. The highest remaining risk is not broad “random bug risk”; it is the accumulation of raw-disk and shell-execution paths that still need to be centralized and validated.

The next defensive improvement should be to centralize all external execution and disk access behind a single safety boundary rather than continuing to scatter OS-specific calls across the codebase.

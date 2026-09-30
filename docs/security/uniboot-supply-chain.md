# uniboot Supply Chain Security Analysis

## Overview

This document analyzes the supply chain risks associated with uniboot's default registry and provides mitigation strategies.

## Risk Analysis

### 1. Implicit Registry Redirection

**Risk**: uniboot's built-in registry can silently redirect tool installations to different backends.

**Example**:

```toml
# You specify:
"github:checkmake/checkmake" = "v0.3.2"

# But uniboot's registry maps 'checkmake' to:
aqua:mrtazz/checkmake
```

**Impact**:

- Different maintainer (`mrtazz` vs `checkmake` organization)
- Additional layer (aqua registry) increases attack surface
- Potential for supply chain attacks if registry is comprounibootd

### 2. Affected Tools in This Project

Based on uniboot registry inspection, the following tools have registry mappings:

```bash
checkmake                     aqua:mrtazz/checkmake
gitleaks                      aqua:gitleaks/gitleaks
hadolint                      aqua:hadolint/hadolint
```

## Mitigation Strategies

### ✅ Already Implemented

1. **Explicit Backend Specification**: All tools in `.unibootdesktop.toml` use explicit backends:
   - `github:owner/repo` for GitHub releases
   - `npm:package` for npm packages
   - `pipx:package` for Python packages

2. **Tool Spec Mapping**: In `scripts/lib/lint-wrapper.sh`, we explicitly map tool names to full specs:

   ```bash
   checkmake)
     _UNIRTM_TOOL_SPEC="github:checkmake/checkmake"
     _LINTER_BIN="checkmake"
     ;;
   ```

3. **Version Pinning**: All tools are pinned to specific versions in `.unibootdesktop.toml`

### 🔒 Additional Recommendations

#### 1. Disable uniboot Registry (Future)

When uniboot supports it, consider disabling the default registry:

```toml
[settings]
disable_default_registry = true  # Not yet supported
```

#### 2. Audit Tool Sources

Regularly verify that installed tools match expected sources:

```bash
# Check what uniboot actually installed
uniboot list

# Verify binary checksums against official releases
uniboot exec -- <tool> --version
```

#### 3. Use uniboot.lock for Reproducibility

The `uniboot.lock` file ensures consistent installations across environments:

```bash
# Verify lock file matches configuration
uniboot install --frozen
```

#### 4. Monitor uniboot Registry Changes

Watch for changes in uniboot's registry that might affect your tools:

```bash
# Check current registry mappings
uniboot registry | grep -E "(checkmake|gitleaks|hadolint)"
```

## Verification Steps

### Before Deployment

1. **Verify Tool Sources**:

   ```bash
   uniboot list | grep -v "npm:" | grep -v "pipx:"
   ```

2. **Check for Unexpected Backends**:

   ```bash
   uniboot list | grep "aqua:"
   ```

   Should only show tools you explicitly configured with aqua backend.

3. **Validate Binary Integrity**:

   ```bash
   # For GitHub releases, verify against official checksums
   uniboot where github:checkmake/checkmake
   sha256sum $(uniboot where github:checkmake/checkmake)/bin/checkmake
   ```

### During CI/CD

Our CI workflows already implement:

- ✅ Locked uniboot versions (`UNIRTM_LOCKED=1`)
- ✅ Explicit tool specs in lint-wrapper.sh
- ✅ Version pinning in .unibootdesktop.toml
- ✅ uniboot.lock committed to repository

## Related Security Measures

1. **Dependabot**: Monitors uniboot tool versions
2. **Trivy**: Scans for vulnerabilities in binaries
3. **SBOM Generation**: Documents all tool dependencies
4. **Signed Commits**: Ensures code integrity

## References

- [uniboot Registry Documentation](https://github.com/snowdreamtech/UniBootDesktopregistry.html)
- [uniboot Security Policy](https://github.com/jdx/uniboot/blob/main/SECURITY.md)
- [uniboot Paranoid Mode](https://github.com/snowdreamtech/UniBootDesktopparanoid)
- [SLSA Framework](https://slsa.dev/)

## Action Items

- [ ] Monitor uniboot for registry disable feature
- [ ] Set up automated alerts for registry changes
- [ ] Document tool source verification in CI
- [ ] Consider contributing to uniboot for better registry transparency

---

**Last Updated**: 2026-04-16
**Reviewed By**: Security Team
**Next Review**: 2026-07-16

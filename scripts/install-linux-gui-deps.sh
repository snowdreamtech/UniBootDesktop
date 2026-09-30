#!/usr/bin/env sh
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# install-linux-gui-deps.sh — Cross-platform GUI dependency installer for Linux
# Supports: Debian/Ubuntu, RHEL/CentOS/Fedora/Rocky/Alma, Alpine, Arch, openSUSE

set -eu

# Check if dependencies are already satisfied (Idempotency)
if command -v pkg-config >/dev/null 2>&1 &&
  pkg-config --exists gtk+-3.0 &&
  (pkg-config --exists webkit2gtk-4.0 || pkg-config --exists webkit2gtk-4.1); then
  echo "✓ Linux GUI build dependencies (GTK3 & WebKit2GTK) are already installed."
  exit 0
fi

# Determine privilege escalation helper
SUDO=""
if [ "$(id -u)" -ne 0 ]; then
  if command -v sudo >/dev/null 2>&1; then
    SUDO="sudo"
  else
    echo "ERROR: Root privileges or sudo required to install system dependencies." >&2
    exit 1
  fi
fi

# Detect package manager and install platform-specific packages
if command -v apt-get >/dev/null 2>&1; then
  echo "Detected Debian/Ubuntu-based system (apt-get)."
  export DEBIAN_FRONTEND=noninteractive
  retry=0
  while [ "$retry" -lt 3 ]; do
    if $SUDO apt-get update -qq; then
      break
    fi
    retry=$((retry + 1))
    sleep 3
  done
  $SUDO apt-get install -y --no-install-recommends \
    build-essential \
    pkg-config \
    libgtk-3-dev
  if ! $SUDO apt-get install -y --no-install-recommends libwebkit2gtk-4.0-dev; then
    $SUDO apt-get install -y --no-install-recommends libwebkit2gtk-4.1-dev
  fi

elif command -v dnf >/dev/null 2>&1 || command -v yum >/dev/null 2>&1; then
  PKG_MGR="dnf"
  if ! command -v dnf >/dev/null 2>&1; then
    PKG_MGR="yum"
  fi
  echo "Detected RedHat/Fedora/CentOS-based system ($PKG_MGR)."

  # Enable PowerTools (CentOS/RHEL 8) or CRB (CentOS/RHEL 9) if available
  if command -v dnf >/dev/null 2>&1; then
    $SUDO dnf config-manager --set-enabled powertools 2>/dev/null || true
    $SUDO dnf config-manager --set-enabled crb 2>/dev/null || true
    $SUDO dnf config-manager --enable powertools 2>/dev/null || true
    $SUDO dnf config-manager --enable crb 2>/dev/null || true
  fi
  $SUDO "$PKG_MGR" install -y epel-release 2>/dev/null || true

  # Install build toolchain and GTK3
  $SUDO "$PKG_MGR" install -y gcc gcc-c++ make gtk3-devel

  # Install pkg-config with backward-compatible package name fallback
  if ! $SUDO "$PKG_MGR" install -y pkgconf-pkg-config 2>/dev/null; then
    $SUDO "$PKG_MGR" install -y pkgconfig 2>/dev/null || $SUDO "$PKG_MGR" install -y pkgconf 2>/dev/null || true
  fi

  # Attempt WebKit2GTK candidate packages across different RHEL/CentOS/Fedora versions
  WEBKIT_INSTALLED=0
  for pkg in webkit2gtk4.1-devel webkit2gtk4.0-devel webkit2gtk3-devel webkitgtk4-devel; do
    if $SUDO "$PKG_MGR" install -y "$pkg" 2>/dev/null; then
      WEBKIT_INSTALLED=1
      break
    fi
  done
  if [ "$WEBKIT_INSTALLED" -eq 0 ]; then
    echo "WARNING: Could not install WebKit2GTK devel packages via $PKG_MGR standard repositories." >&2
  fi

elif command -v apk >/dev/null 2>&1; then
  echo "Detected Alpine Linux (apk)."
  $SUDO apk update
  $SUDO apk add --no-cache \
    build-base \
    pkgconf \
    gtk+3.0-dev
  if ! $SUDO apk add --no-cache webkit2gtk-dev; then
    $SUDO apk add --no-cache webkit2gtk-4.1-dev || true
  fi

elif command -v pacman >/dev/null 2>&1; then
  echo "Detected Arch Linux (pacman)."
  $SUDO pacman -Sy --noconfirm --needed \
    base-devel \
    pkgconf \
    gtk3
  if ! $SUDO pacman -S --noconfirm --needed webkit2gtk; then
    $SUDO pacman -S --noconfirm --needed webkit2gtk-4.1 || true
  fi

elif command -v zypper >/dev/null 2>&1; then
  echo "Detected openSUSE/SLES (zypper)."
  $SUDO zypper --non-interactive refresh
  $SUDO zypper --non-interactive install -y \
    patterns-devel-base-devel_basis \
    pkg-config \
    gtk3-devel
  if ! $SUDO zypper --non-interactive install -y webkit2gtk3-devel; then
    $SUDO zypper --non-interactive install -y webkit2gtk-4_1-devel || true
  fi

else
  echo "WARNING: Unsupported Linux package manager. Please ensure GTK3, WebKit2GTK, and pkg-config are installed." >&2
fi

# Final verification
if command -v pkg-config >/dev/null 2>&1 &&
  pkg-config --exists gtk+-3.0 &&
  (pkg-config --exists webkit2gtk-4.0 || pkg-config --exists webkit2gtk-4.1); then
  echo "✓ Linux GUI build dependencies (GTK3 & WebKit2GTK) verified."
  exit 0
else
  echo "ERROR: Linux GUI build dependencies could not be verified via pkg-config." >&2
  exit 1
fi

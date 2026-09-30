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
  $SUDO "$PKG_MGR" install -y epel-release || true
  $SUDO "$PKG_MGR" install -y \
    gcc \
    gcc-c++ \
    make \
    pkgconf-pkg-config \
    gtk3-devel
  if ! $SUDO "$PKG_MGR" install -y webkit2gtk3-devel; then
    $SUDO "$PKG_MGR" install -y webkit2gtk4.1-devel || true
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

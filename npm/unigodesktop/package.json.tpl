{
  "name": "@snowdreamtech/unigodesktop",
  "version": "{{VERSION}}",
  "description": "UniGoDesktop - Universal Runtime Manager: enterprise-grade foundational toolchain for multi-AI IDE collaboration",
  "license": "MIT",
  "homepage": "https://github.com/snowdreamtech/unigodesktop",
  "repository": {
    "type": "git",
    "url": "git+https://github.com/snowdreamtech/unigodesktop.git"
  },
  "bugs": {
    "url": "https://github.com/snowdreamtech/unigodesktop/issues"
  },
  "keywords": [
    "unigodesktop",
    "runtime",
    "manager",
    "toolchain",
    "ai",
    "ide"
  ],
  "bin": {
    "unigodesktop": "install.js"
  },
  "scripts": {
    "postinstall": "node install.js"
  },
  "files": [
    "install.js",
    "LICENSE",
    "README.md",
    "README_zh-CN.md"
  ],
  "optionalDependencies": {
    "@snowdreamtech/unigodesktop-darwin-arm64": "{{VERSION}}",
    "@snowdreamtech/unigodesktop-darwin-x64": "{{VERSION}}",
    "@snowdreamtech/unigodesktop-linux-x64": "{{VERSION}}",
    "@snowdreamtech/unigodesktop-linux-arm64": "{{VERSION}}",
    "@snowdreamtech/unigodesktop-linux-ia32": "{{VERSION}}",
    "@snowdreamtech/unigodesktop-linux-arm": "{{VERSION}}",
    "@snowdreamtech/unigodesktop-linux-loong64": "{{VERSION}}",
    "@snowdreamtech/unigodesktop-linux-ppc64le": "{{VERSION}}",
    "@snowdreamtech/unigodesktop-linux-riscv64": "{{VERSION}}",
    "@snowdreamtech/unigodesktop-linux-s390x": "{{VERSION}}",
    "@snowdreamtech/unigodesktop-windows-x64": "{{VERSION}}",
    "@snowdreamtech/unigodesktop-windows-arm64": "{{VERSION}}",
    "@snowdreamtech/unigodesktop-windows-ia32": "{{VERSION}}"
  },
  "engines": {
    "node": ">=18"
  }
}

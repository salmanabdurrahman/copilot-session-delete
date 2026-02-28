# Changelog

All notable changes to this project will be documented in this file.

This file is updated automatically by the release workflow after each tag is published.
For the full release history, see the [GitHub Releases page](https://github.com/salmanabdurrahman/copilot-session-delete/releases).

<!-- CHANGELOG_ENTRIES -->
## [v0.1.0] — 2026-02-28

## Install

**Quick install (recommended)**
```bash
curl -fsSL https://raw.githubusercontent.com/salmanabdurrahman/copilot-session-delete/main/scripts/install.sh | bash
```

**Manual install**

macOS (Apple Silicon):
```bash
curl -L https://github.com/salmanabdurrahman/copilot-session-delete/releases/download/v0.1.0/copilot-session-delete_Darwin_arm64.tar.gz | tar xz
mkdir -p ~/.local/bin && mv copilot-session-delete ~/.local/bin/
export PATH="$HOME/.local/bin:$PATH"
```

macOS (Intel):
```bash
curl -L https://github.com/salmanabdurrahman/copilot-session-delete/releases/download/v0.1.0/copilot-session-delete_Darwin_x86_64.tar.gz | tar xz
mkdir -p ~/.local/bin && mv copilot-session-delete ~/.local/bin/
export PATH="$HOME/.local/bin:$PATH"
```

Linux (amd64):
```bash
curl -L https://github.com/salmanabdurrahman/copilot-session-delete/releases/download/v0.1.0/copilot-session-delete_Linux_x86_64.tar.gz | tar xz
mkdir -p ~/.local/bin && mv copilot-session-delete ~/.local/bin/
export PATH="$HOME/.local/bin:$PATH"
```

**go install**
```bash
go install github.com/salmanabdurrahman/copilot-session-delete/cmd/copilot-session-delete@v0.1.0
```

<!-- Release notes generated using configuration in .github/release.yml at v0.1.0 -->



**Full Changelog**: https://github.com/salmanabdurrahman/copilot-session-delete/commits/v0.1.0

## Verify checksums

```bash
sha256sum --check checksums.txt
```

**Full Changelog**: https://github.com/salmanabdurrahman/copilot-session-delete/compare/...v0.1.0



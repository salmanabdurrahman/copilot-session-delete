# copilot-session-delete

> Browse and safely delete local [GitHub Copilot CLI](https://github.com/github/copilot-cli) sessions stored on your machine.

[![CI](https://github.com/salmanabdurrahman/copilot-session-delete/actions/workflows/ci.yml/badge.svg)](https://github.com/salmanabdurrahman/copilot-session-delete/actions/workflows/ci.yml)
[![Go version](https://img.shields.io/github/go-mod/go-version/salmanabdurrahman/copilot-session-delete)](go.mod)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

## Overview

Copilot CLI stores its chat sessions locally under `~/.copilot/session-state/` (macOS/Linux) or `%USERPROFILE%\.copilot\session-state\` (Windows). Over time these directories can accumulate gigabytes of data.

**copilot-session-delete** gives you a fast, safe way to review and remove those sessions — either through an interactive terminal UI or a scriptable non-interactive CLI.

**Key design principles:**

- 🔒 **Safety first** — UUID validation + canonical path check prevent any deletion outside the declared root.
- 🧪 **Dry-run by default** — the `delete` subcommand previews what would be removed unless you explicitly opt in.
- ✅ **Confirmation required** — a multi-step confirmation modal prevents accidental bulk deletes.

## Demo

<!-- Replace this block with an animated GIF once the tool is packaged:
![TUI demo](docs/assets/demo.gif)
-->

```
╭─ copilot-session-delete  ~/.copilot/session-state ─────────────────── 2/14 ─╮
│                                                                               │
│   / to search                                                                 │
│                                                                               │
│      SESSION ID       UPDATED AT        CWD/REPO               EVENTS        │
│   ────────────────────────────────────────────────────────────────────────   │
│ > [✓] 86334621-8152…  2026-02-28 10:47  github/copilot-cli        150        │
│   [ ] c0c723f4-08d2…  2026-02-28 09:18  my-project                 42        │
│   [✓] d1e2f3a4-0000…  2026-02-28 08:03  /home/user/work             7        │
│   [ ] e5f6a7b8-1111…  2026-02-27 22:51  another-repo               98        │
│                                                                               │
│  ✓ 2 session(s) deleted.                                                     │
│                                                                               │
│  [↑/↓] navigate  [/] search  [space] select  [a] all  [d] delete  [q] quit  │
╰───────────────────────────────────────────────────────────────────────────────╯
```

## Features

| Feature | Details |
|---|---|
| **Interactive TUI** | Bubble Tea–powered list with real-time search, multi-select, detail panel |
| **Safe deletion** | Confirmation modal, path safety checks, atomic sequential removal |
| **Dry-run mode** | Preview exactly what would be deleted — no files touched |
| **Non-interactive CLI** | `list` and `delete` subcommands for scripting / CI pipelines |
| **JSON output** | `list --json` produces stable, machine-readable output |
| **Cross-platform** | macOS, Linux, Windows (amd64 + arm64) |

## Installation

### Pre-built binaries (recommended)

Download the latest release for your platform from the [Releases page](https://github.com/salmanabdurrahman/copilot-session-delete/releases).

```bash
# macOS (Apple Silicon)
curl -L https://github.com/salmanabdurrahman/copilot-session-delete/releases/latest/download/copilot-session-delete_Darwin_arm64.tar.gz | tar xz
sudo mv copilot-session-delete /usr/local/bin/

# macOS (Intel)
curl -L https://github.com/salmanabdurrahman/copilot-session-delete/releases/latest/download/copilot-session-delete_Darwin_x86_64.tar.gz | tar xz
sudo mv copilot-session-delete /usr/local/bin/

# Linux (amd64)
curl -L https://github.com/salmanabdurrahman/copilot-session-delete/releases/latest/download/copilot-session-delete_Linux_x86_64.tar.gz | tar xz
sudo mv copilot-session-delete /usr/local/bin/
```

Verify checksums against the `checksums.txt` file published with each release.

### Build from source

Requires **Go 1.21+**.

```bash
git clone https://github.com/salmanabdurrahman/copilot-session-delete.git
cd copilot-session-delete
go build -o copilot-session-delete ./cmd/copilot-session-delete
```

### `go install`

```bash
go install github.com/salmanabdurrahman/copilot-session-delete/cmd/copilot-session-delete@latest
```

## Usage

### Interactive TUI (default)

```bash
copilot-session-delete
```

Launches the full-screen TUI. Sessions are loaded from the platform-default directory (`~/.copilot/session-state/`).

```bash
# Use a custom session directory
copilot-session-delete --session-dir /path/to/session-state

# Launch TUI in dry-run mode — 'd' confirm shows preview only, no files removed
copilot-session-delete --dry-run
```

### TUI Keybindings

| Key | Action |
|---|---|
| `↑` / `k` | Move cursor up |
| `↓` / `j` | Move cursor down |
| `g` | Jump to first session |
| `G` | Jump to last session |
| `PgUp` / `PgDn` | Scroll one page |
| `space` | Toggle selection |
| `a` | Select / deselect all visible |
| `enter` | Open session detail panel |
| `d` | Delete selected (or cursor row if none selected) |
| `/` | Focus search — filters list in real-time |
| `esc` | Close search / detail / cancel modal |
| `r` | Refresh session list |
| `q` / `ctrl+c` | Quit |

### Non-interactive CLI

#### List sessions

```bash
# Human-readable table
copilot-session-delete list

# Machine-readable JSON (stable schema)
copilot-session-delete list --json
```

JSON output schema:

```json
[
  {
    "id":          "86334621-8152-4e67-b322-9f139d6c0a57",
    "created_at":  "2026-02-28T09:46:29Z",
    "updated_at":  "2026-02-28T09:47:53Z",
    "cwd":         "/home/user/copilot-cli",
    "repository":  "github/copilot-cli",
    "branch":      "main",
    "summary":     "Explore source code and running instructions",
    "event_count": 150,
    "size_bytes":  2097152
  }
]
```

#### Delete sessions

> ⚠️ **Destructive operation.** Deleted sessions cannot be recovered. Always verify with `--dry-run` (the default) before committing.

```bash
# Preview what would be deleted (default — no files removed)
copilot-session-delete delete --id 86334621-8152-4e67-b322-9f139d6c0a57

# Delete a single session for real
copilot-session-delete delete \
  --id 86334621-8152-4e67-b322-9f139d6c0a57 \
  --dry-run=false

# Delete multiple sessions by ID
copilot-session-delete delete \
  --ids 86334621-8152-4e67-b322-9f139d6c0a57,c0c723f4-08d2-4257-9b30-5d2fd728dc45 \
  --dry-run=false

# Delete ALL sessions (requires --yes)
copilot-session-delete delete --all --dry-run=false --yes
```

### Global flags

| Flag | Default | Description |
|---|---|---|
| `--session-dir` | `~/.copilot/session-state` | Override the session directory path |
| `--dry-run` | `false` (TUI) / `true` (delete cmd) | Preview without removing files |
| `--version` | — | Print version and exit |

### Environment variables

| Variable | Description |
|---|---|
| `COPILOT_SESSION_DIR` | Override default session directory (same as `--session-dir`) |

## ⚠️ Safety Warnings

1. **Deleted sessions cannot be recovered.** There is no recycle bin or undo. Use `--dry-run` to preview before deleting.
2. **Only sessions inside the declared session-state directory can be deleted.** The tool validates that every target path is a canonical descendant of the root; symlink escapes and path-traversal attempts are rejected with an error.
3. **Only directories whose name is a valid UUID v4 are eligible for deletion.** Non-UUID entries are skipped automatically.
4. **The `delete --all` subcommand requires `--yes`** to prevent accidental bulk deletion in scripts.

## Session directory location

| Platform | Default path |
|---|---|
| macOS / Linux | `~/.copilot/session-state/` |
| Windows | `%USERPROFILE%\.copilot\session-state\` |

Override at runtime:
```bash
export COPILOT_SESSION_DIR=/custom/path
copilot-session-delete
```

## Troubleshooting

**`could not determine session directory`**  
Copilot CLI may not have created the directory yet, or the `HOME` / `USERPROFILE` environment variable is unset. Use `--session-dir` to point to the directory manually.

**`session not found: <id>`**  
The session ID you provided does not match any entry in the scanned directory. Run `copilot-session-delete list` to see available IDs.

**`safety check failed for session <id>`**  
The resolved path of the session falls outside the declared session-state root. This may indicate a symlink or a manually crafted path. The deletion is rejected as a safety measure.

**TUI shows blank screen on launch**  
Ensure your terminal is at least 40 columns wide. The TUI requires a minimum width to render.

**Permission denied when deleting**  
Check that the session directories are owned by your user. If Copilot CLI was run with elevated privileges the directories may be owned by root.

## Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) before opening a PR.

**Quick start:**

```bash
git clone https://github.com/salmanabdurrahman/copilot-session-delete.git
cd copilot-session-delete
go test ./...           # run all tests
bash scripts/check-local.sh   # run full quality gate
```

## Inspiration

This project was inspired by similar tools for managing Claude CLI sessions:

- [claude-chats-delete](https://github.com/ataleckij/claude-chats-delete) — Delete Claude AI chat sessions
- [claude-session-viewer](https://github.com/jtklinger/claude-session-viewer) — Browse and manage Claude session data

## License

[MIT](LICENSE) © Salman Abdurrahman

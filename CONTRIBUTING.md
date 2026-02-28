# Contributing to copilot-session-delete

Thank you for your interest in contributing! This document explains how to set up a development environment, run tests, and open a pull request.

## Code of Conduct

This project follows the [Contributor Covenant Code of Conduct](CODE_OF_CONDUCT.md). By participating you agree to uphold its standards.

## Prerequisites

| Tool | Minimum version |
|---|---|
| Go | 1.25 |
| git | 2.x |
| make (optional) | any |

No external services or accounts are required for local development.

## Development setup

```bash
# 1. Fork the repository on GitHub, then clone your fork
git clone https://github.com/<your-username>/copilot-session-delete.git
cd copilot-session-delete

# 2. Install dependencies
go mod download

# 3. Build the binary
go build -o bin/copilot-session-delete ./cmd/copilot-session-delete

# 4. Run all tests
go test ./...

# 5. Run the full local quality gate (build + test + race + vet + coverage)
bash scripts/check-local.sh
```

The binary is written to `bin/`. This directory is git-ignored.

## Project layout

```
cmd/copilot-session-delete/   Main entrypoint (Cobra CLI)
internal/
  app/
    cli/                      Non-interactive list / delete commands
    tui/                      Bubble Tea interactive TUI
  core/
    deletion/                 Planner + Executor (safety-checked removal)
    safety/                   UUID validation + path-traversal protection
    session/                  Scanner, metadata parser, enrichment
  adapters/
    output/                   Human / JSON formatters
test/
  fixtures/                   Fixture session data for tests
  integration/                End-to-end integration tests
docs/                         Release notes, release checklist
scripts/                      Dev helper scripts
configs/                      Static assets (ANSI demos, etc.)
```

## Running tests

```bash
# All tests
go test ./...

# With race detector
go test -race ./...

# Single package
go test ./internal/app/tui/...

# Coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

Coverage must stay **≥ 70%** for CI to pass.

## Code style

- Follow standard Go formatting: run `gofmt -w .` or `goimports -w .` before committing.
- Use `go vet ./...` to catch common issues.
- Comments should explain *why*, not just *what*.
- Test function names follow the convention `TestFunctionName_Scenario` — no test matrix IDs.
- All user-facing messages must be written in **English**.

## Commit style

Use short, imperative present-tense commit messages:

```
Add dry-run preview to delete command
Fix path-safety check for symlinked session dirs
Update keybinding table in README
```

One logical change per commit. Squash fixup commits before opening a PR.

## Pull request process

1. Open an issue first for non-trivial changes to discuss the approach.
2. Create a feature branch from `main`:
   ```bash
   git checkout -b feat/short-description
   ```
3. Make your changes, write/update tests, update documentation as needed.
4. Run the full quality gate:
   ```bash
   bash scripts/check-local.sh
   ```
5. Push your branch and open a PR against `main`.
6. Fill out the PR template completely.
7. At least one maintainer review is required before merging.

## Adding a new feature

- New behavior that is user-visible must be documented in `README.md`.
- New public functions must have Go doc comments.
- All new code paths must have corresponding tests. Aim for ≥ 80% coverage in new files.

## Reporting bugs

Use the [bug report template](.github/ISSUE_TEMPLATE/bug_report.yml) on GitHub Issues.

## Questions?

Open a [discussion](https://github.com/salmanabdurrahman/copilot-session-delete/discussions) on GitHub. Please avoid using Issues for general questions.

# Release Process

This document describes how to cut a new release of copilot-session-delete.

## End-to-end flow

```
Developer                   GitHub Actions (automatic)           Output
─────────                   ──────────────────────────           ──────

1. Open a Pull Request
         │
         ▼
   ┌─────────────────────────────────────────┐
   │  ci.yml  (triggers on PR + push/main)   │
   │                                         │
   │  build-and-test  [ubuntu / macos / win] │
   │    • go mod verify                      │
   │    • go vet + go build                  │
   │    • go test -v                         │
   │    • go test -race                      │
   │                                         │
   │  coverage  [ubuntu]                     │
   │    • go test -coverprofile              │
   │    • threshold >=70% enforced           │
   │    • artifact: coverage.out             │
   │                                         │
   │  lint  [ubuntu]                         │
   │    • golangci-lint                      │
   └─────────────────────────────────────────┘
         │ all green?
         ▼
2. Merge PR to main (manual)
   (Label PR with bug / enhancement / etc.
    for grouped changelog — see below)
         │
         ▼
   ci.yml runs again on main push
         │
3. Ready to release? Create and push a tag:
   git tag -s v0.x.0 -m "Release v0.x.0"
   git push origin v0.x.0   (manual)
         │
         ▼
   ┌─────────────────────────────────────────────────────────────┐
   │  release.yml  (triggers on v* tag push)                     │
   │                                                             │
   │  test  [ubuntu + macos, parallel]                           │
   │    • go test -race ./...                                    │
   │                    │                                        │
   │                    ▼  (both pass)                           │
   │  release  [ubuntu]                                          │
   │    • GoReleaser                                             │
   │      ├── go mod tidy                                        │
   │      ├── build 6 binaries (cross-platform):                 │
   │      │     linux/amd64   linux/arm64                        │
   │      │     darwin/amd64  darwin/arm64                       │
   │      │     windows/amd64 windows/arm64                      │
   │      ├── create .tar.gz/.zip archives                       │
   │      ├── generate checksums.txt (sha256)                    │
   │      ├── create GitHub Release                              │
   │      │     • header: install commands per platform          │
   │      │     • body: github-native auto-changelog             │
   │      │         (grouped by labels via .github/release.yml)  │
   │      │     • footer: checksum verify instructions           │
   │      └── attach all artifacts to release                    │
   │                    │                                        │
   │                    ▼  (release published)                   │
   │  update-changelog  [ubuntu]                                 │
   │    • gh release view v0.x.0 → fetch release body           │
   │    • prepend new entry to CHANGELOG.md                      │
   │    • git commit "chore: update CHANGELOG [skip ci]"         │
   │    • git push → main                                        │
   └─────────────────────────────────────────────────────────────┘
         │
         ▼
   GitHub Release page live  ──► users can download binaries
   CHANGELOG.md updated in repo ──► visible on GitHub file view
```

## Otomatis vs Manual

| Step | Mode | Who |
|---|---|---|
| Run CI on PR | Automatic | GitHub Actions |
| Run CI on merge to main | Automatic | GitHub Actions |
| Create version tag | Manual | Maintainer |
| Run tests before release | Automatic | GitHub Actions |
| Build cross-platform binaries | Automatic | GoReleaser |
| Create GitHub Release + attach binaries | Automatic | GoReleaser |
| Generate changelog from merged PRs | Automatic | GitHub (github-native) |
| Group changelog by category | Automatic | GitHub (.github/release.yml + PR labels) |
| Update CHANGELOG.md in repo | Automatic | update-changelog job |

## PR Labels for Grouped Changelog

Apply these labels to PRs before merging. GitHub will group them automatically in the release notes.

| Label | Changelog section |
|---|---|
| `enhancement`, `feature` | New Features |
| `bug`, `fix` | Bug Fixes |
| `security` | Security |
| `documentation` | Documentation |
| `breaking-change` | Breaking Changes |
| `skip-changelog` | (excluded from changelog) |
| *(no label)* | Other Changes |

## Versioning

This project follows [Semantic Versioning 2.0.0](https://semver.org/):

| Version part | When to bump |
|---|---|
| **MAJOR** (`x.0.0`) | Breaking changes to CLI flags, JSON output schema, or public Go API |
| **MINOR** (`0.x.0`) | New features that are backwards-compatible |
| **PATCH** (`0.0.x`) | Bug fixes, security patches, documentation updates |

Pre-release builds use the suffix `-alpha.N`, `-beta.N`, or `-rc.N` (e.g., `v0.2.0-rc.1`).

## Release checklist

### Before tagging

- [ ] All tests pass on `main`: `go test -race ./...`
- [ ] Coverage >= 70%: `bash scripts/check-local.sh`
- [ ] CI is green on `main` (all matrix jobs passing)
- [ ] `go.mod` and `go.sum` are up to date: `go mod tidy`
- [ ] `README.md` reflects any new flags, commands, or behavior
- [ ] Version is set automatically by GoReleaser via ldflags — no manual edit needed

### Tagging

```bash
# Create and push a signed annotated tag
git tag -s v0.x.0 -m "Release v0.x.0"
git push origin v0.x.0
```

Pushing the tag triggers the release workflow automatically.

### After tagging

- [ ] GitHub Actions `release` workflow completes (all 3 jobs: test → release → update-changelog)
- [ ] Release page has binaries for all 6 platforms
- [ ] `checksums.txt` is attached to the release
- [ ] `CHANGELOG.md` on `main` branch is updated automatically

### If a release fails

1. Delete the remote tag: `git push origin :v0.x.0`
2. Fix the issue on `main`
3. Re-tag and push

## GoReleaser — local dry run

```bash
# Install GoReleaser (one-time)
go install github.com/goreleaser/goreleaser/v2@latest

# Dry-run snapshot build (no git tag needed)
goreleaser release --snapshot --clean
```

Binaries are written to `dist/`.

## Backporting

Security fixes may be backported to the previous minor release. Create a `release/v0.x` branch and cherry-pick the relevant commits. Tag as `v0.x.Y`.

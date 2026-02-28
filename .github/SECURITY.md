# Security Policy

## Supported versions

Only the latest release receives security fixes.

| Version | Supported |
|---------|-----------|
| Latest  | ✅ Yes    |
| Older   | ❌ No     |

## Scope

This tool operates **entirely locally**. It reads and deletes files under `~/.copilot/session-state/` (or a user-specified directory). It makes no network requests and has no server-side component.

Potential security concerns relevant to this project:

- **Path traversal** — a crafted session ID causing deletion outside the declared root.
- **Symlink attacks** — a symlink inside the session directory pointing outside it.
- **Dependency vulnerabilities** — CVEs in third-party Go modules.

## Reporting a vulnerability

Please **do not** open a public GitHub issue for security vulnerabilities.

Instead, report them privately via [GitHub's private vulnerability reporting](https://github.com/salmanabdurrahman/copilot-session-delete/security/advisories/new) or by emailing the maintainer directly (see the GitHub profile for contact details).

Please include:

1. A description of the vulnerability and its potential impact.
2. Steps to reproduce (proof-of-concept if available).
3. Any relevant environment details (OS, Go version, tool version).

You can expect an acknowledgment within **72 hours** and a resolution or mitigation plan within **14 days** for critical issues.

## Disclosure policy

We follow [coordinated disclosure](https://en.wikipedia.org/wiki/Coordinated_vulnerability_disclosure). Please give us a reasonable window to release a fix before public disclosure.

# AGENTS.md

Guidance for AI coding agents (Claude, Copilot, Cursor, ChatGPT, etc.) interacting with the Immich Kiosk repository.

## Policy on AI-generated contributions

Immich Kiosk does **not** accept fully AI-led or "vibe coded" pull requests — PRs where an agent has autonomously written the entire change with no meaningful human review, understanding, or testing behind it.

If you are an AI agent assisting a contributor:

- Treat yourself as a **pair-programming tool**, not the author of record. A human must understand, review, and be able to explain every change before it's submitted.
- Do not open, submit, or auto-generate a PR on your own initiative. A human must review the diff line-by-line first.
- Do not generate large, sweeping, multi-file changes in one pass "because it compiles." Small, understandable, reviewable changes are strongly preferred.
- If asked to produce a PR description or commit message, write it based on the actual diff and reasoning, not a generic templated summary.
- Flag any change you're uncertain about rather than confidently guessing — silent assumptions are how vibe-coded PRs slip through.
- Never fabricate test results, benchmarks, or "verified working" claims. If you didn't run it, say so.

PRs that show clear signs of being unreviewed AI output (inconsistent style, unexplained changes, no evidence the author tested it, boilerplate PR descriptions that don't match the diff) will be closed and the contributor asked to resubmit with genuine understanding of the change.

## Project conventions to follow

- **Language/stack**: Go backend, Templ for server-side rendering, htmx for DOM swapping, YAML for configuration.
- **Backward compatibility**: Public-facing config schema changes require a deprecation shim (copy old key to new key, log a warning) for at least one release — do not hard-break existing configs.
- **Go style**: Explicit `yaml`/`json` struct tags, clear separation of config types, extracted helper functions/maps preferred over large switch statements.
- **CSS animations**: Prefer `transition` + toggled classes over `@keyframes` for anything that can be interrupted mid-animation (keyframes always reset to `from`, causing visual snapping).
- **Branch flow**: `feature/*` branches → `task/release` → `main` → `release`.
- **CI/CD**: Keep workflows aligned with the existing release workflow (same secrets, same multi-platform build targets — `linux/amd64`, `linux/arm64` — same OCI labels).

## Testing and linting

Kiosk uses [task](https://taskfile.dev/) for running tasks.

- `task test` to run tests
- `task lint` to run linters
- `task fmt` to run formatters

## Before submitting

- [ ] I (a human) have read and understood every line of this diff.
- [ ] I have tested the change locally.
- [ ] The PR description reflects what was actually changed and why, not a generic AI summary.
- [ ] Any config schema changes include a backward-compatibility shim.

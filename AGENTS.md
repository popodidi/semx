# semx

Semantic executor for assertions powered by coding agents.

<!-- BEGIN HARVIS SHARED RUNTIME PRESET -->
## Shared Runtime Preset

- Detect the runtime from `HARVIS_RUNTIME`; unset or empty means `local`.
- Supported runtimes are `local` and `remote-codex`. A bot must reject `harvis-agent`, `amux-tn`, and unknown values unless its own prompt explicitly adds support.
- In `local`, treat the session as interactive and manual. Do not create external resources, send messages, push branches, or mutate remote systems unless the owner asks or the active workflow explicitly authorizes the action.
- In `remote-codex`, use the same behavioral defaults as `local`. The sandbox may fail with `bwrap: Failed to make / slave: Permission denied`; retry the necessary command with approved escalation. Escalated commands may ignore `workdir`, so anchor commands to absolute repository paths, for example `git -C /abs/repo ...` and `uv --directory /abs/repo run ...`.
- For authenticated data access, treat the first IAM or permission error as an authentication-context diagnostic. Re-run the documented credential activation, project selection, and one bounded read-only smoke query before using a fallback source; report the exact failing step if it still fails.
- Canonical workspace and `_repo/` dependency clones are read-only. Fetch and fast-forward clean canonical checkouts before analysis; preserve dirty work. Make changes in worktrees and deliver them through PRs. Same-repository worktrees live under the workspace parent’s `.worktrees/<workspace>/<context>/`; dependency worktrees live under `contexts/<context>/_repo/<name>/`. Never force-push or push directly to a default branch.
- Use timestamps in `YYYYMMDD` form and commit messages in `<type>: <summary>` form.
- Never commit secrets, credentials, telemetry endpoints, machine-specific configuration, or agent-internal scratch artifacts.
- Before completion, inspect the affected repository's CI configuration and run the exact matching test command. If no CI exists, run the documented top-level test command.
<!-- END HARVIS SHARED RUNTIME PRESET -->

## Workspace

Read relevant docs and repository-local skills before acting. This repository owns
its application, project knowledge, and personal context. Shared infrastructure
and shared skill sources remain in Harvis; the pinned local snapshot works without
a Harvis checkout. Use `.agents/skills/codebase` for repository operations.

This is a public repository. Keep workspace instructions public-safe; do not retain private operational context, credentials, or raw agent transcripts.

## Development

Use a sibling `.worktrees/semx/<context>/` checkout for changes to this repository.
Keep external canonical clones in `_repo/` and their changes in context worktrees.
Run the exact commands from `.github/workflows/ci.yml` before opening a PR. Never
merge, deploy, trade, send messages, or run live E2E without the corresponding owner
authorization. Runtime support is local and remote-codex; fleet execution and
Amux-TN integration are unsupported.

Maintain adopted knowledge in `docs/`, completed investigations in `research/`,
and curated lessons in `memory/`. Temporary task material belongs in ignored
`contexts/`; do not commit raw logs, secrets, or machine-specific configuration.

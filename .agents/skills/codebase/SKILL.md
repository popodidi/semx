---
name: codebase
description: "Manage repositories in Harvis projects and standalone bot workspaces. Use for clone, sync, status, context worktrees, dependency setup, analysis, and PRs. Subcommands: add, remove, list, clone, sync, status, worktree, worktree-remove, analyze, pr, sync-deps."
allowed-tools:
  - Read
  - Write
  - Edit
  - Glob
  - Grep
  - AskUserQuestion
  - "Bash(git *)"
  - "Bash(mkdir *)"
  - "Bash(ln *)"
  - "Bash(cp *)"
  - "Bash(gh *)"
  - "Bash(ls *)"
---
# Codebase

Manage the current workspace repository and its explicitly registered external dependencies.
A workspace may contain applications, personal context, or both. Its own code, docs,
and instructions belong in the same PR.

## Resolve Before Operating

Read the target repository's AGENTS.md. Run the repository-local resolver:

```bash
uv run .agents/skills/codebase/scripts/registry.py --command worktree --context <context>
uv run .agents/skills/codebase/scripts/registry.py --command worktree --name <dependency> --context <context>
```

From Harvis use `projects/self/skills/codebase/scripts/registry.py`.
`--project-root <directory>` overrides the starting directory. Root detection uses
Git and root AGENTS.md; the nearest dependency registry and Harvis's self registry
remain supported. Never discover registries outside the current Git repository.

- No target or `.` means the current repository, including when it is a linked worktree.
- A named target must be declared in the current workspace's optional `repo.yaml`.
- In Harvis `projects/self`, active entries in `bot.yaml` are also available.
- `list` includes the current repository and declared targets. `--all` explicitly
  selects declared external repositories for inventory or sequential batch operations.
- Archived entries are inventory-only. Unknown targets fail; never guess URLs.
- `clone` requires a named external target or `--all`; do not clone a workspace into itself.

## Paths and Ownership

The resolver returns `path`, `canonical_path`, `url`, and `default_branch`.
With `--context`, it also returns `worktree_path`.

- Registered workspace clones are siblings of Harvis. Their parent is derived
  from the canonical Git directory; `HARVIS_WORKSPACE_PARENT` may override it locally.
- Current-repository and registered-workspace worktrees use
  `<workspace-parent>/.worktrees/<workspace>/<context>/`.
- External clones use `<workspace>/_repo/<name>/`; their worktrees use
  `<workspace>/contexts/<context>/_repo/<name>/`.
- `repo.yaml` lists external dependencies only. Omit it when there are none.
- Canonical checkouts are read-only for implementation. Preserve dirty files,
  untracked tasks, active sessions, and existing worktree registrations.

## Workflow

1. Resolve the target and inspect its status and worktree registrations.
2. Fetch origin; fast-forward a clean canonical default branch. Never reset,
   stash, or overwrite user work. If dirty, fetch and base an isolated worktree
   on the fetched default-branch revision instead.
3. Create a `codex/<context>` branch at `origin/<default_branch>` in the resolved
   worktree path before editing. Existing branches/worktrees resume only after
   checking they belong to the intended task.
4. Read instructions and CI from that worktree. Make minimal changes, run the
   exact CI checks, and commit meaningful changes using `<type>: <summary>`.
5. Push the branch and create a PR in the target repository. Never push directly
   to a default branch or force-push. Use a file for multiline PR descriptions.
6. Merge only on explicit owner approval, with `gh pr merge --repo <owner>/<repo>`.
   If merging errors, inspect remote merge state before retrying.

All commands must be anchored to absolute paths under remote-codex.

## Commands

| Command | Behavior |
|---|---|
| `add/remove <name>` | Edit the current workspace's external dependency registry |
| `list` | Inventory current and declared repositories |
| `clone <name>` / `clone --all` | Clone missing declared repositories at resolved paths |
| `sync [target]` | Fetch and fast-forward the clean canonical checkout |
| `status [target]` | Inspect checkout, branch, and worktrees |
| `worktree [target]` | Create or resume the resolved context worktree |
| `worktree-remove [target]` | Remove a clean, inactive worktree; never delete dirty work |
| `sync-deps [target]` | Apply the target's declared local dependency setup |
| `analyze [target]` | Investigate current code and report file:line evidence |
| `pr [target]` | Commit, push, and open/update the target repository's PR |

External entries retain the existing URL, default_branch, description, key_paths,
and deps fields. `deps.symlink`, `deps.copy`, and `deps.install` are optional;
never commit local credentials or copied environment files.

Delegation is optional for independent substantial work. One writer per worktree;
include its absolute root and require repository-local instructions. Read-only
explorers use refreshed checkouts. Never let a delegated writer modify a canonical clone.

See [implementation](references/implementation.md) for command mechanics.

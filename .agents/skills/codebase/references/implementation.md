# Codebase command mechanics

Resolve the target through `scripts/registry.py` before every operation. Omitted
names and `.` select the current repository. `--all` selects declared dependencies
only, and batch operations run sequentially with failures reported per target.

## Worktree

1. Read `canonical_path`, `url`, `default_branch`, and `worktree_path` from the
   resolver using `--command worktree --context <context>` and an optional name.
2. Inspect `git -C <canonical> status --porcelain` and `git worktree list`.
3. Fetch origin. Fast-forward only a clean canonical default branch; otherwise
   preserve local work and use the fetched revision directly.
4. Run `git -C <canonical> worktree add -b codex/<context> <worktree_path>
   origin/<default_branch>`. Never nest a same-repository worktree in itself.
5. Resume an existing worktree only after checking its branch, task, and status.

## Clone and dependencies

A current workspace cannot clone itself. Clone a named dependency at its resolved
canonical path; an existing directory must be verified before reuse. Optional
`repo.yaml` dependency setup retains symlink/copy/install behavior. Keep env files
and installed dependencies ignored. Harvis's bot registry entries use sibling
canonical directories, not downstream `_repo` directories.

## PR and removal

Use the resolved target's own repository for push and PR creation. Use a body file
for multiline descriptions. Merging needs explicit owner approval; inspect remote
merge state after errors before retrying. Remove worktrees through `git worktree
remove` only after checking dirty state and active sessions. Never recursively
remove old canonical checkout directories during a migration.

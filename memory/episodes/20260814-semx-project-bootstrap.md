---
title: "Semx project and Go CLI bootstrap"
date: "2026-08-14"
runtime: "remote-codex"
project: "semx"
tags: ["semx", "go", "repo-bootstrap", "github-actions"]
status: "complete"
---

# Semx Project and Go CLI Bootstrap

## Summary
- Created the active Harvis `semx` project for `popodidi/semx`, using the positioning “semantic assertions powered by coding agents” and “`assert()` for things that need an agent.”
- Bootstrapped the empty Semx repository with a deliberately minimal Go CLI skeleton derived from the established TikTape and Amux layouts.
- Published `f913c2e9e3204501e74cd5b0af8760edca3e5b17` as the parentless first commit on `main`; the push-triggered GitHub CI run passed formatting, build, tests, vet, and lint.

## Durable Learnings
- Semx now has `main` as its default branch and should use the normal `/codebase worktree semx` plus PR flow for all future changes.
- The initial repository intentionally contains only the `cmd/semx` CLI, `internal/version` build metadata, tests, README, Makefile, pinned `golangci-lint` policy, and GitHub CI. The CLI exposes help and version behavior; assertion syntax and coding-agent integrations remain intentionally undefined for later work.
- When the explicit requirement is that the feature skeleton itself be the first commit, the normal empty-repository PR bootstrap would violate that requirement. Prepare and fully validate a parentless commit locally, obtain explicit approval for the one-time direct `main` push, then verify default-branch state, root ancestry, remote commit count, and push-triggered CI.
- The implementation session detected `HARVIS_RUNTIME=remote-codex`. During distillation on 2026-08-17 the variable was unset despite the shell still exhibiting the remote sandbox mount failure pattern; treat that as an environment-configuration mismatch rather than changing the recorded session runtime.

## Memory Updates
- `memory/episodes/20260814-semx-project-bootstrap.md`
- `memory/knowledge/projects.md`
- `memory/knowledge/patterns.md`
- `memory/index.md`

## Candidates
- Docs: none
- Skills: none; the existing project and codebase workflows cover future work.
- Prompt/preset: none

## Provenance
- Harvis project PR: `https://github.com/popodidi/harvis/pull/2231`; squash merge `349f43fe4379614e0972d3de6e37f2a66df3e974`.
- Semx root commit: `https://github.com/popodidi/semx/commit/f913c2e9e3204501e74cd5b0af8760edca3e5b17`.
- Passing Semx CI: `https://github.com/popodidi/semx/actions/runs/31784349996`.
- Harvis sources: `projects/semx/AGENTS.md` and `projects/semx/repo.yaml`.
- Reference scaffold commits inspected: Amux `413a430625712c97e4bd5d1b214d6422d21156cb` and TikTape `bc508b17e0557831f0e90fecf4594a4b48dfdec8`.

## Excluded
- Raw chat and command logs, the full browser-extracted shared conversation, generated binaries under `bin/`, credentials, and unrelated untracked workspace files.

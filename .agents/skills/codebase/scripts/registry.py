#!/usr/bin/env -S uv run --script
# /// script
# dependencies = ["pyyaml>=6,<7"]
# ///
"""Resolve effective codebase entries for Harvis projects and standalone bots."""

from __future__ import annotations

import argparse
import json
import os
import subprocess
import re
import sys
from pathlib import Path
from typing import Any

import yaml

NAME_RE = re.compile(r"^[a-z0-9]+(?:-[a-z0-9]+)*$")
COMMANDS = {
    "list",
    "clone",
    "sync",
    "status",
    "worktree",
    "worktree-remove",
    "sync-deps",
    "analyze",
    "pr",
}


class UniqueLoader(yaml.SafeLoader):
    pass


def unique_mapping(loader: UniqueLoader, node: yaml.MappingNode, deep: bool = False) -> dict:
    result: dict[Any, Any] = {}
    for key_node, value_node in node.value:
        key = loader.construct_object(key_node, deep=deep)
        if key in result:
            raise ValueError(f"duplicate YAML key: {key}")
        result[key] = loader.construct_object(value_node, deep=deep)
    return result


UniqueLoader.add_constructor(
    yaml.resolver.BaseResolver.DEFAULT_MAPPING_TAG, unique_mapping
)


def load(path: Path) -> dict[str, Any]:
    if not path.exists():
        return {}
    data = yaml.load(path.read_text(), Loader=UniqueLoader)
    if not isinstance(data, dict):
        raise ValueError(f"{path} must contain a mapping")
    return data


def is_self_project(root: Path) -> bool:
    return root.name == "self" and root.parent.name == "projects"


def effective_repos(root: Path, include_archived: bool = False) -> dict[str, Any]:
    repos = load(root / "repo.yaml").get("repos", {})
    if repos is None:
        repos = {}
    if not isinstance(repos, dict):
        raise ValueError("repo.yaml repos must be a mapping")
    result = dict(repos)
    if not is_self_project(root):
        return result
    registry = load(root / "bot.yaml")
    if registry and registry.get("version") != 1:
        raise ValueError("bot.yaml version must be 1")
    bots = registry.get("bots", {})
    if not isinstance(bots, dict):
        raise ValueError("bot.yaml bots must be a mapping")
    conflicts = set(result) & set(bots)
    if conflicts:
        raise ValueError(f"duplicate names across repo.yaml and bot.yaml: {sorted(conflicts)}")
    for name, bot in bots.items():
        if not NAME_RE.fullmatch(name) or not isinstance(bot, dict):
            raise ValueError(f"invalid bot entry: {name}")
        if include_archived or bot.get("status") == "active":
            result[name] = {
                "url": bot.get("url"),
                "default_branch": bot.get("default_branch"),
                "description": bot.get("description"),
                "_source": "bot.yaml",
                "_status": bot.get("status"),
            }
    return result


def git(root: Path, *args: str) -> str:
    return subprocess.check_output(["git", "-C", str(root), *args], text=True, stderr=subprocess.PIPE).strip()


def canonical_root(root: Path) -> Path:
    common = Path(git(root, "rev-parse", "--path-format=absolute", "--git-common-dir"))
    return common.parent


def workspace_parent(root: Path) -> Path:
    configured = os.environ.get("HARVIS_WORKSPACE_PARENT")
    return Path(configured).expanduser().resolve() if configured else canonical_root(root).parent


def find_workspace_root(start: Path) -> Path:
    start = start.resolve()
    top = Path(git(start, "rev-parse", "--show-toplevel"))
    for candidate in (start, *start.parents):
        if is_self_project(candidate) and (candidate / "bot.yaml").is_file():
            return candidate
        if (candidate / "repo.yaml").is_file():
            return candidate
        if candidate == top:
            break
    if not (top / "AGENTS.md").is_file():
        raise ValueError(f"workspace needs AGENTS.md: {top}")
    return top


def resolve_target(root: Path, name: str | None = None) -> dict[str, Any]:
    if name in (None, "."):
        top = Path(git(root, "rev-parse", "--show-toplevel"))
        try:
            branch = git(root, "symbolic-ref", "--short", "refs/remotes/origin/HEAD").removeprefix("origin/")
        except subprocess.CalledProcessError:
            branch = "main"
        return {"url": git(root, "remote", "get-url", "origin"), "default_branch": branch,
                "path": str(top), "canonical_path": str(canonical_root(root)), "_source": "current"}
    repos = effective_repos(root)
    if name not in repos:
        raise ValueError(f"repository not found: {name}")
    if not NAME_RE.fullmatch(name):
        raise ValueError(f"invalid repository name: {name}")
    entry = dict(repos[name])
    path = workspace_parent(root) / name if entry.get("_source") == "bot.yaml" else root / "_repo" / name
    entry.update(path=str(path), canonical_path=str(path))
    return entry


def worktree_path(root: Path, name: str | None, context: str) -> Path:
    if not re.fullmatch(r"[A-Za-z0-9][A-Za-z0-9._-]*", context):
        raise ValueError("context must be a single safe directory name")
    entry = resolve_target(root, name)
    if name in (None, ".") or entry.get("_source") == "bot.yaml":
        return workspace_parent(root) / ".worktrees" / Path(entry["canonical_path"]).name / context
    return root / "contexts" / context / "_repo" / name


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--project-root", type=Path, default=Path.cwd())
    parser.add_argument("--include-archived", action="store_true")
    parser.add_argument("--command", choices=sorted(COMMANDS), default="list")
    targets = parser.add_mutually_exclusive_group()
    targets.add_argument("--name")
    targets.add_argument("--all", action="store_true", help="select declared external repositories")
    parser.add_argument("--context", help="also resolve the context worktree path")
    args = parser.parse_args()
    try:
        root = find_workspace_root(args.project_root)
        if args.include_archived and (args.command not in {"list", "status"} or args.name or args.context):
            raise ValueError("archived entries are inventory only")
        if args.all or (args.command == "list" and not args.name):
            repos = effective_repos(root, args.include_archived)
            output = repos if args.include_archived else {name: resolve_target(root, name) for name in repos}
            if not args.all:
                output = {".": resolve_target(root), **output}
        else:
            output = resolve_target(root, args.name)
            if args.command == "clone" and args.name in (None, "."):
                raise ValueError("current workspace already exists; name a dependency or use --all")
            if args.context:
                output["worktree_path"] = str(worktree_path(root, args.name, args.context))
        print(json.dumps(output, indent=2, sort_keys=True))
    except (ValueError, OSError, subprocess.CalledProcessError, yaml.YAMLError) as exc:
        print(f"ERROR: {exc}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

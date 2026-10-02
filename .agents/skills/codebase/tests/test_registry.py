from __future__ import annotations

import importlib.util
import tempfile
import subprocess
import os
from unittest.mock import patch
import unittest
from pathlib import Path

SOURCE = Path(__file__).parents[1] / "scripts/registry.py"
SPEC = importlib.util.spec_from_file_location("registry_module", SOURCE)
registry = importlib.util.module_from_spec(SPEC)
assert SPEC.loader
SPEC.loader.exec_module(registry)


class RegistryTest(unittest.TestCase):
    def setUp(self) -> None:
        self.temp = tempfile.TemporaryDirectory()
        self.root = Path(self.temp.name)

    def tearDown(self) -> None:
        self.temp.cleanup()

    def test_self_exposes_active_bot_to_complete_command_surface(self) -> None:
        project = self.root / "projects/self"
        project.mkdir(parents=True)
        (project / "bot.yaml").write_text(
            "version: 1\nbots:\n"
            "  tn-eng:\n"
            "    url: git@github.com:popodidi/tn-eng.git\n"
            "    default_branch: main\n"
            "    description: Engineering bot\n"
            "    status: active\n"
        )
        for command in registry.COMMANDS:
            repos = registry.effective_repos(project)
            self.assertIn("tn-eng", repos, command)
            self.assertEqual("bot.yaml", repos["tn-eng"]["_source"])

    def test_self_ignores_archived_bot(self) -> None:
        project = self.root / "projects/self"
        project.mkdir(parents=True)
        (project / "bot.yaml").write_text(
            "version: 1\nbots:\n"
            "  old-bot:\n"
            "    url: git@github.com:popodidi/old-bot.git\n"
            "    default_branch: main\n"
            "    description: Old\n"
            "    status: archived\n"
        )
        self.assertNotIn("old-bot", registry.effective_repos(project))
        self.assertIn("old-bot", registry.effective_repos(project, include_archived=True))

    def test_duplicate_across_registries_fails(self) -> None:
        project = self.root / "projects/self"
        project.mkdir(parents=True)
        (project / "repo.yaml").write_text("repos:\n  tn-eng: {}\n")
        (project / "bot.yaml").write_text(
            "version: 1\nbots:\n  tn-eng:\n    status: active\n"
        )
        with self.assertRaisesRegex(ValueError, "duplicate names"):
            registry.effective_repos(project)

    def test_other_harvis_project_never_loads_bot_registry(self) -> None:
        project = self.root / "projects/truenorth"
        project.mkdir(parents=True)
        (project / "repo.yaml").write_text("repos:\n  service: {}\n")
        (project / "bot.yaml").write_text(
            "version: 1\nbots:\n  leaked:\n    status: active\n"
        )
        self.assertEqual({"service": {}}, registry.effective_repos(project))

    def test_standalone_bot_uses_root_repo_yaml(self) -> None:
        standalone = self.root / "tn-eng"
        standalone.mkdir()
        (standalone / "repo.yaml").write_text("repos:\n  service: {}\n")
        self.assertEqual({"service": {}}, registry.effective_repos(standalone))


class WorkspaceResolutionTest(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.parent = Path(self.temp.name)
        self.root = self.parent / "application"
        self.root.mkdir()
        subprocess.run(["git", "init", "-q", "-b", "main", str(self.root)], check=True)
        subprocess.run(["git", "-C", str(self.root), "remote", "add", "origin", "git@github.com:test/application.git"], check=True)
        (self.root / "AGENTS.md").write_text("Application workspace")
        subprocess.run(["git", "-C", str(self.root), "add", "."], check=True)
        subprocess.run(["git", "-C", str(self.root), "-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "-qm", "init"], check=True)

    def tearDown(self):
        self.temp.cleanup()

    def test_application_without_dependency_registry(self):
        child = self.root / "src"
        child.mkdir()
        self.assertEqual(registry.find_workspace_root(child), self.root)
        self.assertEqual(registry.resolve_target(self.root), registry.resolve_target(self.root, "."))
        self.assertEqual(registry.effective_repos(self.root), {})
        self.assertEqual(registry.worktree_path(self.root, None, "task-1"), self.parent / ".worktrees/application/task-1")

    def test_linked_worktree_keeps_canonical_parent(self):
        target = self.parent / ".worktrees/application/task-1"
        subprocess.run(["git", "-C", str(self.root), "worktree", "add", "-qb", "task-1", str(target)], check=True)
        self.assertEqual(registry.canonical_root(target), self.root)
        self.assertEqual(registry.worktree_path(target, ".", "task-2"), self.parent / ".worktrees/application/task-2")
        self.assertEqual(registry.resolve_target(target)["path"], str(target))

    def test_dependency_and_harvis_registry_paths(self):
        (self.root / "repo.yaml").write_text("repos:\n  service:\n    url: git@github.com:test/service.git\n")
        self.assertEqual(registry.resolve_target(self.root, "service")["path"], str(self.root / "_repo/service"))
        self.assertEqual(registry.worktree_path(self.root, "service", "task"), self.root / "contexts/task/_repo/service")
        project = self.root / "projects/self"
        project.mkdir(parents=True)
        (project / "bot.yaml").write_text("version: 1\nbots:\n  desk:\n    status: active\n    url: git@github.com:test/desk.git\n")
        self.assertEqual(registry.find_workspace_root(project), project)
        self.assertEqual(registry.resolve_target(project, "desk")["path"], str(self.parent / "desk"))
        with patch.dict(os.environ, {"HARVIS_WORKSPACE_PARENT": str(self.parent / "configured")}):
            self.assertEqual(registry.resolve_target(project, "desk")["path"], str(self.parent / "configured/desk"))

    def test_bad_context_and_unknown_dependency_fail(self):
        with self.assertRaises(ValueError):
            registry.worktree_path(self.root, ".", "../other")
        with self.assertRaises(ValueError):
            registry.resolve_target(self.root, "missing")

    def test_cli_default_dot_and_explicit_all(self):
        import json
        result = subprocess.check_output(["python", str(SOURCE), "--project-root", str(self.root), "--command", "worktree", "--context", "task"], text=True)
        self.assertEqual(json.loads(result)["path"], str(self.root))
        result = subprocess.check_output(["python", str(SOURCE), "--project-root", str(self.root), "--all"], text=True)
        self.assertEqual(json.loads(result), {})


if __name__ == "__main__":
    unittest.main()

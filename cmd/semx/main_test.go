package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunHelp(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if err := run(context.Background(), []string{"--help"}, &stdout, &stderr); err != nil {
		t.Fatalf("run(--help) error = %v", err)
	}
	if !strings.Contains(stdout.String(), "run         execute") || !strings.Contains(stdout.String(), "validate    validate") {
		t.Fatalf("help output = %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunWithoutCommandShowsUsage(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := run(context.Background(), nil, &stdout, &stderr)
	if !errors.Is(err, errUsage) {
		t.Fatalf("run() error = %v, want errUsage", err)
	}
	if !strings.Contains(stderr.String(), "usage: semx <command>") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestValidatePrintsValid(t *testing.T) {
	configPath := writeFile(t, "semx.yaml", `runner:
  type: codex
prompt:
  user: Is it factual?
output:
  format: text
`, 0o644)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if err := run(context.Background(), []string{"validate", configPath}, &stdout, &stderr); err != nil {
		t.Fatalf("run(validate) error = %v", err)
	}
	if stdout.String() != "valid\n" {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestValidateCollectsErrors(t *testing.T) {
	configPath := writeFile(t, "semx.yaml", "{}\n", 0o644)
	err := run(context.Background(), []string{"validate", configPath}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("run(validate) error = nil")
	}
	for _, want := range []string{"runner.type is required", "prompt.user is required", "output.format is required"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error = %q, missing %q", err, want)
		}
	}
}

func TestRunCommandEndToEndWithAliasesAndOverrides(t *testing.T) {
	script := writeFile(t, "fake-runner", "#!/bin/sh\nprintf '{\"pass\":true,\"reason\":\"factual\"}\\n'\n", 0o755)
	schema := writeFile(t, "schema.json", `{
  "type": "object",
  "required": ["pass", "reason"],
  "properties": {
    "pass": {"type": "boolean"},
    "reason": {"type": "string"}
  },
  "additionalProperties": false
}`, 0o644)
	configPath := writeFile(t, "semx.yaml", `runner:
  type: codex
prompt:
  user: Is it factual?
output:
  format: text
`, 0o644)
	corpus := t.TempDir()
	output := filepath.Join(t.TempDir(), "output")
	args := []string{
		"run", configPath,
		"--corpus", corpus,
		"--out", output,
		"--output.format=json",
		"--output.schema", schema,
		"--runner.command", script,
		"--runner.args.model=test",
	}
	if err := run(context.Background(), args, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
		t.Fatalf("run(run) error = %v", err)
	}
	for _, name := range []string{"result.json", "semx.resolved.yaml", "runner.stdout", "runner.stderr"} {
		if _, err := os.Stat(filepath.Join(output, name)); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
	}
}

func writeFile(t *testing.T, name, contents string, mode os.FileMode) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(contents), mode); err != nil {
		t.Fatal(err)
	}
	return path
}

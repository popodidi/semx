package execution

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/popodidi/semx/internal/config"
)

func TestRunWritesValidatedArtifacts(t *testing.T) {
	script := executable(t, `#!/bin/sh
printf '{"pass":true,"reason":"factual"}\n'
printf 'runner debug\n' >&2
`)
	cfg := validConfig(t, script)
	if err := Run(context.Background(), cfg); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	assertFileContains(t, filepath.Join(cfg.Output.Dir, "result.json"), `{"pass":true,"reason":"factual"}`)
	assertFileContains(t, filepath.Join(cfg.Output.Dir, "semx.resolved.yaml"), "type: codex")
	assertFileContains(t, filepath.Join(cfg.Output.Dir, "runner.stdout"), `{"pass":true`)
	assertFileContains(t, filepath.Join(cfg.Output.Dir, "runner.stderr"), "runner debug")
}

func TestRunRejectsInvalidSemanticResults(t *testing.T) {
	cases := []struct {
		name   string
		output string
		want   string
	}{
		{"malformed JSON", "not-json", "runner output is not valid JSON"},
		{"schema mismatch", `{"pass":"yes","reason":"wrong type"}`, "runner output does not match output schema"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			script := executable(t, "#!/bin/sh\nprintf '%s\\n' '"+tc.output+"'\n")
			cfg := validConfig(t, script)
			err := Run(context.Background(), cfg)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("Run() error = %v, want %q", err, tc.want)
			}
			if _, statErr := os.Stat(filepath.Join(cfg.Output.Dir, "result.json")); !os.IsNotExist(statErr) {
				t.Fatalf("result.json stat error = %v, want not exist", statErr)
			}
		})
	}
}

func TestRunCapturesDiagnosticsWhenChildFails(t *testing.T) {
	script := executable(t, "#!/bin/sh\nprintf 'backend failed\\n' >&2\nexit 9\n")
	cfg := validConfig(t, script)
	err := Run(context.Background(), cfg)
	if err == nil || !strings.Contains(err.Error(), "exit status 9") {
		t.Fatalf("Run() error = %v", err)
	}
	assertFileContains(t, filepath.Join(cfg.Output.Dir, "runner.stderr"), "backend failed")
	assertFileContains(t, filepath.Join(cfg.Output.Dir, "semx.resolved.yaml"), "type: codex")
}

func validConfig(t *testing.T, command string) config.Config {
	t.Helper()
	schema := filepath.Join(t.TempDir(), "schema.json")
	if err := os.WriteFile(schema, []byte(`{
  "type": "object",
  "required": ["pass", "reason"],
  "properties": {
    "pass": {"type": "boolean"},
    "reason": {"type": "string"}
  },
  "additionalProperties": false
}`), 0o644); err != nil {
		t.Fatal(err)
	}
	return config.Config{
		Runner: config.RunnerConfig{Type: "codex", Command: command, Args: map[string]any{"model": "test"}},
		Corpus: config.CorpusConfig{Path: t.TempDir()},
		Prompt: config.PromptConfig{System: "You are a judge.", User: "Is it factual?"},
		Output: config.OutputConfig{Dir: filepath.Join(t.TempDir(), "output"), Format: "json", Schema: schema},
	}
}

func executable(t *testing.T, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "fake-runner")
	if err := os.WriteFile(path, []byte(contents), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

func assertFileContains(t *testing.T, path, want string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), want) {
		t.Fatalf("%s = %q, want substring %q", path, data, want)
	}
}

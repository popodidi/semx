package runner

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/popodidi/semx/internal/config"
	"github.com/popodidi/semx/internal/request"
)

func TestSerializeArgs(t *testing.T) {
	got, err := SerializeArgs(map[string]any{
		"model":     "gpt-5.6",
		"full-auto": true,
		"disabled":  false,
		"retries":   3,
		"add-dir":   []any{"./foo", "./bar"},
	})
	if err != nil {
		t.Fatalf("SerializeArgs() error = %v", err)
	}
	want := []string{
		"--add-dir", "./foo", "--add-dir", "./bar",
		"--full-auto",
		"--model", "gpt-5.6",
		"--retries", "3",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("SerializeArgs() = %#v, want %#v", got, want)
	}
}

func TestCodexRunnerLowersRequestAndCapturesStreams(t *testing.T) {
	argsFile := filepath.Join(t.TempDir(), "args")
	t.Setenv("SEMX_ARGS_FILE", argsFile)
	script := executable(t, `#!/bin/sh
printf '%s\n' "$@" > "$SEMX_ARGS_FILE"
printf 'result output\n'
printf 'debug output\n' >&2
`)
	backend, err := New(config.RunnerConfig{
		Type:    "codex",
		Command: script,
		Args:    map[string]any{"model": "gpt-5.6"},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	req := request.RunRequest{
		Corpus: request.Corpus{Path: "/corpus"},
		Prompt: request.Prompt{System: "judge", User: "is it factual?"},
		Output: request.OutputContract{Format: "json", Schema: "/schema.json"},
	}
	result, err := backend.Run(context.Background(), req)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.Output != "result output\n" || result.Stdout != result.Output || result.Stderr != "debug output\n" {
		t.Fatalf("Run() result = %#v", result)
	}
	data, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatal(err)
	}
	got := string(data)
	for _, want := range []string{
		"exec\n",
		"--model\ngpt-5.6\n",
		"--cd\n/corpus\n",
		"--skip-git-repo-check\n",
		"--ephemeral\n",
		"--color\nnever\n",
		"--output-schema\n/schema.json\n",
		"Corpus directory:",
		"is it factual?",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("argv = %q, missing %q", got, want)
		}
	}
}

func TestClaudeRunnerLowersStructuredRequestAndUnwrapsOutput(t *testing.T) {
	argsFile := filepath.Join(t.TempDir(), "args")
	cwdFile := filepath.Join(t.TempDir(), "cwd")
	t.Setenv("SEMX_ARGS_FILE", argsFile)
	t.Setenv("SEMX_CWD_FILE", cwdFile)
	script := executable(t, `#!/bin/sh
pwd > "$SEMX_CWD_FILE"
printf '%s\n' "$@" > "$SEMX_ARGS_FILE"
printf '{"structured_output":{"pass":true,"reason":"factual"},"result":""}\n'
`)
	corpus := t.TempDir()
	schema := filepath.Join(t.TempDir(), "schema.json")
	schemaData := `{"type":"object","required":["pass","reason"],"properties":{"pass":{"type":"boolean"},"reason":{"type":"string"}}}`
	if err := os.WriteFile(schema, []byte(schemaData), 0o644); err != nil {
		t.Fatal(err)
	}
	backend, err := New(config.RunnerConfig{
		Type:    "claude",
		Command: script,
		Args:    map[string]any{"model": "sonnet", "safe-mode": true},
	})
	if err != nil {
		t.Fatal(err)
	}
	req := request.RunRequest{
		Corpus: request.Corpus{Path: corpus},
		Prompt: request.Prompt{System: "judge", User: "is it factual?"},
		Output: request.OutputContract{Format: "json", Schema: schema},
	}
	result, err := backend.Run(context.Background(), req)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.Output != `{"pass":true,"reason":"factual"}` {
		t.Fatalf("Output = %q", result.Output)
	}
	if !strings.Contains(result.Stdout, `"structured_output"`) {
		t.Fatalf("Stdout = %q", result.Stdout)
	}
	data, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatal(err)
	}
	got := string(data)
	for _, want := range []string{
		"--model\nsonnet\n",
		"--safe-mode\n",
		"--permission-mode\ndontAsk\n",
		"--print\n",
		"--no-session-persistence\n",
		"--output-format\njson\n",
		"--json-schema\n",
		`"required"`,
		corpus,
		"is it factual?",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("argv = %q, missing %q", got, want)
		}
	}
	cwd, err := os.ReadFile(cwdFile)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(cwd)) != corpus {
		t.Fatalf("working directory = %q, want %q", cwd, corpus)
	}
}

func TestOpenCodeRunnerLowersRequestAndExtractsTextEvents(t *testing.T) {
	argsFile := filepath.Join(t.TempDir(), "args")
	t.Setenv("SEMX_ARGS_FILE", argsFile)
	script := executable(t, `#!/bin/sh
printf '%s\n' "$@" > "$SEMX_ARGS_FILE"
printf '%s\n' \
  '{"type":"step_start","part":{"type":"step-start"}}' \
  '{"type":"text","part":{"type":"text","text":"result "}}' \
  '{"type":"text","part":{"type":"text","text":"output\n"}}' \
  '{"type":"step_finish","part":{"type":"step-finish","reason":"stop"}}'
printf 'debug output\n' >&2
`)
	corpus := t.TempDir()
	backend, err := New(config.RunnerConfig{
		Type:    "opencode",
		Command: script,
		Args:    map[string]any{"model": "ollama/test"},
	})
	if err != nil {
		t.Fatal(err)
	}
	result, err := backend.Run(context.Background(), request.RunRequest{
		Corpus: request.Corpus{Path: corpus},
		Prompt: request.Prompt{User: "judge this"},
		Output: request.OutputContract{Format: "text"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Output != "result output\n" || !strings.Contains(result.Stdout, `"step_start"`) || result.Stderr != "debug output\n" {
		t.Fatalf("Run() result = %#v", result)
	}
	data, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatal(err)
	}
	got := string(data)
	for _, want := range []string{
		"run\n",
		"--model\nollama/test\n",
		"--format\njson\n",
		"--dir\n" + corpus + "\n",
		"Corpus directory:\n",
		corpus,
		"judge this",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("argv = %q, missing %q", got, want)
		}
	}
}

func TestOpenCodeRunnerRejectsMissingTextResponse(t *testing.T) {
	script := executable(t, "#!/bin/sh\nprintf '%s\\n' '{\"type\":\"step_finish\",\"part\":{\"type\":\"step-finish\",\"reason\":\"stop\"}}'\n")
	backend, err := New(config.RunnerConfig{Type: "opencode", Command: script, Args: map[string]any{}})
	if err != nil {
		t.Fatal(err)
	}
	_, err = backend.Run(context.Background(), request.RunRequest{Corpus: request.Corpus{Path: t.TempDir()}})
	if err == nil || !strings.Contains(err.Error(), "did not include a text response") {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestNewRejectsUnsupportedPiRunner(t *testing.T) {
	if _, err := New(config.RunnerConfig{Type: "pi", Args: map[string]any{}}); err == nil || !strings.Contains(err.Error(), `unsupported runner type "pi"`) {
		t.Fatalf("New() error = %v", err)
	}
}

func TestCodexRunnerSurfacesChildFailure(t *testing.T) {
	script := executable(t, "#!/bin/sh\nprintf 'failure details\\n' >&2\nexit 7\n")
	backend, err := New(config.RunnerConfig{Type: "codex", Command: script, Args: map[string]any{}})
	if err != nil {
		t.Fatal(err)
	}
	result, err := backend.Run(context.Background(), request.RunRequest{})
	if err == nil || !strings.Contains(err.Error(), "exit status 7") {
		t.Fatalf("Run() error = %v", err)
	}
	if result.Stderr != "failure details\n" {
		t.Fatalf("stderr = %q", result.Stderr)
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

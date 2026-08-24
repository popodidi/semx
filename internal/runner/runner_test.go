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

func TestCommandRunnerLowersRequestAndCapturesStreams(t *testing.T) {
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
	for _, want := range []string{"--model\ngpt-5.6\n", "Corpus directory:", "/corpus", "is it factual?", "/schema.json"} {
		if !strings.Contains(got, want) {
			t.Fatalf("argv = %q, missing %q", got, want)
		}
	}
}

func TestCommandRunnerSurfacesChildFailure(t *testing.T) {
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

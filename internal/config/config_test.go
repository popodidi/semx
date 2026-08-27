package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadStrictYAMLAndDefaults(t *testing.T) {
	t.Run("valid configuration and dynamic runner arguments", func(t *testing.T) {
		path := writeConfig(t, `runner:
  type: codex
  args:
    model: gpt-5.6
    backend-specific: true
prompt:
  user: Is it factual?
output:
  format: text
`)
		cfg, err := Load(path)
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if cfg.Runner.Args["model"] != "gpt-5.6" || cfg.Runner.Args["backend-specific"] != true {
			t.Fatalf("Runner.Args = %#v", cfg.Runner.Args)
		}
	})

	t.Run("malformed YAML", func(t *testing.T) {
		path := writeConfig(t, "runner: [\n")
		if _, err := Load(path); err == nil || !strings.Contains(err.Error(), "decode configuration") {
			t.Fatalf("Load() error = %v", err)
		}
	})

	t.Run("unknown field", func(t *testing.T) {
		path := writeConfig(t, "ouptut:\n  format: text\n")
		if _, err := Load(path); err == nil || !strings.Contains(err.Error(), "field ouptut not found") {
			t.Fatalf("Load() error = %v", err)
		}
	})

	t.Run("defaults initialize runner args", func(t *testing.T) {
		path := writeConfig(t, "{}\n")
		cfg, err := Load(path)
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if cfg.Runner.Args == nil {
			t.Fatal("Runner.Args is nil")
		}
	})
}

func TestApplyOverridesUsesReflectedFieldsAndAliases(t *testing.T) {
	cfg := Defaults()
	args := []string{
		"--runner.type=claude",
		"--runner.command", "/bin/claude",
		"--corpus.path", "/canonical-corpus",
		"--prompt.system=system",
		"--prompt.user", "user",
		"--output.dir=/canonical-output",
		"--output.format=json",
		"--output.schema", "/schema.json",
		"--corpus", "/alias-corpus",
		"--out=/alias-output",
		"--runner.args.model=opus",
		"--runner.args.permission-mode", "bypassPermissions",
		"--runner.args.ignore-user-config=true",
	}
	if err := ApplyOverrides(&cfg, args); err != nil {
		t.Fatalf("ApplyOverrides() error = %v", err)
	}
	if cfg.Runner.Type != "claude" || cfg.Runner.Command != "/bin/claude" {
		t.Fatalf("Runner = %#v", cfg.Runner)
	}
	if cfg.Corpus.Path != "/alias-corpus" || cfg.Output.Dir != "/alias-output" {
		t.Fatalf("aliases did not update canonical fields: %#v %#v", cfg.Corpus, cfg.Output)
	}
	if cfg.Prompt.System != "system" || cfg.Prompt.User != "user" {
		t.Fatalf("Prompt = %#v", cfg.Prompt)
	}
	if cfg.Output.Format != "json" || cfg.Output.Schema != "/schema.json" {
		t.Fatalf("Output = %#v", cfg.Output)
	}
	if cfg.Runner.Args["model"] != "opus" || cfg.Runner.Args["permission-mode"] != "bypassPermissions" {
		t.Fatalf("Runner.Args = %#v", cfg.Runner.Args)
	}
	if cfg.Runner.Args["ignore-user-config"] != true {
		t.Fatalf("Runner.Args = %#v, want typed boolean", cfg.Runner.Args)
	}
}

func TestResolveCLIOverridesYAML(t *testing.T) {
	path := writeConfig(t, `runner:
  type: codex
prompt:
  user: factual?
output:
  format: text
`)
	cfg, err := Resolve(path, []string{"--runner.type=claude", "--output.format=json", "--output.schema=/schema.json"})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if cfg.Runner.Type != "claude" || cfg.Output.Format != "json" || cfg.Output.Schema != "/schema.json" {
		t.Fatalf("Resolve() = %#v", cfg)
	}
}

func TestValidation(t *testing.T) {
	corpus := t.TempDir()
	schema := writeSchema(t, `{"type":"object","required":["pass"],"properties":{"pass":{"type":"boolean"}}}`)
	valid := Config{
		Runner: RunnerConfig{Type: "codex", Args: map[string]any{}},
		Corpus: CorpusConfig{Path: corpus},
		Prompt: PromptConfig{User: "Is it factual?"},
		Output: OutputConfig{Dir: filepath.Join(t.TempDir(), "output"), Format: "json", Schema: schema},
	}

	if errs := ValidateConfig(Config{
		Runner: RunnerConfig{Type: "codex", Args: map[string]any{}},
		Prompt: PromptConfig{User: "Can be supplied a corpus at run time"},
		Output: OutputConfig{Format: "json", Schema: schema},
	}); len(errs) != 0 {
		t.Fatalf("ValidateConfig(static) = %v", errs)
	}
	if errs := ValidateRun(valid); len(errs) != 0 {
		t.Fatalf("ValidateRun(valid) = %v", errs)
	}

	cases := []struct {
		name string
		edit func(*Config)
		want string
	}{
		{"missing runner", func(c *Config) { c.Runner.Type = "" }, "runner.type is required"},
		{"invalid runner", func(c *Config) { c.Runner.Type = "other" }, "runner.type must be one of"},
		{"unsupported pi runner", func(c *Config) { c.Runner.Type = "pi" }, "runner.type must be one of: claude, codex, opencode"},
		{"missing prompt", func(c *Config) { c.Prompt.User = "" }, "prompt.user is required"},
		{"invalid format", func(c *Config) { c.Output.Format = "yaml" }, "output.format must be one of"},
		{"json without schema", func(c *Config) { c.Output.Schema = "" }, "output.schema is required"},
		{"invalid schema", func(c *Config) { c.Output.Schema = writeSchema(t, `{"type":42}`) }, "output.schema:"},
		{"schema for text", func(c *Config) { c.Output.Format = "text" }, "output.schema may only"},
		{"nonexistent corpus", func(c *Config) { c.Corpus.Path = filepath.Join(t.TempDir(), "missing") }, "corpus.path:"},
		{"missing output directory", func(c *Config) { c.Output.Dir = "" }, "output.dir is required"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := valid
			tc.edit(&cfg)
			errs := ValidateRun(cfg)
			if !errorsContain(errs, tc.want) {
				t.Fatalf("ValidateRun() = %v, want substring %q", errs, tc.want)
			}
		})
	}
}

func writeConfig(t *testing.T, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "semx.yaml")
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func writeSchema(t *testing.T, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "schema.json")
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func errorsContain(errs []error, substring string) bool {
	for _, err := range errs {
		if strings.Contains(err.Error(), substring) {
			return true
		}
	}
	return false
}

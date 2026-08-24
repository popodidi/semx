package request

import (
	"reflect"
	"testing"

	"github.com/popodidi/semx/internal/config"
)

func TestFromConfigKeepsRunnerSettingsOutOfSemanticRequest(t *testing.T) {
	cfg := config.Config{
		Runner: config.RunnerConfig{
			Type: "codex",
			Args: map[string]any{"sandbox": "workspace-write"},
		},
		Corpus: config.CorpusConfig{Path: "/corpus"},
		Prompt: config.PromptConfig{System: "judge", User: "factual?"},
		Output: config.OutputConfig{Format: "json", Schema: "/schema.json"},
	}
	want := RunRequest{
		Corpus: Corpus{Path: "/corpus"},
		Prompt: Prompt{System: "judge", User: "factual?"},
		Output: OutputContract{Format: "json", Schema: "/schema.json"},
	}
	if got := FromConfig(cfg); !reflect.DeepEqual(got, want) {
		t.Fatalf("FromConfig() = %#v, want %#v", got, want)
	}
}

package runner

import (
	"fmt"

	"github.com/popodidi/semx/internal/config"
)

func New(cfg config.RunnerConfig) (Runner, error) {
	switch cfg.Type {
	case "codex":
		return &CodexRunner{cfg: cfg}, nil
	case "claude":
		return &ClaudeRunner{cfg: cfg}, nil
	case "opencode", "pi":
		return &CommandRunner{cfg: cfg}, nil
	default:
		return nil, fmt.Errorf("unsupported runner type %q", cfg.Type)
	}
}

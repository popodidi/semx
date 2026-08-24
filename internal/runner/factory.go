package runner

import (
	"fmt"

	"github.com/popodidi/semx/internal/config"
)

func New(cfg config.RunnerConfig) (Runner, error) {
	switch cfg.Type {
	case "codex", "claude", "opencode", "pi":
		return &CommandRunner{cfg: cfg}, nil
	default:
		return nil, fmt.Errorf("unsupported runner type %q", cfg.Type)
	}
}

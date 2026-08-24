package execution

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/popodidi/semx/internal/config"
	"github.com/popodidi/semx/internal/request"
	"github.com/popodidi/semx/internal/runner"
)

func Run(ctx context.Context, cfg config.Config) error {
	backend, err := runner.New(cfg.Runner)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(cfg.Output.Dir, 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}
	resolved, err := config.Marshal(cfg)
	if err != nil {
		return err
	}
	if err := writeArtifact(cfg.Output.Dir, "semx.resolved.yaml", resolved); err != nil {
		return err
	}

	result, runErr := backend.Run(ctx, request.FromConfig(cfg))
	if err := writeArtifact(cfg.Output.Dir, "runner.stdout", []byte(result.Stdout)); err != nil {
		return err
	}
	if err := writeArtifact(cfg.Output.Dir, "runner.stderr", []byte(result.Stderr)); err != nil {
		return err
	}
	if runErr != nil {
		return runErr
	}

	resultData := []byte(result.Output)
	resultName := "result.txt"
	if cfg.Output.Format == "json" {
		var value any
		if err := json.Unmarshal(resultData, &value); err != nil {
			return fmt.Errorf("runner output is not valid JSON: %w", err)
		}
		schema, err := config.CompileSchema(cfg.Output.Schema)
		if err != nil {
			return fmt.Errorf("compile output schema: %w", err)
		}
		if err := schema.Validate(value); err != nil {
			return fmt.Errorf("runner output does not match output schema: %w", err)
		}
		resultName = "result.json"
	}
	if err := writeArtifact(cfg.Output.Dir, resultName, resultData); err != nil {
		return err
	}
	return nil
}

func writeArtifact(dir, name string, data []byte) error {
	if !strings.HasSuffix(string(data), "\n") {
		data = append(data, '\n')
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", name, err)
	}
	return nil
}

package runner

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/popodidi/semx/internal/config"
	"github.com/popodidi/semx/internal/request"
)

type CodexRunner struct {
	cfg config.RunnerConfig
}

func (r *CodexRunner) Run(ctx context.Context, req request.RunRequest) (RunResult, error) {
	command := r.cfg.Command
	if command == "" {
		command = "codex"
	}
	runnerArgs, err := SerializeArgs(r.cfg.Args)
	if err != nil {
		return RunResult{}, err
	}
	corpusPath, err := filepath.Abs(req.Corpus.Path)
	if err != nil {
		return RunResult{}, fmt.Errorf("resolve corpus path: %w", err)
	}
	req.Corpus.Path = corpusPath

	args := append([]string{"exec"}, runnerArgs...)
	args = append(args,
		"--cd", corpusPath,
		"--skip-git-repo-check",
		"--ephemeral",
		"--color", "never",
	)
	if req.Output.Format == "json" {
		schemaPath, err := filepath.Abs(req.Output.Schema)
		if err != nil {
			return RunResult{}, fmt.Errorf("resolve output schema path: %w", err)
		}
		req.Output.Schema = schemaPath
		args = append(args, "--output-schema", schemaPath)
	}
	args = append(args, BuildPrompt(req))
	return runCommand(ctx, command, args)
}

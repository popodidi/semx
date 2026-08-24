package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/popodidi/semx/internal/config"
	"github.com/popodidi/semx/internal/request"
)

type ClaudeRunner struct {
	cfg config.RunnerConfig
}

func (r *ClaudeRunner) Run(ctx context.Context, req request.RunRequest) (RunResult, error) {
	command := r.cfg.Command
	if command == "" {
		command = "claude"
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

	args := runnerArgs
	if _, configured := r.cfg.Args["permission-mode"]; !configured {
		args = append(args, "--permission-mode", "dontAsk")
	}
	args = append(args,
		"--print",
		"--no-session-persistence",
	)
	if req.Output.Format == "json" {
		schemaPath, err := filepath.Abs(req.Output.Schema)
		if err != nil {
			return RunResult{}, fmt.Errorf("resolve output schema path: %w", err)
		}
		schema, err := os.ReadFile(schemaPath)
		if err != nil {
			return RunResult{}, fmt.Errorf("read output schema: %w", err)
		}
		req.Output.Schema = schemaPath
		args = append(args,
			"--output-format", "json",
			"--json-schema", string(schema),
		)
	} else {
		args = append(args, "--output-format", "text")
	}
	args = append(args, BuildPrompt(req))

	result, err := runCommandInDir(ctx, command, args, corpusPath)
	if err != nil || req.Output.Format != "json" {
		return result, err
	}
	var envelope struct {
		StructuredOutput json.RawMessage `json:"structured_output"`
	}
	if err := json.Unmarshal([]byte(result.Stdout), &envelope); err != nil {
		return result, fmt.Errorf("decode Claude runner output: %w", err)
	}
	if len(envelope.StructuredOutput) == 0 || string(envelope.StructuredOutput) == "null" {
		return result, fmt.Errorf("claude runner output did not include structured_output")
	}
	result.Output = string(envelope.StructuredOutput)
	return result, nil
}

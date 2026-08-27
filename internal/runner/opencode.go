package runner

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/popodidi/semx/internal/config"
	"github.com/popodidi/semx/internal/request"
)

type OpenCodeRunner struct {
	cfg config.RunnerConfig
}

func (r *OpenCodeRunner) Run(ctx context.Context, req request.RunRequest) (RunResult, error) {
	command := r.cfg.Command
	if command == "" {
		command = "opencode"
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

	args := append([]string{"run"}, runnerArgs...)
	args = append(args,
		"--format", "json",
		"--dir", corpusPath,
		BuildPrompt(req),
	)
	result, err := runCommand(ctx, command, args)
	if err != nil {
		return result, err
	}
	result.Output, err = decodeOpenCodeOutput(result.Stdout)
	if err != nil {
		return result, err
	}
	return result, nil
}

func decodeOpenCodeOutput(stdout string) (string, error) {
	decoder := json.NewDecoder(strings.NewReader(stdout))
	var output strings.Builder
	for {
		var event struct {
			Type string `json:"type"`
			Part struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"part"`
		}
		if err := decoder.Decode(&event); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return "", fmt.Errorf("decode OpenCode runner output: %w", err)
		}
		if event.Type == "text" && event.Part.Type == "text" {
			output.WriteString(event.Part.Text)
		}
	}
	if output.Len() == 0 {
		return "", errors.New("opencode runner output did not include a text response")
	}
	return output.String(), nil
}

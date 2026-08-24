package runner

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"

	"github.com/popodidi/semx/internal/config"
	"github.com/popodidi/semx/internal/request"
)

type CommandRunner struct {
	cfg config.RunnerConfig
}

func (r *CommandRunner) Run(ctx context.Context, req request.RunRequest) (RunResult, error) {
	command := r.cfg.Command
	if command == "" {
		command = r.cfg.Type
	}
	args, err := SerializeArgs(r.cfg.Args)
	if err != nil {
		return RunResult{}, err
	}
	args = append(args, BuildPrompt(req))
	return runCommand(ctx, command, args)
}

func runCommand(ctx context.Context, command string, args []string) (RunResult, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	result := RunResult{
		Output: stdout.String(),
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}
	if err != nil {
		return result, fmt.Errorf("runner command %q failed: %w", command, err)
	}
	return result, nil
}

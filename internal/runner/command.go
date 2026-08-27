package runner

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
)

func runCommand(ctx context.Context, command string, args []string) (RunResult, error) {
	return runCommandInDir(ctx, command, args, "")
}

func runCommandInDir(ctx context.Context, command string, args []string, dir string) (RunResult, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = dir
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

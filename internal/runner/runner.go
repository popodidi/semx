package runner

import (
	"context"

	"github.com/popodidi/semx/internal/request"
)

type RunResult struct {
	Output string
	Stdout string
	Stderr string
}

type Runner interface {
	Run(context.Context, request.RunRequest) (RunResult, error)
}

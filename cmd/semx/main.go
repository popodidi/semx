package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"

	"github.com/popodidi/semx/internal/config"
	"github.com/popodidi/semx/internal/execution"
)

var errUsage = errors.New("usage requested")

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	if err := run(ctx, os.Args[1:], os.Stdout, os.Stderr); err != nil {
		if !errors.Is(err, errUsage) {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, stdout io.Writer, stderr io.Writer) error {
	if len(args) == 0 {
		usage(stderr)
		return errUsage
	}
	if args[0] == "help" || args[0] == "-h" || args[0] == "--help" {
		usage(stdout)
		return nil
	}
	if args[0] != "run" && args[0] != "validate" {
		usage(stderr)
		return fmt.Errorf("unknown command %q", args[0])
	}
	if len(args) < 2 || strings.HasPrefix(args[1], "-") {
		return fmt.Errorf("usage: semx %s <config.yaml> [flags]", args[0])
	}

	cfg, err := config.Resolve(args[1], args[2:])
	if err != nil {
		return err
	}
	if args[0] == "validate" {
		if errs := config.ValidateConfig(cfg); len(errs) > 0 {
			return validationError(errs)
		}
		fmt.Fprintln(stdout, "valid")
		return nil
	}
	if errs := config.ValidateRun(cfg); len(errs) > 0 {
		return validationError(errs)
	}
	return execution.Run(ctx, cfg)
}

func validationError(errs []error) error {
	var message strings.Builder
	message.WriteString("invalid configuration:\n")
	for _, err := range errs {
		fmt.Fprintf(&message, "\n- %s", err)
	}
	return errors.New(message.String())
}

func usage(w io.Writer) {
	fmt.Fprintln(w, "semx — semantic assertions powered by coding agents")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "usage: semx <command> <config.yaml> [flags]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "commands:")
	fmt.Fprintln(w, "  run         execute a semantic assertion")
	fmt.Fprintln(w, "  validate    validate a semantic assertion configuration")
}

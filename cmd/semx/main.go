package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/popodidi/semx/internal/version"
)

var errUsage = errors.New("usage requested")

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		if !errors.Is(err, errUsage) {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}
}

func run(args []string, stdout io.Writer, stderr io.Writer) error {
	if len(args) == 0 {
		usage(stderr)
		return errUsage
	}

	switch args[0] {
	case "help", "-h", "--help":
		usage(stdout)
		return nil
	case "version", "--version":
		if len(args) != 1 {
			return errors.New("usage: semx version")
		}
		info := version.Get()
		fmt.Fprintf(
			stdout,
			"semx %s\ncommit: %s\nbuild_time: %s\n",
			info.Version,
			info.Commit,
			info.BuildTime,
		)
		return nil
	default:
		usage(stderr)
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func usage(w io.Writer) {
	fmt.Fprintln(w, "semx — semantic assertions powered by coding agents")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "usage: semx <command>")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "commands:")
	fmt.Fprintln(w, "  version    print build information")
}

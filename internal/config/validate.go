package config

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

var runnerTypes = map[string]struct{}{
	"claude":   {},
	"codex":    {},
	"opencode": {},
}

func ValidateConfig(cfg Config) []error {
	var errs []error
	if strings.TrimSpace(cfg.Runner.Type) == "" {
		errs = append(errs, errors.New("runner.type is required"))
	} else if _, ok := runnerTypes[cfg.Runner.Type]; !ok {
		errs = append(errs, fmt.Errorf("runner.type must be one of: %s", strings.Join(SupportedRunnerTypes(), ", ")))
	}
	if strings.TrimSpace(cfg.Prompt.User) == "" {
		errs = append(errs, errors.New("prompt.user is required"))
	}
	switch cfg.Output.Format {
	case "":
		errs = append(errs, errors.New("output.format is required"))
	case "text":
		if cfg.Output.Schema != "" {
			errs = append(errs, errors.New("output.schema may only be specified when output.format=json"))
		}
	case "json":
		if cfg.Output.Schema == "" {
			errs = append(errs, errors.New("output.schema is required when output.format=json"))
		} else if _, err := CompileSchema(cfg.Output.Schema); err != nil {
			errs = append(errs, fmt.Errorf("output.schema: %w", err))
		}
	default:
		errs = append(errs, errors.New("output.format must be one of: text, json"))
	}
	if cfg.Corpus.Path != "" {
		if err := validateCorpus(cfg.Corpus.Path); err != nil {
			errs = append(errs, fmt.Errorf("corpus.path: %w", err))
		}
	}
	if err := validateRunnerArgs(cfg.Runner.Args); err != nil {
		errs = append(errs, fmt.Errorf("runner.args: %w", err))
	}
	return errs
}

func ValidateRun(cfg Config) []error {
	errs := ValidateConfig(cfg)
	if cfg.Corpus.Path == "" {
		errs = append(errs, errors.New("corpus.path is required"))
	}
	if cfg.Output.Dir == "" {
		errs = append(errs, errors.New("output.dir is required"))
	} else if info, err := os.Stat(cfg.Output.Dir); err == nil && !info.IsDir() {
		errs = append(errs, errors.New("output.dir must be a directory"))
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		errs = append(errs, fmt.Errorf("output.dir: %w", err))
	}
	return errs
}

func SupportedRunnerTypes() []string {
	types := make([]string, 0, len(runnerTypes))
	for runnerType := range runnerTypes {
		types = append(types, runnerType)
	}
	sort.Strings(types)
	return types
}

func validateCorpus(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return errors.New("must be a directory")
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Readdirnames(1)
	if errors.Is(err, io.EOF) {
		return nil
	}
	return err
}

func CompileSchema(path string) (*jsonschema.Schema, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(abs)
	if err != nil {
		return nil, err
	}
	if err := f.Close(); err != nil {
		return nil, err
	}
	compiler := jsonschema.NewCompiler()
	schema, err := compiler.Compile(abs)
	if err != nil {
		return nil, err
	}
	return schema, nil
}

func validateRunnerArgs(args map[string]any) error {
	for key, value := range args {
		if strings.TrimSpace(key) == "" {
			return errors.New("keys must not be empty")
		}
		if err := validateRunnerArgValue(value); err != nil {
			return fmt.Errorf("%s: %w", key, err)
		}
	}
	return nil
}

func validateRunnerArgValue(value any) error {
	if value == nil {
		return errors.New("value must not be null")
	}
	kind := reflect.ValueOf(value).Kind()
	switch kind {
	case reflect.String, reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return nil
	case reflect.Slice, reflect.Array:
		valueOf := reflect.ValueOf(value)
		for i := 0; i < valueOf.Len(); i++ {
			if err := validateRunnerArgValue(valueOf.Index(i).Interface()); err != nil {
				return fmt.Errorf("item %d: %w", i, err)
			}
		}
		return nil
	default:
		return fmt.Errorf("unsupported value type %T", value)
	}
}

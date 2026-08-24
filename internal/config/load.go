package config

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"go.yaml.in/yaml/v3"
)

func Resolve(path string, args []string) (Config, error) {
	cfg, err := Load(path)
	if err != nil {
		return Config{}, err
	}
	if err := ApplyOverrides(&cfg, args); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func Load(path string) (Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return Config{}, fmt.Errorf("load configuration: %w", err)
	}
	defer f.Close()

	cfg := Defaults()
	decoder := yaml.NewDecoder(f)
	decoder.KnownFields(true)
	if err := decoder.Decode(&cfg); err != nil {
		return Config{}, fmt.Errorf("decode configuration: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return Config{}, errors.New("decode configuration: multiple YAML documents are not supported")
		}
		return Config{}, fmt.Errorf("decode configuration: %w", err)
	}
	if cfg.Runner.Args == nil {
		cfg.Runner.Args = make(map[string]any)
	}

	base, err := filepath.Abs(filepath.Dir(path))
	if err != nil {
		return Config{}, fmt.Errorf("resolve configuration directory: %w", err)
	}
	cfg.Corpus.Path = resolveYAMLPath(base, cfg.Corpus.Path)
	cfg.Output.Dir = resolveYAMLPath(base, cfg.Output.Dir)
	cfg.Output.Schema = resolveYAMLPath(base, cfg.Output.Schema)
	if cfg.Runner.Command != "" && strings.ContainsRune(cfg.Runner.Command, filepath.Separator) {
		cfg.Runner.Command = resolveYAMLPath(base, cfg.Runner.Command)
	}
	return cfg, nil
}

func resolveYAMLPath(base, value string) string {
	if value == "" || filepath.IsAbs(value) {
		return value
	}
	return filepath.Clean(filepath.Join(base, value))
}

func Marshal(cfg Config) ([]byte, error) {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("marshal resolved configuration: %w", err)
	}
	return data, nil
}

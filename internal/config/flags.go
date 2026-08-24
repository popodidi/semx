package config

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"go.yaml.in/yaml/v3"
)

var aliases = map[string]string{
	"corpus": "corpus.path",
	"out":    "output.dir",
}

func ApplyOverrides(cfg *Config, args []string) error {
	fields := reflectedStringFields(reflect.ValueOf(cfg).Elem(), "")
	for i := 0; i < len(args); i++ {
		raw := args[i]
		if !strings.HasPrefix(raw, "--") || raw == "--" {
			return fmt.Errorf("unexpected argument %q", raw)
		}
		name, value, hasValue := strings.Cut(strings.TrimPrefix(raw, "--"), "=")
		if canonical, ok := aliases[name]; ok {
			name = canonical
		}
		if !hasValue {
			i++
			if i >= len(args) || strings.HasPrefix(args[i], "--") {
				return fmt.Errorf("flag --%s requires a value", name)
			}
			value = args[i]
		}
		if strings.HasPrefix(name, "runner.args.") {
			key := strings.TrimPrefix(name, "runner.args.")
			if key == "" {
				return errors.New("runner argument name is required")
			}
			parsed, err := parseRunnerArgValue(value)
			if err != nil {
				return fmt.Errorf("invalid value for --%s: %w", name, err)
			}
			cfg.Runner.Args[key] = parsed
			continue
		}
		field, ok := fields[name]
		if !ok {
			return fmt.Errorf("unknown flag --%s", name)
		}
		field.SetString(value)
	}
	return nil
}

func parseRunnerArgValue(value string) (any, error) {
	var parsed any
	if err := yaml.Unmarshal([]byte(value), &parsed); err != nil {
		return nil, err
	}
	return parsed, nil
}

func reflectedStringFields(value reflect.Value, prefix string) map[string]reflect.Value {
	fields := make(map[string]reflect.Value)
	typeOf := value.Type()
	for i := 0; i < value.NumField(); i++ {
		fieldType := typeOf.Field(i)
		name := strings.Split(fieldType.Tag.Get("yaml"), ",")[0]
		if name == "" || name == "-" {
			continue
		}
		path := name
		if prefix != "" {
			path = prefix + "." + name
		}
		field := value.Field(i)
		switch field.Kind() {
		case reflect.Struct:
			for childPath, child := range reflectedStringFields(field, path) {
				fields[childPath] = child
			}
		case reflect.String:
			fields[path] = field
		}
	}
	return fields
}

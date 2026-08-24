package runner

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
)

func SerializeArgs(values map[string]any) ([]string, error) {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var args []string
	for _, key := range keys {
		serialized, err := serializeArg("--"+key, values[key])
		if err != nil {
			return nil, fmt.Errorf("runner argument %q: %w", key, err)
		}
		args = append(args, serialized...)
	}
	return args, nil
}

func serializeArg(flag string, value any) ([]string, error) {
	if value == nil {
		return nil, errors.New("null values are not supported")
	}
	valueOf := reflect.ValueOf(value)
	switch valueOf.Kind() {
	case reflect.Bool:
		if valueOf.Bool() {
			return []string{flag}, nil
		}
		return nil, nil
	case reflect.String:
		return []string{flag, valueOf.String()}, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return []string{flag, fmt.Sprint(value)}, nil
	case reflect.Slice, reflect.Array:
		var args []string
		for i := 0; i < valueOf.Len(); i++ {
			item, err := serializeArg(flag, valueOf.Index(i).Interface())
			if err != nil {
				return nil, fmt.Errorf("item %d: %w", i, err)
			}
			args = append(args, item...)
		}
		return args, nil
	default:
		return nil, fmt.Errorf("unsupported value type %T", value)
	}
}

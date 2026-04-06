package config

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type Updater interface {
	Update() error
}

type structMeta struct {
	fieldValue reflect.Value
	fieldName  string
	path       string
	envList    []string
	defValue   *string
	required   bool
	updatable  bool
	separator  string
	layout     string
}

func (m *structMeta) isFieldValueZero() bool {
	return m.fieldValue.IsZero()
}

func readStructMetadata(cfg interface{}) ([]structMeta, error) {
	v := reflect.ValueOf(cfg)
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil, fmt.Errorf("config pointer is nil")
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("config must be a struct or pointer to struct, got %s", v.Kind())
	}
	return collectStructMeta(v, ""), nil
}

func collectStructMeta(v reflect.Value, path string) []structMeta {
	t := v.Type()
	var result []structMeta

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldVal := v.Field(i)

		if !fieldVal.CanSet() {
			continue
		}

		ft := fieldVal.Type()
		if ft.Kind() == reflect.Ptr {
			ft = ft.Elem()
		}

		if ft.Kind() == reflect.Struct && ft != reflect.TypeOf(time.Time{}) {
			var nested reflect.Value
			if fieldVal.Kind() == reflect.Ptr {
				if fieldVal.IsNil() {
					fieldVal.Set(reflect.New(ft))
				}
				nested = fieldVal.Elem()
			} else {
				nested = fieldVal
			}

			subPath := path
			if !field.Anonymous {
				if subPath != "" {
					subPath += "."
				}
				subPath += field.Name
			}
			result = append(result, collectStructMeta(nested, subPath)...)
			continue
		}

		meta := structMeta{
			fieldValue: fieldVal,
			fieldName:  field.Name,
			path:       maybeAddDot(path),
			separator:  ",",
		}

		if envTag, ok := field.Tag.Lookup("env"); ok && envTag != "" {
			meta.envList = splitTrimmed(envTag, ",")
		}

		if _, ok := field.Tag.Lookup("env-required"); ok {
			meta.required = true
		}

		if def, ok := field.Tag.Lookup("env-default"); ok {
			meta.defValue = &def
		}

		if sep, ok := field.Tag.Lookup("env-separator"); ok {
			meta.separator = sep
		}

		if layout, ok := field.Tag.Lookup("env-layout"); ok {
			meta.layout = layout
		}

		if _, ok := field.Tag.Lookup("env-upd"); ok {
			meta.updatable = true
		}

		result = append(result, meta)
	}

	return result
}

func parseValue(fieldVal reflect.Value, rawValue string, separator string, layout string) error {
	if fieldVal.Kind() == reflect.Ptr {
		newVal := reflect.New(fieldVal.Type().Elem())
		if err := parseValue(newVal.Elem(), rawValue, separator, layout); err != nil {
			return err
		}
		fieldVal.Set(newVal)
		return nil
	}

	switch fieldVal.Kind() {
	case reflect.String:
		fieldVal.SetString(rawValue)

	case reflect.Bool:
		b, err := strconv.ParseBool(rawValue)
		if err != nil {
			return fmt.Errorf("invalid bool %q: %w", rawValue, err)
		}
		fieldVal.SetBool(b)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
		n, err := strconv.ParseInt(rawValue, 10, fieldVal.Type().Bits())
		if err != nil {
			return fmt.Errorf("invalid int %q: %w", rawValue, err)
		}
		fieldVal.SetInt(n)

	case reflect.Int64:
		if fieldVal.Type() == reflect.TypeOf(time.Duration(0)) {
			d, err := time.ParseDuration(rawValue)
			if err != nil {
				return fmt.Errorf("invalid duration %q: %w", rawValue, err)
			}
			fieldVal.SetInt(int64(d))
		} else {
			n, err := strconv.ParseInt(rawValue, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid int64 %q: %w", rawValue, err)
			}
			fieldVal.SetInt(n)
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(rawValue, 10, fieldVal.Type().Bits())
		if err != nil {
			return fmt.Errorf("invalid uint %q: %w", rawValue, err)
		}
		fieldVal.SetUint(n)

	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(rawValue, fieldVal.Type().Bits())
		if err != nil {
			return fmt.Errorf("invalid float %q: %w", rawValue, err)
		}
		fieldVal.SetFloat(f)

	case reflect.Slice:
		if separator == "" {
			separator = ","
		}
		parts := strings.Split(rawValue, separator)
		slice := reflect.MakeSlice(fieldVal.Type(), len(parts), len(parts))
		for i, part := range parts {
			if err := parseValue(slice.Index(i), strings.TrimSpace(part), separator, layout); err != nil {
				return fmt.Errorf("slice element %d: %w", i, err)
			}
		}
		fieldVal.Set(slice)

	case reflect.Struct:
		if fieldVal.Type() == reflect.TypeOf(time.Time{}) {
			if layout == "" {
				layout = time.RFC3339
			}
			t, err := time.Parse(layout, rawValue)
			if err != nil {
				return fmt.Errorf("invalid time %q (layout %q): %w", rawValue, layout, err)
			}
			fieldVal.Set(reflect.ValueOf(t))
		} else {
			return fmt.Errorf("unsupported struct type %s", fieldVal.Type())
		}

	default:
		return fmt.Errorf("unsupported field type %s", fieldVal.Kind())
	}

	return nil
}

func readEnvVars(cfg interface{}, update bool) error {
	metaInfo, err := readStructMetadata(cfg)
	if err != nil {
		return err
	}

	if updater, ok := cfg.(Updater); ok {
		if err = updater.Update(); err != nil {
			return err
		}
	}

	for _, meta := range metaInfo {
		if update && !meta.updatable {
			continue
		}

		var rawValue *string

		for _, env := range meta.envList {
			if value, ok := os.LookupEnv(env); ok {
				rawValue = &value
				break
			}
		}

		var envName string
		if len(meta.envList) > 0 {
			envName = meta.envList[0]
		}

		if rawValue == nil && meta.required && meta.isFieldValueZero() {
			return fmt.Errorf("field %q is required but the value is not provided",
				meta.path+meta.fieldName,
			)
		}

		if rawValue == nil && meta.isFieldValueZero() {
			rawValue = meta.defValue
		}

		if rawValue == nil {
			continue
		}

		if err = parseValue(meta.fieldValue, *rawValue, meta.separator, meta.layout); err != nil {
			return fmt.Errorf("parsing field %q env %q: %v",
				meta.path+meta.fieldName, envName, err,
			)
		}
	}

	return nil
}

func maybeAddDot(path string) string {
	if path == "" {
		return ""
	}
	return path + "."
}

func splitTrimmed(s, sep string) []string {
	parts := strings.Split(s, sep)
	for i, p := range parts {
		parts[i] = strings.TrimSpace(p)
	}
	return parts
}

package app

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
)

var (
	errNotPointer      = errors.New("argument must be a pointer")
	errEmptyEnvVar     = errors.New("environment variable is set but empty")
	errEmptyInput      = errors.New("value cannot be empty")
	errUnsupportedType = errors.New("unsupported field type")
)

// forEachSensitiveField calls fn for each field tagged sensitive:"true" on the
// struct that v points to. v must be a pointer to a struct. Returns an error if
// v is not a pointer to a struct, or if fn returns an error.
func forEachSensitiveField(v any, fn func(reflect.StructField, reflect.Value) error) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer {
		return fmt.Errorf("forEachSensitiveField: %w", errNotPointer)
	}
	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return fmt.Errorf("forEachSensitiveField: argument must be a pointer to a struct")
	}

	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		if field.Tag.Get("sensitive") != "true" {
			continue
		}
		if err := fn(field, rv.Field(i)); err != nil {
			return err
		}
	}

	return nil
}

// promptSensitiveFields fills fields tagged `sensitive:"true"` on the given
// struct pointer. For each such field, it first checks for a matching
// environment variable (case-insensitive, prefixed with "STC_" + json tag or
// field name). If no env var is found, it prompts the user interactively using
// the `prompt` tag value, or the field name if no prompt tag is set.
//
// Returns an error if a sensitive field has a type other than string, int, float, or bool.
// Also returns an error if v is not a pointer, or if the user enters an empty value.
func promptSensitiveFields(v any, r io.Reader, w io.Writer) error {
	return forEachSensitiveField(v, func(field reflect.StructField, fieldVal reflect.Value) error {
		value, err := resolveFieldValue(field, r, w)
		if err != nil {
			return err
		}
		return setSensitiveField(fieldVal, field, value)
	})
}

// resolveFieldValue returns the value for a sensitive field, sourced from an
// environment variable if present, or from an interactive prompt otherwise.
func resolveFieldValue(field reflect.StructField, r io.Reader, w io.Writer) (string, error) {
	key := fieldEnvKey(field)
	if value, found := lookupEnvInsensitive(key); found {
		if value == "" {
			return "", fmt.Errorf("environment variable %q: %w", key, errEmptyEnvVar)
		}
		return value, nil
	}
	return readPromptValue(r, w, promptLabel(field))
}

// fieldJSONKey returns the json tag name for field, or the field name if no
// json tag is set.
func fieldJSONKey(field reflect.StructField) string {
	if jsonTag := field.Tag.Get("json"); jsonTag != "" {
		return strings.Split(jsonTag, ",")[0]
	}
	return field.Name
}

// fieldEnvKey returns "STC_" + the json tag name, or "STC_" + the field name
// if no json tag is set.
func fieldEnvKey(field reflect.StructField) string {
	return "STC_" + fieldJSONKey(field)
}

// isSetByEnvVar reports whether the env var corresponding to field is present
// and non-empty.
func isSetByEnvVar(field reflect.StructField) bool {
	value, found := lookupEnvInsensitive(fieldEnvKey(field))
	return found && value != ""
}

// appendSensitiveDiagnostics appends a diagnostic entry for each
// sensitive-tagged field of v, indicating whether the field can be resolved
// from an environment variable ("loaded") or will require an interactive prompt
// ("will prompt"). v must be a pointer to a struct.
func appendSensitiveDiagnostics(diagnostics *Diagnostics, v any) error {
	return forEachSensitiveField(v, func(field reflect.StructField, _ reflect.Value) error {
		status := statusWillPrompt
		if isSetByEnvVar(field) {
			status = statusLoaded
		}
		diagnostics.append(diagnostic{Field: fieldJSONKey(field), Status: status})
		return nil
	})
}

// promptLabel returns the prompt tag value, or the field name if no prompt tag is set.
func promptLabel(field reflect.StructField) string {
	if label := field.Tag.Get("prompt"); label != "" {
		return label
	}
	return field.Name
}

// lookupEnvInsensitive does a case-insensitive search for key in os.Environ.
// Returns the value and true if found, or ("", false) if not present.
func lookupEnvInsensitive(key string) (string, bool) {
	for _, entry := range os.Environ() {
		parts := strings.SplitN(entry, "=", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], key) {
			return parts[1], true
		}
	}
	return "", false
}

// readPromptValue writes a prompt to w, reads one line from r, and returns the
// trimmed value. Returns errEmptyInput if the user enters nothing.
func readPromptValue(r io.Reader, w io.Writer, label string) (string, error) {
	_, _ = fmt.Fprintf(w, "Enter a value for %q: ", label)
	input, err := bufio.NewReader(r).ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("error reading input for %q: %v", label, err)
	}
	if value := strings.TrimSpace(input); value != "" {
		return value, nil
	}
	return "", fmt.Errorf("field %q: %w", label, errEmptyInput)
}

func setSensitiveField(fieldVal reflect.Value, field reflect.StructField, value string) error {
	switch fieldVal.Kind() {
	case reflect.String:
		fieldVal.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("field %q: cannot parse %q as int: %w", field.Name, value, err)
		}
		fieldVal.SetInt(n)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("field %q: cannot parse %q as float: %w", field.Name, value, err)
		}
		fieldVal.SetFloat(f)
	case reflect.Bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("field %q: cannot parse %q as bool: %w", field.Name, value, err)
		}
		fieldVal.SetBool(b)
	default:
		return fmt.Errorf("promptSensitiveFields: field %q: %w: %s", field.Name, errUnsupportedType, fieldVal.Kind())
	}
	return nil
}

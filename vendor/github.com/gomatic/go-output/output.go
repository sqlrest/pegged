// Package output encodes a command result to a writer in a selectable format
// (JSON or YAML). It is CLI-agnostic — it takes a writer, a format, and a value —
// so a daemon, a CLI, or a test can all reuse the same encoding. It owns the one
// error intrinsic to its own machinery, declared on the shared sentinel type.
package output

import (
	"encoding/json"
	"io"

	"gopkg.in/yaml.v3"
)

// Format selects how a result is encoded.
type Format string

// FormatJSON and FormatYAML are the supported encodings.
const (
	FormatJSON Format = "json"
	FormatYAML Format = "yaml"
)

// encoder writes data to w in a specific format.
type encoder func(w io.Writer, data any) error

var encoders = map[Format]encoder{
	FormatJSON: encodeJSON,
	FormatYAML: encodeYAML,
}

func encodeJSON(w io.Writer, data any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(data)
}

func encodeYAML(w io.Writer, data any) error {
	enc := yaml.NewEncoder(w)
	defer enc.Close()
	return enc.Encode(data)
}

// Write encodes data to w in the requested format, defaulting to JSON when the
// format is empty, and returning ErrUnsupportedFormat for any unknown format.
func Write(w io.Writer, f Format, data any) error {
	if f == "" {
		f = FormatJSON
	}
	enc, ok := encoders[f]
	if !ok {
		return ErrUnsupportedFormat.With(nil, string(f))
	}
	return enc(w, data)
}

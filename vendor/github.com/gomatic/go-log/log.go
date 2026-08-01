// Package log provides CLI-agnostic structured-logging setup over log/slog:
// textual level and format types bound from configuration, and a constructor
// that builds a *slog.Logger over any writer. It knows nothing about command-line
// frameworks; the binding of these types to flags lives in a consumer.
package log

import (
	"io"
	"log/slog"
)

type (
	// Level is the textual logging level (debug, info, warn, error).
	Level string
	// Format selects the log encoding (text or json).
	Format string
)

// FormatText and FormatJSON are the supported log encodings.
const (
	FormatText Format = "text"
	FormatJSON Format = "json"
)

// handlerFunc constructs a slog handler over a writer.
type handlerFunc func(io.Writer, *slog.HandlerOptions) slog.Handler

var handlers = map[Format]handlerFunc{
	FormatText: func(w io.Writer, o *slog.HandlerOptions) slog.Handler { return slog.NewTextHandler(w, o) },
	FormatJSON: func(w io.Writer, o *slog.HandlerOptions) slog.Handler { return slog.NewJSONHandler(w, o) },
}

// level parses the textual level, defaulting to info when empty or invalid.
func (l Level) level() slog.Level {
	var lvl slog.Level
	if err := lvl.UnmarshalText([]byte(l)); err != nil {
		return slog.LevelInfo
	}
	return lvl
}

// handler returns the slog handler for the format, defaulting to text when unknown.
func (f Format) handler(w io.Writer, opts *slog.HandlerOptions) slog.Handler {
	if h, ok := handlers[f]; ok {
		return h(w, opts)
	}
	return slog.NewTextHandler(w, opts)
}

// LoggerConfig holds the logging configuration bound from a consumer's flags.
type LoggerConfig struct {
	Level  Level
	Format Format
}

// NewLogger builds a logger writing to w using the level and format in cfg.
func (cfg LoggerConfig) NewLogger(w io.Writer) *slog.Logger {
	return slog.New(cfg.Format.handler(w, &slog.HandlerOptions{Level: cfg.Level.level()}))
}

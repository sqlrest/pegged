package app

import (
	"context"
	"log/slog"

	"github.com/urfave/cli/v3"
)

// LoggerMetadataKey is the cli.Command metadata key under which the configured
// logger is stored for retrieval by command actions.
const LoggerMetadataKey = "logger"

// GetLoggerFunc retrieves a logger for a command; injected so mains and tests can
// substitute one.
type GetLoggerFunc func(*cli.Command) *slog.Logger

// GetLogger returns the logger stored in the root command metadata, or
// slog.Default when absent or the command is nil.
func GetLogger(c *cli.Command) *slog.Logger {
	if c == nil {
		return slog.Default()
	}
	if logger, ok := c.Root().Metadata[LoggerMetadataKey].(*slog.Logger); ok {
		return logger
	}
	return slog.Default()
}

// LoggerBefore returns a cli Before hook that stores the logger built by
// getLogger in the root command metadata, where GetLogger then finds it.
func LoggerBefore(getLogger GetLoggerFunc) func(context.Context, *cli.Command) (context.Context, error) {
	return func(ctx context.Context, c *cli.Command) (context.Context, error) {
		c.Root().Metadata[LoggerMetadataKey] = getLogger(c)
		return ctx, nil
	}
}

var _ GetLoggerFunc = GetLogger

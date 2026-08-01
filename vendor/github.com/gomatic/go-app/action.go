// Package app is the urfave/cli/v3 framework shared by gomatic CLIs. It
// supplies the generic, command-agnostic glue — the runner/action combinators,
// the logger-in-metadata convention, the standard global flags, and the run
// harness — and nothing tied to any one command. Command-specific flags, errors,
// and trees stay in their own repositories.
package app

import (
	"context"
	"log/slog"

	output "github.com/gomatic/go-output"
	"github.com/urfave/cli/v3"
)

// outputFlagName is the global flag whose value selects the result encoding.
const outputFlagName = "output"

// Runner executes a command's work: it receives the bound config and positional
// arguments and returns a result to be encoded. Implementations live in the
// orchestration (domain) packages of each consumer.
type Runner[CONFIG any, RESULT any] func(context.Context, *slog.Logger, CONFIG, ...string) (RESULT, error)

// getLogger is a package indirection so tests can substitute a logger source.
var getLogger = GetLogger

// action runs the runner with the bound config and encodes its result in the
// format selected by the root command's output flag.
func action[C, R any](ctx context.Context, c *cli.Command, cfg C, runner Runner[C, R]) error {
	result, err := runner(ctx, getLogger(c), cfg, c.Args().Slice()...)
	if err != nil {
		return err
	}
	return output.Write(c.Root().Writer, output.Format(c.Root().String(outputFlagName)), result)
}

// Default binds a config pointer and runner into a cli action function.
func Default[C, R any](cfg *C, runner Runner[C, R]) func(context.Context, *cli.Command) error {
	return func(ctx context.Context, c *cli.Command) error {
		return action(ctx, c, *cfg, runner)
	}
}

// Interactive binds a config pointer and runner into a cli action function for
// a command that owns its output — an interactive REPL, say — and so has no
// result for the app tier to render. The runner's result is intentionally
// discarded; only its error propagates. Use [Default] whenever the command
// produces a result to encode; result rendering is the common case, not a
// requirement.
func Interactive[C, R any](cfg *C, runner Runner[C, R]) func(context.Context, *cli.Command) error {
	return func(ctx context.Context, c *cli.Command) error {
		_, err := runner(ctx, getLogger(c), *cfg, c.Args().Slice()...)
		return err
	}
}

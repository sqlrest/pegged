package app

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/urfave/cli/v3"
)

// Run executes cmd with signal-aware cancellation (SIGINT/SIGTERM), logging a
// non-nil error under the command's name and exiting non-zero via exit. The args
// and exit function are injected so a main can be exercised in tests.
func Run(ctx context.Context, cmd *cli.Command, args []string, exit func(int)) {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := cmd.Run(ctx, args); err != nil {
		slog.Error(cmd.Name+" error", "error", err)
		exit(1)
	}
}

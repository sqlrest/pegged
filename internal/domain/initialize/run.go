package initialize

import (
	"context"
	"log/slog"

	docker "github.com/gomatic/go-docker"
	pgdocker "github.com/gomatic/go-pgdocker"

	"github.com/sqlrest/pegged/internal/constants"
	"github.com/sqlrest/pegged/internal/domain"
	"github.com/sqlrest/pegged/internal/manage"
)

// newManager is the manager seam; tests substitute a manager built over a
// fake engine.
var newManager = manage.Manager

// Result is the outcome of the init command.
type Result struct {
	ExitCode docker.ExitCode `json:"exit_code"`
}

// Run executes the --image container against the resolved port's running
// instance. The first positional argument is the port only when it is
// numeric (a digits-only token is an intended port — an out-of-range one
// fails with ErrInvalidPort rather than becoming a command); every
// remaining argument overrides the container command. A non-zero exit from
// the init command returns both the exit code and ErrInitFailed, so the
// process exits non-zero.
func Run(ctx context.Context, logger *slog.Logger, cfg Config, args ...domain.Argument) (Result, error) {
	if cfg.Image == "" {
		return Result{}, constants.ErrMissingArgument.With(nil, "flag", "image")
	}
	port, command, err := splitArgs(args, manage.FlagPort(cfg.Port))
	if err != nil {
		return Result{}, err
	}
	manager, err := newManager(ctx)
	if err != nil {
		return Result{}, err
	}
	exit, err := manager.Init(ctx, initSpec(cfg, port, command))
	if err != nil {
		return Result{}, err
	}
	if exit != 0 {
		return Result{ExitCode: exit}, constants.ErrInitFailed.With(nil, "exit_code", int64(exit))
	}
	logger.Info("Initialization complete.", "port", int(port), "image", string(cfg.Image))
	return Result{ExitCode: exit}, nil
}

// initSpec renders the library's InitSpec from the resolved inputs.
func initSpec(cfg Config, port docker.Port, command []domain.Argument) pgdocker.InitSpec {
	return pgdocker.InitSpec{
		Output:   cfg.Output,
		Image:    docker.ImageRef(cfg.Image),
		Database: pgdocker.DatabaseName(cfg.Database),
		User:     pgdocker.UserName(cfg.User),
		Password: pgdocker.Password(cfg.Password),
		Command:  docker.Command(command),
		Port:     port,
	}
}

// splitArgs peels a leading port argument off the init argv: a numeric
// first argument is an intended port (ResolvePort rejects one out of
// range); everything else is the container command override.
func splitArgs(args []domain.Argument, flagged manage.FlagPort) (docker.Port, []domain.Argument, error) {
	if len(args) == 0 || !isNumeric(args[0]) {
		return docker.Port(flagged), args, nil
	}
	port, err := manage.ResolvePort(args[:1], flagged)
	return port, args[1:], err
}

// isNumeric reports whether an argument is a non-empty digits-only token —
// an intended port, whether or not it is in range.
func isNumeric(arg domain.Argument) bool {
	if arg == "" {
		return false
	}
	for _, character := range arg {
		if character < '0' || character > '9' {
			return false
		}
	}
	return true
}

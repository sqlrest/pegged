package stop

import (
	"context"
	"log/slog"

	docker "github.com/gomatic/go-docker"
	pgdocker "github.com/gomatic/go-pgdocker"

	"github.com/sqlrest/pegged/internal/domain"
	"github.com/sqlrest/pegged/internal/manage"
)

// newManager is the manager seam; tests substitute a manager built over a
// fake engine.
var newManager = manage.Manager

// Result is the outcome of the stop command.
type Result struct {
	Report pgdocker.StopReport `json:"report"`
}

// Run stops the resolved port's managed instance and applies volume
// retention. All lifecycle behavior lives in go-pgdocker; Run only resolves
// flags into a StopSpec.
func Run(ctx context.Context, logger *slog.Logger, cfg Config, args ...domain.Argument) (Result, error) {
	port, err := manage.ResolvePort(args, manage.FlagPort(cfg.Port))
	if err != nil {
		return Result{}, err
	}
	retain, err := manage.RetentionFor(manage.FlagValue(cfg.Retain))
	if err != nil {
		return Result{}, err
	}
	manager, err := newManager(ctx)
	if err != nil {
		return Result{}, err
	}
	report, err := manager.Stop(ctx, pgdocker.StopSpec{
		Retain: retain,
		Grace:  docker.StopSeconds(cfg.Grace),
		Port:   port,
	})
	if err != nil {
		return Result{}, err
	}
	logger.Info("Instance stopped.",
		"port", int(port), "stopped", len(report.Stopped), "removed_volumes", len(report.RemovedVolumes))
	return Result{Report: report}, nil
}

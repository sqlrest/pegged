package inspect

import (
	"context"
	"log/slog"

	pgdocker "github.com/gomatic/go-pgdocker"

	"github.com/sqlrest/pegged/internal/domain"
	"github.com/sqlrest/pegged/internal/manage"
)

// newManager is the manager seam; tests substitute a manager built over a
// fake engine.
var newManager = manage.Manager

// Result is the outcome of the inspect command.
type Result struct {
	Report pgdocker.Report `json:"report"`
}

// Run reports the resolved port's managed container and volumes.
func Run(ctx context.Context, logger *slog.Logger, cfg Config, args ...domain.Argument) (Result, error) {
	port, err := manage.ResolvePort(args, manage.FlagPort(cfg.Port))
	if err != nil {
		return Result{}, err
	}
	manager, err := newManager(ctx)
	if err != nil {
		return Result{}, err
	}
	report, err := manager.Inspect(ctx, port)
	if err != nil {
		return Result{}, err
	}
	logger.Info("Port inspected.",
		"port", int(report.Port), "has_container", report.HasContainer, "volumes", len(report.Volumes))
	return Result{Report: report}, nil
}

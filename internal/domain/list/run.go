package list

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

// Result is the outcome of the list command.
type Result struct {
	Reports []pgdocker.Report `json:"reports"`
}

// Run reports every managed port on the daemon, ordered by port.
func Run(ctx context.Context, logger *slog.Logger, _ Config, _ ...domain.Argument) (Result, error) {
	manager, err := newManager(ctx)
	if err != nil {
		return Result{}, err
	}
	reports, err := manager.List(ctx)
	if err != nil {
		return Result{}, err
	}
	logger.Info("Managed ports listed.", "ports", len(reports))
	return Result{Reports: reports}, nil
}

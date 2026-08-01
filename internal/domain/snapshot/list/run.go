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

// Result is the outcome of the "snapshot list" command.
type Result struct {
	Snapshots []pgdocker.Snapshot `json:"snapshots"`
}

// Run reports every stored snapshot in pegged's namespace.
func Run(ctx context.Context, logger *slog.Logger, _ Config, _ ...domain.Argument) (Result, error) {
	manager, err := newManager(ctx)
	if err != nil {
		return Result{}, err
	}
	snapshots, err := manager.SnapshotList(ctx)
	if err != nil {
		return Result{}, err
	}
	logger.Info("Snapshots listed.", "snapshots", len(snapshots))
	return Result{Snapshots: snapshots}, nil
}

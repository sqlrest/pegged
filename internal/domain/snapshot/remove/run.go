package remove

import (
	"context"
	"log/slog"

	pgdocker "github.com/gomatic/go-pgdocker"

	"github.com/sqlrest/pegged/internal/constants"
	"github.com/sqlrest/pegged/internal/domain"
	"github.com/sqlrest/pegged/internal/manage"
)

// newManager is the manager seam; tests substitute a manager built over a
// fake engine.
var newManager = manage.Manager

// Result is the outcome of the "snapshot delete" command.
type Result struct {
	Deleted pgdocker.SnapshotName `json:"deleted"`
}

// Run removes the named snapshot's volume. The snapshot name is required.
func Run(ctx context.Context, logger *slog.Logger, _ Config, args ...domain.Argument) (Result, error) {
	if len(args) < 1 {
		return Result{}, constants.ErrMissingArgument.With(nil, "argument", "name")
	}
	name := pgdocker.SnapshotName(args[0])
	manager, err := newManager(ctx)
	if err != nil {
		return Result{}, err
	}
	if err := manager.SnapshotDelete(ctx, name); err != nil {
		return Result{}, err
	}
	logger.Info("Snapshot deleted.", "name", string(name))
	return Result{Deleted: name}, nil
}

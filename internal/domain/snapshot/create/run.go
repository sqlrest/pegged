package create

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

// provenanceLabels is the provenance seam; tests substitute a fixed stamp
// so the spec's label content is pinned against the seam's output.
var provenanceLabels = manage.ProvenanceLabels

// Result is the outcome of the "snapshot create" command.
type Result struct {
	Snapshot pgdocker.Snapshot `json:"snapshot"`
}

// Run snapshots the resolved port's newest data volume. The first positional
// argument is the port, the second the snapshot name; an omitted name gets a
// timestamp. Snapshot semantics — stopped-instance requirement, major-version
// binding — live in go-pgdocker.
func Run(ctx context.Context, logger *slog.Logger, cfg Config, args ...domain.Argument) (Result, error) {
	port, err := manage.ResolvePort(args, manage.FlagPort(cfg.Port))
	if err != nil {
		return Result{}, err
	}
	manager, err := newManager(ctx)
	if err != nil {
		return Result{}, err
	}
	snapshot, err := manager.SnapshotCreate(ctx, pgdocker.SnapshotSpec{
		Labels:    provenanceLabels(),
		Name:      nameFrom(args),
		CopyImage: docker.ImageRef(cfg.CopyImage),
		Port:      port,
	})
	if err != nil {
		return Result{}, err
	}
	logger.Info("Snapshot created.",
		"name", string(snapshot.Name), "volume", string(snapshot.Volume), "major", string(snapshot.Major))
	return Result{Snapshot: snapshot}, nil
}

// nameFrom reads the optional snapshot name from the second positional
// argument; empty defers to the library's timestamp default.
func nameFrom(args []domain.Argument) pgdocker.SnapshotName {
	if len(args) < 2 {
		return ""
	}
	return pgdocker.SnapshotName(args[1])
}

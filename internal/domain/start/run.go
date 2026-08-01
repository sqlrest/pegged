package start

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

// Result is the outcome of the start command.
type Result struct {
	Instance pgdocker.Instance `json:"instance"`
}

// Run starts a managed PostgreSQL instance on the resolved port. All
// lifecycle behavior — port policy, volume selection, readiness — lives in
// go-pgdocker; Run only resolves flags into a StartSpec.
func Run(ctx context.Context, logger *slog.Logger, cfg Config, args ...domain.Argument) (Result, error) {
	spec, err := specFor(cfg, args)
	if err != nil {
		return Result{}, err
	}
	manager, err := newManager(ctx)
	if err != nil {
		return Result{}, err
	}
	instance, err := manager.Start(ctx, spec)
	if err != nil {
		return Result{}, err
	}
	logger.Info("Instance started.",
		"port", int(instance.Port), "container", string(instance.Name))
	return Result{Instance: instance}, nil
}

// specFor resolves the port and validates the enumerated flags into the
// library's StartSpec.
func specFor(cfg Config, args []domain.Argument) (pgdocker.StartSpec, error) {
	port, err := manage.ResolvePort(args, manage.FlagPort(cfg.Port))
	if err != nil {
		return pgdocker.StartSpec{}, err
	}
	volume, err := manage.VolumeSourceFor(manage.FlagValue(cfg.Volume))
	if err != nil {
		return pgdocker.StartSpec{}, err
	}
	retain, err := manage.RetentionFor(manage.FlagValue(cfg.Retain))
	if err != nil {
		return pgdocker.StartSpec{}, err
	}
	return pgdocker.StartSpec{
		Labels:   provenanceLabels(),
		Volume:   manage.PolicyVolume(port, volume),
		Retain:   manage.PolicyRetention(port, retain),
		Image:    docker.ImageRef(cfg.Image),
		Platform: docker.Platform(cfg.Platform),
		Snapshot: pgdocker.SnapshotName(cfg.Snapshot),
		Database: pgdocker.DatabaseName(cfg.Database),
		Listen:   pgdocker.ListenAddress(cfg.Listen),
		Password: pgdocker.Password(cfg.Password),
		User:     pgdocker.UserName(cfg.User),
		InitSQL:  pgdocker.InitSQLDir(cfg.InitSQL),
		Port:     port,
	}, nil
}

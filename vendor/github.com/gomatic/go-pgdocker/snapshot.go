package pgdocker

import (
	"context"
	"strconv"

	docker "github.com/gomatic/go-docker"
)

// Snapshot describes one stored snapshot.
type Snapshot struct {
	Labels     docker.Labels
	Name       SnapshotName
	Volume     docker.VolumeName
	Major      PostgresMajor
	CreatedAt  docker.VolumeCreatedAt
	SourcePort docker.Port
}

// SnapshotSpec describes a snapshot to create.
type SnapshotSpec struct {
	// Labels ride onto the snapshot volume for the caller's own
	// provenance or discovery; namespaced managed keys always win a
	// collision.
	Labels    docker.Labels
	Name      SnapshotName
	CopyImage docker.ImageRef
	Port      docker.Port
}

// SnapshotCreate clones the port's newest data volume into a labeled
// snapshot volume. Snapshots are physical PGDATA copies: consistent only
// when the database is stopped, so a running instance on the port is
// refused, and a snapshot boots only on the PostgreSQL major that wrote it.
func (m Manager) SnapshotCreate(ctx context.Context, spec SnapshotSpec) (Snapshot, error) {
	spec = m.snapshotDefaults(spec)
	if err := m.refuseRunning(ctx, spec.Port); err != nil {
		return Snapshot{}, err
	}
	if err := m.refuseExisting(ctx, spec.Name); err != nil {
		return Snapshot{}, err
	}
	source, err := m.sourceVolume(ctx, spec.Port)
	if err != nil {
		return Snapshot{}, err
	}
	created, err := m.createSnapshotVolume(ctx, spec, source)
	if err != nil {
		return Snapshot{}, err
	}
	if err := m.clone(ctx, m.cloneImageFor(spec, source), cloneCommand(), source.Name, created); err != nil {
		removeErr := m.engine.RemoveVolume(ctx, created, docker.VolumeRemoval{})
		if removeErr != nil {
			return Snapshot{}, ErrCloneFailed.With(err, "orphaned_volume", string(created))
		}
		return Snapshot{}, err
	}
	return m.snapshotByName(ctx, spec.Name)
}

// snapshotDefaults resolves the spec's zero fields.
func (m Manager) snapshotDefaults(spec SnapshotSpec) SnapshotSpec {
	if spec.Port == 0 {
		spec.Port = DefaultPort
	}
	if spec.Name == "" {
		spec.Name = SnapshotName(m.clock().UTC().Format(compactTimestamp))
	}
	return spec
}

// SnapshotList reports every snapshot in the namespace, sorted by name.
func (m Manager) SnapshotList(ctx context.Context) ([]Snapshot, error) {
	volumes, err := m.engine.Volumes(ctx, docker.VolumeQuery{Labels: m.snapshotFilter("")})
	if err != nil {
		return nil, err
	}
	snapshots := make([]Snapshot, 0, len(volumes))
	for _, volume := range volumes {
		snapshots = append(snapshots, m.snapshotFrom(volume))
	}
	return snapshots, nil
}

// snapshotFrom maps a labeled volume onto the snapshot shape.
func (m Manager) snapshotFrom(volume docker.VolumeDetails) Snapshot {
	sourcePort, _ := strconv.ParseUint(m.labelValue(volume.Labels, labelSourcePort), 10, 16)
	return Snapshot{
		Name:       SnapshotName(m.labelValue(volume.Labels, labelSnapshot)),
		Volume:     volume.Name,
		Major:      PostgresMajor(m.labelValue(volume.Labels, labelMajor)),
		SourcePort: docker.Port(sourcePort),
		CreatedAt:  volume.CreatedAt,
		Labels:     volume.Labels,
	}
}

// SnapshotDelete removes a snapshot's volume.
func (m Manager) SnapshotDelete(ctx context.Context, name SnapshotName) error {
	snapshot, err := m.snapshotByName(ctx, name)
	if err != nil {
		return err
	}
	return m.engine.RemoveVolume(ctx, snapshot.Volume, docker.VolumeRemoval{})
}

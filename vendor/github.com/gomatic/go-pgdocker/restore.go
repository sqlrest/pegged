package pgdocker

import (
	"context"

	docker "github.com/gomatic/go-docker"
)

// snapshotByName finds one snapshot, failing with ErrSnapshotNotFound.
func (m Manager) snapshotByName(ctx context.Context, name SnapshotName) (Snapshot, error) {
	volumes, err := m.engine.Volumes(ctx, docker.VolumeQuery{Labels: m.snapshotFilter(name)})
	if err != nil {
		return Snapshot{}, err
	}
	if len(volumes) == 0 {
		return Snapshot{}, ErrSnapshotNotFound.With(nil, "name", string(name))
	}
	return m.snapshotFrom(volumes[0]), nil
}

// cloneSnapshot clones a snapshot into a fresh data volume for a starting
// instance, refusing a PostgreSQL major mismatch unless the spec forces it.
func (m Manager) cloneSnapshot(ctx context.Context, spec StartSpec) (docker.VolumeName, error) {
	snapshot, err := m.snapshotByName(ctx, spec.Snapshot)
	if err != nil {
		return "", err
	}
	if err = checkSkew(snapshot, spec); err != nil {
		return "", err
	}
	destination, err := m.freshVolume(ctx, spec)
	if err != nil {
		return "", err
	}
	if err = m.clone(ctx, spec.Image, restoreCommand(), snapshot.Volume, destination); err != nil {
		return "", err
	}
	return destination, nil
}

// checkSkew refuses a snapshot restore across PostgreSQL majors unless the
// spec explicitly allows the skew.
func checkSkew(snapshot Snapshot, spec StartSpec) error {
	if spec.SnapshotSkewEnabled {
		return nil
	}
	requested := imageMajor(spec.Image)
	if snapshot.Major == "" || requested == "" || snapshot.Major == requested {
		return nil
	}
	return ErrSnapshotSkew.With(nil,
		"snapshot", string(snapshot.Name),
		"snapshot_major", string(snapshot.Major),
		"image_major", string(requested))
}

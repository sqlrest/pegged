package pgdocker

import (
	"context"
	"strconv"
	"time"

	docker "github.com/gomatic/go-docker"
)

// refuseRunning rejects snapshotting under a live server: a PGDATA copy of
// a running database is torn and useless.
func (m Manager) refuseRunning(ctx context.Context, port docker.Port) error {
	report, err := m.Inspect(ctx, port)
	if err != nil {
		return err
	}
	if report.HasContainer && report.Instance.IsRunning {
		return ErrAlreadyRunning.With(nil,
			"port", int(port), "reason", "stop the instance before snapshotting")
	}
	return nil
}

// refuseExisting rejects a duplicate snapshot name.
func (m Manager) refuseExisting(ctx context.Context, name SnapshotName) error {
	existing, err := m.engine.Volumes(ctx, docker.VolumeQuery{Labels: m.snapshotFilter(name)})
	if err != nil {
		return err
	}
	if len(existing) > 0 {
		return ErrSnapshotExists.With(nil, "name", string(name))
	}
	return nil
}

// sourceVolume finds the newest data volume to snapshot.
func (m Manager) sourceVolume(ctx context.Context, port docker.Port) (docker.VolumeDetails, error) {
	name, err := m.newestVolume(ctx, port)
	if err != nil {
		return docker.VolumeDetails{}, err
	}
	if name == "" {
		return docker.VolumeDetails{}, ErrNoDataVolume.With(nil, "port", int(port))
	}
	volumes, err := m.engine.Volumes(ctx, docker.VolumeQuery{Labels: m.portFilter(port)})
	if err != nil {
		return docker.VolumeDetails{}, err
	}
	for _, volume := range volumes {
		if volume.Name == name {
			return volume, nil
		}
	}
	return docker.VolumeDetails{}, ErrNoDataVolume.With(nil, "port", int(port))
}

// createSnapshotVolume creates the labeled destination volume, carrying the
// source's postgres image and major so restores can verify compatibility.
func (m Manager) createSnapshotVolume(
	ctx context.Context,
	spec SnapshotSpec,
	source docker.VolumeDetails,
) (docker.VolumeName, error) {
	labels := docker.Labels{
		m.labelKey(labelManaged):    labelTrue,
		m.labelKey(labelSnapshot):   string(spec.Name),
		m.labelKey(labelSourcePort): strconv.Itoa(int(spec.Port)),
		m.labelKey(labelImage):      m.labelValue(source.Labels, labelImage),
		m.labelKey(labelMajor):      m.labelValue(source.Labels, labelMajor),
		m.labelKey(labelCreated):    m.clock().UTC().Format(time.RFC3339),
	}
	created, err := m.engine.CreateVolume(ctx, docker.VolumeSpec{
		Name:   m.snapshotVolumeName(spec.Name),
		Labels: m.mergeLabels(labels, spec.Labels),
	})
	if err != nil {
		return "", err
	}
	return created.Name, nil
}

// cloneImageFor picks the helper image for the copy: the caller's choice,
// the source volume's recorded postgres image, or the package default.
func (m Manager) cloneImageFor(spec SnapshotSpec, source docker.VolumeDetails) docker.ImageRef {
	if spec.CopyImage != "" {
		return spec.CopyImage
	}
	if recorded := m.labelValue(source.Labels, labelImage); recorded != "" {
		return docker.ImageRef(recorded)
	}
	return DefaultImage
}

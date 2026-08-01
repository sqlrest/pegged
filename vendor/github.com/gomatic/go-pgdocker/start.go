package pgdocker

import (
	"context"

	docker "github.com/gomatic/go-docker"
)

// Instance is a started (or inspected) managed PostgreSQL instance. It
// states facts about the instance only — no connection URL is derived,
// because roles, databases, and passwords belong to initialization and
// later administration, not to the lifecycle record.
type Instance struct {
	Labels    docker.Labels
	Name      docker.ContainerName
	ID        docker.ContainerID
	Image     docker.ImageRef
	Volume    docker.VolumeName
	Status    docker.ContainerStatus
	Port      docker.Port
	IsRunning bool
}

// Start provisions and starts a PostgreSQL instance for the spec's port,
// waits until the server answers pg_isready, and returns the running
// instance. Starting a port that already has a managed container fails
// with ErrAlreadyRunning.
//
// Start is atomic on its container: if the launch or readiness wait fails,
// the container it created is force-removed before the error returns, so a
// failed Start never leaves a container running. The data volume is kept —
// it is durable and reusable, and inspecting it is often how a failure is
// diagnosed.
func (m Manager) Start(ctx context.Context, spec StartSpec) (Instance, error) {
	spec = spec.withDefaults()
	if err := m.ensureVacant(ctx, spec.Port); err != nil {
		return Instance{}, err
	}
	volume, err := m.resolveVolume(ctx, spec)
	if err != nil {
		return Instance{}, err
	}
	if err = m.ensureImage(ctx, spec.Image, spec.Platform); err != nil {
		return Instance{}, err
	}
	rendered := m.containerSpec(spec, volume)
	id, err := m.engine.CreateContainer(ctx, rendered)
	if err != nil {
		return Instance{}, err
	}
	if err = m.launch(ctx, id, spec); err != nil {
		m.removeFailedStart(ctx, id)
		return Instance{}, err
	}
	return m.instanceFor(spec, id, volume, rendered.Labels), nil
}

// launch starts the created container and waits until it answers readiness.
func (m Manager) launch(ctx context.Context, id docker.ContainerID, spec StartSpec) error {
	if err := m.engine.StartContainer(ctx, id); err != nil {
		return err
	}
	return m.awaitReady(ctx, id, spec)
}

// removeFailedStart force-removes a container whose launch or readiness
// wait failed. Removal is best-effort: a container the daemon already
// reaped (an image that exited on boot auto-removes itself) is not an error
// worth surfacing over the launch failure that prompted the cleanup.
func (m Manager) removeFailedStart(ctx context.Context, id docker.ContainerID) {
	_ = m.engine.RemoveContainer(ctx, id, docker.RemovalOptions{ForceEnabled: true})
}

// ensureVacant fails with ErrAlreadyRunning when any managed container —
// running or stopped — already occupies the port.
func (m Manager) ensureVacant(ctx context.Context, port docker.Port) error {
	occupants, err := m.engine.Containers(ctx, docker.ContainerQuery{
		Labels:     m.portFilter(port),
		AllEnabled: true,
	})
	if err != nil {
		return err
	}
	if len(occupants) > 0 {
		return ErrAlreadyRunning.With(nil,
			"port", int(port), "container", string(occupants[0].Name))
	}
	return nil
}

// ensureImage pulls the image only when the daemon does not hold it.
func (m Manager) ensureImage(
	ctx context.Context,
	image docker.ImageRef,
	platform docker.Platform,
) error {
	held, err := m.engine.ImageExists(ctx, image)
	if err != nil || held {
		return err
	}
	return m.engine.PullImage(ctx, image, docker.PullOptions{Platform: platform})
}

// resolveVolume produces the data volume the instance mounts: a snapshot
// clone, the newest reusable volume, or a fresh one.
func (m Manager) resolveVolume(ctx context.Context, spec StartSpec) (docker.VolumeName, error) {
	if spec.Snapshot != "" {
		return m.cloneSnapshot(ctx, spec)
	}
	if sourceOrDefault(spec.Volume) == VolumeReuse {
		existing, reuseErr := m.newestVolume(ctx, spec.Port)
		if reuseErr != nil {
			return "", reuseErr
		}
		if existing != "" {
			return existing, nil
		}
	}
	return m.freshVolume(ctx, spec)
}

// newestVolume finds the most recent managed data volume for a port; empty
// when none exists. Volume names embed a sortable UTC timestamp, so the
// lexicographically greatest name is the newest.
func (m Manager) newestVolume(ctx context.Context, port docker.Port) (docker.VolumeName, error) {
	volumes, err := m.engine.Volumes(ctx, docker.VolumeQuery{Labels: m.portFilter(port)})
	if err != nil {
		return "", err
	}
	newest := docker.VolumeName("")
	for _, volume := range volumes {
		if volume.Name > newest {
			newest = volume.Name
		}
	}
	return newest, nil
}

// freshVolume creates a new labeled data volume for the spec, carrying the
// caller's labels alongside the managed set (managed keys win a collision).
func (m Manager) freshVolume(ctx context.Context, spec StartSpec) (docker.VolumeName, error) {
	name, err := m.volumeName(spec.Port)
	if err != nil {
		return "", err
	}
	created, err := m.engine.CreateVolume(ctx, docker.VolumeSpec{
		Name:   name,
		Labels: m.mergeLabels(m.managedLabels(spec.Port, spec.Image), spec.Labels),
	})
	if err != nil {
		return "", err
	}
	return created.Name, nil
}

// instanceFor assembles the caller-facing instance record, carrying the
// exact label set the container was created with — creation-time truth,
// not a recomputation.
func (m Manager) instanceFor(
	spec StartSpec,
	id docker.ContainerID,
	volume docker.VolumeName,
	labels docker.Labels,
) Instance {
	return Instance{
		Labels:    labels,
		Name:      m.containerName(spec.Port),
		ID:        id,
		Port:      spec.Port,
		Image:     spec.Image,
		Volume:    volume,
		Status:    statusRunning,
		IsRunning: true,
	}
}

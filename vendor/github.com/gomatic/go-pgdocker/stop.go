package pgdocker

import (
	"context"
	"errors"

	docker "github.com/gomatic/go-docker"
)

// StopSpec adjusts an instance stop.
type StopSpec struct {
	// Retain overrides retention; zero defers to the label the container
	// recorded at start, then to the non-destructive keep default.
	Retain Retention
	Grace  docker.StopSeconds
	Port   docker.Port
}

// defaultStopGrace allows postgres a clean shutdown checkpoint.
const defaultStopGrace docker.StopSeconds = 10

// StopReport is the outcome of a stop: what was stopped and what happened
// to each managed volume on the port.
type StopReport struct {
	Stopped        []docker.ContainerName
	RemovedVolumes []docker.VolumeName
	KeptVolumes    []docker.VolumeName
}

// Stop stops the port's managed containers and applies volume retention:
// the explicit spec choice first, then the container's recorded label, then
// the non-destructive keep default. Stopping a port with nothing running
// still applies retention to leftover volumes.
func (m Manager) Stop(ctx context.Context, spec StopSpec) (StopReport, error) {
	if spec.Port == 0 {
		spec.Port = DefaultPort
	}
	if spec.Grace <= 0 {
		spec.Grace = defaultStopGrace
	}
	stopped, labeled, err := m.stopContainers(ctx, spec)
	if err != nil {
		return StopReport{}, err
	}
	retention := m.stopRetention(spec, labeled)
	report, err := m.applyRetention(ctx, spec.Port, retention)
	if err != nil {
		return StopReport{}, err
	}
	report.Stopped = stopped
	return report, nil
}

// stopContainers stops every managed container on the port, waiting for the
// daemon's auto-removal, and reports the retention label recorded at start
// (from the first labeled container found).
func (m Manager) stopContainers(
	ctx context.Context,
	spec StopSpec,
) ([]docker.ContainerName, Retention, error) {
	containers, err := m.engine.Containers(ctx, docker.ContainerQuery{
		Labels:     m.portFilter(spec.Port),
		AllEnabled: true,
	})
	if err != nil {
		return nil, "", err
	}
	stopped := make([]docker.ContainerName, 0, len(containers))
	labeled := Retention("")
	for _, container := range containers {
		if labeled == "" {
			labeled = Retention(m.labelValue(container.Labels, labelVolumeAction))
		}
		if err := m.stopOne(ctx, container.ID, spec.Grace); err != nil {
			return nil, "", err
		}
		stopped = append(stopped, container.Name)
	}
	return stopped, labeled, nil
}

// stopOne stops a container and waits for its auto-removal; a container
// already gone counts as stopped.
func (m Manager) stopOne(
	ctx context.Context,
	id docker.ContainerID,
	grace docker.StopSeconds,
) error {
	if err := m.engine.StopContainer(ctx, id, grace); err != nil {
		if errors.Is(err, docker.ErrNotFound) {
			return nil
		}
		return err
	}
	if _, err := m.engine.WaitContainer(ctx, id, docker.WaitRemoved); err != nil {
		if errors.Is(err, docker.ErrNotFound) {
			return nil
		}
		return err
	}
	return nil
}

// stopRetention resolves the effective retention: explicit spec choice,
// then the container's start-time label, then the non-destructive default.
func (m Manager) stopRetention(spec StopSpec, labeled Retention) Retention {
	if spec.Retain != "" {
		return spec.Retain
	}
	if labeled == RetainKeep || labeled == RetainRemove {
		return labeled
	}
	return retentionOrDefault("")
}

// applyRetention removes or keeps every managed data volume on the port.
func (m Manager) applyRetention(
	ctx context.Context,
	port docker.Port,
	retention Retention,
) (StopReport, error) {
	volumes, err := m.engine.Volumes(ctx, docker.VolumeQuery{Labels: m.portFilter(port)})
	if err != nil {
		return StopReport{}, err
	}
	report := StopReport{}
	for _, volume := range volumes {
		if retention == RetainKeep {
			report.KeptVolumes = append(report.KeptVolumes, volume.Name)
			continue
		}
		if err := m.engine.RemoveVolume(ctx, volume.Name, docker.VolumeRemoval{}); err != nil {
			return StopReport{}, err
		}
		report.RemovedVolumes = append(report.RemovedVolumes, volume.Name)
	}
	return report, nil
}

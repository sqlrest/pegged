package pgdocker

import (
	"context"

	docker "github.com/gomatic/go-docker"
)

// Clone mount points inside the helper container.
const (
	cloneSource docker.MountTarget = "/pgdocker-source"
	cloneTarget docker.MountTarget = "/pgdocker-target"
)

// cloneCommand copies the whole source tree preserving ownership, modes,
// and timestamps — a physical PGDATA copy.
func cloneCommand() docker.Command {
	return docker.Command{
		"sh", "-c",
		"cp -a " + string(cloneSource) + "/. " + string(cloneTarget) + "/",
	}
}

// restoreCommand copies the snapshot tree, then normalizes ownership to the
// destination image's postgres user. The locked-down instance runs without
// the entrypoint's root chown phase that used to absorb a UID mismatch
// (snapshots taken under a different image flavor), so this throwaway root
// helper — running the exact image about to boot — performs that
// normalization instead.
func restoreCommand() docker.Command {
	return docker.Command{
		"sh", "-c",
		"cp -a " + string(cloneSource) + "/. " + string(cloneTarget) + "/" +
			" && chown -R postgres:postgres " + string(cloneTarget),
	}
}

// clone runs command to copy one volume onto another inside a throwaway
// helper container and fails with ErrCloneFailed when the copy exits
// non-zero or dies before reporting.
func (m Manager) clone(
	ctx context.Context,
	image docker.ImageRef,
	command docker.Command,
	source, target docker.VolumeName,
) error {
	if err := m.ensureImage(ctx, image, ""); err != nil {
		return err
	}
	id, err := m.engine.CreateContainer(ctx, docker.ContainerSpec{
		Image:      image,
		Entrypoint: command,
		Labels:     docker.Labels{m.labelKey(labelHelper): labelTrue},
		Mounts: []docker.Mount{
			{
				Kind:            docker.MountVolume,
				Source:          docker.MountSource(source),
				Target:          cloneSource,
				ReadOnlyEnabled: true,
			},
			{
				Kind:   docker.MountVolume,
				Source: docker.MountSource(target),
				Target: cloneTarget,
			},
		},
		AutoRemoveEnabled: true,
	})
	if err != nil {
		return err
	}
	if err = m.engine.StartContainer(ctx, id); err != nil {
		return err
	}
	exit, err := m.engine.WaitContainer(ctx, id, docker.WaitNotRunning)
	if err != nil {
		return err
	}
	if exit != 0 {
		return ErrCloneFailed.With(nil, "exit", int64(exit),
			"source", string(source), "target", string(target))
	}
	return nil
}

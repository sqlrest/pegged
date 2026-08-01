package pgdocker

import (
	"context"
	"errors"
	"time"

	docker "github.com/gomatic/go-docker"
)

// awaitReady polls pg_isready inside the instance container until the
// server accepts connections, the container dies, or the spec's timeout
// elapses (ErrNotReady, carrying the last observed exit code).
func (m Manager) awaitReady(ctx context.Context, id docker.ContainerID, spec StartSpec) error {
	bounded, cancel := context.WithTimeout(ctx, spec.ReadyTimeout)
	defer cancel()
	ticker := time.NewTicker(spec.ReadyInterval)
	defer ticker.Stop()
	for {
		exit, err := m.probeReady(bounded, id, spec)
		if err != nil {
			return err
		}
		if exit == 0 {
			return nil
		}
		if err := awaitTick(bounded, ticker.C, id, exit); err != nil {
			return err
		}
	}
}

// awaitTick waits for the next readiness poll, converting timeout into
// ErrNotReady carrying the last probe's exit code.
func awaitTick(
	ctx context.Context,
	tick <-chan time.Time,
	id docker.ContainerID,
	lastExit docker.ExitCode,
) error {
	select {
	case <-ctx.Done():
		return ErrNotReady.With(ctx.Err(),
			"container", string(id), "last_exit", int64(lastExit))
	case <-tick:
		return nil
	}
}

// probeReady runs one pg_isready check, first confirming the container is
// still alive so a crashed server fails fast with its exit code instead of
// polling into the timeout. A container that died and was auto-remove
// reaped before the probe inspects it is still "died before ready", so the
// not-found race maps to ErrNotReady rather than leaking a raw lookup
// failure.
func (m Manager) probeReady(
	ctx context.Context,
	id docker.ContainerID,
	spec StartSpec,
) (docker.ExitCode, error) {
	details, err := m.engine.InspectContainer(ctx, id)
	if errors.Is(err, docker.ErrNotFound) {
		return 0, ErrNotReady.With(err, "container", string(id), "status", "removed")
	}
	if err != nil {
		return 0, err
	}
	if !details.IsRunning {
		return 0, ErrNotReady.With(nil,
			"container", string(id), "status", string(details.Status),
			"exit", int64(details.ExitCode))
	}
	return m.engine.Exec(ctx, id, readyCommand(spec), docker.ExecOptions{})
}

// readyProbeBinary is the server-shipped readiness checker.
const readyProbeBinary = "pg_isready"

// readyCommand is the pg_isready invocation for the spec's role and
// database.
func readyCommand(spec StartSpec) docker.Command {
	return docker.Command{
		readyProbeBinary,
		"-U", string(spec.User),
		"-d", string(spec.Database),
	}
}

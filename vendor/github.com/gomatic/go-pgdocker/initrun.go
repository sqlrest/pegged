package pgdocker

import (
	"context"
	"io"
	"strconv"

	docker "github.com/gomatic/go-docker"
)

// InitSpec describes an initialization run: a caller-supplied image and
// command executed against a running instance — migrations, seeds, schema
// loads. The init container joins the instance's network namespace, so it
// reaches the server at localhost:5432 on every platform; no
// host.docker.internal, no host networking.
type InitSpec struct {
	Output   io.Writer
	Image    docker.ImageRef
	Database DatabaseName
	User     UserName
	Password Password
	Command  docker.Command
	Env      docker.Env
	Port     docker.Port
}

// Init runs the spec's container against the port's running instance and
// returns the command's exit code. Infrastructure failures are errors; a
// non-zero exit from the init command itself is a result, not an error.
func (m Manager) Init(ctx context.Context, spec InitSpec) (docker.ExitCode, error) {
	spec = initDefaults(spec)
	if spec.Image == "" {
		return 0, ErrInvalidSpec.With(nil, "reason", "an init image is required")
	}
	instance, err := m.runningInstance(ctx, spec.Port)
	if err != nil {
		return 0, err
	}
	if err := m.ensureImage(ctx, spec.Image, ""); err != nil {
		return 0, err
	}
	return m.runInit(ctx, spec, instance)
}

// initDefaults resolves the spec's zero connection fields to the same
// defaults StartSpec uses.
func initDefaults(spec InitSpec) InitSpec {
	if spec.Port == 0 {
		spec.Port = DefaultPort
	}
	if spec.Database == "" {
		spec.Database = defaultDatabase
	}
	if spec.User == "" {
		spec.User = defaultUser
	}
	return spec
}

// runInit creates, runs, drains, and removes the init container. The
// container is kept (not auto-removed) until its logs are read, so output
// is complete rather than racing removal.
func (m Manager) runInit(
	ctx context.Context,
	spec InitSpec,
	instance Instance,
) (docker.ExitCode, error) {
	id, err := m.engine.CreateContainer(ctx, m.initContainerSpec(spec, instance))
	if err != nil {
		return 0, err
	}
	defer m.removeInit(ctx, id)
	if err = m.engine.StartContainer(ctx, id); err != nil {
		return 0, err
	}
	exit, err := m.engine.WaitContainer(ctx, id, docker.WaitNotRunning)
	if err != nil {
		return 0, err
	}
	if err := m.drainInitLogs(ctx, id, spec.Output); err != nil {
		return 0, err
	}
	return exit, nil
}

// initContainerSpec renders the init container: shared network namespace,
// standard PG* variables pointed at localhost, and PGDOCKER_* metadata.
func (m Manager) initContainerSpec(spec InitSpec, instance Instance) docker.ContainerSpec {
	env := docker.Env{
		docker.EnvVar("PGHOST=localhost"),
		docker.EnvVar("PGPORT=" + strconv.Itoa(int(DefaultPort))),
		docker.EnvVar("PGUSER=" + string(spec.User)),
		docker.EnvVar("PGDATABASE=" + string(spec.Database)),
		docker.EnvVar("PGDOCKER_NAMESPACE=" + string(m.namespace)),
		docker.EnvVar("PGDOCKER_PORT=" + strconv.Itoa(int(spec.Port))),
		docker.EnvVar("PGDOCKER_CONTAINER=" + string(instance.Name)),
		docker.EnvVar("PGDOCKER_DATABASE=" + string(spec.Database)),
	}
	if spec.Password != "" {
		env = append(env, docker.EnvVar("PGPASSWORD="+string(spec.Password)))
	}
	return docker.ContainerSpec{
		Image:   spec.Image,
		Command: spec.Command,
		Env:     append(env, spec.Env...),
		Labels:  docker.Labels{m.labelKey(labelHelper): labelTrue},
		Network: docker.NetworkMode("container:" + string(instance.ID)),
	}
}

// drainInitLogs copies the finished run's combined output to the spec's
// writer; a nil writer discards.
func (m Manager) drainInitLogs(
	ctx context.Context,
	id docker.ContainerID,
	output io.Writer,
) error {
	if output == nil {
		output = io.Discard
	}
	return m.engine.ContainerLogs(ctx, id, docker.LogOptions{}, output, output)
}

// removeInit force-removes the finished init container; removal failure is
// not the run's failure.
func (m Manager) removeInit(ctx context.Context, id docker.ContainerID) {
	_ = m.engine.RemoveContainer(ctx, id, docker.RemovalOptions{ForceEnabled: true})
}

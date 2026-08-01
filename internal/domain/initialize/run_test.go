package initialize

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"testing"

	docker "github.com/gomatic/go-docker"
	errs "github.com/gomatic/go-error"
	pgdocker "github.com/gomatic/go-pgdocker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sqlrest/pegged/internal/constants"
	"github.com/sqlrest/pegged/internal/domain"
)

// errSeam is the failure the scripted seams and fakes emit.
const errSeam errs.Const = "seam failure"

// stubEngine is a benign pgdocker.Engine: no containers or volumes exist.
// Scenario fakes embed it and override only the calls they script.
type stubEngine struct{}

func (stubEngine) CreateContainer(context.Context, docker.ContainerSpec) (docker.ContainerID, error) {
	return "container-1", nil
}

func (stubEngine) StartContainer(context.Context, docker.ContainerID) error { return nil }

func (stubEngine) StopContainer(context.Context, docker.ContainerID, docker.StopSeconds) error {
	return nil
}

func (stubEngine) RemoveContainer(context.Context, docker.ContainerID, docker.RemovalOptions) error {
	return nil
}

func (stubEngine) InspectContainer(context.Context, docker.ContainerID) (docker.ContainerDetails, error) {
	return docker.ContainerDetails{Status: "running", IsRunning: true}, nil
}

func (stubEngine) Containers(context.Context, docker.ContainerQuery) ([]docker.ContainerSummary, error) {
	return nil, nil
}

func (stubEngine) WaitContainer(context.Context, docker.ContainerID, docker.WaitCondition) (docker.ExitCode, error) {
	return 0, nil
}

func (stubEngine) ContainerLogs(context.Context, docker.ContainerID, docker.LogOptions, io.Writer, io.Writer) error {
	return nil
}

func (stubEngine) Exec(
	context.Context,
	docker.ContainerID,
	docker.Command,
	docker.ExecOptions,
) (docker.ExitCode, error) {
	return 0, nil
}

func (stubEngine) PullImage(context.Context, docker.ImageRef, docker.PullOptions) error { return nil }

func (stubEngine) ImageExists(context.Context, docker.ImageRef) (bool, error) { return true, nil }

func (stubEngine) CreateVolume(_ context.Context, spec docker.VolumeSpec) (docker.VolumeDetails, error) {
	return docker.VolumeDetails{Name: spec.Name, Labels: spec.Labels}, nil
}

func (stubEngine) Volumes(context.Context, docker.VolumeQuery) ([]docker.VolumeDetails, error) {
	return nil, nil
}

func (stubEngine) RemoveVolume(context.Context, docker.VolumeName, docker.VolumeRemoval) error {
	return nil
}

// runningEngine hosts one running managed instance on 5432, captures the
// init container spec, scripts the init exit code, and streams a line of
// container output.
type runningEngine struct {
	stubEngine
	spec *docker.ContainerSpec
	exit docker.ExitCode
}

func (runningEngine) Containers(context.Context, docker.ContainerQuery) ([]docker.ContainerSummary, error) {
	return []docker.ContainerSummary{{
		Labels: docker.Labels{"pegged.port": "5432"},
		ID:     "instance-1",
		Name:   "pegged-5432",
		Status: "running",
	}}, nil
}

func (e runningEngine) CreateContainer(_ context.Context, spec docker.ContainerSpec) (docker.ContainerID, error) {
	*e.spec = spec
	return "init-1", nil
}

func (e runningEngine) WaitContainer(
	context.Context,
	docker.ContainerID,
	docker.WaitCondition,
) (docker.ExitCode, error) {
	return e.exit, nil
}

func (runningEngine) ContainerLogs(
	_ context.Context,
	_ docker.ContainerID,
	_ docker.LogOptions,
	stdout, _ io.Writer,
) error {
	_, err := stdout.Write([]byte("initialized\n"))
	return err
}

// stubManager reroutes the manager seam to one assembled over the engine,
// restoring the production seam when the test ends; a nil engine scripts a
// construction failure. Tests using it reassign a package var, so they stay
// serial.
func stubManager(t *testing.T, engine pgdocker.Engine) {
	t.Helper()
	original := newManager
	t.Cleanup(func() { newManager = original })
	newManager = func(context.Context) (pgdocker.Manager, error) {
		if engine == nil {
			return pgdocker.Manager{}, errSeam
		}
		return pgdocker.New(pgdocker.Namespace("pegged"), pgdocker.WithEngine{Engine: engine})
	}
}

// discardLogger keeps domain logging out of test output.
func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestRun(t *testing.T) {
	tests := []struct {
		wantErr      error
		engine       pgdocker.Engine
		name         string
		args         []domain.Argument
		config       Config
		wantExit     docker.ExitCode
		wantExitCode bool
	}{
		{
			name:   "runs the init image against the running instance",
			engine: runningEngine{spec: new(docker.ContainerSpec)},
			config: Config{Image: "migrate/migrate:4"},
			args:   []domain.Argument{"5432", "migrate", "up"},
		},
		{
			name:         "a non-zero init exit fails with the exit code",
			engine:       runningEngine{spec: new(docker.ContainerSpec), exit: 7},
			config:       Config{Image: "migrate/migrate:4"},
			wantErr:      constants.ErrInitFailed,
			wantExit:     7,
			wantExitCode: true,
		},
		{
			name:    "the image flag is required",
			engine:  stubEngine{},
			wantErr: constants.ErrMissingArgument,
		},
		{
			name:    "positional port conflicts with the flag",
			engine:  stubEngine{},
			config:  Config{Image: "migrate/migrate:4", Port: 5432},
			args:    []domain.Argument{"5433"},
			wantErr: constants.ErrPortConflict,
		},
		{
			name:    "a numeric first argument above the port range is a typoed port, not a command",
			engine:  stubEngine{},
			config:  Config{Image: "migrate/migrate:4"},
			args:    []domain.Argument{"70000", "echo", "hi"},
			wantErr: constants.ErrInvalidPort,
		},
		{
			name:    "a vacant port has no running instance",
			engine:  stubEngine{},
			config:  Config{Image: "migrate/migrate:4"},
			wantErr: pgdocker.ErrNotRunning,
		},
		{
			name:    "manager construction failure",
			config:  Config{Image: "migrate/migrate:4"},
			wantErr: errSeam,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want, must := assert.New(t), require.New(t)
			stubManager(t, tt.engine)

			result, err := Run(context.Background(), discardLogger(), tt.config, tt.args...)

			if tt.wantErr != nil {
				must.Error(err)
				want.ErrorIs(err, tt.wantErr)
				if tt.wantExitCode {
					want.Equal(tt.wantExit, result.ExitCode,
						"a failed init still reports its exit code")
				}
				return
			}
			must.NoError(err)
			want.Equal(tt.wantExit, result.ExitCode)
		})
	}
}

// TestRunSpecMapping asserts the argv split and the flags land on the init
// container: a leading port peels off, the rest overrides the command, and
// the CLI-injected output writer receives the container's output.
func TestRunSpecMapping(t *testing.T) {
	want, must := assert.New(t), require.New(t)
	spec := new(docker.ContainerSpec)
	var output bytes.Buffer
	stubManager(t, runningEngine{spec: spec})

	config := Config{
		Output:   &output,
		Image:    "migrate/migrate:4",
		Database: "app",
		User:     "dev",
	}
	_, err := Run(context.Background(), discardLogger(), config, "5432", "migrate", "up")
	must.NoError(err)

	want.Equal(docker.ImageRef("migrate/migrate:4"), spec.Image)
	want.Equal(docker.Command{"migrate", "up"}, spec.Command)
	want.Equal(docker.NetworkMode("container:instance-1"), spec.Network)
	want.Contains(spec.Env, docker.EnvVar("PGHOST=localhost"))
	want.Contains(spec.Env, docker.EnvVar("PGDATABASE=app"))
	want.Contains(spec.Env, docker.EnvVar("PGUSER=dev"))
	want.Equal("initialized\n", output.String())
}

// TestRunCommandWithoutPort asserts a non-numeric first argument is treated
// as the command, not a port — including a token with a digit boundary.
func TestRunCommandWithoutPort(t *testing.T) {
	want, must := assert.New(t), require.New(t)
	spec := new(docker.ContainerSpec)
	stubManager(t, runningEngine{spec: spec})

	_, err := Run(context.Background(), discardLogger(),
		Config{Image: "postgres:17-alpine"}, "psql", "-c", "select 1")
	must.NoError(err)
	want.Equal(docker.Command{"psql", "-c", "select 1"}, spec.Command)

	_, err = Run(context.Background(), discardLogger(),
		Config{Image: "postgres:17-alpine"}, "5432x", "up")
	must.NoError(err)
	want.Equal(docker.Command{"5432x", "up"}, spec.Command,
		"a token that is not digits-only is a command, never a port")

	_, err = Run(context.Background(), discardLogger(),
		Config{Image: "postgres:17-alpine"}, "", "up")
	must.NoError(err)
	want.Equal(docker.Command{"", "up"}, spec.Command,
		"an empty token is not a port")
}

package start

import (
	"context"
	"io"
	"log/slog"
	"strings"
	"testing"

	docker "github.com/gomatic/go-docker"
	errs "github.com/gomatic/go-error"
	pgdocker "github.com/gomatic/go-pgdocker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sqlrest/pegged/internal/constants"
	"github.com/sqlrest/pegged/internal/domain"
	"github.com/sqlrest/pegged/internal/manage"
)

// errSeam is the failure the scripted seams and fakes emit.
const errSeam errs.Const = "seam failure"

// stubEngine is a benign pgdocker.Engine: no containers exist, images are
// held, volumes are created as asked, and the readiness probe answers
// immediately. Scenario fakes embed it and override only the calls they
// script.
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

// failingEngine fails the first engine call Start issues.
type failingEngine struct{ stubEngine }

func (failingEngine) Containers(context.Context, docker.ContainerQuery) ([]docker.ContainerSummary, error) {
	return nil, errSeam
}

// captureEngine records the container and volume specs the manager submits.
type captureEngine struct {
	stubEngine
	container *docker.ContainerSpec
	volume    *docker.VolumeSpec
}

func (e captureEngine) CreateContainer(_ context.Context, spec docker.ContainerSpec) (docker.ContainerID, error) {
	*e.container = spec
	return "container-1", nil
}

func (e captureEngine) CreateVolume(_ context.Context, spec docker.VolumeSpec) (docker.VolumeDetails, error) {
	*e.volume = spec
	return docker.VolumeDetails{Name: spec.Name, Labels: spec.Labels}, nil
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
		wantErr error
		engine  pgdocker.Engine
		name    string
		args    []domain.Argument
		config  Config
	}{
		{
			name:   "starts on the default port",
			engine: stubEngine{},
		},
		{
			name:    "non-numeric positional port",
			engine:  stubEngine{},
			args:    []domain.Argument{"nope"},
			wantErr: constants.ErrInvalidPort,
		},
		{
			name:    "positional port conflicts with the flag",
			engine:  stubEngine{},
			args:    []domain.Argument{"5434"},
			config:  Config{Port: 5433},
			wantErr: constants.ErrPortConflict,
		},
		{
			name:    "invalid volume source",
			engine:  stubEngine{},
			config:  Config{Volume: "recycle"},
			wantErr: constants.ErrInvalidValue,
		},
		{
			name:    "invalid retention",
			engine:  stubEngine{},
			config:  Config{Retain: "discard"},
			wantErr: constants.ErrInvalidValue,
		},
		{
			name:    "manager construction failure",
			wantErr: errSeam,
		},
		{
			name:    "engine failure propagates",
			engine:  failingEngine{},
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
				return
			}
			must.NoError(err)
			want.True(result.Instance.IsRunning)
			want.Equal(pgdocker.DefaultPort, result.Instance.Port)
		})
	}
}

// TestRunSpecMapping asserts the flags land on the engine as the intended
// container and volume operations — the CLI-to-library contract.
func TestRunSpecMapping(t *testing.T) {
	want, must := assert.New(t), require.New(t)
	var container docker.ContainerSpec
	var volume docker.VolumeSpec
	stubManager(t, captureEngine{container: &container, volume: &volume})

	config := Config{
		Image:    "postgres:17-alpine",
		Database: "app",
		User:     "dev",
		Retain:   "keep",
	}
	result, err := Run(context.Background(), discardLogger(), config, "5433")
	must.NoError(err)

	want.True(strings.HasPrefix(string(volume.Name), "pegged-5433-"), "volume name %q", volume.Name)
	want.Equal(docker.ContainerName("pegged-5433"), container.Name)
	want.Equal(docker.ImageRef("postgres:17-alpine"), container.Image)
	want.Equal([]docker.PortBinding{{HostAddress: "127.0.0.1", Host: 5433, Container: 5432}},
		container.Ports)
	want.Contains(container.Env, docker.EnvVar("POSTGRES_USER=dev"))
	want.Contains(container.Env, docker.EnvVar("POSTGRES_DB=app"))
	want.Contains(container.Env, docker.EnvVar("POSTGRES_HOST_AUTH_METHOD=trust"))
	want.Equal("keep", container.Labels["pegged.volume.action"])
	must.NotEmpty(container.Mounts)
	want.Equal(docker.MountSource(volume.Name), container.Mounts[0].Source)
	want.Equal(docker.MountTarget("/var/lib/postgresql/data"), container.Mounts[0].Target)
	want.Equal(docker.Port(5433), result.Instance.Port)
	want.Equal(volume.Name, result.Instance.Volume)
}

// TestRunAppliesPortPolicy pins pegged's product policy: with no flags, a
// port above the durable ceiling records remove-on-stop, and one at the
// ceiling records keep — the library itself holds no port opinion.
func TestRunAppliesPortPolicy(t *testing.T) {
	tests := []struct {
		name       string
		port       string
		wantAction string
	}{
		{name: "ephemeral port records remove", port: "5433", wantAction: "remove"},
		{name: "durable port records keep", port: "5432", wantAction: "keep"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want, must := assert.New(t), require.New(t)
			var container docker.ContainerSpec
			var volume docker.VolumeSpec
			stubManager(t, captureEngine{container: &container, volume: &volume})
			_, err := Run(context.Background(), discardLogger(), Config{}, tt.port)
			must.NoError(err)
			want.Equal(tt.wantAction, container.Labels["pegged.volume.action"])
		})
	}
}

// TestRunStampsProvenance pins the provenance contract: the container and
// its data volume both carry exactly the label set the provenance seam
// produces — a stubbed stamp, so a hardcoded value in the spec builder
// cannot pass. Seam-reassigning, so serial with t.Cleanup restore.
func TestRunStampsProvenance(t *testing.T) {
	want, must := assert.New(t), require.New(t)
	previous := provenanceLabels
	provenanceLabels = func() docker.Labels { return docker.Labels{manage.VersionLabel: "vSTAMP"} }
	t.Cleanup(func() { provenanceLabels = previous })
	var container docker.ContainerSpec
	var volume docker.VolumeSpec
	stubManager(t, captureEngine{container: &container, volume: &volume})
	_, err := Run(context.Background(), discardLogger(), Config{}, "5433")
	must.NoError(err)
	want.Equal("vSTAMP", container.Labels[manage.VersionLabel])
	want.Equal("vSTAMP", volume.Labels[manage.VersionLabel])
}

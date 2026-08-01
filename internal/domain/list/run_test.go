package list

import (
	"context"
	"io"
	"log/slog"
	"testing"

	docker "github.com/gomatic/go-docker"
	errs "github.com/gomatic/go-error"
	pgdocker "github.com/gomatic/go-pgdocker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// occupiedEngine hosts running managed instances on two ports.
type occupiedEngine struct{ stubEngine }

func (occupiedEngine) Containers(context.Context, docker.ContainerQuery) ([]docker.ContainerSummary, error) {
	return []docker.ContainerSummary{
		{
			Labels: docker.Labels{"pegged.port": "5433"},
			ID:     "container-2",
			Name:   "pegged-5433",
			Status: "exited",
		},
		{
			Labels: docker.Labels{"pegged.port": "5432"},
			ID:     "container-1",
			Name:   "pegged-5432",
			Status: "running",
		},
	}, nil
}

// failingEngine fails the first engine call List issues.
type failingEngine struct{ stubEngine }

func (failingEngine) Containers(context.Context, docker.ContainerQuery) ([]docker.ContainerSummary, error) {
	return nil, errSeam
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
		wantErr   error
		engine    pgdocker.Engine
		name      string
		wantPorts []docker.Port
	}{
		{
			name:      "reports every managed port ordered by port",
			engine:    occupiedEngine{},
			wantPorts: []docker.Port{5432, 5433},
		},
		{
			name:      "an empty daemon yields no reports",
			engine:    stubEngine{},
			wantPorts: []docker.Port{},
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

			result, err := Run(context.Background(), discardLogger(), Config{})

			if tt.wantErr != nil {
				must.Error(err)
				want.ErrorIs(err, tt.wantErr)
				return
			}
			must.NoError(err)
			ports := make([]docker.Port, 0, len(result.Reports))
			for _, report := range result.Reports {
				ports = append(ports, report.Port)
			}
			want.Equal(tt.wantPorts, ports)
		})
	}
}

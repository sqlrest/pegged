package create

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

	"github.com/sqlrest/pegged/internal/constants"
	"github.com/sqlrest/pegged/internal/domain"
	"github.com/sqlrest/pegged/internal/manage"
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

// snapshotEngine hosts a stopped instance's data volume on 5432 and records
// the snapshot volume and clone container the manager creates. Volumes
// answers the snapshot-name filter with the created snapshot (once created)
// and the port filter with the source data volume.
type snapshotEngine struct {
	stubEngine
	created    *docker.VolumeSpec
	cloneImage *docker.ImageRef
}

func (e snapshotEngine) CreateVolume(_ context.Context, spec docker.VolumeSpec) (docker.VolumeDetails, error) {
	*e.created = spec
	return docker.VolumeDetails{Name: spec.Name, Labels: spec.Labels}, nil
}

func (e snapshotEngine) CreateContainer(_ context.Context, spec docker.ContainerSpec) (docker.ContainerID, error) {
	*e.cloneImage = spec.Image
	return "clone-1", nil
}

func (e snapshotEngine) Volumes(_ context.Context, query docker.VolumeQuery) ([]docker.VolumeDetails, error) {
	if _, bySnapshot := query.Labels["pegged.snapshot.name"]; bySnapshot {
		if e.created.Name == "" {
			return nil, nil
		}
		return []docker.VolumeDetails{{Labels: e.created.Labels, Name: e.created.Name}}, nil
	}
	return []docker.VolumeDetails{{
		Labels: docker.Labels{
			"pegged.port":           "5432",
			"pegged.postgres.image": "postgres:17-alpine",
			"pegged.postgres.major": "17",
		},
		Name: "pegged-5432-20260731T120000-000000",
	}}, nil
}

// failingEngine fails the first engine call SnapshotCreate issues.
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
		wantErr error
		engine  pgdocker.Engine
		name    string
		args    []domain.Argument
		config  Config
	}{
		{
			name:    "a port with no data volume cannot be snapshotted",
			engine:  stubEngine{},
			wantErr: pgdocker.ErrNoDataVolume,
		},
		{
			name:    "non-numeric positional port",
			engine:  stubEngine{},
			args:    []domain.Argument{"nope"},
			wantErr: constants.ErrInvalidPort,
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

			_, err := Run(context.Background(), discardLogger(), tt.config, tt.args...)

			must.Error(err)
			want.ErrorIs(err, tt.wantErr)
		})
	}
}

// TestRunCreatesSnapshot asserts the happy path: the named snapshot is
// cloned from the port's data volume with the source's postgres lineage.
func TestRunCreatesSnapshot(t *testing.T) {
	want, must := assert.New(t), require.New(t)
	var created docker.VolumeSpec
	var cloneImage docker.ImageRef
	stubManager(t, snapshotEngine{created: &created, cloneImage: &cloneImage})

	result, err := Run(context.Background(), discardLogger(), Config{}, "5432", "golden")
	must.NoError(err)

	want.Equal(pgdocker.SnapshotName("golden"), result.Snapshot.Name)
	want.Equal(docker.VolumeName("pegged-snapshot-golden"), result.Snapshot.Volume)
	want.Equal(pgdocker.PostgresMajor("17"), result.Snapshot.Major)
	want.Equal(docker.Port(5432), result.Snapshot.SourcePort)
	want.Equal("golden", created.Labels["pegged.snapshot.name"])
	want.Equal(docker.ImageRef("postgres:17-alpine"), cloneImage,
		"the clone helper runs the source volume's recorded image")
}

// TestRunDefaultName asserts an omitted snapshot name defers to the
// library's UTC-timestamp default rather than any fixed value.
func TestRunDefaultName(t *testing.T) {
	want, must := assert.New(t), require.New(t)
	var created docker.VolumeSpec
	var cloneImage docker.ImageRef
	stubManager(t, snapshotEngine{created: &created, cloneImage: &cloneImage})

	result, err := Run(context.Background(), discardLogger(), Config{}, "5432")
	must.NoError(err)
	want.Regexp(`^\d{8}T\d{6}$`, string(result.Snapshot.Name))
}

// TestRunCopyImage asserts --copy-image overrides the clone helper image.
func TestRunCopyImage(t *testing.T) {
	want, must := assert.New(t), require.New(t)
	var created docker.VolumeSpec
	var cloneImage docker.ImageRef
	stubManager(t, snapshotEngine{created: &created, cloneImage: &cloneImage})

	_, err := Run(context.Background(), discardLogger(),
		Config{CopyImage: "alpine:3.20"}, "5432", "golden")
	must.NoError(err)
	want.Equal(docker.ImageRef("alpine:3.20"), cloneImage)
}

// TestRunStampsProvenance pins the provenance contract: the snapshot volume
// carries exactly the label the provenance seam produces — a stubbed
// stamp, so a hardcoded value in the spec builder cannot pass.
// Seam-reassigning, so serial with t.Cleanup restore.
func TestRunStampsProvenance(t *testing.T) {
	want, must := assert.New(t), require.New(t)
	previous := provenanceLabels
	provenanceLabels = func() docker.Labels { return docker.Labels{manage.VersionLabel: "vSTAMP"} }
	t.Cleanup(func() { provenanceLabels = previous })
	var created docker.VolumeSpec
	var cloneImage docker.ImageRef
	stubManager(t, snapshotEngine{created: &created, cloneImage: &cloneImage})
	_, err := Run(context.Background(), discardLogger(), Config{}, "5432", "golden")
	must.NoError(err)
	want.Equal("vSTAMP", created.Labels[manage.VersionLabel])
}

package pgdocker

import (
	"context"
	"io"

	docker "github.com/gomatic/go-docker"
)

// Engine is the slice of the Docker Engine client this package drives. It is
// the package's one external seam: production passes a docker.Client, tests
// pass a fake, and nothing here ever reaches the daemon another way.
type Engine interface {
	CreateContainer(ctx context.Context, spec docker.ContainerSpec) (docker.ContainerID, error)
	StartContainer(ctx context.Context, id docker.ContainerID) error
	StopContainer(ctx context.Context, id docker.ContainerID, grace docker.StopSeconds) error
	RemoveContainer(ctx context.Context, id docker.ContainerID, options docker.RemovalOptions) error
	InspectContainer(ctx context.Context, id docker.ContainerID) (docker.ContainerDetails, error)
	Containers(ctx context.Context, query docker.ContainerQuery) ([]docker.ContainerSummary, error)
	WaitContainer(
		ctx context.Context,
		id docker.ContainerID,
		condition docker.WaitCondition,
	) (docker.ExitCode, error)
	ContainerLogs(
		ctx context.Context,
		id docker.ContainerID,
		options docker.LogOptions,
		stdout, stderr io.Writer,
	) error
	Exec(
		ctx context.Context,
		id docker.ContainerID,
		command docker.Command,
		options docker.ExecOptions,
	) (docker.ExitCode, error)
	PullImage(ctx context.Context, image docker.ImageRef, options docker.PullOptions) error
	ImageExists(ctx context.Context, image docker.ImageRef) (bool, error)
	CreateVolume(ctx context.Context, spec docker.VolumeSpec) (docker.VolumeDetails, error)
	Volumes(ctx context.Context, query docker.VolumeQuery) ([]docker.VolumeDetails, error)
	RemoveVolume(ctx context.Context, name docker.VolumeName, options docker.VolumeRemoval) error
}

// The production client satisfies the seam by construction.
var _ Engine = docker.Client{}

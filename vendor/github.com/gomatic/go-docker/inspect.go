package docker

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

// containerDetailsWire mirrors the subset of the Engine's container inspect
// document this package exposes. Wire keys are the daemon's; decode-only.
type containerDetailsWire struct {
	Config containerConfWire  `json:"Config"`
	ID     string             `json:"Id"`
	Name   string             `json:"Name"`
	Mounts []mountWire        `json:"Mounts"`
	State  containerStateWire `json:"State"`
}

// containerStateWire mirrors the inspect State document.
type containerStateWire struct {
	Status    string
	ExitCode  int64 `json:"ExitCode"`
	IsRunning bool  `json:"Running"`
}

// containerConfWire mirrors the inspect Config document.
type containerConfWire struct {
	Labels map[string]string
	Image  string
}

// mountWire mirrors one inspect Mounts entry.
type mountWire struct {
	Kind        string `json:"Type"`
	Name        string
	Source      string
	Destination string
}

// ContainerDetails is the inspected state of one container.
type ContainerDetails struct {
	Labels    Labels
	ID        ContainerID
	Name      ContainerName
	Image     ImageRef
	Status    ContainerStatus
	Mounts    []MountDetails
	ExitCode  ExitCode
	IsRunning bool
}

// ContainerStatus is the Engine's lifecycle state word ("created",
// "running", "exited", ...).
type ContainerStatus string

// ExitCode is a container or exec process exit status.
type ExitCode int64

// MountDetails describes one mount attached to an inspected container.
type MountDetails struct {
	Kind   MountKind
	Volume VolumeName
	Source MountSource
	Target MountTarget
}

// InspectContainer returns the current state of one container.
func (c Client) InspectContainer(ctx context.Context, id ContainerID) (ContainerDetails, error) {
	response, err := c.request(
		ctx, http.MethodGet, containerPath(id)+"/json", nil, nil, http.StatusOK,
	)
	if err != nil {
		return ContainerDetails{}, err
	}
	defer func() { _ = response.Body.Close() }()
	raw, err := readBody(response.Body)
	if err != nil {
		return ContainerDetails{}, err
	}
	var wire containerDetailsWire
	if err := json.Unmarshal(raw, &wire); err != nil {
		return ContainerDetails{}, ErrDecodeResponse.With(err)
	}
	return detailsFrom(wire), nil
}

// detailsFrom maps the wire document onto the public details.
func detailsFrom(wire containerDetailsWire) ContainerDetails {
	return ContainerDetails{
		ID:        ContainerID(wire.ID),
		Name:      ContainerName(strings.TrimPrefix(wire.Name, "/")),
		Image:     ImageRef(wire.Config.Image),
		Status:    ContainerStatus(wire.State.Status),
		ExitCode:  ExitCode(wire.State.ExitCode),
		Labels:    Labels(wire.Config.Labels),
		Mounts:    mountDetails(wire.Mounts),
		IsRunning: wire.State.IsRunning,
	}
}

// mountDetails maps wire mounts onto the public shape. Volume carries the
// volume name for volume mounts and is empty for binds.
func mountDetails(wire []mountWire) []MountDetails {
	if len(wire) == 0 {
		return nil
	}
	mounts := make([]MountDetails, len(wire))
	for i, mount := range wire {
		mounts[i] = MountDetails{
			Kind:   MountKind(mount.Kind),
			Volume: VolumeName(mount.Name),
			Source: MountSource(mount.Source),
			Target: MountTarget(mount.Destination),
		}
	}
	return mounts
}

package docker

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

// VolumeName identifies a named Docker volume.
type VolumeName string

// VolumeSpec describes a volume to create.
type VolumeSpec struct {
	Labels Labels
	Name   VolumeName
}

// volumeWire mirrors the Engine's volume document; decode-only.
type volumeWire struct {
	Labels     map[string]string
	Name       string
	Mountpoint string
	CreatedAt  string
}

// VolumeDetails is one volume's state.
type VolumeDetails struct {
	Labels     Labels
	Name       VolumeName
	Mountpoint MountSource
	CreatedAt  VolumeCreatedAt
}

// VolumeCreatedAt is the Engine's volume creation timestamp, as reported
// (RFC 3339 text).
type VolumeCreatedAt string

// CreateVolume creates a named volume; creating an existing name returns the
// existing volume unchanged (Engine semantics).
func (c Client) CreateVolume(ctx context.Context, spec VolumeSpec) (VolumeDetails, error) {
	body := wireBody{string(wireKeyName): string(spec.Name)}
	setLabels(body, spec.Labels)
	response, err := c.request(
		ctx, http.MethodPost, "/volumes/create", nil, body, http.StatusCreated,
	)
	if err != nil {
		return VolumeDetails{}, err
	}
	defer func() { _ = response.Body.Close() }()
	raw, err := readBody(response.Body)
	if err != nil {
		return VolumeDetails{}, err
	}
	var wire volumeWire
	if err := json.Unmarshal(raw, &wire); err != nil {
		return VolumeDetails{}, ErrDecodeResponse.With(err)
	}
	return volumeFrom(wire), nil
}

// volumeFrom maps the wire document onto the public shape.
func volumeFrom(wire volumeWire) VolumeDetails {
	return VolumeDetails{
		Name:       VolumeName(wire.Name),
		Mountpoint: MountSource(wire.Mountpoint),
		CreatedAt:  VolumeCreatedAt(wire.CreatedAt),
		Labels:     Labels(wire.Labels),
	}
}

// InspectVolume returns one volume's state.
func (c Client) InspectVolume(ctx context.Context, name VolumeName) (VolumeDetails, error) {
	response, err := c.request(ctx, http.MethodGet, volumePath(name), nil, nil, http.StatusOK)
	if err != nil {
		return VolumeDetails{}, err
	}
	defer func() { _ = response.Body.Close() }()
	raw, err := readBody(response.Body)
	if err != nil {
		return VolumeDetails{}, err
	}
	var wire volumeWire
	if err := json.Unmarshal(raw, &wire); err != nil {
		return VolumeDetails{}, ErrDecodeResponse.With(err)
	}
	return volumeFrom(wire), nil
}

// volumePath is the API path for one volume.
func volumePath(name VolumeName) requestPath {
	return requestPath("/volumes/" + url.PathEscape(string(name)))
}

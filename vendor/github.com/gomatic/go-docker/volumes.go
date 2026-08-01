package docker

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
)

// VolumeQuery filters volume listings by label terms; zero value lists all.
type VolumeQuery struct {
	Labels Labels
}

// volumeListWire mirrors the Engine's volume list envelope; decode-only.
type volumeListWire struct {
	Volumes []volumeWire
}

// Volumes lists volumes matching the query.
func (c Client) Volumes(ctx context.Context, query VolumeQuery) ([]VolumeDetails, error) {
	values := url.Values{}
	if filters, err := listFilters(query.Labels, ""); err != nil {
		return nil, err
	} else if filters != "" {
		values.Set("filters", string(filters))
	}
	response, err := c.request(ctx, http.MethodGet, "/volumes", values, nil, http.StatusOK)
	if err != nil {
		return nil, err
	}
	defer func() { _ = response.Body.Close() }()
	raw, err := readBody(response.Body)
	if err != nil {
		return nil, err
	}
	var wire volumeListWire
	if err := json.Unmarshal(raw, &wire); err != nil {
		return nil, ErrDecodeResponse.With(err)
	}
	return volumesFrom(wire.Volumes), nil
}

// volumesFrom maps wire volumes onto the public shape.
func volumesFrom(wire []volumeWire) []VolumeDetails {
	volumes := make([]VolumeDetails, len(wire))
	for i, entry := range wire {
		volumes[i] = volumeFrom(entry)
	}
	return volumes
}

// VolumeRemoval adjusts volume removal.
type VolumeRemoval struct {
	// ForceEnabled removes the volume even when the daemon believes it is
	// in use.
	ForceEnabled bool
}

// RemoveVolume deletes a named volume.
func (c Client) RemoveVolume(ctx context.Context, name VolumeName, options VolumeRemoval) error {
	query := url.Values{}
	query.Set("force", strconv.FormatBool(options.ForceEnabled))
	response, err := c.request(
		ctx, http.MethodDelete, volumePath(name), query, nil, http.StatusNoContent,
	)
	if err != nil {
		return err
	}
	defer func() { _ = response.Body.Close() }()
	drainBody(response.Body)
	return nil
}

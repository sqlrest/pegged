package docker

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

// StartContainer starts a created container.
func (c Client) StartContainer(ctx context.Context, id ContainerID) error {
	return c.simpleContainerPost(ctx, id, "/start", nil)
}

// StopSeconds is the grace period the daemon allows a container before
// killing it on stop.
type StopSeconds int

// StopContainer stops a running container, allowing it grace seconds to exit.
// Stopping an already-stopped container succeeds.
func (c Client) StopContainer(ctx context.Context, id ContainerID, grace StopSeconds) error {
	query := url.Values{}
	query.Set("t", strconv.Itoa(int(grace)))
	return c.simpleContainerPost(ctx, id, "/stop", query)
}

// simpleContainerPost issues a body-less container action accepting the
// Engine's no-content and not-modified outcomes.
func (c Client) simpleContainerPost(
	ctx context.Context,
	id ContainerID,
	action requestPath,
	query url.Values,
) error {
	response, err := c.request(
		ctx, http.MethodPost, containerPath(id)+action, query, nil,
		http.StatusNoContent, http.StatusNotModified,
	)
	if err != nil {
		return err
	}
	defer func() { _ = response.Body.Close() }()
	drainBody(response.Body)
	return nil
}

// containerPath is the API path prefix for one container.
func containerPath(id ContainerID) requestPath {
	return requestPath("/containers/" + url.PathEscape(string(id)))
}

// RemovalOptions adjusts container removal.
type RemovalOptions struct {
	ForceEnabled         bool
	RemoveVolumesEnabled bool
}

// RemoveContainer removes a container. Force removal stops a running
// container first; volume removal also deletes its anonymous volumes.
func (c Client) RemoveContainer(ctx context.Context, id ContainerID, options RemovalOptions) error {
	query := url.Values{}
	query.Set("force", strconv.FormatBool(options.ForceEnabled))
	query.Set("v", strconv.FormatBool(options.RemoveVolumesEnabled))
	response, err := c.request(
		ctx, http.MethodDelete, containerPath(id), query, nil, http.StatusNoContent,
	)
	if err != nil {
		return err
	}
	defer func() { _ = response.Body.Close() }()
	drainBody(response.Body)
	return nil
}

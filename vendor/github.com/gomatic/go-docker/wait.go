package docker

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// WaitCondition names the container state transition WaitContainer blocks on.
type WaitCondition string

// Wait conditions the Engine accepts.
const (
	WaitNotRunning WaitCondition = "not-running"
	WaitNextExit   WaitCondition = "next-exit"
	WaitRemoved    WaitCondition = "removed"
)

// waitResultWire mirrors the Engine's wait response; decode-only.
type waitResultWire struct {
	Failure    *waitErrorWire `json:"Error"`
	StatusCode int64          `json:"StatusCode"`
}

// waitErrorWire mirrors the wait response's error document.
type waitErrorWire struct {
	Message string
}

// WaitContainer blocks until the container reaches the condition and returns
// its exit code. Cancellation is the caller's context; there is no internal
// polling or recursion.
func (c Client) WaitContainer(
	ctx context.Context,
	id ContainerID,
	condition WaitCondition,
) (ExitCode, error) {
	query := url.Values{}
	query.Set("condition", string(condition))
	response, err := c.request(
		ctx, http.MethodPost, containerPath(id)+"/wait", query, nil, http.StatusOK,
	)
	if err != nil {
		return 0, err
	}
	defer func() { _ = response.Body.Close() }()
	raw, err := readBody(response.Body)
	if err != nil {
		return 0, err
	}
	var wire waitResultWire
	if err := json.Unmarshal(raw, &wire); err != nil {
		return 0, ErrDecodeResponse.With(err)
	}
	if wire.Failure != nil && wire.Failure.Message != "" {
		return 0, ErrServer.With(nil, "message", wire.Failure.Message)
	}
	return ExitCode(wire.StatusCode), nil
}

// LogOptions adjusts container log retrieval.
type LogOptions struct {
	// Since limits logs to entries after this time; zero means from the
	// container's start.
	Since time.Time
	// FollowEnabled streams logs until the container stops instead of
	// returning the current backlog.
	FollowEnabled bool
}

// ContainerLogs copies a container's stdout and stderr onto the two writers,
// demultiplexing the Engine's framed stream. It returns when the backlog is
// exhausted — or, when following, when the container stops or the context is
// canceled.
func (c Client) ContainerLogs(
	ctx context.Context,
	id ContainerID,
	options LogOptions,
	stdout, stderr io.Writer,
) error {
	response, err := c.request(
		ctx, http.MethodGet, containerPath(id)+"/logs", logQuery(options), nil, http.StatusOK,
	)
	if err != nil {
		return err
	}
	defer func() { _ = response.Body.Close() }()
	return DemuxStream(response.Body, stdout, stderr)
}

// logQuery renders the logs endpoint's query parameters.
func logQuery(options LogOptions) url.Values {
	query := url.Values{}
	query.Set("stdout", "true")
	query.Set("stderr", "true")
	query.Set("follow", strconv.FormatBool(options.FollowEnabled))
	if !options.Since.IsZero() {
		query.Set("since", strconv.FormatInt(options.Since.Unix(), 10))
	}
	return query
}

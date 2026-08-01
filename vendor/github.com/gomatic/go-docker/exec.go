package docker

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// Exec runs a command inside a running container and waits for its exit
// code by polling the exec inspection endpoint — no connection hijacking, so
// it works over every transport this package speaks. Output is not captured;
// callers needing output stream ContainerLogs or run a dedicated container.

// ExecOptions adjusts an in-container command run.
type ExecOptions struct {
	// User runs the command as this user; empty inherits the container's
	// configured user.
	User ContainerUser
	// Env adds environment entries visible to the command.
	Env Env
	// PollInterval is the delay between exit-code polls; zero uses
	// defaultExecPollInterval.
	PollInterval time.Duration
}

// defaultExecPollInterval paces exec completion polling when the caller does
// not choose a pace.
const defaultExecPollInterval = 100 * time.Millisecond

// execCreatedWire mirrors the exec create response; decode-only.
type execCreatedWire struct {
	ID string `json:"Id"`
}

// execStateWire mirrors the exec inspect response; decode-only.
type execStateWire struct {
	IsRunning bool  `json:"Running"`
	ExitCode  int64 `json:"ExitCode"`
}

// Exec runs command inside the container and returns its exit code, polling
// until the command finishes or the context is canceled.
func (c Client) Exec(
	ctx context.Context,
	id ContainerID,
	command Command,
	options ExecOptions,
) (ExitCode, error) {
	execID, err := c.createExec(ctx, id, command, options)
	if err != nil {
		return 0, err
	}
	if err := c.startExec(ctx, execID); err != nil {
		return 0, err
	}
	return c.awaitExec(ctx, execID, pollInterval(options.PollInterval))
}

// execID identifies one exec instance on the daemon.
type execID string

// createExec registers the command with the daemon.
func (c Client) createExec(
	ctx context.Context,
	id ContainerID,
	command Command,
	options ExecOptions,
) (execID, error) {
	body := wireBody{string(wireKeyCmd): []string(command)}
	setEnv(body, options.Env)
	if options.User != "" {
		body["User"] = string(options.User)
	}
	response, err := c.request(
		ctx, http.MethodPost, containerPath(id)+"/exec", nil, body, http.StatusCreated,
	)
	if err != nil {
		return "", err
	}
	defer func() { _ = response.Body.Close() }()
	raw, err := readBody(response.Body)
	if err != nil {
		return "", err
	}
	var created execCreatedWire
	if err := json.Unmarshal(raw, &created); err != nil {
		return "", ErrDecodeResponse.With(err)
	}
	return execID(created.ID), nil
}

// startExec launches a registered exec detached; completion is observed by
// polling, not by holding the connection.
func (c Client) startExec(ctx context.Context, id execID) error {
	response, err := c.request(
		ctx, http.MethodPost, execPath(id)+"/start", nil,
		wireBody{"Detach": true}, http.StatusOK, http.StatusNoContent,
	)
	if err != nil {
		return err
	}
	defer func() { _ = response.Body.Close() }()
	drainBody(response.Body)
	return nil
}

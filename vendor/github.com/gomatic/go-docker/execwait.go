package docker

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"time"
)

// pollInterval applies the default poll pace when the caller chose none.
func pollInterval(chosen time.Duration) time.Duration {
	if chosen <= 0 {
		return defaultExecPollInterval
	}
	return chosen
}

// awaitExec polls the exec state until the command exits, pacing polls with
// the interval and aborting on context cancellation.
func (c Client) awaitExec(
	ctx context.Context,
	id execID,
	interval time.Duration,
) (ExitCode, error) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		state, err := c.inspectExec(ctx, id)
		if err != nil {
			return 0, err
		}
		if !state.IsRunning {
			return ExitCode(state.ExitCode), nil
		}
		select {
		case <-ctx.Done():
			return 0, ErrCanceled.With(ctx.Err(), "exec", string(id))
		case <-ticker.C:
		}
	}
}

// inspectExec reads one exec instance's state.
func (c Client) inspectExec(ctx context.Context, id execID) (execStateWire, error) {
	response, err := c.request(ctx, http.MethodGet, execPath(id)+"/json", nil, nil, http.StatusOK)
	if err != nil {
		return execStateWire{}, err
	}
	defer func() { _ = response.Body.Close() }()
	raw, err := readBody(response.Body)
	if err != nil {
		return execStateWire{}, err
	}
	var state execStateWire
	if err := json.Unmarshal(raw, &state); err != nil {
		return execStateWire{}, ErrDecodeResponse.With(err)
	}
	return state, nil
}

// execPath is the API path prefix for one exec instance. Exec identifiers
// are daemon-generated hex, but they are escaped anyway so a corrupt value
// cannot reshape the request path.
func execPath(id execID) requestPath {
	return requestPath("/exec/" + url.PathEscape(string(id)))
}

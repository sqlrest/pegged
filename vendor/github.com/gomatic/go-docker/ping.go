package docker

import (
	"context"
	"net/http"
)

// Ping verifies daemon connectivity and returns the API version the daemon
// reports (empty when the daemon predates the header).
func (c Client) Ping(ctx context.Context) (APIVersion, error) {
	response, err := c.request(ctx, http.MethodGet, "/_ping", nil, nil, http.StatusOK)
	if err != nil {
		return "", ErrPing.With(err)
	}
	defer func() { _ = response.Body.Close() }()
	drainBody(response.Body)
	return APIVersion(response.Header.Get("Api-Version")), nil
}

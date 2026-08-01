package docker

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
)

// RegistryAuth is a pre-encoded X-Registry-Auth header value (the Engine's
// base64url auth document). Empty means anonymous.
type RegistryAuth string

// PullOptions adjusts an image pull.
type PullOptions struct {
	Progress StreamHandler
	Platform Platform
	Auth     RegistryAuth
}

// PullImage pulls an image, scanning the Engine's progress stream so a pull
// that fails mid-stream — auth failure, missing manifest, platform mismatch —
// returns an error instead of a silent HTTP 200 success.
func (c Client) PullImage(ctx context.Context, image ImageRef, options PullOptions) error {
	response, err := c.pullRequest(ctx, image, options)
	if err != nil {
		return ErrPull.With(err, "image", string(image))
	}
	defer func() { _ = response.Body.Close() }()
	if err := ScanStream(response.Body, options.Progress); err != nil {
		return ErrPull.With(err, "image", string(image))
	}
	return nil
}

// pullRequest issues the images/create call with the pull query and auth
// header.
func (c Client) pullRequest(
	ctx context.Context,
	image ImageRef,
	options PullOptions,
) (*http.Response, error) {
	name, tag := splitImageRef(image)
	query := url.Values{}
	query.Set("fromImage", name)
	query.Set("tag", tag)
	if options.Platform != "" {
		query.Set("platform", string(options.Platform))
	}
	return c.requestWithAuth(ctx, "/images/create", query, options.Auth)
}

// requestWithAuth issues a POST carrying the registry auth header when
// present.
func (c Client) requestWithAuth(
	ctx context.Context,
	path requestPath,
	query url.Values,
	auth RegistryAuth,
) (*http.Response, error) {
	request, err := http.NewRequestWithContext(
		ctx, http.MethodPost, c.apiURL(path, query), nil,
	)
	if err != nil {
		return nil, ErrConnect.With(err, "path", string(path))
	}
	if auth != "" {
		request.Header.Set("X-Registry-Auth", string(auth))
	}
	response, err := c.doer.Do(request)
	if err != nil {
		return nil, ErrConnect.With(err, "path", string(path))
	}
	return acceptResponse(response, []HTTPStatus{http.StatusOK})
}

// defaultImageTag is the tag assumed for an untagged image reference.
const defaultImageTag = "latest"

// splitImageRef separates an image reference into the Engine's fromImage and
// tag query values. A digest reference ("name@sha256:...") yields the digest
// as the tag; an untagged reference defaults to "latest".
func splitImageRef(image ImageRef) (string, string) {
	ref := string(image)
	if name, digest, found := strings.Cut(ref, "@"); found {
		return name, digest
	}
	slash := strings.LastIndex(ref, "/")
	if colon := strings.LastIndex(ref, ":"); colon > slash {
		return ref[:colon], ref[colon+1:]
	}
	return ref, defaultImageTag
}

// ImageExists reports whether the daemon holds the image locally.
func (c Client) ImageExists(ctx context.Context, image ImageRef) (bool, error) {
	response, err := c.request(
		ctx, http.MethodGet, imagePath(image)+"/json", nil, nil, http.StatusOK,
	)
	if errors.Is(err, ErrNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	defer func() { _ = response.Body.Close() }()
	drainBody(response.Body)
	return true, nil
}

// imagePath is the API path prefix for one image. Image references carry
// literal slashes and colons (registry ports, repository paths) that the
// Engine router expects unescaped, so only the remaining reserved characters
// are escaped.
func imagePath(image ImageRef) requestPath {
	escaped := strings.ReplaceAll(url.PathEscape(string(image)), "%2F", "/")
	return requestPath("/images/" + escaped)
}

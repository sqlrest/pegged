package docker

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
)

// requestPath is an Engine API path relative to the negotiated version prefix,
// like "/containers/create".
type requestPath string

// wireBody is a JSON request body assembled with the Engine's exact wire
// keys. Bodies are built as maps rather than tagged structs, keeping the
// daemon-owned PascalCase wire vocabulary out of this module's type surface.
type wireBody map[string]any

// marshalBody is the JSON encoder seam for request bodies. Production uses
// json.Marshal; tests reassign it to exercise the encode failure branch,
// which is otherwise unreachable with valid bodies.
var marshalBody = json.Marshal

// apiURL joins the endpoint base, version prefix, path, and query.
func (c Client) apiURL(path requestPath, query url.Values) string {
	versioned := string(c.endpoint.base) + "/v" + string(c.version) + string(path)
	if len(query) == 0 {
		return versioned
	}
	return versioned + "?" + query.Encode()
}

// request issues one Engine API call and returns the response when its status
// is one of wanted. Any other status consumes the body and maps the Engine's
// error envelope onto this package's sentinels. The caller owns the returned
// body and must close it.
func (c Client) request(
	ctx context.Context,
	method HTTPMethod,
	path requestPath,
	query url.Values,
	body wireBody,
	wanted ...HTTPStatus,
) (*http.Response, error) {
	encoded, err := encodeBody(body)
	if err != nil {
		return nil, err
	}
	response, err := c.send(ctx, method, path, query, encoded)
	if err != nil {
		return nil, err
	}
	return acceptResponse(response, wanted)
}

// HTTPMethod is an HTTP request method.
type HTTPMethod string

// HTTPStatus is an HTTP response status code.
type HTTPStatus int

// encodeBody marshals a wire body, or yields a nil reader for a nil body.
func encodeBody(body wireBody) (io.Reader, error) {
	if body == nil {
		return nil, nil
	}
	encoded, err := marshalBody(body)
	if err != nil {
		return nil, ErrEncodeRequest.With(err)
	}
	return bytes.NewReader(encoded), nil
}

// send builds and issues the HTTP request, wrapping transport failures in
// ErrConnect.
func (c Client) send(
	ctx context.Context,
	method HTTPMethod,
	path requestPath,
	query url.Values,
	body io.Reader,
) (*http.Response, error) {
	request, err := http.NewRequestWithContext(ctx, string(method), c.apiURL(path, query), body)
	if err != nil {
		return nil, ErrConnect.With(err, "path", string(path))
	}
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	response, err := c.doer.Do(request)
	if err != nil {
		return nil, ErrConnect.With(err, "path", string(path))
	}
	return response, nil
}

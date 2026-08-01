package docker

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	errs "github.com/gomatic/go-error"
)

// errorEnvelope mirrors the Engine's error response document.
type errorEnvelope struct {
	Message string `json:"message"`
}

// acceptResponse passes wanted statuses through and converts everything else
// into a sentinel error carrying the Engine's envelope message.
func acceptResponse(response *http.Response, wanted []HTTPStatus) (*http.Response, error) {
	for _, status := range wanted {
		if response.StatusCode == int(status) {
			return response, nil
		}
	}
	defer func() { _ = response.Body.Close() }()
	return nil, statusError(HTTPStatus(response.StatusCode), envelopeMessage(response.Body))
}

// envelopeMessage extracts the Engine error envelope's message, falling back
// to the raw (trimmed) body when the envelope does not parse.
func envelopeMessage(body io.Reader) daemonMessage {
	raw, err := io.ReadAll(io.LimitReader(body, maxErrorBodyBytes))
	if err != nil {
		return ""
	}
	var envelope errorEnvelope
	if err := json.Unmarshal(raw, &envelope); err != nil || envelope.Message == "" {
		return daemonMessage(strings.TrimSpace(string(raw)))
	}
	return daemonMessage(envelope.Message)
}

// maxErrorBodyBytes bounds how much of an error response is read for its
// message; envelopes are small and an unbounded read of a hostile response
// would be a memory hazard.
const maxErrorBodyBytes int64 = 64 * 1024

// daemonMessage is the Engine error envelope's human-readable message.
type daemonMessage string

// statusError maps an unexpected Engine status onto the matching sentinel,
// attaching the daemon's message.
func statusError(status HTTPStatus, message daemonMessage) error {
	sentinel := sentinelFor(status)
	if message == "" {
		return sentinel.With(nil, "status", int(status))
	}
	return sentinel.With(nil, "status", int(status), "message", string(message))
}

// sentinelFor chooses the sentinel for an unexpected Engine status.
func sentinelFor(status HTTPStatus) errs.Const {
	switch {
	case status == http.StatusNotFound:
		return ErrNotFound
	case status == http.StatusConflict:
		return ErrConflict
	case status == http.StatusBadRequest:
		return ErrBadParameter
	case status >= http.StatusInternalServerError:
		return ErrServer
	default:
		return ErrUnexpectedStatus
	}
}

// maxDocumentBytes bounds a decoded response document; Engine documents are
// small and an unbounded read of a hostile response would be a memory hazard.
const maxDocumentBytes int64 = 16 * 1024 * 1024

// readBody reads a response body; the caller owns closing it. Each caller
// unmarshals the returned bytes into its own concrete wire type at its own
// call site — that keeps every decode-only mirror type visibly rooted at a
// json.Unmarshal call, which is the sanctioned shape for externally-owned
// wire vocabularies.
func readBody(body io.Reader) ([]byte, error) {
	raw, err := io.ReadAll(io.LimitReader(body, maxDocumentBytes))
	if err != nil {
		return nil, ErrDecodeResponse.With(err)
	}
	return raw, nil
}

// drainBody consumes a response body so the underlying connection can be
// reused; the caller owns closing it.
func drainBody(body io.Reader) {
	_, _ = io.Copy(io.Discard, io.LimitReader(body, maxErrorBodyBytes))
}

package docker

import (
	"bufio"
	"encoding/json"
	"io"
)

// The Engine reports pull (and build) progress as a stream of JSON documents,
// one per line. A failure mid-stream arrives as a document carrying an error
// field while the HTTP status stays 200 — a scanner that ignores it reports
// false success, so this one never does.

// StreamMessage is one progress document from an Engine JSON stream.
type StreamMessage struct {
	Status  string            `json:"status"`
	Failure string            `json:"error"`
	Detail  StreamErrorDetail `json:"errorDetail"`
}

// StreamErrorDetail is the structured error a failed stream operation
// reports.
type StreamErrorDetail struct {
	Message string `json:"message"`
}

// StreamHandler observes each stream message in order. Returning an error
// stops the scan and surfaces that error.
type StreamHandler func(message StreamMessage) error

// maxStreamLineBytes bounds one stream document; Engine progress lines are
// small, and an unbounded line from a hostile stream would be a memory
// hazard.
const maxStreamLineBytes = 1024 * 1024

// ScanStream decodes an Engine JSON stream, invoking handle per message, and
// converts an in-band error document into ErrStreamError. A nil handler just
// watches for failure.
func ScanStream(source io.Reader, handle StreamHandler) error {
	scanner := bufio.NewScanner(source)
	scanner.Buffer(make([]byte, 0, bufio.MaxScanTokenSize), maxStreamLineBytes)
	for scanner.Scan() {
		if err := scanLine(scanner.Bytes(), handle); err != nil {
			return err
		}
	}
	if err := scanner.Err(); err != nil {
		return ErrStream.With(err)
	}
	return nil
}

// scanLine decodes and dispatches one stream line, skipping blank lines.
func scanLine(line []byte, handle StreamHandler) error {
	if len(line) == 0 {
		return nil
	}
	var message StreamMessage
	if err := json.Unmarshal(line, &message); err != nil {
		return ErrDecodeResponse.With(err)
	}
	if err := failureOf(message); err != nil {
		return err
	}
	if handle == nil {
		return nil
	}
	return handle(message)
}

// failureOf surfaces a message's in-band error, preferring the structured
// detail message.
func failureOf(message StreamMessage) error {
	if message.Detail.Message != "" {
		return ErrStreamError.With(nil, "message", message.Detail.Message)
	}
	if message.Failure != "" {
		return ErrStreamError.With(nil, "message", message.Failure)
	}
	return nil
}

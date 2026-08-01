package docker

import (
	"encoding/binary"
	"errors"
	"io"
)

// The Engine multiplexes a non-TTY container's stdout and stderr onto one
// connection as frames: an 8-byte header — stream byte, three zero bytes, a
// big-endian uint32 payload size — followed by the payload. DemuxStream
// splits that stream back into its two writers.

// streamByte identifies a frame's origin stream.
type streamByte byte

// Frame origin streams the Engine emits.
const (
	streamStdin  streamByte = 0
	streamStdout streamByte = 1
	streamStderr streamByte = 2
	streamSystem streamByte = 3
)

// frameHeaderSize is the fixed multiplexed frame header length.
const frameHeaderSize = 8

// maxFramePayloadBytes bounds a single frame's payload. The Engine emits
// frames far below this; a larger claim indicates a corrupt or hostile
// stream and is rejected rather than buffered.
const maxFramePayloadBytes uint32 = 32 * 1024 * 1024

// DemuxStream copies a multiplexed Engine stream onto stdout and stderr
// until EOF. A system frame (the daemon's in-band error channel) surfaces as
// ErrStreamError carrying the frame payload; a malformed frame surfaces as
// ErrStream; an oversize frame as ErrFrameTooLarge.
func DemuxStream(source io.Reader, stdout, stderr io.Writer) error {
	for {
		header, err := readHeader(source)
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		if err := copyFrame(source, header, stdout, stderr); err != nil {
			return err
		}
	}
}

// frameHeader is one decoded frame header.
type frameHeader struct {
	stream streamByte
	size   frameSize
}

// readHeader reads and validates one frame header. io.EOF passes through
// untouched so the demuxer can distinguish clean stream end from a torn
// header, which is reported as ErrStream.
func readHeader(source io.Reader) (frameHeader, error) {
	var raw [frameHeaderSize]byte
	if _, err := io.ReadFull(source, raw[:]); err != nil {
		if errors.Is(err, io.EOF) {
			return frameHeader{}, io.EOF
		}
		return frameHeader{}, ErrStream.With(err)
	}
	header := frameHeader{
		stream: streamByte(raw[0]),
		size:   frameSize(binary.BigEndian.Uint32(raw[4:frameHeaderSize])),
	}
	if header.size > frameSize(maxFramePayloadBytes) {
		return frameHeader{}, ErrFrameTooLarge.With(nil, "size", uint32(header.size))
	}
	return frameHeader{stream: header.stream, size: header.size}, nil
}

// copyFrame routes one frame's payload to its destination writer.
func copyFrame(source io.Reader, header frameHeader, stdout, stderr io.Writer) error {
	destination, err := frameDestination(source, header, stdout, stderr)
	if err != nil || destination == nil {
		return err
	}
	if _, err := io.CopyN(destination, source, int64(header.size)); err != nil {
		return ErrStream.With(err, "stream", int(header.stream))
	}
	return nil
}

// frameDestination selects where a frame's payload goes: stdout, stderr,
// discarded (stdin echo), or — for a system frame — surfaced as the daemon's
// in-band error. A nil writer with nil error means the frame is skipped.
func frameDestination(
	source io.Reader,
	header frameHeader,
	stdout, stderr io.Writer,
) (io.Writer, error) {
	switch header.stream {
	case streamStdout:
		return stdout, nil
	case streamStderr:
		return stderr, nil
	case streamStdin:
		return io.Discard, nil
	case streamSystem:
		return nil, systemFrameError(source, header.size)
	default:
		return nil, ErrStream.With(nil, "stream", int(header.stream))
	}
}

// frameSize is a multiplexed frame's payload length in bytes.
type frameSize uint32

// systemFrameError reads a system frame's payload and returns it as the
// daemon's reported stream error.
func systemFrameError(source io.Reader, size frameSize) error {
	payload := make([]byte, size)
	if _, err := io.ReadFull(source, payload); err != nil {
		return ErrStream.With(err)
	}
	return ErrStreamError.With(nil, "message", string(payload))
}

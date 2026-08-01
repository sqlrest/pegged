package docker

import errs "github.com/gomatic/go-error"

// Sentinel errors this package emits, matchable with errors.Is. The Const
// mechanism is owned by gomatic/go-error. Keep sorted alphabetically.
//
// Daemon failures carry the Engine's error envelope message via With, so a
// wrapped sentinel still matches while preserving the daemon's diagnostic.
const (
	ErrBadParameter     errs.Const = "daemon rejected the request"
	ErrCanceled         errs.Const = "operation canceled"
	ErrConflict         errs.Const = "conflicting resource state"
	ErrConnect          errs.Const = "connecting to the docker daemon"
	ErrDecodeResponse   errs.Const = "decoding daemon response"
	ErrEncodeRequest    errs.Const = "encoding request body"
	ErrFrameTooLarge    errs.Const = "stream frame exceeds the size limit"
	ErrHost             errs.Const = "unsupported docker host"
	ErrNotFound         errs.Const = "no such resource"
	ErrPing             errs.Const = "pinging the docker daemon"
	ErrPull             errs.Const = "pulling image"
	ErrServer           errs.Const = "daemon internal error"
	ErrStream           errs.Const = "malformed multiplexed stream"
	ErrStreamError      errs.Const = "daemon reported a stream error"
	ErrUnexpectedStatus errs.Const = "unexpected daemon status"
)

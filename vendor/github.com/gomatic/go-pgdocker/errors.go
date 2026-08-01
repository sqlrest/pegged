package pgdocker

import errs "github.com/gomatic/go-error"

// Sentinel errors this package emits, matchable with errors.Is. The Const
// mechanism is owned by gomatic/go-error. Keep sorted alphabetically.
//
// Engine-level failures (daemon unreachable, conflicts, missing resources)
// pass through from go-docker and stay matchable against that package's
// sentinels; the constants here name this package's own lifecycle semantics.
const (
	ErrAlreadyRunning   errs.Const = "an instance already occupies this port"
	ErrCloneFailed      errs.Const = "cloning the data volume failed"
	ErrInvalidSpec      errs.Const = "invalid specification"
	ErrNoDataVolume     errs.Const = "no data volume exists for this port"
	ErrNotReady         errs.Const = "the database did not become ready in time"
	ErrNotRunning       errs.Const = "no running instance on this port"
	ErrSnapshotExists   errs.Const = "a snapshot with this name already exists"
	ErrSnapshotNotFound errs.Const = "no such snapshot"
	ErrSnapshotSkew     errs.Const = "snapshot postgres major version differs from the requested image"
)

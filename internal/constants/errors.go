// Package constants declares pegged's sentinel error values. The error
// mechanism (the matchable string type) lives in the shared gomatic/go-error
// library; these values are this CLI's own.
package constants

// Imported bare (the package is named error); this file declares only sentinels
// and uses no builtin error type, so each declaration reads errs.Const.
import errs "github.com/gomatic/go-error"

// Keep these constants sorted alphabetically.
const (
	ErrInitFailed      errs.Const = "initialization command failed"
	ErrInvalidPort     errs.Const = "invalid port"
	ErrInvalidValue    errs.Const = "invalid value"
	ErrMissingArgument errs.Const = "missing required argument"
	ErrPortConflict    errs.Const = "conflicting port arguments"
)

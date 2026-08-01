package manage

import (
	pgdocker "github.com/gomatic/go-pgdocker"

	"github.com/sqlrest/pegged/internal/constants"
)

// FlagValue is the raw text of an enumerated flag (--volume, --retain).
type FlagValue string

// VolumeSourceFor maps a --volume flag value onto the library's volume
// source. Only empty (follow pegged's port policy), "reuse", and "fresh"
// are valid; anything else fails with ErrInvalidValue.
func VolumeSourceFor(value FlagValue) (pgdocker.VolumeSource, error) {
	switch source := pgdocker.VolumeSource(value); source {
	case "", pgdocker.VolumeReuse, pgdocker.VolumeFresh:
		return source, nil
	default:
		return "", constants.ErrInvalidValue.With(nil, "flag", "volume", "value", string(value))
	}
}

// RetentionFor maps a --retain flag value onto the library's retention
// policy. Only empty (defer to the label recorded at start, then keep),
// "keep", and "remove" are valid; anything else fails with ErrInvalidValue.
func RetentionFor(value FlagValue) (pgdocker.Retention, error) {
	switch retention := pgdocker.Retention(value); retention {
	case "", pgdocker.RetainKeep, pgdocker.RetainRemove:
		return retention, nil
	default:
		return "", constants.ErrInvalidValue.With(nil, "flag", "retain", "value", string(value))
	}
}

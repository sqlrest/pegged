package constants

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// The error mechanism (Error/With) is exercised in gomatic/go-error; this test
// only verifies that this package's sentinels carry their text and remain
// matchable with errors.Is once wrapped — the contract consumers rely on.
func TestSentinels(t *testing.T) {
	t.Parallel()
	want := assert.New(t)

	want.Equal("invalid port", ErrInvalidPort.Error())
	want.Equal("conflicting port arguments", ErrPortConflict.Error())

	wrapped := ErrInitFailed.With(nil, "exit_code", 7)
	want.ErrorIs(wrapped, ErrInitFailed)
	want.NotErrorIs(wrapped, ErrInvalidValue)
}

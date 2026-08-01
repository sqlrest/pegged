package manage

import (
	"testing"

	pgdocker "github.com/gomatic/go-pgdocker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sqlrest/pegged/internal/constants"
)

func TestVolumeSourceFor(t *testing.T) {
	t.Parallel()
	tests := []struct {
		wantErr error
		name    string
		value   FlagValue
		want    pgdocker.VolumeSource
	}{
		{name: "empty follows the port policy", value: "", want: pgdocker.VolumeSource("")},
		{name: "reuse", value: "reuse", want: pgdocker.VolumeReuse},
		{name: "fresh", value: "fresh", want: pgdocker.VolumeFresh},
		{name: "anything else is invalid", value: "recycle", wantErr: constants.ErrInvalidValue},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			want, must := assert.New(t), require.New(t)

			source, err := VolumeSourceFor(tt.value)

			if tt.wantErr != nil {
				must.Error(err)
				want.ErrorIs(err, tt.wantErr)
				return
			}
			must.NoError(err)
			want.Equal(tt.want, source)
		})
	}
}

func TestRetentionFor(t *testing.T) {
	t.Parallel()
	tests := []struct {
		wantErr error
		name    string
		value   FlagValue
		want    pgdocker.Retention
	}{
		{name: "empty follows the port policy", value: "", want: pgdocker.Retention("")},
		{name: "keep", value: "keep", want: pgdocker.RetainKeep},
		{name: "remove", value: "remove", want: pgdocker.RetainRemove},
		{name: "anything else is invalid", value: "discard", wantErr: constants.ErrInvalidValue},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			want, must := assert.New(t), require.New(t)

			retention, err := RetentionFor(tt.value)

			if tt.wantErr != nil {
				must.Error(err)
				want.ErrorIs(err, tt.wantErr)
				return
			}
			must.NoError(err)
			want.Equal(tt.want, retention)
		})
	}
}

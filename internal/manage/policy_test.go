package manage

import (
	"testing"

	docker "github.com/gomatic/go-docker"
	pgdocker "github.com/gomatic/go-pgdocker"
	"github.com/stretchr/testify/assert"
)

func TestPolicyVolume(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		chosen pgdocker.VolumeSource
		want   pgdocker.VolumeSource
		port   docker.Port
	}{
		{name: "durable ceiling reuses", port: 5432, want: pgdocker.VolumeReuse},
		{name: "below the ceiling reuses", port: 5000, want: pgdocker.VolumeReuse},
		{name: "above the ceiling is fresh", port: 5433, want: pgdocker.VolumeFresh},
		{
			name:   "explicit fresh wins on a durable port",
			port:   5432,
			chosen: pgdocker.VolumeFresh,
			want:   pgdocker.VolumeFresh,
		},
		{
			name:   "explicit reuse wins on an ephemeral port",
			port:   9999,
			chosen: pgdocker.VolumeReuse,
			want:   pgdocker.VolumeReuse,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			want := assert.New(t)
			want.Equal(tt.want, PolicyVolume(tt.port, tt.chosen))
		})
	}
}

func TestPolicyRetention(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		chosen pgdocker.Retention
		want   pgdocker.Retention
		port   docker.Port
	}{
		{name: "durable ceiling keeps", port: 5432, want: pgdocker.RetainKeep},
		{name: "above the ceiling removes", port: 5433, want: pgdocker.RetainRemove},
		{
			name:   "explicit remove wins on a durable port",
			port:   5432,
			chosen: pgdocker.RetainRemove,
			want:   pgdocker.RetainRemove,
		},
		{
			name:   "explicit keep wins on an ephemeral port",
			port:   9999,
			chosen: pgdocker.RetainKeep,
			want:   pgdocker.RetainKeep,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			want := assert.New(t)
			want.Equal(tt.want, PolicyRetention(tt.port, tt.chosen))
		})
	}
}

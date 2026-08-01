package manage

import (
	"testing"

	docker "github.com/gomatic/go-docker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sqlrest/pegged/internal/constants"
	"github.com/sqlrest/pegged/internal/domain"
)

func TestResolvePort(t *testing.T) {
	t.Parallel()
	tests := []struct {
		wantErr error
		name    string
		args    []domain.Argument
		want    docker.Port
		flagged FlagPort
	}{
		{
			name: "no argument and no flag resolves to zero",
			args: []domain.Argument{},
			want: 0,
		},
		{
			name:    "flag alone wins",
			args:    []domain.Argument{},
			flagged: 5433,
			want:    5433,
		},
		{
			name: "positional argument alone wins",
			args: []domain.Argument{"5434"},
			want: 5434,
		},
		{
			name:    "matching argument and flag agree",
			args:    []domain.Argument{"5433"},
			flagged: 5433,
			want:    5433,
		},
		{
			name:    "argument and flag disagree",
			args:    []domain.Argument{"5434"},
			flagged: 5433,
			wantErr: constants.ErrPortConflict,
		},
		{
			name:    "non-numeric argument is invalid",
			args:    []domain.Argument{"postgres"},
			wantErr: constants.ErrInvalidPort,
		},
		{
			name:    "argument above the port range is invalid",
			args:    []domain.Argument{"65536"},
			wantErr: constants.ErrInvalidPort,
		},
		{
			name:    "negative argument is invalid",
			args:    []domain.Argument{"-1"},
			wantErr: constants.ErrInvalidPort,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			want, must := assert.New(t), require.New(t)

			port, err := ResolvePort(tt.args, tt.flagged)

			if tt.wantErr != nil {
				must.Error(err)
				want.ErrorIs(err, tt.wantErr)
				return
			}
			must.NoError(err)
			want.Equal(tt.want, port)
		})
	}
}

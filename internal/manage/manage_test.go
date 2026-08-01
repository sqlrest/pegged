package manage

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	docker "github.com/gomatic/go-docker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestManager drives the production path end to end through DOCKER_HOST: a
// fake daemon proves defaultEngine and Manager assemble a working manager,
// and an unreachable daemon proves the engine failure propagates. It stays
// serial because it rewrites the process environment.
func TestManager(t *testing.T) {
	daemon := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Api-Version", "1.43")
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(daemon.Close)

	tests := []struct {
		wantErr error
		name    string
		host    string
	}{
		{
			name: "assembles a manager over the daemon",
			host: daemon.URL,
		},
		{
			name:    "unsupported daemon scheme fails",
			host:    "bogus://daemon",
			wantErr: docker.ErrHost,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want, must := assert.New(t), require.New(t)
			t.Setenv("DOCKER_HOST", tt.host)

			_, err := Manager(context.Background())

			if tt.wantErr != nil {
				must.Error(err)
				want.ErrorIs(err, tt.wantErr)
				return
			}
			must.NoError(err)
		})
	}
}

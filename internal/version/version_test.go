package version

import (
	"runtime/debug"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Seam-reassigning tests stay serial and restore via t.Cleanup.

// stubBuildInfo installs a scripted build-info reading.
func stubBuildInfo(t *testing.T, info *debug.BuildInfo, ok bool) {
	t.Helper()
	previous := readBuildInfo
	readBuildInfo = func() (*debug.BuildInfo, bool) { return info, ok }
	t.Cleanup(func() { readBuildInfo = previous })
}

// stubInjected installs a scripted ldflags value.
func stubInjected(t *testing.T, value string) {
	t.Helper()
	previous := injected
	injected = value
	t.Cleanup(func() { injected = previous })
}

func TestResolve(t *testing.T) {
	tests := []struct {
		info     *debug.BuildInfo
		name     string
		injected string
		want     Resolved
		infoOK   bool
	}{
		{
			name:     "injected ldflags value wins",
			injected: "v1.0.0",
			want:     "v1.0.0",
		},
		{
			name:     "injected dev defers to module version",
			injected: "dev",
			info:     &debug.BuildInfo{Main: debug.Module{Version: "v1.2.3"}},
			infoOK:   true,
			want:     "v1.2.3",
		},
		{
			name:   "module version",
			info:   &debug.BuildInfo{Main: debug.Module{Version: "v0.4.0"}},
			infoOK: true,
			want:   "v0.4.0",
		},
		{
			name: "devel module falls back to long revision, shortened",
			info: &debug.BuildInfo{
				Main: debug.Module{Version: "(devel)"},
				Settings: []debug.BuildSetting{
					{Key: "vcs.modified", Value: "false"},
					{Key: revisionKey, Value: "0123456789abcdef0123"},
				},
			},
			infoOK: true,
			want:   "0123456789ab",
		},
		{
			name: "short revision kept whole",
			info: &debug.BuildInfo{
				Settings: []debug.BuildSetting{{Key: revisionKey, Value: "abc123"}},
			},
			infoOK: true,
			want:   "abc123",
		},
		{
			name:   "no version and no revision resolves dev",
			info:   &debug.BuildInfo{Settings: []debug.BuildSetting{{Key: revisionKey, Value: ""}}},
			infoOK: true,
			want:   "dev",
		},
		{
			name: "unreadable build info resolves dev",
			want: "dev",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want := assert.New(t)
			stubInjected(t, tt.injected)
			stubBuildInfo(t, tt.info, tt.infoOK)
			want.Equal(tt.want, Resolve())
		})
	}
}

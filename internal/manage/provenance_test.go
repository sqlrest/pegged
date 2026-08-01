package manage

import (
	"testing"

	docker "github.com/gomatic/go-docker"
	"github.com/stretchr/testify/assert"

	"github.com/sqlrest/pegged/internal/version"
)

// TestProvenanceLabels pins the provenance contract: exactly one label,
// keyed pegged.version, carrying the resolved version. Seam-reassigning,
// so serial with t.Cleanup restore.
func TestProvenanceLabels(t *testing.T) {
	want := assert.New(t)
	previous := resolveVersion
	resolveVersion = func() version.Resolved { return "v9.9.9" }
	t.Cleanup(func() { resolveVersion = previous })
	want.Equal(docker.Labels{"pegged.version": "v9.9.9"}, ProvenanceLabels())
}

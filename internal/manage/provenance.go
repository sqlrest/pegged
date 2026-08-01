package manage

// Provenance labels: every container, data volume, and snapshot volume
// pegged creates records which pegged version created it, so posture drift
// (instances created by an older pegged) is detectable by inspection.

import (
	docker "github.com/gomatic/go-docker"

	"github.com/sqlrest/pegged/internal/version"
)

// VersionLabel is the provenance label key.
const VersionLabel = "pegged.version"

// resolveVersion is the version seam.
var resolveVersion = version.Resolve

// ProvenanceLabels are the labels recording which pegged created a
// resource; go-pgdocker transports them onto containers and volumes.
func ProvenanceLabels() docker.Labels {
	return docker.Labels{VersionLabel: string(resolveVersion())}
}

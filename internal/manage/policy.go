// pegged's port policy. The library holds no opinion (its defaults are
// fresh volumes and non-destructive retention); the durable-vs-ephemeral
// port convention is this product's, applied at start time so the choice is
// recorded on the container and honored by stop.
package manage

import (
	docker "github.com/gomatic/go-docker"
	pgdocker "github.com/gomatic/go-pgdocker"
)

// DurablePortCeiling is the highest port treated as a durable development
// database: at or below it, start reuses the newest volume and stop keeps
// it; above it, instances are ephemeral.
const DurablePortCeiling docker.Port = 5432

// PolicyVolume resolves the start-time volume source: the user's explicit
// choice, else the port policy.
func PolicyVolume(port docker.Port, chosen pgdocker.VolumeSource) pgdocker.VolumeSource {
	if chosen != "" {
		return chosen
	}
	if port <= DurablePortCeiling {
		return pgdocker.VolumeReuse
	}
	return pgdocker.VolumeFresh
}

// PolicyRetention resolves the start-time retention recorded on the
// container: the user's explicit choice, else the port policy.
func PolicyRetention(port docker.Port, chosen pgdocker.Retention) pgdocker.Retention {
	if chosen != "" {
		return chosen
	}
	if port <= DurablePortCeiling {
		return pgdocker.RetainKeep
	}
	return pgdocker.RetainRemove
}

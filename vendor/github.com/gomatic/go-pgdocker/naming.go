package pgdocker

import (
	"encoding/hex"
	"io"
	"strconv"
	"strings"

	docker "github.com/gomatic/go-docker"
)

// compactTimestamp is the sortable UTC timestamp embedded in volume names.
const compactTimestamp = "20060102T150405"

// nameSuffixBytes is the entropy width of a volume-name suffix.
const nameSuffixBytes = 3

// containerName derives the instance container name for a port.
func (m Manager) containerName(port docker.Port) docker.ContainerName {
	return docker.ContainerName(string(m.namespace) + "-" + strconv.Itoa(int(port)))
}

// volumeName derives a fresh, unique data-volume name for a port: namespace,
// port, sortable UTC timestamp, and a short random suffix.
func (m Manager) volumeName(port docker.Port) (docker.VolumeName, error) {
	suffix := make([]byte, nameSuffixBytes)
	if _, err := io.ReadFull(m.entropy, suffix); err != nil {
		return "", ErrInvalidSpec.With(err, "reason", "entropy source failed")
	}
	stamp := m.clock().UTC().Format(compactTimestamp)
	name := string(m.namespace) + "-" + strconv.Itoa(int(port)) + "-" + stamp + "-" + hex.EncodeToString(suffix)
	return docker.VolumeName(name), nil
}

// snapshotVolumeName derives the volume name holding a snapshot.
func (m Manager) snapshotVolumeName(name SnapshotName) docker.VolumeName {
	return docker.VolumeName(string(m.namespace) + "-snapshot-" + string(name))
}

// imageMajor extracts the PostgreSQL major version from an image reference's
// tag ("postgres:17-alpine" → "17"); empty when it cannot be determined.
func imageMajor(image docker.ImageRef) PostgresMajor {
	ref := string(image)
	colon := strings.LastIndex(ref, ":")
	if colon < strings.LastIndex(ref, "/") || colon == -1 {
		return ""
	}
	tag := ref[colon+1:]
	end := 0
	for end < len(tag) && tag[end] >= '0' && tag[end] <= '9' {
		end++
	}
	return PostgresMajor(tag[:end])
}

// PostgresMajor is a PostgreSQL major version ("17"); empty means unknown.
type PostgresMajor string

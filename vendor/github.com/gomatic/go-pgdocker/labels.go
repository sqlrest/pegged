package pgdocker

import (
	"strconv"
	"time"

	docker "github.com/gomatic/go-docker"
)

// Label suffixes appended to the manager's namespace. All discovery is
// label-first: names are a human convenience, labels are the contract.
const (
	labelManaged      labelSuffix = "managed"
	labelPort         labelSuffix = "port"
	labelVolumeAction labelSuffix = "volume.action"
	labelImage        labelSuffix = "postgres.image"
	labelMajor        labelSuffix = "postgres.major"
	labelHelper       labelSuffix = "helper"
	labelSnapshot     labelSuffix = "snapshot.name"
	labelSourcePort   labelSuffix = "snapshot.source-port"
	labelCreated      labelSuffix = "created"
)

// labelSuffix is the namespace-relative part of one label key.
type labelSuffix string

// labelTrue marks a boolean label as set.
const labelTrue = "true"

// labelKey renders one fully-namespaced label key.
func (m Manager) labelKey(suffix labelSuffix) string {
	return string(m.namespace) + "." + string(suffix)
}

// managedLabels is the base label set every managed container and volume
// carries.
func (m Manager) managedLabels(port docker.Port, image docker.ImageRef) docker.Labels {
	return docker.Labels{
		m.labelKey(labelManaged): labelTrue,
		m.labelKey(labelPort):    strconv.Itoa(int(port)),
		m.labelKey(labelImage):   string(image),
		m.labelKey(labelMajor):   string(imageMajor(image)),
		m.labelKey(labelCreated): m.clock().UTC().Format(time.RFC3339),
	}
}

// reservedSuffixes is every label suffix the library writes. mergeLabels
// drops caller keys naming any of these, whichever resource they would
// ride on — otherwise a data volume could be made to answer the snapshot
// filter (and be destroyed by SnapshotDelete), or a snapshot volume the
// port filter.
var reservedSuffixes = []labelSuffix{
	labelManaged, labelPort, labelVolumeAction, labelImage, labelMajor,
	labelHelper, labelSnapshot, labelSourcePort, labelCreated,
}

// reserved reports whether a caller label key collides with the namespaced
// discovery contract.
func (m Manager) reserved(key string) bool {
	for _, suffix := range reservedSuffixes {
		if key == m.labelKey(suffix) {
			return true
		}
	}
	return false
}

// mergeLabels overlays caller labels onto managed labels. Every reserved
// namespaced key comes from the library alone: a caller key naming ANY
// reserved suffix is dropped — not just the keys present in this call's
// managed set — so callers cannot forge the discovery contract.
func (m Manager) mergeLabels(managed, extra docker.Labels) docker.Labels {
	merged := docker.Labels{}
	for key, value := range extra {
		if !m.reserved(key) {
			merged[key] = value
		}
	}
	for key, value := range managed {
		merged[key] = value
	}
	return merged
}

// portFilter is the label filter selecting one port's managed resources.
func (m Manager) portFilter(port docker.Port) docker.Labels {
	return docker.Labels{m.labelKey(labelPort): strconv.Itoa(int(port))}
}

// managedFilter is the label filter selecting every managed resource.
func (m Manager) managedFilter() docker.Labels {
	return docker.Labels{m.labelKey(labelManaged): labelTrue}
}

// snapshotFilter is the label filter selecting snapshots — by name when one
// is given, otherwise all snapshots.
func (m Manager) snapshotFilter(name SnapshotName) docker.Labels {
	if name == "" {
		return docker.Labels{m.labelKey(labelSnapshot): ""}
	}
	return docker.Labels{m.labelKey(labelSnapshot): string(name)}
}

// labelValue reads one namespaced label from a label map.
func (m Manager) labelValue(labels docker.Labels, suffix labelSuffix) string {
	return labels[m.labelKey(suffix)]
}

// portOf reads the port label from a label map, zero when absent or
// unparseable.
func (m Manager) portOf(labels docker.Labels) docker.Port {
	parsed, err := strconv.ParseUint(m.labelValue(labels, labelPort), 10, 16)
	if err != nil {
		return 0
	}
	return docker.Port(parsed)
}

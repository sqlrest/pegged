package pgdocker

import (
	"time"

	docker "github.com/gomatic/go-docker"
)

// VolumeSource selects where a starting instance's data volume comes from.
// The zero value defaults to VolumeFresh; reusing existing state is an
// explicit choice.
type VolumeSource string

// Volume sources.
const (
	VolumeReuse VolumeSource = "reuse"
	VolumeFresh VolumeSource = "fresh"
)

// Retention selects what happens to an instance's data volume on stop. The
// zero value defaults to RetainKeep: the library never destroys data unless
// told to. A container label records the choice made at start so stop
// honors it without being told again.
type Retention string

// Retention policies.
const (
	RetainKeep   Retention = "keep"
	RetainRemove Retention = "remove"
)

// SnapshotName identifies a snapshot within the manager's namespace.
type SnapshotName string

// InitSQLDir is a host directory of *.sql/*.sh first-boot provisioning
// scripts, mounted read-only at /docker-entrypoint-initdb.d.
type InitSQLDir string

// DataDir is the PGDATA mount target inside the container. The default
// matches the official postgres images through major 17; postgres 18+
// images move the volume to /var/lib/postgresql.
type DataDir docker.MountTarget

// defaultDataDir matches the official postgres image through major 17.
const defaultDataDir DataDir = "/var/lib/postgresql/data"

// Connection identity defaults, matching the official postgres image.
const (
	defaultDatabase DatabaseName = "postgres"
	defaultUser     UserName     = "postgres"
)

// ListenAddress is the host interface the instance's port binds. The
// default is loopback — a managed development database is not reachable
// from other machines unless the caller widens this deliberately.
type ListenAddress docker.HostIP

// defaultListenAddress keeps instances loopback-only.
const defaultListenAddress ListenAddress = "127.0.0.1"

// StartSpec describes an instance to start. The zero value starts a
// loopback trust-auth postgres on the default port with the tuned
// development settings.
type StartSpec struct {
	Settings            Settings
	Labels              docker.Labels
	Volume              VolumeSource
	Retain              Retention
	Image               docker.ImageRef
	Platform            docker.Platform
	Snapshot            SnapshotName
	Database            DatabaseName
	Listen              ListenAddress
	Password            Password
	User                UserName
	InitSQL             InitSQLDir
	Data                DataDir
	Env                 docker.Env
	ReadyTimeout        time.Duration
	ReadyInterval       time.Duration
	Port                docker.Port
	SnapshotSkewEnabled bool
}

// Readiness pacing defaults.
const (
	defaultReadyTimeout  = 30 * time.Second
	defaultReadyInterval = 500 * time.Millisecond
)

// withDefaults resolves every zero field to its documented default.
func (s StartSpec) withDefaults() StartSpec {
	if s.Port == 0 {
		s.Port = DefaultPort
	}
	if s.Database == "" {
		s.Database = defaultDatabase
	}
	if s.User == "" {
		s.User = defaultUser
	}
	if s.Image == "" {
		s.Image = DefaultImage
	}
	return s.withRuntimeDefaults()
}

// withRuntimeDefaults resolves the runtime posture fields.
func (s StartSpec) withRuntimeDefaults() StartSpec {
	if s.Data == "" {
		s.Data = defaultDataDir
	}
	if s.Listen == "" {
		s.Listen = defaultListenAddress
	}
	if s.ReadyTimeout <= 0 {
		s.ReadyTimeout = defaultReadyTimeout
	}
	if s.ReadyInterval <= 0 {
		s.ReadyInterval = defaultReadyInterval
	}
	return s
}

// sourceOrDefault resolves the volume source's non-destructive default.
func sourceOrDefault(chosen VolumeSource) VolumeSource {
	if chosen == "" {
		return VolumeFresh
	}
	return chosen
}

// retentionOrDefault resolves retention's non-destructive default.
func retentionOrDefault(chosen Retention) Retention {
	if chosen == "" {
		return RetainKeep
	}
	return chosen
}

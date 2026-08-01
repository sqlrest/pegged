// Package pgdocker manages PostgreSQL instances in Docker: start, stop,
// inspect, list, snapshot, and initialize — the engine behind a
// database-lifecycle CLI. The host port is an instance's identity; every
// container and volume the package touches carries namespaced labels, and
// all discovery filters on those labels server-side. Snapshots are physical
// PGDATA volume clones, so they are bound to the PostgreSQL major version
// that wrote them.
//
// The library holds no port-based policy: defaults are explicit and
// non-destructive — a fresh volume on start, volumes kept on stop — and a
// consumer that wants an opinionated posture (a CLI's durable-vs-ephemeral
// port convention, for instance) passes explicit VolumeSource and Retention
// choices per call.
package pgdocker

import (
	"crypto/rand"
	"io"
	"time"

	docker "github.com/gomatic/go-docker"
)

// Namespace prefixes every container name, volume name, and label key this
// package writes, so multiple consumers of the library coexist on one
// daemon without seeing each other's instances.
type Namespace string

// DefaultNamespace marks instances managed by an unconfigured Manager.
const DefaultNamespace Namespace = "pgdocker"

// DefaultPort is the PostgreSQL default port, used when a spec names none.
const DefaultPort docker.Port = 5432

// DefaultImage is the PostgreSQL image started when a spec names none.
const DefaultImage docker.ImageRef = "postgres:17-alpine"

// Clock supplies the current time; injectable so naming and labels are
// deterministic under test. It matches time.Now.
type Clock func() time.Time

// Entropy supplies random bytes for volume-name suffixes; injectable so
// naming is deterministic under test. Production uses crypto/rand.
type Entropy io.Reader

// Manager drives PostgreSQL instances on one Docker daemon. The zero value
// is not usable; construct with New. Manager is a value: it holds only
// configuration and the injected engine, so it is safe to copy.
type Manager struct {
	engine    Engine
	entropy   Entropy
	clock     Clock
	namespace Namespace
}

// ManagerOption adjusts Manager construction. Options are concrete types —
// Namespace, WithEngine, WithClock, WithEntropy.
type ManagerOption interface {
	apply(assembled Manager) Manager
}

// apply sets the manager's namespace.
func (n Namespace) apply(assembled Manager) Manager {
	assembled.namespace = n
	return assembled
}

// WithEngine injects the Docker Engine seam; tests pass a fake.
type WithEngine struct {
	Engine Engine
}

// apply installs the engine.
func (w WithEngine) apply(assembled Manager) Manager {
	assembled.engine = w.Engine
	return assembled
}

// WithClock injects the time source.
type WithClock struct {
	Clock Clock
}

// apply installs the clock.
func (w WithClock) apply(assembled Manager) Manager {
	assembled.clock = w.Clock
	return assembled
}

// WithEntropy injects the randomness source for name suffixes.
type WithEntropy struct {
	Entropy Entropy
}

// apply installs the entropy source.
func (w WithEntropy) apply(assembled Manager) Manager {
	assembled.entropy = w.Entropy
	return assembled
}

// New assembles a Manager. An engine must be supplied — production wraps a
// docker.Client in WithEngine; there is no implicit daemon connection here,
// so construction is deterministic and never touches the network.
func New(options ...ManagerOption) (Manager, error) {
	assembled := Manager{
		namespace: DefaultNamespace,
		clock:     time.Now,
		entropy:   rand.Reader,
	}
	for _, option := range options {
		assembled = option.apply(assembled)
	}
	if assembled.engine == nil {
		return Manager{}, ErrInvalidSpec.With(nil, "reason", "an Engine is required")
	}
	if assembled.namespace == "" {
		return Manager{}, ErrInvalidSpec.With(nil, "reason", "namespace must not be empty")
	}
	return assembled, nil
}

// Package manage is the shared plumbing every pegged domain package builds
// on: it assembles the pegged-namespaced go-pgdocker Manager over the local
// Docker daemon and resolves the port and enumerated flag values the
// commands share. All lifecycle behavior lives in go-pgdocker; this package
// only wires the CLI to it.
package manage

import (
	"context"

	docker "github.com/gomatic/go-docker"
	pgdocker "github.com/gomatic/go-pgdocker"
)

// namespace prefixes every container name, volume name, and label key
// pegged writes, isolating pegged's instances from other go-pgdocker
// consumers on the same daemon.
const namespace pgdocker.Namespace = "pegged"

// newEngine is the Docker engine seam Manager resolves its engine through.
var newEngine = defaultEngine

// defaultEngine connects to the local Docker daemon (honoring DOCKER_HOST).
func defaultEngine(ctx context.Context) (pgdocker.Engine, error) {
	client, err := docker.New(ctx)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// Manager assembles the pegged-namespaced go-pgdocker Manager over the
// engine seam. Every domain Run obtains its Manager through this function.
func Manager(ctx context.Context) (pgdocker.Manager, error) {
	engine, err := newEngine(ctx)
	if err != nil {
		return pgdocker.Manager{}, err
	}
	return pgdocker.New(namespace, pgdocker.WithEngine{Engine: engine})
}

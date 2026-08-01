// Package docker is a dependency-light client for the Docker Engine REST API:
// containers, images, volumes, exec, and the Engine's two stream formats
// (stdcopy multiplexing and JSON progress streams). It speaks HTTP directly
// over the daemon socket — honoring DOCKER_HOST — so consumers inherit no
// Docker SDK module graph, and every part of it is testable against httptest
// fakes: the transport (Doer), host, environment, and API version are all
// injectable options.
package docker

import (
	"context"
	"os"
)

// APIVersion is a Docker Engine API version like "1.43". Supplying it as an
// option pins the client and skips negotiation; otherwise the daemon's
// reported version is adopted at construction.
type APIVersion string

// minimumAPIVersion is the fallback when the daemon does not report a
// version; every operation this package issues exists at this version.
const minimumAPIVersion APIVersion = "1.43"

// Client speaks the Engine API for one daemon. The zero value is not usable;
// construct with New. Client is a value: it holds only immutable
// configuration plus the injected transport, so it is safe to copy and share.
type Client struct {
	doer     Doer
	endpoint endpoint
	version  APIVersion
}

// Option adjusts client construction. Options are concrete types — HostAddress,
// APIVersion, Connection, Environment — applied in order.
type Option interface {
	apply(assembled settings) settings
}

// settings accumulates option effects before the client is assembled.
type settings struct {
	doer    Doer
	env     EnvironmentLookup
	host    HostAddress
	version APIVersion
}

// apply sets the daemon host, overriding DOCKER_HOST and the default socket.
func (h HostAddress) apply(assembled settings) settings {
	assembled.host = h
	return assembled
}

// apply pins the Engine API version and skips negotiation.
func (v APIVersion) apply(assembled settings) settings {
	assembled.version = v
	return assembled
}

// Connection injects the HTTP transport used to reach the daemon; tests pass
// an httptest client or a hand-rolled fake.
type Connection struct {
	Doer Doer
}

// apply installs the injected transport.
func (t Connection) apply(assembled settings) settings {
	assembled.doer = t.Doer
	return assembled
}

// Environment injects the environment lookup consulted for DOCKER_HOST;
// production defaults to os.LookupEnv.
type Environment EnvironmentLookup

// apply installs the environment lookup.
func (e Environment) apply(assembled settings) settings {
	assembled.env = EnvironmentLookup(e)
	return assembled
}

// New resolves the daemon endpoint, builds the transport, and negotiates the
// API version (unless one was pinned). The returned Client is ready for use
// and safe to copy.
func New(ctx context.Context, options ...Option) (Client, error) {
	assembled := settings{env: os.LookupEnv}
	for _, option := range options {
		assembled = option.apply(assembled)
	}
	resolved, err := resolveHost(assembled.host, assembled.env)
	if err != nil {
		return Client{}, err
	}
	client := Client{doer: assembled.doer, endpoint: resolved, version: assembled.version}
	if client.doer == nil {
		client.doer = doerFor(resolved)
	}
	return client.negotiate(ctx)
}

// negotiate adopts the daemon's reported API version when none was pinned.
func (c Client) negotiate(ctx context.Context) (Client, error) {
	if c.version != "" {
		return c, nil
	}
	c.version = minimumAPIVersion
	reported, err := c.Ping(ctx)
	if err != nil {
		return Client{}, err
	}
	if reported != "" {
		c.version = reported
	}
	return c, nil
}

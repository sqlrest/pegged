package docker

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"strings"
)

// HostAddress locates a Docker daemon: a unix://, tcp://, http://, or
// https:// URL. The zero value defers to the DOCKER_HOST environment
// variable, falling back to the platform default socket.
type HostAddress string

// defaultSocket is the platform-default daemon socket consulted when neither
// an option nor DOCKER_HOST names a host.
const defaultSocket HostAddress = "unix:///var/run/docker.sock"

// hostEnvVar is the environment variable the Docker CLI family uses to locate
// the daemon; this package honors the same contract.
const hostEnvVar = "DOCKER_HOST"

// endpoint is a resolved daemon location: the base URL requests are built
// against and, for unix sockets, the dialer that reaches the socket.
type endpoint struct {
	base   baseURL
	socket socketPath
}

// baseURL is the scheme://authority prefix every API request path is appended
// to. Unix-socket endpoints use a fixed placeholder authority; the socket
// itself is dialed by the transport.
type baseURL string

// socketPath is the filesystem path of a unix daemon socket; empty for TCP
// endpoints.
type socketPath string

// unixAuthority is the placeholder authority used in URLs for unix-socket
// endpoints. The HTTP layer requires a syntactically valid host; the dialer
// ignores it and connects to the socket instead.
const unixAuthority = "http://docker.sock.invalid"

// resolveHost turns a HostAddress (or, when empty, the environment/default)
// into a concrete endpoint, rejecting schemes this package does not speak
// with ErrHost.
func resolveHost(host HostAddress, env EnvironmentLookup) (endpoint, error) {
	if host == "" {
		host = hostFromEnvironment(env)
	}
	parsed, err := url.Parse(string(host))
	if err != nil {
		return endpoint{}, ErrHost.With(err, "host", string(host))
	}
	return endpointFor(host, parsed)
}

// hostFromEnvironment consults DOCKER_HOST, falling back to the platform
// default socket.
func hostFromEnvironment(env EnvironmentLookup) HostAddress {
	if env == nil {
		return defaultSocket
	}
	if fromEnv, ok := env(hostEnvVar); ok && fromEnv != "" {
		return HostAddress(fromEnv)
	}
	return defaultSocket
}

// endpointFor maps a parsed host URL onto an endpoint by scheme.
func endpointFor(host HostAddress, parsed *url.URL) (endpoint, error) {
	switch parsed.Scheme {
	case "unix":
		return endpoint{base: unixAuthority, socket: socketPath(parsed.Path)}, nil
	case "tcp":
		return endpoint{base: baseURL("http://" + parsed.Host)}, nil
	case "http", "https":
		return endpoint{base: baseURL(strings.TrimSuffix(string(host), "/"))}, nil
	default:
		return endpoint{}, ErrHost.With(nil, "host", string(host), "scheme", parsed.Scheme)
	}
}

// EnvironmentLookup reads one environment variable, reporting whether it is
// set. It matches os.LookupEnv so production passes that function directly
// while tests substitute a fake.
type EnvironmentLookup func(key string) (string, bool)

// Doer issues one HTTP request. It matches (*http.Client).Do so production
// wraps a real client while tests substitute a fake or an httptest client.
type Doer interface {
	Do(request *http.Request) (*http.Response, error)
}

// dialerFunc opens one network connection; it matches the http.Transport
// DialContext seam so unix-socket endpoints can route every request to the
// daemon socket.
type dialerFunc func(ctx context.Context, network, address string) (net.Conn, error)

// doerFor builds the HTTP client for an endpoint: unix-socket endpoints get a
// transport whose dialer ignores the placeholder authority and connects to
// the socket; TCP endpoints use a default client.
func doerFor(resolved endpoint) Doer {
	if resolved.socket == "" {
		return &http.Client{}
	}
	return &http.Client{Transport: &http.Transport{DialContext: socketDialer(resolved.socket)}}
}

// socketDialer yields a dialer bound to one unix socket path.
func socketDialer(socket socketPath) dialerFunc {
	dialer := net.Dialer{}
	return func(ctx context.Context, _, _ string) (net.Conn, error) {
		return dialer.DialContext(ctx, "unix", string(socket))
	}
}

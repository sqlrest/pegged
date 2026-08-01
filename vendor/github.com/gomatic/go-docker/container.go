package docker

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
)

// Identifier and vocabulary types shared across container operations.
type (
	// ContainerID is the Engine-assigned container identifier.
	ContainerID string
	// ContainerName is the caller-chosen container name.
	ContainerName string
	// ImageRef names an image, like "postgres:17-alpine".
	ImageRef string
	// Command is an argv vector (entrypoint or command override).
	Command []string
	// EnvVar is one KEY=value environment entry.
	EnvVar string
	// Env is the container environment.
	Env []EnvVar
	// Labels is a label map applied to containers and volumes.
	Labels map[string]string
	// WorkDir is the container working directory.
	WorkDir string
	// NetworkMode selects the container network ("", "host", or
	// "container:<id>" to share another container's network namespace).
	NetworkMode string
	// Platform is an OS/architecture selector like "linux/arm64"; empty
	// defers to the daemon default.
	Platform string
	// Port is a TCP port number.
	Port uint16
	// HostIP is the host interface address a port binding listens on.
	HostIP string
	// ContainerUser is the run-as user for a container or exec — "name",
	// "uid", "name:group", or "uid:gid"; empty defers to the image default.
	ContainerUser string
	// Capability names one Linux capability in the Engine's vocabulary,
	// like "ALL" or "NET_BIND_SERVICE".
	Capability string
	// SecurityOption is one HostConfig security option, like
	// "no-new-privileges".
	SecurityOption string
	// TmpfsOptions is the mount-option string for one tmpfs target, like
	// "rw,nosuid,noexec,mode=1777"; empty applies the daemon defaults.
	TmpfsOptions string
	// Tmpfs maps container paths to tmpfs mounts by their options.
	Tmpfs map[MountTarget]TmpfsOptions
)

// Mount attaches a volume or host path into a container.
type Mount struct {
	Kind            MountKind
	Source          MountSource
	Target          MountTarget
	ReadOnlyEnabled bool
}

// MountKind is the mount flavor: MountVolume or MountBind.
type MountKind string

// Sanctioned mount kinds.
const (
	MountVolume MountKind = "volume"
	MountBind   MountKind = "bind"
)

// MountSource is the volume name or absolute host path being mounted.
type MountSource string

// MountTarget is the absolute path inside the container.
type MountTarget string

// PortBinding publishes one container TCP port on the host.
type PortBinding struct {
	HostAddress HostIP
	Host        Port
	Container   Port
}

// ContainerSpec describes a container to create. Zero fields defer to image
// or daemon defaults.
type ContainerSpec struct {
	Name                  ContainerName
	Image                 ImageRef
	Command               Command
	Entrypoint            Command
	Env                   Env
	Labels                Labels
	WorkingDir            WorkDir
	User                  ContainerUser
	Network               NetworkMode
	Platform              Platform
	Mounts                []Mount
	Ports                 []PortBinding
	Tmpfs                 Tmpfs
	CapDrop               []Capability
	SecurityOptions       []SecurityOption
	AutoRemoveEnabled     bool
	ReadOnlyRootfsEnabled bool
}

// containerCreated mirrors the Engine's create response.
type containerCreated struct {
	ID string `json:"Id"`
}

// CreateContainer creates a container from spec and returns its identifier.
func (c Client) CreateContainer(ctx context.Context, spec ContainerSpec) (ContainerID, error) {
	response, err := c.request(
		ctx, http.MethodPost, "/containers/create",
		createQuery(spec), createBody(spec), http.StatusCreated,
	)
	if err != nil {
		return "", err
	}
	defer func() { _ = response.Body.Close() }()
	raw, err := readBody(response.Body)
	if err != nil {
		return "", err
	}
	var created containerCreated
	if err := json.Unmarshal(raw, &created); err != nil {
		return "", ErrDecodeResponse.With(err)
	}
	return ContainerID(created.ID), nil
}

// createQuery carries the name and platform, which the create endpoint takes
// as query parameters rather than body fields.
func createQuery(spec ContainerSpec) url.Values {
	query := url.Values{}
	if spec.Name != "" {
		query.Set("name", string(spec.Name))
	}
	if spec.Platform != "" {
		query.Set("platform", string(spec.Platform))
	}
	return query
}

// createBody assembles the Engine's container configuration document. Wire
// keys are the Engine's exact vocabulary, kept here and nowhere else.
func createBody(spec ContainerSpec) wireBody {
	body := wireBody{string(wireKeyImage): string(spec.Image), string(wireKeyHostConfig): hostConfig(spec)}
	setArgv(body, wireKeyCmd, spec.Command)
	setArgv(body, "Entrypoint", spec.Entrypoint)
	setEnv(body, spec.Env)
	setLabels(body, spec.Labels)
	if spec.WorkingDir != "" {
		body["WorkingDir"] = string(spec.WorkingDir)
	}
	if spec.User != "" {
		body["User"] = string(spec.User)
	}
	if ports := exposedPorts(spec.Ports); len(ports) > 0 {
		body["ExposedPorts"] = ports
	}
	return body
}

// wireKey is one Engine wire-vocabulary key in a request body.
type wireKey string

// Wire keys shared by more than one body builder.
const (
	wireKeyCmd        wireKey = "Cmd"
	wireKeyHostConfig wireKey = "HostConfig"
	wireKeyImage      wireKey = "Image"
	wireKeyName       wireKey = "Name"
	wireKeyReadOnly   wireKey = "ReadOnly"
	wireKeySource     wireKey = "Source"
	wireKeyTarget     wireKey = "Target"
	wireKeyType       wireKey = "Type"
)

// setArgv sets an argv-shaped wire field when the vector is non-empty.
func setArgv(body wireBody, key wireKey, argv Command) {
	if len(argv) == 0 {
		return
	}
	body[string(key)] = []string(argv)
}

// setEnv sets the wire environment when non-empty.
func setEnv(body wireBody, env Env) {
	if len(env) == 0 {
		return
	}
	entries := make([]string, len(env))
	for i, entry := range env {
		entries[i] = string(entry)
	}
	body["Env"] = entries
}

// setLabels sets the wire label map when non-empty.
func setLabels(body wireBody, labels Labels) {
	if len(labels) == 0 {
		return
	}
	body["Labels"] = map[string]string(labels)
}

// hostConfig assembles the Engine's HostConfig document.
func hostConfig(spec ContainerSpec) wireBody {
	host := wireBody{}
	if spec.AutoRemoveEnabled {
		host["AutoRemove"] = true
	}
	if spec.Network != "" {
		host["NetworkMode"] = string(spec.Network)
	}
	if len(spec.Mounts) > 0 {
		host["Mounts"] = wireMounts(spec.Mounts)
	}
	if bindings := portBindings(spec.Ports); len(bindings) > 0 {
		host["PortBindings"] = bindings
	}
	hardenConfig(host, spec)
	return host
}

// wireMounts renders mounts in the Engine's Mounts vocabulary.
func wireMounts(mounts []Mount) []wireBody {
	rendered := make([]wireBody, len(mounts))
	for i, mount := range mounts {
		rendered[i] = wireBody{
			string(wireKeyType):     string(mount.Kind),
			string(wireKeySource):   string(mount.Source),
			string(wireKeyTarget):   string(mount.Target),
			string(wireKeyReadOnly): mount.ReadOnlyEnabled,
		}
	}
	return rendered
}

// portKey renders a container port in the Engine's "port/proto" key form.
func portKey(port Port) string {
	return strconv.Itoa(int(port)) + "/tcp"
}

// exposedPorts renders the ExposedPorts document.
func exposedPorts(bindings []PortBinding) wireBody {
	exposed := wireBody{}
	for _, binding := range bindings {
		exposed[portKey(binding.Container)] = struct{}{}
	}
	return exposed
}

// portBindings renders the HostConfig PortBindings document.
func portBindings(bindings []PortBinding) wireBody {
	rendered := wireBody{}
	for _, binding := range bindings {
		rendered[portKey(binding.Container)] = []wireBody{{
			"HostIp":   string(binding.HostAddress),
			"HostPort": strconv.Itoa(int(binding.Host)),
		}}
	}
	return rendered
}

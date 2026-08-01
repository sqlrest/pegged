package pgdocker

import docker "github.com/gomatic/go-docker"

// Lockdown posture for every instance container: the server runs as the
// image's postgres user end to end (no root phase, no setuid step-down),
// nothing can regain privileges, and nothing is writable except the data
// volume postgres owns and the tmpfs scratch it needs.
const (
	// instanceUser is the image's unprivileged OS account — distinct from
	// StartSpec.User, which names the database role.
	instanceUser docker.ContainerUser = "postgres"
	// scratchOptions locks tmpfs scratch to sticky world-writable data
	// with no setuid and no executables.
	scratchOptions docker.TmpfsOptions = "rw,nosuid,noexec,mode=1777"
	// socketDir is where the server keeps its Unix socket and lock files.
	socketDir docker.MountTarget = "/var/run/postgresql"
	// tempDir is the server's temporary-file scratch.
	tempDir docker.MountTarget = "/tmp"
	// noNewPrivileges disables every privilege-escalation path.
	noNewPrivileges docker.SecurityOption = "no-new-privileges"
	// capabilityAll names the full capability set for dropping.
	capabilityAll docker.Capability = "ALL"
)

// containerSpec renders the instance container: locked down, labeled,
// loopback-bound by default, auto-removed on stop, configured entirely
// through argv and env — no configuration files are staged.
func (m Manager) containerSpec(spec StartSpec, volume docker.VolumeName) docker.ContainerSpec {
	labels := m.mergeLabels(m.managedLabels(spec.Port, spec.Image), spec.Labels)
	labels[m.labelKey(labelVolumeAction)] = string(retentionOrDefault(spec.Retain))
	mounts := []docker.Mount{{
		Kind:   docker.MountVolume,
		Source: docker.MountSource(volume),
		Target: docker.MountTarget(spec.Data),
	}}
	if spec.InitSQL != "" {
		mounts = append(mounts, docker.Mount{
			Kind:            docker.MountBind,
			Source:          docker.MountSource(spec.InitSQL),
			Target:          "/docker-entrypoint-initdb.d",
			ReadOnlyEnabled: true,
		})
	}
	return docker.ContainerSpec{
		Name:     m.containerName(spec.Port),
		Image:    spec.Image,
		Command:  settingsArgs(mergeSettings(spec.Settings)),
		Env:      append(authEnv(spec), spec.Env...),
		Labels:   labels,
		User:     instanceUser,
		Platform: spec.Platform,
		Mounts:   mounts,
		Ports: []docker.PortBinding{{
			HostAddress: docker.HostIP(spec.Listen),
			Host:        spec.Port,
			Container:   DefaultPort,
		}},
		Tmpfs:                 docker.Tmpfs{socketDir: scratchOptions, tempDir: scratchOptions},
		CapDrop:               []docker.Capability{capabilityAll},
		SecurityOptions:       []docker.SecurityOption{noNewPrivileges},
		AutoRemoveEnabled:     true,
		ReadOnlyRootfsEnabled: true,
	}
}

// authEnv renders the official image's authentication contract: trust when
// no password is chosen (development posture), password auth otherwise.
func authEnv(spec StartSpec) docker.Env {
	env := docker.Env{
		docker.EnvVar("POSTGRES_USER=" + string(spec.User)),
		docker.EnvVar("POSTGRES_DB=" + string(spec.Database)),
	}
	if spec.Password == "" {
		return append(env, "POSTGRES_HOST_AUTH_METHOD=trust")
	}
	return append(env, docker.EnvVar("POSTGRES_PASSWORD="+string(spec.Password)))
}

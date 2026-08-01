package docker

// Container lockdown rendering: the HostConfig fields that reduce a
// container's privileges — read-only rootfs, tmpfs scratch mounts,
// capability drops, and security options. Zero values emit nothing, so an
// unhardened spec renders a byte-identical create request.

// hardenConfig adds the lockdown fields when the spec chooses them.
func hardenConfig(host wireBody, spec ContainerSpec) {
	if spec.ReadOnlyRootfsEnabled {
		host["ReadonlyRootfs"] = true
	}
	if len(spec.Tmpfs) > 0 {
		host["Tmpfs"] = wireTmpfs(spec.Tmpfs)
	}
	if len(spec.CapDrop) > 0 {
		host["CapDrop"] = wireStrings(spec.CapDrop)
	}
	if len(spec.SecurityOptions) > 0 {
		host["SecurityOpt"] = wireStrings(spec.SecurityOptions)
	}
}

// wireTmpfs renders tmpfs mounts in the Engine's path-to-options form.
func wireTmpfs(mounts Tmpfs) map[string]string {
	rendered := make(map[string]string, len(mounts))
	for target, options := range mounts {
		rendered[string(target)] = string(options)
	}
	return rendered
}

// wireStrings renders a vocabulary-typed list as the Engine's plain strings.
func wireStrings[T ~string](values []T) []string {
	rendered := make([]string, len(values))
	for i, value := range values {
		rendered[i] = string(value)
	}
	return rendered
}

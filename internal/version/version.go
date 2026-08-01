// Package version resolves the pegged build version recorded as provenance
// on every container, data volume, and snapshot volume pegged creates.
package version

import "runtime/debug"

// Resolved is the version provenance records.
type Resolved string

// injected is the ldflags anchor:
//
//	-X github.com/sqlrest/pegged/internal/version.injected=v1.0.0
//
// Empty or "dev" (no injection) defers to the module's build information.
var injected string

// readBuildInfo is the build-info seam.
var readBuildInfo = debug.ReadBuildInfo

// unversioned is the resolution of a build carrying no version at all.
const unversioned Resolved = "dev"

// develVersion is the toolchain's placeholder for an unversioned main module.
const develVersion = "(devel)"

// revisionDigits bounds a VCS revision recorded as a version.
const revisionDigits = 12

// revisionKey is the build-info setting carrying the VCS revision.
const revisionKey = "vcs.revision"

// revision is a VCS revision hash from the build information.
type revision string

// Resolve picks the provenance version: the injected ldflags value when
// set, else the module version, else the stamped VCS revision, else "dev".
func Resolve() Resolved {
	if injected != "" && injected != string(unversioned) {
		return Resolved(injected)
	}
	info, ok := readBuildInfo()
	if !ok {
		return unversioned
	}
	return fromBuildInfo(info)
}

// fromBuildInfo prefers the module version, then the VCS revision.
func fromBuildInfo(info *debug.BuildInfo) Resolved {
	if info.Main.Version != "" && info.Main.Version != develVersion {
		return Resolved(info.Main.Version)
	}
	return fromSettings(info.Settings)
}

// fromSettings resolves the stamped VCS revision, shortened for label use.
func fromSettings(settings []debug.BuildSetting) Resolved {
	for _, setting := range settings {
		if setting.Key == revisionKey && setting.Value != "" {
			return shorten(revision(setting.Value))
		}
	}
	return unversioned
}

// shorten trims a full VCS revision to a label-friendly prefix.
func shorten(full revision) Resolved {
	if len(full) > revisionDigits {
		return Resolved(full[:revisionDigits])
	}
	return Resolved(full)
}

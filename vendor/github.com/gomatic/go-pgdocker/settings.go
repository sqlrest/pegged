package pgdocker

import (
	"slices"

	docker "github.com/gomatic/go-docker"
)

// Settings are PostgreSQL server parameters delivered as "-c key=value"
// command arguments — no configuration files are staged or mounted. Caller
// settings overlay the defaults; an empty value removes the key entirely.
type Settings map[SettingKey]SettingValue

// serverBinary is the PostgreSQL server executable the argv addresses.
const serverBinary = "postgres"

// Setting keys named in the default profile and its tests. Untyped, so the
// SettingKey constant set does not read as a closed enum.
const (
	settingJIT           = "jit"
	settingSharedBuffers = "shared_buffers"
)

// SettingKey and SettingValue are the halves of one server parameter.
type (
	// SettingKey is a PostgreSQL parameter name.
	SettingKey string
	// SettingValue is a PostgreSQL parameter value.
	SettingValue string
)

// defaultSettings is the tuned development profile: fast, single-node, and
// UTC-pinned. Every entry is overridable, and durability trades are safe
// here because the data is a disposable development volume.
func defaultSettings() Settings {
	return Settings{
		settingJIT:           "off",
		"log_timezone":       "UTC",
		"max_wal_senders":    "0",
		settingSharedBuffers: "256MB",
		"synchronous_commit": "off",
		"timezone":           "UTC",
		"wal_level":          "minimal",
	}
}

// mergeSettings overlays caller settings onto the defaults; an empty caller
// value deletes the key.
func mergeSettings(overrides Settings) Settings {
	merged := defaultSettings()
	for key, value := range overrides {
		if value == "" {
			delete(merged, key)
			continue
		}
		merged[key] = value
	}
	return merged
}

// settingsArgs renders settings as a deterministic postgres argv: sorted
// "-c key=value" pairs after the server binary name.
func settingsArgs(merged Settings) docker.Command {
	keys := make([]string, 0, len(merged))
	for key := range merged {
		keys = append(keys, string(key))
	}
	slices.Sort(keys)
	command := make(docker.Command, 0, 1+2*len(keys))
	command = append(command, serverBinary)
	for _, key := range keys {
		command = append(command, "-c", key+"="+string(merged[SettingKey(key)]))
	}
	return command
}

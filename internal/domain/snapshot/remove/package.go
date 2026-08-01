// Package remove orchestrates the "snapshot delete" command.
//
// It defines the command's Config (empty — the command takes no flags) and
// Run (the orchestration entry point the CLI invokes). Run validates the
// required snapshot name and removes the snapshot through go-pgdocker; it
// contains no CLI, flag, or output-formatting logic. This is the domain tier
// between the app tier (internal/app/commands/snapshot/remove) and the
// go-pgdocker library.
package remove

// Package create orchestrates the "snapshot create" command.
//
// It defines the command's Config (the flags and arguments the CLI binds) and
// Run (the orchestration entry point the CLI invokes). Run resolves the port
// and snapshot name and clones the port's data volume through go-pgdocker; it
// contains no CLI, flag, or output-formatting logic. This is the domain tier
// between the app tier (internal/app/commands/snapshot/create) and the
// go-pgdocker library.
package create

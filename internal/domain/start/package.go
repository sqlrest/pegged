// Package start orchestrates the start command.
//
// It defines the command's Config (the flags and arguments the CLI binds) and
// Run (the orchestration entry point the CLI invokes). Run resolves the port,
// validates the enumerated flags, renders the go-pgdocker StartSpec, and
// starts the instance; it contains no CLI, flag, or output-formatting logic.
// This is the domain tier between the app tier
// (internal/app/commands/start) and the go-pgdocker library.
package start

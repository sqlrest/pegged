// Package stop orchestrates the stop command.
//
// It defines the command's Config (the flags and arguments the CLI binds) and
// Run (the orchestration entry point the CLI invokes). Run resolves the port,
// validates the retention flag, and stops the instance through go-pgdocker;
// it contains no CLI, flag, or output-formatting logic. This is the domain
// tier between the app tier (internal/app/commands/stop) and the go-pgdocker
// library.
package stop

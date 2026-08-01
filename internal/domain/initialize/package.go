// Package initialize orchestrates the init command. (The package is named
// initialize because init is reserved in Go; the CLI command is "init".)
//
// It defines the command's Config (the flags and arguments the CLI binds) and
// Run (the orchestration entry point the CLI invokes). Run resolves the port,
// splits the command override off the argv, and runs the initialization
// container against the port's running instance through go-pgdocker; it
// contains no CLI, flag, or output-formatting logic. This is the domain tier
// between the app tier (internal/app/commands/initialize) and the go-pgdocker
// library.
package initialize

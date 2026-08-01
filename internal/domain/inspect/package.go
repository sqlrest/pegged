// Package inspect orchestrates the inspect command.
//
// It defines the command's Config (the flags and arguments the CLI binds) and
// Run (the orchestration entry point the CLI invokes). Run resolves the port
// and reports its managed container and volumes through go-pgdocker; it
// contains no CLI, flag, or output-formatting logic. This is the domain tier
// between the app tier (internal/app/commands/inspect) and the go-pgdocker
// library.
package inspect

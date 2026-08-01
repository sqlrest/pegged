// Package list orchestrates the list command.
//
// It defines the command's Config (empty — list takes no flags) and Run (the
// orchestration entry point the CLI invokes). Run reports every managed port
// through go-pgdocker; it contains no CLI, flag, or output-formatting logic.
// This is the domain tier between the app tier (internal/app/commands/list)
// and the go-pgdocker library.
package list

// Package list orchestrates the "snapshot list" command.
//
// It defines the command's Config (empty — the command takes no flags) and
// Run (the orchestration entry point the CLI invokes). Run reports every
// stored snapshot through go-pgdocker; it contains no CLI, flag, or
// output-formatting logic. This is the domain tier between the app tier
// (internal/app/commands/snapshot/list) and the go-pgdocker library.
package list

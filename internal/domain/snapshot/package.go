// Package snapshot is the domain counterpart of the snapshot parent
// command. The parent only groups its children — create, list, remove — so
// this package carries no Config or Run; each child owns its own domain
// package beneath this one.
package snapshot

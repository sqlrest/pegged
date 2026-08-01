package snapshot

import (
	"github.com/urfave/cli/v3"

	"github.com/sqlrest/pegged/internal/app/commands/snapshot/create"
	"github.com/sqlrest/pegged/internal/app/commands/snapshot/list"
	"github.com/sqlrest/pegged/internal/app/commands/snapshot/remove"
)

const (
	name        = `snapshot`
	usage       = `Manage PostgreSQL data-volume snapshots.`
	description = `Manage snapshots of managed instances' data volumes:

  snapshot create [port] [name] - Snapshot a stopped port's data volume
  snapshot list                 - List every stored snapshot
  snapshot delete <name>        - Remove a stored snapshot

Snapshots are physical PGDATA copies, so they are bound to the PostgreSQL
major version that wrote them: a snapshot boots only on the major that
created it. They are taken from stopped instances only.`
)

// Command returns the parent CLI command definition.
// Parent commands have Commands, not Action.
func Command() *cli.Command {
	return &cli.Command{
		Name:        name,
		Usage:       usage,
		Description: description,
		Commands: []*cli.Command{
			create.Command(),
			remove.Command(),
			list.Command(),
		},
	}
}

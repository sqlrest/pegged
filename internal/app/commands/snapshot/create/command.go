package create

import (
	app "github.com/gomatic/go-app"
	"github.com/urfave/cli/v3"

	domain "github.com/sqlrest/pegged/internal/domain/snapshot/create"
)

const (
	name        = `create`
	usage       = `Snapshot a port's data volume.`
	argUsage    = `[port] [name]`
	description = `Clone the given port's newest data volume into a named snapshot. The
first positional argument is the port (or --port; they must agree), the
second the snapshot name; an omitted name gets a UTC timestamp.

Snapshots are physical PGDATA copies: the port's instance must be stopped,
and the snapshot boots only on the PostgreSQL major version that wrote it.`
)

const (
	copyImageFlag = "copy-image"
	portFlag      = "port"
)

// envPrefix namespaces every flag's environment source.
const envPrefix = "PEGGED_"

var (
	cfg       domain.Config
	runAction = domain.Run
)

// Command returns the CLI command definition.
func Command() *cli.Command {
	return &cli.Command{
		Name:        name,
		Usage:       usage,
		ArgsUsage:   argUsage,
		Description: description,
		Action:      app.Default(&cfg, runAction),
		Flags: []cli.Flag{
			&cli.Uint16Flag{
				Name:        portFlag,
				Sources:     cli.EnvVars(envPrefix+"PORT", "PGPORT"),
				Value:       0,
				Usage:       "Host port to snapshot (alternative to the positional argument)",
				Destination: (*uint16)(&cfg.Port),
			},
			&cli.StringFlag{
				Name:        copyImageFlag,
				Sources:     cli.EnvVars(envPrefix + "COPY_IMAGE"),
				Value:       "",
				Usage:       "Helper image for the volume copy (default: the source's recorded image)",
				Destination: (*string)(&cfg.CopyImage),
			},
		},
	}
}

package list

import (
	app "github.com/gomatic/go-app"
	"github.com/urfave/cli/v3"

	domain "github.com/sqlrest/pegged/internal/domain/list"
)

const (
	name        = `list`
	usage       = `List every managed PostgreSQL port.`
	argUsage    = ``
	description = `Report every managed port on the Docker daemon, ordered by port: its
container (when one exists) and its managed data volumes, newest first.`
)

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
	}
}

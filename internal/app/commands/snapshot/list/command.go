package list

import (
	app "github.com/gomatic/go-app"
	"github.com/urfave/cli/v3"

	domain "github.com/sqlrest/pegged/internal/domain/snapshot/list"
)

const (
	name        = `list`
	usage       = `List every stored snapshot.`
	argUsage    = ``
	description = `Report every stored snapshot in pegged's namespace: its name, volume,
source port, and the PostgreSQL major version that wrote it.`
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

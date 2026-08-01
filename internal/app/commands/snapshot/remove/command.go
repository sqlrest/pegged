package remove

import (
	app "github.com/gomatic/go-app"
	"github.com/urfave/cli/v3"

	domain "github.com/sqlrest/pegged/internal/domain/snapshot/remove"
)

const (
	name        = `delete`
	usage       = `Remove a stored snapshot.`
	argUsage    = `<name>`
	description = `Remove the named snapshot's volume. The name is required.`
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

package inspect

import (
	app "github.com/gomatic/go-app"
	"github.com/urfave/cli/v3"

	domain "github.com/sqlrest/pegged/internal/domain/inspect"
)

const (
	name        = `inspect`
	usage       = `Report one port's managed container and volumes.`
	argUsage    = `[port]`
	description = `Report the managed container (when one exists) and every managed data
volume on the given port (positional argument or --port; they must agree).
A port with volumes but no container is a stopped-but-retained database; a
port with neither yields an empty report.`
)

const (
	portFlag = "port"
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
				Usage:       "Host port to inspect (alternative to the positional argument)",
				Destination: (*uint16)(&cfg.Port),
			},
		},
	}
}

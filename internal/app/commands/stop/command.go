package stop

import (
	app "github.com/gomatic/go-app"
	"github.com/urfave/cli/v3"

	domain "github.com/sqlrest/pegged/internal/domain/stop"
)

const (
	name        = `stop`
	usage       = `Stop a managed PostgreSQL instance.`
	argUsage    = `[port]`
	description = `Stop the managed instance on the given port (positional argument or
--port; they must agree) and apply volume retention: the explicit --retain
choice first, then the retention recorded at start, then the port policy
(at or below 5432 keep, above remove).`
)

const (
	graceFlag  = "grace"
	portFlag   = "port"
	retainFlag = "retain"
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
				Usage:       "Host port of the instance (alternative to the positional argument)",
				Destination: (*uint16)(&cfg.Port),
			},
			&cli.StringFlag{
				Name:        retainFlag,
				Sources:     cli.EnvVars(envPrefix + "RETAIN"),
				Usage:       "Volume retention: keep or remove (default follows the recorded choice, then the port policy)",
				Destination: (*string)(&cfg.Retain),
			},
			&cli.IntFlag{
				Name:        graceFlag,
				Sources:     cli.EnvVars(envPrefix + "GRACE"),
				Usage:       "Shutdown grace period in seconds (default 10)",
				Destination: (*int)(&cfg.Grace),
			},
		},
	}
}

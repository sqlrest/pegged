package initialize

import (
	"os"

	app "github.com/gomatic/go-app"
	"github.com/urfave/cli/v3"

	domain "github.com/sqlrest/pegged/internal/domain/initialize"
)

const (
	name        = `init`
	usage       = `Run an initialization container against a running instance.`
	argUsage    = `[port] [--] [command...]`
	description = `Run the --image container — migrations, seeds, schema loads — against
the running instance on the given port. The first positional argument is
the port only when it is numeric (or use --port); every remaining argument
overrides the container command. When the command override carries flags of
its own, separate it with "--":

  pegged init --image migrate/migrate:4 5432 -- migrate -path /m up

The init container joins the instance's network namespace, so it reaches
the server at localhost:5432 with the standard PG* variables preset. Its
output streams to stderr.

A non-zero exit from the init command fails this command: the process
exits 1 and no JSON result is printed; the exit code is carried in the
error.`
)

const (
	databaseFlag = "database"
	imageFlag    = "image"
	passwordFlag = "password"
	portFlag     = "port"
	userFlag     = "user"
)

// envPrefix namespaces every flag's environment source.
const envPrefix = "PEGGED_"

var (
	cfg       = domain.Config{Output: os.Stderr}
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
				Usage:       "Host port of the instance (alternative to the positional argument)",
				Destination: (*uint16)(&cfg.Port),
			},
			&cli.StringFlag{
				Name:        imageFlag,
				Sources:     cli.EnvVars(envPrefix + "IMAGE"),
				Value:       "",
				Usage:       "Initialization image to run (required)",
				Destination: (*string)(&cfg.Image),
			},
			&cli.StringFlag{
				Name:        databaseFlag,
				Sources:     cli.EnvVars(envPrefix + "DATABASE"),
				Value:       "",
				Usage:       "Target database name (default postgres)",
				Destination: (*string)(&cfg.Database),
			},
			&cli.StringFlag{
				Name:        userFlag,
				Sources:     cli.EnvVars(envPrefix + "USER"),
				Value:       "",
				Usage:       "Connecting role name (default postgres)",
				Destination: (*string)(&cfg.User),
			},
			&cli.StringFlag{
				Name:        passwordFlag,
				Sources:     cli.EnvVars(envPrefix + "PASSWORD"),
				Value:       "",
				Usage:       "Password for the connecting role",
				Destination: (*string)(&cfg.Password),
			},
		},
	}
}

package start

import (
	app "github.com/gomatic/go-app"
	"github.com/urfave/cli/v3"

	domain "github.com/sqlrest/pegged/internal/domain/start"
)

const (
	name        = `start`
	usage       = `Start a managed PostgreSQL instance.`
	argUsage    = `[port]`
	description = `Start a PostgreSQL instance in Docker on the given port (positional
argument or --port; they must agree). Without a port, PostgreSQL's default
5432 is used.

Port policy: at or below 5432 an instance is durable — its data volume is
reused on start and kept on stop; above 5432 it is ephemeral — a fresh
volume per start, removed on stop. Override per call with --volume
(reuse|fresh) and --retain (keep|remove).

Without --password the instance runs trust auth (development posture) and
binds loopback only; widen deliberately with --listen.`
)

const (
	databaseFlag = "database"
	imageFlag    = "image"
	initSQLFlag  = "init-sql"
	listenFlag   = "listen"
	passwordFlag = "password"
	platformFlag = "platform"
	portFlag     = "port"
	retainFlag   = "retain"
	snapshotFlag = "snapshot"
	userFlag     = "user"
	volumeFlag   = "volume"
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
				Usage:       "Host port for the instance (alternative to the positional argument)",
				Destination: (*uint16)(&cfg.Port),
			},
			&cli.StringFlag{
				Name:        imageFlag,
				Sources:     cli.EnvVars(envPrefix + "IMAGE"),
				Value:       "",
				Usage:       "PostgreSQL image to run (default postgres:17-alpine)",
				Destination: (*string)(&cfg.Image),
			},
			&cli.StringFlag{
				Name:        databaseFlag,
				Sources:     cli.EnvVars(envPrefix + "DATABASE"),
				Value:       "",
				Usage:       "Initial database name (default postgres)",
				Destination: (*string)(&cfg.Database),
			},
			&cli.StringFlag{
				Name:        userFlag,
				Sources:     cli.EnvVars(envPrefix + "USER"),
				Value:       "",
				Usage:       "Superuser role name (default postgres)",
				Destination: (*string)(&cfg.User),
			},
			&cli.StringFlag{
				Name:        passwordFlag,
				Sources:     cli.EnvVars(envPrefix + "PASSWORD"),
				Value:       "",
				Usage:       "Superuser password; empty runs trust auth",
				Destination: (*string)(&cfg.Password),
			},
			&cli.StringFlag{
				Name:        snapshotFlag,
				Sources:     cli.EnvVars(envPrefix + "SNAPSHOT"),
				Value:       "",
				Usage:       "Seed the data volume from this snapshot",
				Destination: (*string)(&cfg.Snapshot),
			},
			&cli.StringFlag{
				Name:        volumeFlag,
				Sources:     cli.EnvVars(envPrefix + "VOLUME"),
				Value:       "",
				Usage:       "Data volume source: reuse or fresh (default follows the port policy)",
				Destination: (*string)(&cfg.Volume),
			},
			&cli.StringFlag{
				Name:        retainFlag,
				Sources:     cli.EnvVars(envPrefix + "RETAIN"),
				Value:       "",
				Usage:       "Volume retention on stop: keep or remove (default follows the port policy)",
				Destination: (*string)(&cfg.Retain),
			},
			&cli.StringFlag{
				Name:        listenFlag,
				Sources:     cli.EnvVars(envPrefix + "LISTEN"),
				Value:       "",
				Usage:       "Host interface to bind (default 127.0.0.1)",
				Destination: (*string)(&cfg.Listen),
			},
			&cli.StringFlag{
				Name:        initSQLFlag,
				Sources:     cli.EnvVars(envPrefix + "INIT_SQL"),
				Value:       "",
				Usage:       "Host directory of first-boot *.sql/*.sh scripts",
				Destination: (*string)(&cfg.InitSQL),
			},
			&cli.StringFlag{
				Name:        platformFlag,
				Sources:     cli.EnvVars(envPrefix + "PLATFORM"),
				Value:       "",
				Usage:       "Image platform selector, like linux/arm64",
				Destination: (*string)(&cfg.Platform),
			},
		},
	}
}

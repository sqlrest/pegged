package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sort"

	app "github.com/gomatic/go-app"
	"github.com/gomatic/go-log"
	"github.com/urfave/cli/v3"

	"github.com/sqlrest/pegged/internal/app/commands/initialize"
	"github.com/sqlrest/pegged/internal/app/commands/inspect"
	"github.com/sqlrest/pegged/internal/app/commands/list"
	"github.com/sqlrest/pegged/internal/app/commands/snapshot"
	"github.com/sqlrest/pegged/internal/app/commands/start"
	"github.com/sqlrest/pegged/internal/app/commands/stop"
	pkgversion "github.com/sqlrest/pegged/internal/version"
)

const (
	argUsage    = ``
	description = `Manage local PostgreSQL databases in Docker: a thin shell over the
gomatic/go-pgdocker library, which owns all lifecycle behavior.

Available Commands:
  start    - Start a PostgreSQL instance on a port
  stop     - Stop an instance and apply volume retention
  inspect  - Report one port's managed container and volumes
  list     - Report every managed port
  snapshot - Manage data-volume snapshots (create/list/delete)
  init     - Run an initialization container against a running instance

Port Policy:
  The host port is an instance's identity. At or below 5432 an instance is
  durable — its data volume is reused on start and kept on stop; above 5432
  it is ephemeral — a fresh volume per start, removed on stop. Override per
  call with --volume (reuse|fresh) and --retain (keep|remove).

Snapshots:
  Snapshots are physical PGDATA volume copies taken from stopped instances.
  They are bound to the PostgreSQL major version that wrote them — a
  snapshot boots only on that major.

Command Naming Convention:
  Commands follow a "noun... verb" pattern (resource-oriented design), like
  "snapshot create". Never use hyphens in command names - use hierarchy
  levels instead.

Version:
  Use --version flag (built-in urfave/cli support)`
	envName   = "PEGGED"
	envPrefix = envName + "_"
	name      = `pegged`
	usage     = `Manage local PostgreSQL databases in Docker.`
)

var (
	appCreator    = createApp
	loggerConfig  log.LoggerConfig
	loggerCreator = productionLogger
)

// productionLogger builds the application logger from the parsed logging flags.
// It is invoked from the root Before hook, after flag parsing has populated
// loggerConfig, so --log-level and --log-format take effect.
func productionLogger(_ *cli.Command) *slog.Logger {
	return loggerConfig.NewLogger(os.Stderr)
}

// version is the resolved application version — the same value provenance
// labels record. The ldflags anchor lives in internal/version:
//
//	-X github.com/sqlrest/pegged/internal/version.injected=v1.0.0
var version = string(pkgversion.Resolve())

// osExit is indirected so tests can observe the process exit code.
var osExit = os.Exit

func main() { osExit(run(os.Args)) }

// run builds and executes the CLI, returning the process exit code. Keeping the
// exit code as a return value (rather than calling os.Exit here) makes the whole
// run path testable.
func run(args []string) int {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	if err := appCreator(loggerCreator).Run(ctx, args); err != nil {
		slog.Error("Application error", "error", err)
		return 1
	}
	return 0
}

// createApp constructs the definition of the CLI.
func createApp(getLogger app.GetLoggerFunc) *cli.Command {
	cliApp := &cli.Command{
		Name:                  name,
		Usage:                 usage,
		ArgsUsage:             argUsage,
		Description:           description,
		Version:               version,
		EnableShellCompletion: true,
		Commands: []*cli.Command{
			initialize.Command(),
			inspect.Command(),
			list.Command(),
			snapshot.Command(),
			start.Command(),
			stop.Command(),
		},
		Before: func(ctx context.Context, c *cli.Command) (context.Context, error) {
			c.Root().Metadata[app.LoggerMetadataKey] = getLogger(c)
			return ctx, nil
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "log-level",
				Sources:     cli.EnvVars(envPrefix + "LOG_LEVEL"),
				Value:       "info",
				Usage:       "Set the logging level (debug, info, warn, error)",
				Destination: (*string)(&loggerConfig.Level),
			},
			&cli.StringFlag{
				Name:        "log-format",
				Sources:     cli.EnvVars(envPrefix + "LOG_FORMAT"),
				Value:       "text",
				Usage:       "Set the log output format (text, json)",
				Destination: (*string)(&loggerConfig.Format),
			},
		},
	}

	sort.Sort(cli.FlagsByName(cliApp.Flags))

	return cliApp
}

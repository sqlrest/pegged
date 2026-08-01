package app

import (
	"github.com/gomatic/go-log"
	"github.com/urfave/cli/v3"
)

// EnvPrefix is the namespace prepended to each standard flag's environment
// variable source (e.g. "APP_" yields APP_LOG_LEVEL, APP_LOG_FORMAT, APP_OUTPUT).
type EnvPrefix string

// LoggerFlags binds the standard logging flags to a shared logger
// configuration. The config pointer is carried as a field (the binder pattern)
// so the flag constructors take no pointer parameters; each flag's Destination
// still writes through to Config as the command line is parsed.
type LoggerFlags struct {
	// Config receives the parsed flag values; it must be non-nil.
	Config *log.LoggerConfig
	// EnvPrefix namespaces the flags' environment variable sources.
	EnvPrefix EnvPrefix
}

// LevelFlag returns the standard --log-level global flag, sourced from
// <EnvPrefix>LOG_LEVEL and bound to Config.
func (f LoggerFlags) LevelFlag() cli.Flag {
	return &cli.StringFlag{
		Name:        "log-level",
		Sources:     cli.EnvVars(string(f.EnvPrefix) + "LOG_LEVEL"),
		Value:       "info",
		Usage:       "Logging level (debug, info, warn, error)",
		Destination: (*string)(&f.Config.Level),
	}
}

// FormatFlag returns the standard --log-format global flag, sourced from
// <EnvPrefix>LOG_FORMAT, bound to Config, defaulting to def (consumers differ:
// a CLI defaults to text, a daemon to json).
func (f LoggerFlags) FormatFlag(def log.Format) cli.Flag {
	return &cli.StringFlag{
		Name:        "log-format",
		Sources:     cli.EnvVars(string(f.EnvPrefix) + "LOG_FORMAT"),
		Value:       string(def),
		Usage:       "Log output format (text, json)",
		Destination: (*string)(&f.Config.Format),
	}
}

// OutputFlag returns the standard --output/-o global flag selecting the result
// encoding, sourced from <envPrefix>OUTPUT and defaulting to json. The action
// combinator reads it to encode each command's result.
func OutputFlag(envPrefix EnvPrefix) cli.Flag {
	return &cli.StringFlag{
		Name:    outputFlagName,
		Aliases: []string{"o"},
		Sources: cli.EnvVars(string(envPrefix) + "OUTPUT"),
		Value:   "json",
		Usage:   "Result output format (json, yaml)",
	}
}

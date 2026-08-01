package manage

import (
	"strconv"

	docker "github.com/gomatic/go-docker"

	"github.com/sqlrest/pegged/internal/constants"
	"github.com/sqlrest/pegged/internal/domain"
)

// FlagPort is the port a --port flag supplied; zero means the flag was not
// set.
type FlagPort uint16

// ResolvePort resolves the effective port from the positional argument and
// the --port flag: the positional argument wins when present, then the flag,
// then zero (go-pgdocker applies its own default). A positional argument
// that is not a number in the port range fails with ErrInvalidPort; a
// positional argument and a flag that disagree fail with ErrPortConflict.
func ResolvePort(args []domain.Argument, flagged FlagPort) (docker.Port, error) {
	if len(args) == 0 {
		return docker.Port(flagged), nil
	}
	parsed, err := strconv.ParseUint(args[0], 10, 16)
	if err != nil {
		return 0, constants.ErrInvalidPort.With(err, "argument", args[0])
	}
	positional := docker.Port(parsed)
	if flagged != 0 && docker.Port(flagged) != positional {
		return 0, constants.ErrPortConflict.With(nil,
			"argument", int(positional), "flag", int(flagged))
	}
	return positional, nil
}

package stop

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	app "github.com/gomatic/go-app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"

	domain "github.com/sqlrest/pegged/internal/domain/stop"
)

// stubRun reroutes the domain seam to capture the bound config and args,
// restoring it (and the shared config) when the test ends. Tests using it
// reassign package vars, so they stay serial.
func stubRun(t *testing.T, got *domain.Config, gotArgs *[]string) {
	t.Helper()
	original := runAction
	t.Cleanup(func() { runAction = original; cfg = domain.Config{} })
	runAction = func(_ context.Context, _ *slog.Logger, bound domain.Config, args ...string) (domain.Result, error) {
		*got = bound
		*gotArgs = args
		return domain.Result{}, nil
	}
}

// testApp hosts the command under a root that supplies the logger and
// swallows output.
func testApp() *cli.Command {
	return &cli.Command{
		Name:     "pegged",
		Writer:   &bytes.Buffer{},
		Commands: []*cli.Command{Command()},
		Metadata: map[string]any{
			app.LoggerMetadataKey: slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)),
		},
	}
}

func TestCommand(t *testing.T) {
	t.Parallel()
	want := assert.New(t)

	command := Command()
	want.Equal("stop", command.Name)
	want.Equal("[port]", command.ArgsUsage)
	want.NotEmpty(command.Usage)
	want.NotEmpty(command.Description)
	want.NotNil(command.Action)
	want.Len(command.Flags, 3)
}

func TestCommandBinding(t *testing.T) {
	want, must := assert.New(t), require.New(t)
	var got domain.Config
	var gotArgs []string
	stubRun(t, &got, &gotArgs)

	err := testApp().Run(context.Background(), []string{
		"pegged", "stop", "--port", "5433", "--retain", "remove", "--grace", "5", "5433",
	})
	must.NoError(err)

	want.Equal(domain.Config{Retain: "remove", Grace: 5, Port: 5433}, got)
	want.Equal([]string{"5433"}, gotArgs)
}

// TestCommandEnvSources asserts the PEGGED_* (and PGPORT) environment
// contract binds without flags.
func TestCommandEnvSources(t *testing.T) {
	want, must := assert.New(t), require.New(t)
	var got domain.Config
	var gotArgs []string
	stubRun(t, &got, &gotArgs)
	t.Setenv("PEGGED_RETAIN", "keep")
	t.Setenv("PGPORT", "6001")

	err := testApp().Run(context.Background(), []string{"pegged", "stop"})
	must.NoError(err)

	want.Equal(domain.Config{Retain: "keep", Port: 6001}, got)
	want.Empty(gotArgs)
}

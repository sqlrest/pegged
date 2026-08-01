package initialize

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"testing"

	app "github.com/gomatic/go-app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"

	domain "github.com/sqlrest/pegged/internal/domain/initialize"
)

// stubRun reroutes the domain seam to capture the bound config and args,
// restoring it (and the shared config, with its stderr output) when the
// test ends. Tests using it reassign package vars, so they stay serial.
func stubRun(t *testing.T, got *domain.Config, gotArgs *[]string) {
	t.Helper()
	original := runAction
	t.Cleanup(func() { runAction = original; cfg = domain.Config{Output: os.Stderr} })
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
	want.Equal("init", command.Name)
	want.Equal("[port] [--] [command...]", command.ArgsUsage)
	want.NotEmpty(command.Usage)
	want.NotEmpty(command.Description)
	want.NotNil(command.Action)
	want.Len(command.Flags, 5)
}

func TestCommandBinding(t *testing.T) {
	want, must := assert.New(t), require.New(t)
	// The injected output writer is compared by identity against the package
	// var: under `go test -json` the harness swaps os.Stderr mid-run, so the
	// assert-time os.Stderr is not a stable expectation.
	wantOutput := cfg.Output
	var got domain.Config
	var gotArgs []string
	stubRun(t, &got, &gotArgs)

	err := testApp().Run(context.Background(), []string{
		"pegged", "init",
		"--port", "5433", "--image", "migrate/migrate:4", "--database", "app",
		"--user", "dev", "--password", "hunter2",
		"5433", "migrate", "up",
	})
	must.NoError(err)

	want.Equal(domain.Config{
		Output:   wantOutput,
		Image:    "migrate/migrate:4",
		Database: "app",
		User:     "dev",
		Password: "hunter2",
		Port:     5433,
	}, got)
	want.NotNil(got.Output, "the init output writer must be injected")
	want.Equal([]string{"5433", "migrate", "up"}, gotArgs)
}

// TestCommandEnvSources asserts the PEGGED_* (and PGPORT) environment
// contract binds without flags, and that the init output stays wired to
// stderr.
func TestCommandEnvSources(t *testing.T) {
	want, must := assert.New(t), require.New(t)
	wantOutput := cfg.Output
	var got domain.Config
	var gotArgs []string
	stubRun(t, &got, &gotArgs)
	t.Setenv("PEGGED_IMAGE", "migrate/migrate:4")
	t.Setenv("PGPORT", "6001")

	err := testApp().Run(context.Background(), []string{"pegged", "init"})
	must.NoError(err)

	want.Equal(domain.Config{Output: wantOutput, Image: "migrate/migrate:4", Port: 6001}, got)
	want.NotNil(got.Output, "the init output writer must be injected")
	want.Empty(gotArgs)
}

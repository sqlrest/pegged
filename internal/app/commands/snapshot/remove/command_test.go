package remove

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	app "github.com/gomatic/go-app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"

	domain "github.com/sqlrest/pegged/internal/domain/snapshot/remove"
)

// stubRun reroutes the domain seam to capture the args, restoring it when
// the test ends. Tests using it reassign a package var, so they stay serial.
func stubRun(t *testing.T, gotArgs *[]string) {
	t.Helper()
	original := runAction
	t.Cleanup(func() { runAction = original })
	runAction = func(_ context.Context, _ *slog.Logger, _ domain.Config, args ...string) (domain.Result, error) {
		*gotArgs = args
		return domain.Result{}, nil
	}
}

func TestCommand(t *testing.T) {
	t.Parallel()
	want := assert.New(t)

	command := Command()
	want.Equal("delete", command.Name)
	want.Equal("<name>", command.ArgsUsage)
	want.NotEmpty(command.Usage)
	want.NotEmpty(command.Description)
	want.NotNil(command.Action)
	want.Empty(command.Flags)
}

func TestCommandBinding(t *testing.T) {
	want, must := assert.New(t), require.New(t)
	var gotArgs []string
	stubRun(t, &gotArgs)

	testApp := &cli.Command{
		Name:     "pegged",
		Writer:   &bytes.Buffer{},
		Commands: []*cli.Command{Command()},
		Metadata: map[string]any{
			app.LoggerMetadataKey: slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)),
		},
	}
	err := testApp.Run(context.Background(), []string{"pegged", "delete", "golden"})
	must.NoError(err)
	want.Equal([]string{"golden"}, gotArgs)
}

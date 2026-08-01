package list

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	app "github.com/gomatic/go-app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"

	domain "github.com/sqlrest/pegged/internal/domain/snapshot/list"
)

// stubRun reroutes the domain seam to record the invocation, restoring it
// when the test ends. Tests using it reassign a package var, so they stay
// serial.
func stubRun(t *testing.T, invoked *int) {
	t.Helper()
	original := runAction
	t.Cleanup(func() { runAction = original })
	runAction = func(context.Context, *slog.Logger, domain.Config, ...string) (domain.Result, error) {
		*invoked++
		return domain.Result{}, nil
	}
}

func TestCommand(t *testing.T) {
	t.Parallel()
	want := assert.New(t)

	command := Command()
	want.Equal("list", command.Name)
	want.Empty(command.ArgsUsage)
	want.NotEmpty(command.Usage)
	want.NotEmpty(command.Description)
	want.NotNil(command.Action)
	want.Empty(command.Flags)
}

func TestCommandInvokesRun(t *testing.T) {
	want, must := assert.New(t), require.New(t)
	invoked := 0
	stubRun(t, &invoked)

	testApp := &cli.Command{
		Name:     "pegged",
		Writer:   &bytes.Buffer{},
		Commands: []*cli.Command{Command()},
		Metadata: map[string]any{
			app.LoggerMetadataKey: slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)),
		},
	}
	err := testApp.Run(context.Background(), []string{"pegged", "list"})
	must.NoError(err)
	want.Equal(1, invoked)
}

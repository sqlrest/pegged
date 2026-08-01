package main

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

	"github.com/sqlrest/pegged/internal/constants"
)

func TestRun_Version(t *testing.T) {
	tests := []struct {
		name         string
		wantContains string
		args         []string
	}{
		{
			name:         "version flag outputs version",
			args:         []string{"pegged", "--version"},
			wantContains: version,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want, must := assert.New(t), require.New(t)

			var stdout bytes.Buffer

			logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
				Level: slog.LevelWarn,
			}))

			// Create app and run with --version flag
			app := createApp(func(_ *cli.Command) *slog.Logger { return logger })
			app.Writer = &stdout

			err := app.Run(context.Background(), tt.args)
			must.NoError(err)

			output := stdout.String()
			want.Contains(output, tt.wantContains)
		})
	}
}

func TestCreateApp(t *testing.T) {
	tests := []struct {
		name             string
		expectedName     string
		expectedVersion  string
		expectedCommands []string
	}{
		{
			name:             "creates app with correct name and version",
			expectedName:     name,
			expectedVersion:  version,
			expectedCommands: []string{"init", "inspect", "list", "snapshot", "start", "stop"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want, must := assert.New(t), require.New(t)

			logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
			app := createApp(func(_ *cli.Command) *slog.Logger { return logger })

			want.Equal(tt.expectedName, app.Name)
			want.Equal(tt.expectedVersion, app.Version)
			must.NotEmpty(app.Commands, "expected app to have commands")

			// Verify expected commands exist
			for _, expected := range tt.expectedCommands {
				found := false
				for _, cmd := range app.Commands {
					if cmd.Name == expected {
						found = true
						break
					}
				}
				want.True(found, "expected command %q not found", expected)
			}
		})
	}
}

// TestRun_Help runs the root command with no arguments, which prints help
// and exercises the Before hook that installs the logger.
func TestRun_Help(t *testing.T) {
	want, must := assert.New(t), require.New(t)

	var stdout bytes.Buffer
	var logBuf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logBuf, &slog.HandlerOptions{
		Level: slog.LevelWarn,
	}))

	app := createApp(func(_ *cli.Command) *slog.Logger { return logger })
	app.Writer = &stdout

	err := app.Run(context.Background(), []string{"pegged"})
	must.NoError(err)

	output := stdout.String()
	want.Contains(output, "start")
	want.Contains(output, "snapshot")
	want.Contains(output, "init")
}

func TestRun_ExitCodes(t *testing.T) {
	original := appCreator
	t.Cleanup(func() { appCreator = original })

	want := assert.New(t)

	appCreator = func(app.GetLoggerFunc) *cli.Command {
		return &cli.Command{Name: "x", Writer: &bytes.Buffer{}}
	}
	want.Equal(0, run([]string{"x"}), "successful run exits 0")

	appCreator = func(app.GetLoggerFunc) *cli.Command {
		return &cli.Command{
			Name:   "x",
			Writer: &bytes.Buffer{},
			Action: func(context.Context, *cli.Command) error { return constants.ErrInvalidValue },
		}
	}
	want.Equal(1, run([]string{"x"}), "failed run exits 1")
}

func TestMainEntry(t *testing.T) {
	originalCreator, originalExit, originalArgs := appCreator, osExit, os.Args
	t.Cleanup(func() { appCreator, osExit, os.Args = originalCreator, originalExit, originalArgs })

	var code int
	osExit = func(c int) { code = c }
	appCreator = func(app.GetLoggerFunc) *cli.Command {
		return &cli.Command{Name: "x", Writer: &bytes.Buffer{}}
	}
	os.Args = []string{"x"}

	main()
	assert.New(t).Equal(0, code)
}

func TestProductionLogger(t *testing.T) {
	assert.New(t).NotNil(productionLogger(nil))
}

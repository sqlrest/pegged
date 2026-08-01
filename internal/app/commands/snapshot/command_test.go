package snapshot

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommand(t *testing.T) {
	t.Parallel()
	want, must := assert.New(t), require.New(t)

	command := Command()
	want.Equal("snapshot", command.Name)
	want.NotEmpty(command.Usage)
	want.NotEmpty(command.Description)
	must.NotEmpty(command.Commands, "snapshot should expose subcommands")

	names := make([]string, 0, len(command.Commands))
	for _, sub := range command.Commands {
		names = append(names, sub.Name)
	}
	want.ElementsMatch([]string{"create", "delete", "list"}, names)
}

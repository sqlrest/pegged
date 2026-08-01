package create

// Config holds the flags and arguments for the "snapshot create" command.
// Its fields are bound by the CLI tier and read by Run; it carries no
// behavior.
type Config struct {
	CopyImage copyImage
	Port      flagPort
}

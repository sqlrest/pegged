package inspect

// Config holds the flags and arguments for the inspect command. Its fields
// are bound by the CLI tier and read by Run; it carries no behavior.
type Config struct {
	Port flagPort
}

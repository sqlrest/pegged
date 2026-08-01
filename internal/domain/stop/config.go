package stop

// Config holds the flags and arguments for the stop command. Its fields are
// bound by the CLI tier and read by Run; it carries no behavior.
type Config struct {
	Retain retention
	Grace  graceSeconds
	Port   flagPort
}

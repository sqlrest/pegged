package inspect

// Named types for every Config field. The CLI binds flags to these via
// pointer conversion (e.g. (*uint16)(&cfg.Port)); naming the domain concept
// keeps the Config self-describing and avoids bare primitives.
type (
	flagPort uint16 // flagPort is the --port flag; zero means unset.
)

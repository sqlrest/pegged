package stop

// Named types for every Config field. The CLI binds flags to these via pointer
// conversion (e.g. (*int)(&cfg.Grace)); naming the domain concept keeps the
// Config self-describing and avoids bare primitives.
type (
	retention    string // retention selects keep|remove for the data volume (--retain).
	graceSeconds int    // graceSeconds is the shutdown grace period (--grace).
	flagPort     uint16 // flagPort is the --port flag; zero means unset.
)

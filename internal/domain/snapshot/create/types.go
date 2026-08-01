package create

// Named types for every Config field. The CLI binds flags to these via
// pointer conversion (e.g. (*string)(&cfg.CopyImage)); naming the domain
// concept keeps the Config self-describing and avoids bare primitives.
type (
	copyImage string // copyImage is the helper image for the volume copy (--copy-image).
	flagPort  uint16 // flagPort is the --port flag; zero means unset.
)

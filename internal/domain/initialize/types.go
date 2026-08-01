package initialize

import "io"

// Named types for every Config field. The CLI binds flags to these via
// pointer conversion (e.g. (*string)(&cfg.Image)); naming the domain concept
// keeps the Config self-describing and avoids bare primitives.
type (
	initOutput   io.Writer // initOutput streams the init container's output; the CLI injects os.Stderr.
	imageRef     string    // imageRef is the initialization image to run (--image).
	databaseName string    // databaseName is the target database (--database).
	userName     string    // userName is the connecting role (--user).
	password     string    // password authenticates the connection (--password).
	flagPort     uint16    // flagPort is the --port flag; zero means unset.
)

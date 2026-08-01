package start

// Named types for every Config field. The CLI binds flags to these via pointer
// conversion (e.g. (*string)(&cfg.Image)); naming the domain concept keeps
// the Config self-describing and avoids bare primitives.
type (
	imageRef      string // imageRef is the PostgreSQL image to run (--image).
	databaseName  string // databaseName is the initial database (--database).
	userName      string // userName is the superuser role (--user).
	password      string // password enables password auth; empty means trust (--password).
	snapshotName  string // snapshotName seeds the data volume from a snapshot (--snapshot).
	volumeSource  string // volumeSource selects reuse|fresh (--volume).
	retention     string // retention selects keep|remove on stop (--retain).
	listenAddress string // listenAddress is the host interface to bind (--listen).
	initSQLDir    string // initSQLDir mounts first-boot *.sql/*.sh scripts (--init-sql).
	platform      string // platform is the image platform selector (--platform).
	flagPort      uint16 // flagPort is the --port flag; zero means unset.
)

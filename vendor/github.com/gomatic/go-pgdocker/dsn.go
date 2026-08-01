package pgdocker

// Connection vocabulary for a managed instance. The library configures the
// server's bootstrap identity without claiming to know how to connect to
// it: initialization scripts and later administration own the database's
// roles, databases, and passwords, so no connection URL is derived or
// reported — the socket facts (port, listen address) on the instance are
// the library's whole connection statement.
type (
	// DatabaseName is the PostgreSQL database an instance serves.
	DatabaseName string
	// UserName is the PostgreSQL role an instance authenticates.
	UserName string
	// Password is the role's password; empty selects trust auth (a
	// development posture — the instance accepts any local connection).
	Password string
)

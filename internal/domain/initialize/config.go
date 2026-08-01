package initialize

// Config holds the flags and arguments for the init command. Its fields are
// bound (and Output injected) by the CLI tier and read by Run; it carries no
// behavior.
type Config struct {
	Output   initOutput
	Image    imageRef
	Database databaseName
	User     userName
	Password password
	Port     flagPort
}

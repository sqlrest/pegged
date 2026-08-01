package start

// Config holds the flags and arguments for the start command. Its fields are
// bound by the CLI tier and read by Run; it carries no behavior.
type Config struct {
	Image    imageRef
	Database databaseName
	User     userName
	Password password
	Snapshot snapshotName
	Volume   volumeSource
	Retain   retention
	Listen   listenAddress
	InitSQL  initSQLDir
	Platform platform
	Port     flagPort
}

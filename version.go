package main

import "fmt"

const (
	VersionMajor   = 0
	VersionMinor   = 3
	VersionPatch   = 0
	VersionRelease = "-dev" // -dev -release etc.
)

var Version = fmt.Sprintf("%d.%d.%d%s", VersionMajor, VersionMinor, VersionPatch, VersionRelease)

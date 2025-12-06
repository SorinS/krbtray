package main

import "fmt"

const (
	VersionMajor = 0
	VersionMinor = 1
	VersionPatch
	VersionRelease = "-dev" // -dev -release etc.
)

var Version = fmt.Sprintf("%d.%d.%d%s", VersionMajor, VersionMinor, VersionPatch, VersionRelease)

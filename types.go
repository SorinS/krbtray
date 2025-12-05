// Package main provides cross-platform Kerberos service ticket acquisition.
package main

// Global debug flag
var debugMode bool

// SetDebugMode enables or disables debug output
func SetDebugMode(debug bool) {
	debugMode = debug
}

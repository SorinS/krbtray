//go:build linux
// +build linux

package main

import "os/exec"

// openBrowser opens the specified URL in the default browser
func openBrowser(url string) error {
	return exec.Command("xdg-open", url).Start()
}
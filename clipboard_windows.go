//go:build windows
// +build windows

package main

import (
	"os/exec"
)

func copyToClipboardPlatform(text string) error {
	cmd := exec.Command("cmd", "/c", "echo|set /p="+text+"|clip")
	return cmd.Run()
}

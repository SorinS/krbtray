//go:build darwin
// +build darwin

package main

import (
	"os/exec"
	"strings"
)

func copyToClipboardPlatform(text string) error {
	cmd := exec.Command("pbcopy")
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

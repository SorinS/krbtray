//go:build linux
// +build linux

package main

import (
	"os/exec"
	"strings"
)

func copyToClipboardPlatform(text string) error {
	// Try xclip first, then xsel
	cmd := exec.Command("xclip", "-selection", "clipboard")
	cmd.Stdin = strings.NewReader(text)
	if err := cmd.Run(); err == nil {
		return nil
	}

	// Fallback to xsel
	cmd = exec.Command("xsel", "--clipboard", "--input")
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

//go:build !darwin && !windows && !linux
// +build !darwin,!windows,!linux

package main

import "fmt"

func copyToClipboardPlatform(text string) error {
	return fmt.Errorf("clipboard not supported on this platform")
}

// pasteFromClipboard is not implemented on this platform
func pasteFromClipboard() {
	// Not implemented
}

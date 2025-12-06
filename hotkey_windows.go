//go:build windows

package main

import "golang.design/x/hotkey"

// getSnippetHotkeyModifiers returns the platform-specific modifiers for the snippet hotkey
// Windows: Ctrl+Alt+K
func getSnippetHotkeyModifiers() ([]hotkey.Modifier, string) {
	return []hotkey.Modifier{hotkey.ModCtrl, hotkey.ModAlt}, "Ctrl+Alt+K"
}
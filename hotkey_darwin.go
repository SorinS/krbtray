//go:build darwin

package main

import "golang.design/x/hotkey"

// getSnippetHotkeyModifiers returns the platform-specific modifiers for the snippet hotkey
// macOS: Command+Option+K
func getSnippetHotkeyModifiers() ([]hotkey.Modifier, string) {
	return []hotkey.Modifier{hotkey.ModOption, hotkey.ModCmd}, "Cmd+Option+K"
}
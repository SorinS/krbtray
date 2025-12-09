//go:build darwin

package main

import "golang.design/x/hotkey"

// getSnippetHotkeyModifiers returns the platform-specific modifiers for the snippet hotkey
// macOS: Command+Option+[0-9]
func getSnippetHotkeyModifiers() ([]hotkey.Modifier, string) {
	return []hotkey.Modifier{hotkey.ModOption, hotkey.ModCmd}, "Cmd+Option"
}

// getURLHotkeyModifiers returns the platform-specific modifiers for the URL hotkey
// macOS: Control+Command+[0-9]
func getURLHotkeyModifiers() ([]hotkey.Modifier, string) {
	return []hotkey.Modifier{hotkey.ModCtrl, hotkey.ModCmd}, "Ctrl+Cmd"
}

// getSSHHotkeyModifiers returns the platform-specific modifiers for the SSH hotkey
// macOS: Control+Option+[0-9]
func getSSHHotkeyModifiers() ([]hotkey.Modifier, string) {
	return []hotkey.Modifier{hotkey.ModCtrl, hotkey.ModOption}, "Ctrl+Option"
}
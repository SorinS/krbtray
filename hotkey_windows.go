//go:build windows

package main

import "golang.design/x/hotkey"

// getSnippetHotkeyModifiers returns the platform-specific modifiers for the snippet hotkey
// Windows: Ctrl+Alt+[0-9]
func getSnippetHotkeyModifiers() ([]hotkey.Modifier, string) {
	return []hotkey.Modifier{hotkey.ModCtrl, hotkey.ModAlt}, "Ctrl+Alt"
}

// getURLHotkeyModifiers returns the platform-specific modifiers for the URL hotkey
// Windows: Ctrl+Shift+[0-9]
func getURLHotkeyModifiers() ([]hotkey.Modifier, string) {
	return []hotkey.Modifier{hotkey.ModCtrl, hotkey.ModShift}, "Ctrl+Shift"
}

// getSSHHotkeyModifiers returns the platform-specific modifiers for the SSH hotkey
// Windows: Alt+Shift+[0-9]
func getSSHHotkeyModifiers() ([]hotkey.Modifier, string) {
	return []hotkey.Modifier{hotkey.ModAlt, hotkey.ModShift}, "Alt+Shift"
}
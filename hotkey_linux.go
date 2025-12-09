//go:build linux

package main

import "golang.design/x/hotkey"

// getSnippetHotkeyModifiers returns the platform-specific modifiers for the snippet hotkey
// Linux: Ctrl+Alt+[0-9] (Mod1 is typically Alt on X11)
func getSnippetHotkeyModifiers() ([]hotkey.Modifier, string) {
	return []hotkey.Modifier{hotkey.ModCtrl, hotkey.Mod1}, "Ctrl+Alt"
}

// getURLHotkeyModifiers returns the platform-specific modifiers for the URL hotkey
// Linux: Ctrl+Shift+[0-9]
func getURLHotkeyModifiers() ([]hotkey.Modifier, string) {
	return []hotkey.Modifier{hotkey.ModCtrl, hotkey.ModShift}, "Ctrl+Shift"
}

// getSSHHotkeyModifiers returns the platform-specific modifiers for the SSH hotkey
// Linux: Alt+Shift+[0-9] (Mod1 is typically Alt on X11)
func getSSHHotkeyModifiers() ([]hotkey.Modifier, string) {
	return []hotkey.Modifier{hotkey.Mod1, hotkey.ModShift}, "Alt+Shift"
}
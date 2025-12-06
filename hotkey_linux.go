//go:build linux

package main

import "golang.design/x/hotkey"

// getSnippetHotkeyModifiers returns the platform-specific modifiers for the snippet hotkey
// Linux: Ctrl+Alt+K (Mod1 is typically Alt on X11)
func getSnippetHotkeyModifiers() ([]hotkey.Modifier, string) {
	return []hotkey.Modifier{hotkey.ModCtrl, hotkey.Mod1}, "Ctrl+Alt+K"
}
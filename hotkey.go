package main

import (
	"fmt"
	"time"

	"golang.design/x/hotkey"
)

var (
	snippetHotkeys [10]*hotkey.Hotkey // Cmd+Option+0 through Cmd+Option+9
)

// InitHotkeys initializes global hotkey support
// Called from onReady after systray is initialized
func InitHotkeys() {
	go initHotkeysAsync()
}

func initHotkeysAsync() {
	// Small delay to ensure systray is fully initialized
	time.Sleep(500 * time.Millisecond)

	// Get platform-specific modifiers
	mods, hotkeyDesc := getSnippetHotkeyModifiers()

	// Register Cmd+Option+0 through Cmd+Option+9 (or Ctrl+Alt+0-9 on Windows/Linux)
	keys := []hotkey.Key{
		hotkey.Key0, hotkey.Key1, hotkey.Key2, hotkey.Key3, hotkey.Key4,
		hotkey.Key5, hotkey.Key6, hotkey.Key7, hotkey.Key8, hotkey.Key9,
	}

	registeredCount := 0
	var errors []string
	for i, key := range keys {
		hk := hotkey.New(mods, key)
		if err := hk.Register(); err != nil {
			errors = append(errors, fmt.Sprintf("%d: %v", i, err))
			continue
		}
		snippetHotkeys[i] = hk
		registeredCount++

		// Start listener for this hotkey
		go func(num int, h *hotkey.Hotkey) {
			for range h.Keydown() {
				copySnippetByIndex(num)
			}
		}(i, hk)
	}

	// Always show registration result in status
	if registeredCount > 0 {
		mStatus.SetTitle(fmt.Sprintf("Hotkeys ready: %s+[0-9]", hotkeyDesc))
	} else if len(errors) > 0 {
		mStatus.SetTitle(fmt.Sprintf("Hotkey error: %s", errors[0]))
		fmt.Printf("Hotkey registration errors: %v\n", errors)
	}

	if debugMode {
		fmt.Printf("Registered %d snippet hotkeys (%s+0 through %s+9)\n", registeredCount, hotkeyDesc, hotkeyDesc)
		if len(errors) > 0 {
			fmt.Printf("Errors: %v\n", errors)
		}
	}
}

func copySnippetByIndex(num int) {
	// Find snippet with matching index
	if appConfig == nil || len(appConfig.Snippets) == 0 {
		mStatus.SetTitle("No snippets configured")
		return
	}

	for _, snippet := range appConfig.Snippets {
		if snippet.Index == num {
			if err := copyToClipboard(snippet.Value); err != nil {
				mStatus.SetTitle(fmt.Sprintf("Copy failed: %s", snippet.Name))
			} else {
				mStatus.SetTitle(fmt.Sprintf("Copied: [%d] %s", num, snippet.Name))
			}
			return
		}
	}

	mStatus.SetTitle(fmt.Sprintf("No snippet with index %d", num))
}

// CleanupHotkeys unregisters all hotkeys
func CleanupHotkeys() {
	for _, hk := range snippetHotkeys {
		if hk != nil {
			hk.Unregister()
		}
	}
}
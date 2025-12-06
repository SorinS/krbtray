package main

import (
	"fmt"
	"time"

	"golang.design/x/hotkey"
)

var (
	snippetHotkeys [10]*hotkey.Hotkey // Cmd+Option+0 through Cmd+Option+9

	snippetInput   string      // Accumulated digit input (e.g., "1", "15")
	snippetTimeout *time.Timer // Timeout for multi-digit input
	inputTimeout   = 1 * time.Second
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

	// Register Cmd+Option+0 through Cmd+Option+9
	// For multi-digit snippets (e.g., 12), user presses Cmd+Option+1, then Cmd+Option+2
	keys := []hotkey.Key{
		hotkey.Key0, hotkey.Key1, hotkey.Key2, hotkey.Key3, hotkey.Key4,
		hotkey.Key5, hotkey.Key6, hotkey.Key7, hotkey.Key8, hotkey.Key9,
	}

	registeredCount := 0
	for i, key := range keys {
		hk := hotkey.New(mods, key)
		if err := hk.Register(); err != nil {
			if debugMode {
				fmt.Printf("Failed to register hotkey %s+%d: %v\n", hotkeyDesc, i, err)
			}
			continue
		}
		snippetHotkeys[i] = hk
		registeredCount++

		// Start listener for this hotkey
		go func(num int, h *hotkey.Hotkey) {
			for range h.Keydown() {
				handleSnippetDigit(num)
			}
		}(i, hk)
	}

	if registeredCount > 0 {
		mStatus.SetTitle(fmt.Sprintf("Hotkey: %s+[0-9]", hotkeyDesc))
	}

	if debugMode {
		fmt.Printf("Registered %d snippet hotkeys (%s+0 through %s+9)\n", registeredCount, hotkeyDesc, hotkeyDesc)
	}
}

// handleSnippetDigit handles Cmd+Option+N presses and accumulates digits
func handleSnippetDigit(num int) {
	stateMutex.Lock()
	snippetInput += fmt.Sprintf("%d", num)
	currentInput := snippetInput
	stateMutex.Unlock()

	// Update status to show current input
	mStatus.SetTitle(fmt.Sprintf("Snippet #%s...", currentInput))

	// Reset/start timeout - after 1 second of no more digits, select the snippet
	if snippetTimeout != nil {
		snippetTimeout.Stop()
	}
	snippetTimeout = time.AfterFunc(inputTimeout, finalizeSnippetSelection)
}

func finalizeSnippetSelection() {
	stateMutex.Lock()
	input := snippetInput
	snippetInput = "" // Reset for next time
	stateMutex.Unlock()

	if input != "" {
		selectSnippetByInput(input)
	}
}

func selectSnippetByInput(input string) {
	if input == "" {
		mStatus.SetTitle("No snippet number entered")
		return
	}

	// Parse the number
	var num int
	fmt.Sscanf(input, "%d", &num)

	copySnippetByIndex(num)
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
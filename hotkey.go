package main

import (
	"fmt"
	"time"

	"golang.design/x/hotkey"
)

var (
	snippetHotkeys [10]*hotkey.Hotkey // Cmd+Option+0 through Cmd+Option+9
	urlHotkeys     [10]*hotkey.Hotkey // Ctrl+Cmd+0 through Ctrl+Cmd+9

	snippetInput   string      // Accumulated digit input (e.g., "1", "15")
	snippetTimeout *time.Timer // Timeout for multi-digit input

	urlInput   string      // Accumulated digit input for URLs
	urlTimeout *time.Timer // Timeout for multi-digit URL input

	inputTimeout = 1 * time.Second
)

// InitHotkeys initializes global hotkey support
// Called from onReady after systray is initialized
func InitHotkeys() {
	go initHotkeysAsync()
}

func initHotkeysAsync() {
	// Small delay to ensure systray is fully initialized
	time.Sleep(500 * time.Millisecond)

	keys := []hotkey.Key{
		hotkey.Key0, hotkey.Key1, hotkey.Key2, hotkey.Key3, hotkey.Key4,
		hotkey.Key5, hotkey.Key6, hotkey.Key7, hotkey.Key8, hotkey.Key9,
	}

	// Register snippet hotkeys (Cmd+Option+0 through Cmd+Option+9)
	snippetMods, snippetDesc := getSnippetHotkeyModifiers()
	snippetCount := 0
	for i, key := range keys {
		hk := hotkey.New(snippetMods, key)
		if err := hk.Register(); err != nil {
			if debugMode {
				fmt.Printf("Failed to register hotkey %s+%d: %v\n", snippetDesc, i, err)
			}
			continue
		}
		snippetHotkeys[i] = hk
		snippetCount++

		// Start listener for this hotkey
		go func(num int, h *hotkey.Hotkey) {
			for range h.Keydown() {
				handleSnippetDigit(num)
			}
		}(i, hk)
	}

	// Register URL hotkeys (Ctrl+Cmd+0 through Ctrl+Cmd+9)
	urlMods, urlDesc := getURLHotkeyModifiers()
	urlCount := 0
	for i, key := range keys {
		hk := hotkey.New(urlMods, key)
		if err := hk.Register(); err != nil {
			if debugMode {
				fmt.Printf("Failed to register hotkey %s+%d: %v\n", urlDesc, i, err)
			}
			continue
		}
		urlHotkeys[i] = hk
		urlCount++

		// Start listener for this hotkey
		go func(num int, h *hotkey.Hotkey) {
			for range h.Keydown() {
				handleURLDigit(num)
			}
		}(i, hk)
	}

	if snippetCount > 0 || urlCount > 0 {
		mStatus.SetTitle(fmt.Sprintf("Hotkeys: %s (snippets), %s (URLs)", snippetDesc, urlDesc))
	}

	if debugMode {
		fmt.Printf("Registered %d snippet hotkeys (%s+[0-9])\n", snippetCount, snippetDesc)
		fmt.Printf("Registered %d URL hotkeys (%s+[0-9])\n", urlCount, urlDesc)
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

// handleURLDigit handles Ctrl+Cmd+N presses and accumulates digits
func handleURLDigit(num int) {
	stateMutex.Lock()
	urlInput += fmt.Sprintf("%d", num)
	currentInput := urlInput
	stateMutex.Unlock()

	// Update status to show current input
	mStatus.SetTitle(fmt.Sprintf("URL #%s...", currentInput))

	// Reset/start timeout - after 1 second of no more digits, open the URL
	if urlTimeout != nil {
		urlTimeout.Stop()
	}
	urlTimeout = time.AfterFunc(inputTimeout, finalizeURLSelection)
}

func finalizeURLSelection() {
	stateMutex.Lock()
	input := urlInput
	urlInput = "" // Reset for next time
	stateMutex.Unlock()

	if input != "" {
		selectURLByInput(input)
	}
}

func selectURLByInput(input string) {
	if input == "" {
		mStatus.SetTitle("No URL number entered")
		return
	}

	// Parse the number
	var num int
	fmt.Sscanf(input, "%d", &num)

	openURLByIndex(num)
}

func openURLByIndex(num int) {
	// Find URL with matching index
	if appConfig == nil || len(appConfig.URLs) == 0 {
		mStatus.SetTitle("No URLs configured")
		return
	}

	for _, url := range appConfig.URLs {
		if url.Index == num {
			if err := openBrowser(url.URL); err != nil {
				mStatus.SetTitle(fmt.Sprintf("Failed to open: %s", url.Name))
			} else {
				mStatus.SetTitle(fmt.Sprintf("Opened: [%d] %s", num, url.Name))
			}
			return
		}
	}

	mStatus.SetTitle(fmt.Sprintf("No URL with index %d", num))
}

// CleanupHotkeys unregisters all hotkeys
func CleanupHotkeys() {
	for _, hk := range snippetHotkeys {
		if hk != nil {
			hk.Unregister()
		}
	}
	for _, hk := range urlHotkeys {
		if hk != nil {
			hk.Unregister()
		}
	}
}
package main

import (
	"fmt"
	"time"

	"golang.design/x/hotkey"
)

// Note: fmt is still needed for Sscanf and Sprintf for non-logging purposes

var (
	snippetHotkeys [10]*hotkey.Hotkey // Cmd+Option+0 through Cmd+Option+9
	urlHotkeys     [10]*hotkey.Hotkey // Ctrl+Cmd+0 through Ctrl+Cmd+9
	sshHotkeys     [10]*hotkey.Hotkey // Ctrl+Option+0 through Ctrl+Option+9

	snippetInput   string      // Accumulated digit input (e.g., "1", "15")
	snippetTimeout *time.Timer // Timeout for multi-digit input

	urlInput   string      // Accumulated digit input for URLs
	urlTimeout *time.Timer // Timeout for multi-digit URL input

	sshInput   string      // Accumulated digit input for SSH
	sshTimeout *time.Timer // Timeout for multi-digit SSH input

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
			LogDebug("Failed to register hotkey %s+%d: %v", snippetDesc, i, err)
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
			LogDebug("Failed to register hotkey %s+%d: %v", urlDesc, i, err)
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

	// Register SSH hotkeys (Ctrl+Option+0 through Ctrl+Option+9)
	sshMods, sshDesc := getSSHHotkeyModifiers()
	sshCount := 0
	for i, key := range keys {
		hk := hotkey.New(sshMods, key)
		if err := hk.Register(); err != nil {
			LogDebug("Failed to register hotkey %s+%d: %v", sshDesc, i, err)
			continue
		}
		sshHotkeys[i] = hk
		sshCount++

		// Start listener for this hotkey
		go func(num int, h *hotkey.Hotkey) {
			for range h.Keydown() {
				handleSSHDigit(num)
			}
		}(i, hk)
	}

	if snippetCount > 0 || urlCount > 0 || sshCount > 0 {
		mStatus.SetTitle(fmt.Sprintf("Hotkeys: %s (snippets), %s (URLs), %s (SSH)", snippetDesc, urlDesc, sshDesc))
	}

	LogDebug("Registered %d snippet hotkeys (%s+[0-9])", snippetCount, snippetDesc)
	LogDebug("Registered %d URL hotkeys (%s+[0-9])", urlCount, urlDesc)
	LogDebug("Registered %d SSH hotkeys (%s+[0-9])", sshCount, sshDesc)
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
			executeSnippetEntry(snippet)
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
			executeURLEntry(url)
			return
		}
	}

	mStatus.SetTitle(fmt.Sprintf("No URL with index %d", num))
}

// handleSSHDigit handles Ctrl+Option+N presses and accumulates digits
func handleSSHDigit(num int) {
	stateMutex.Lock()
	sshInput += fmt.Sprintf("%d", num)
	currentInput := sshInput
	stateMutex.Unlock()

	// Update status to show current input
	mStatus.SetTitle(fmt.Sprintf("SSH #%s...", currentInput))

	// Reset/start timeout - after 1 second of no more digits, open SSH
	if sshTimeout != nil {
		sshTimeout.Stop()
	}
	sshTimeout = time.AfterFunc(inputTimeout, finalizeSSHSelection)
}

func finalizeSSHSelection() {
	stateMutex.Lock()
	input := sshInput
	sshInput = "" // Reset for next time
	stateMutex.Unlock()

	if input != "" {
		selectSSHByInput(input)
	}
}

func selectSSHByInput(input string) {
	if input == "" {
		mStatus.SetTitle("No SSH number entered")
		return
	}

	// Parse the number
	var num int
	fmt.Sscanf(input, "%d", &num)

	openSSHByIndex(num)
}

func openSSHByIndex(num int) {
	// Find SSH with matching index
	if appConfig == nil || len(appConfig.SSH) == 0 {
		mStatus.SetTitle("No SSH connections configured")
		return
	}

	for _, ssh := range appConfig.SSH {
		if ssh.Index == num {
			executeSSHEntry(ssh)
			return
		}
	}

	mStatus.SetTitle(fmt.Sprintf("No SSH with index %d", num))
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
	for _, hk := range sshHotkeys {
		if hk != nil {
			hk.Unregister()
		}
	}
}
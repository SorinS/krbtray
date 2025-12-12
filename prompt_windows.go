//go:build windows

package main

// PromptForInput shows a dialog asking the user for text input
// Windows implementation - returns error for now (could use Windows API later)
func PromptForInput(title, message, defaultValue string, secure bool) (string, bool) {
	// TODO: Implement using Windows API (e.g., TaskDialog or custom dialog)
	// For now, return empty/cancelled
	LogWarn("Prompt dialog not yet implemented on Windows")
	return "", false
}

// ConfirmDialog shows a Yes/No confirmation dialog
func ConfirmDialog(title, message string) bool {
	// TODO: Implement using MessageBox
	LogWarn("Confirm dialog not yet implemented on Windows")
	return false
}
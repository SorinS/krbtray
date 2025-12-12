//go:build linux

package main

import (
	"os/exec"
	"strings"
)

// PromptForInput shows a dialog asking the user for text input
// Linux implementation using zenity or kdialog
func PromptForInput(title, message, defaultValue string, secure bool) (string, bool) {
	// Try zenity first (GTK)
	if path, err := exec.LookPath("zenity"); err == nil {
		args := []string{"--entry", "--title", title, "--text", message}
		if defaultValue != "" {
			args = append(args, "--entry-text", defaultValue)
		}
		if secure {
			args = append(args, "--hide-text")
		}

		cmd := exec.Command(path, args...)
		output, err := cmd.Output()
		if err != nil {
			// User cancelled or error
			return "", false
		}
		return strings.TrimSpace(string(output)), true
	}

	// Try kdialog (KDE)
	if path, err := exec.LookPath("kdialog"); err == nil {
		args := []string{"--title", title}
		if secure {
			args = append(args, "--password", message)
		} else {
			args = append(args, "--inputbox", message, defaultValue)
		}

		cmd := exec.Command(path, args...)
		output, err := cmd.Output()
		if err != nil {
			return "", false
		}
		return strings.TrimSpace(string(output)), true
	}

	LogWarn("No dialog tool found (install zenity or kdialog)")
	return "", false
}

// ConfirmDialog shows a Yes/No confirmation dialog
func ConfirmDialog(title, message string) bool {
	// Try zenity first
	if path, err := exec.LookPath("zenity"); err == nil {
		cmd := exec.Command(path, "--question", "--title", title, "--text", message)
		err := cmd.Run()
		return err == nil
	}

	// Try kdialog
	if path, err := exec.LookPath("kdialog"); err == nil {
		cmd := exec.Command(path, "--title", title, "--yesno", message)
		err := cmd.Run()
		return err == nil
	}

	LogWarn("No dialog tool found (install zenity or kdialog)")
	return false
}
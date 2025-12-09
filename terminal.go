package main

import (
	"fmt"
	"os/exec"
	"strings"
)

// openTerminal launches the SSH command in the configured terminal
// The terminal field should contain a command template with {cmd} placeholder
// Examples:
//   - macOS Terminal.app: "open -a /Applications/Utilities/Terminal.app {cmd}"
//   - macOS iTerm2: "/Applications/iTerm.app/Contents/MacOS/iTerm2 {cmd}"
//   - Linux gnome-terminal: "/usr/bin/gnome-terminal -- {cmd}"
//   - Linux alacritty: "/usr/bin/alacritty -e {cmd}"
//   - Windows cmd: "C:\\Windows\\System32\\cmd.exe /k {cmd}"
//   - Windows Terminal: "wt.exe {cmd}"
func openTerminal(entry SSHEntry) error {
	if entry.Terminal == "" {
		return fmt.Errorf("no terminal configured for SSH connection")
	}

	// Replace {cmd} placeholder with the actual SSH command
	cmdLine := strings.Replace(entry.Terminal, "{cmd}", entry.Command, -1)

	if debugMode {
		fmt.Printf("Opening terminal: %s\n", cmdLine)
	}

	// Parse the command line into executable and arguments
	args := parseCommandLine(cmdLine)
	if len(args) == 0 {
		return fmt.Errorf("invalid terminal command")
	}

	cmd := exec.Command(args[0], args[1:]...)
	return cmd.Start()
}

// parseCommandLine splits a command line into arguments, respecting quotes
func parseCommandLine(cmdLine string) []string {
	var args []string
	var current strings.Builder
	inQuote := false
	quoteChar := rune(0)

	for _, r := range cmdLine {
		switch {
		case r == '"' || r == '\'':
			if inQuote && r == quoteChar {
				// End of quoted section
				inQuote = false
				quoteChar = 0
			} else if !inQuote {
				// Start of quoted section
				inQuote = true
				quoteChar = r
			} else {
				// Different quote inside quotes, treat as literal
				current.WriteRune(r)
			}
		case r == ' ' && !inQuote:
			// Space outside quotes - end of argument
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}

	// Don't forget the last argument
	if current.Len() > 0 {
		args = append(args, current.String())
	}

	return args
}
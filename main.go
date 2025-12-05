package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/getlantern/systray"
)

var (
	commit    string
	buildDate string
)

var (
	// Global state
	currentSPN    string
	lastToken     string
	lastTokenTime time.Time
	stateMutex    sync.RWMutex

	// Menu items
	mStatus     *systray.MenuItem
	mCopyToken  *systray.MenuItem
	mCopyHeader *systray.MenuItem
	mRefresh    *systray.MenuItem
	mSPN        *systray.MenuItem
	mDebug      *systray.MenuItem
	mQuit       *systray.MenuItem
)

func main() {
	systray.Run(onReady, onExit)
}

func onReady() {
	// Set tray icon and title
	systray.SetIcon(getIcon())
	systray.SetTitle("Krb5")
	systray.SetTooltip("Kerberos Service Ticket Tool")

	// Status display (disabled, just for display)
	mStatus = systray.AddMenuItem("No ticket", "Current ticket status")
	mStatus.Disable()

	systray.AddSeparator()

	// SPN input/display
	mSPN = systray.AddMenuItem("SPN: (not set)", "Click to set SPN")

	systray.AddSeparator()

	// Actions
	mRefresh = systray.AddMenuItem("Get Ticket", "Request service ticket for SPN")
	mRefresh.Disable() // Disabled until SPN is set

	mCopyHeader = systray.AddMenuItem("Copy HTTP Header", "Copy 'Negotiate <token>' to clipboard")
	mCopyHeader.Disable()

	mCopyToken = systray.AddMenuItem("Copy Token", "Copy base64 token to clipboard")
	mCopyToken.Disable()

	systray.AddSeparator()

	// Settings
	mDebug = systray.AddMenuItemCheckbox("Debug Mode", "Enable debug output", false)

	systray.AddSeparator()

	// Quit
	mQuit = systray.AddMenuItem("Quit", "Quit the application")

	// Handle menu clicks
	go handleMenuClicks()

	// Check for initial SPN from environment
	if spn := os.Getenv("KRB5_SPN"); spn != "" {
		setSPN(spn)
	}

	// Show platform info in status
	updatePlatformStatus()
}

func onExit() {
	// Cleanup
}

func handleMenuClicks() {
	for {
		select {
		case <-mSPN.ClickedCh:
			// Prompt for SPN (platform-specific dialog or use environment)
			promptSPN()

		case <-mRefresh.ClickedCh:
			refreshToken()

		case <-mCopyHeader.ClickedCh:
			copyHTTPHeader()

		case <-mCopyToken.ClickedCh:
			copyToken()

		case <-mDebug.ClickedCh:
			toggleDebug()

		case <-mQuit.ClickedCh:
			systray.Quit()
			return
		}
	}
}

func updatePlatformStatus() {
	var platform string
	switch runtime.GOOS {
	case "darwin":
		if IsMacOS11OrLater() {
			platform = "macOS (GSS API)"
		} else {
			platform = "macOS (unsupported version)"
		}
	case "windows":
		platform = "Windows (SSPI)"
	case "linux":
		platform = "Linux (gokrb5)"
	default:
		platform = runtime.GOOS + " (unsupported)"
	}
	mStatus.SetTitle(fmt.Sprintf("Platform: %s", platform))
}

func setSPN(spn string) {
	stateMutex.Lock()
	currentSPN = spn
	stateMutex.Unlock()

	mSPN.SetTitle(fmt.Sprintf("SPN: %s", spn))
	mRefresh.Enable()

	// Auto-refresh token when SPN is set
	refreshToken()
}

func promptSPN() {
	// For now, check environment variable or use a default
	// In a full implementation, we'd show a native dialog
	spn := os.Getenv("KRB5_SPN")
	if spn == "" {
		// Try to read from stdin or show notification
		mStatus.SetTitle("Set KRB5_SPN environment variable")
		return
	}
	setSPN(spn)
}

func refreshToken() {
	stateMutex.RLock()
	spn := currentSPN
	stateMutex.RUnlock()

	if spn == "" {
		mStatus.SetTitle("Error: No SPN set")
		return
	}

	mStatus.SetTitle("Requesting ticket...")

	// Get the service ticket
	token, err := getServiceTicket(spn)
	if err != nil {
		mStatus.SetTitle(fmt.Sprintf("Error: %v", truncateError(err)))
		mCopyHeader.Disable()
		mCopyToken.Disable()
		return
	}

	// Store token
	stateMutex.Lock()
	lastToken = base64.StdEncoding.EncodeToString(token)
	lastTokenTime = time.Now()
	stateMutex.Unlock()

	// Update UI
	mStatus.SetTitle(fmt.Sprintf("Ticket OK (%d bytes) - %s", len(token), lastTokenTime.Format("15:04:05")))
	mCopyHeader.Enable()
	mCopyToken.Enable()
}

func getServiceTicket(spn string) ([]byte, error) {
	// Check platform support
	if !IsMacOS11OrLater() && !IsWindows() && !IsLinux() {
		return nil, fmt.Errorf("unsupported platform")
	}

	transport := NewGSSCredTransport()
	transport.SetDebug(debugMode)

	// On Linux, check for ccache
	if IsLinux() {
		ccache := os.Getenv("KRB5CCNAME")
		if ccache != "" {
			transport.SetCCachePath(ccache)
		}
	}

	if err := transport.Connect(); err != nil {
		return nil, err
	}
	defer transport.Close()

	return transport.GetServiceTicket(spn)
}

func copyHTTPHeader() {
	stateMutex.RLock()
	token := lastToken
	stateMutex.RUnlock()

	if token == "" {
		return
	}

	header := "Negotiate " + token
	if err := copyToClipboard(header); err != nil {
		mStatus.SetTitle(fmt.Sprintf("Copy failed: %v", err))
		return
	}
	mStatus.SetTitle("Copied HTTP header to clipboard")
}

func copyToken() {
	stateMutex.RLock()
	token := lastToken
	stateMutex.RUnlock()

	if token == "" {
		return
	}

	if err := copyToClipboard(token); err != nil {
		mStatus.SetTitle(fmt.Sprintf("Copy failed: %v", err))
		return
	}
	mStatus.SetTitle("Copied token to clipboard")
}

func toggleDebug() {
	debugMode = !debugMode
	if debugMode {
		mDebug.Check()
	} else {
		mDebug.Uncheck()
	}
	SetDebugMode(debugMode)
}

func truncateError(err error) string {
	s := err.Error()
	if len(s) > 40 {
		return s[:40] + "..."
	}
	return s
}

// getIcon returns the tray icon bytes
// Using a simple placeholder - in production, embed a proper icon
func getIcon() []byte {
	// This is a minimal 16x16 ICO/PNG placeholder
	// For production, use go:embed with actual icon files
	return defaultIcon
}

// Platform-specific clipboard implementation
func copyToClipboard(text string) error {
	return copyToClipboardPlatform(text)
}

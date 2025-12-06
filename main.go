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
	appConfig     *Config

	// Menu items
	mStatus       *systray.MenuItem
	mSPNMenu      *systray.MenuItem
	mSecretsMenu  *systray.MenuItem
	mURLsMenu     *systray.MenuItem
	mSnippetsMenu *systray.MenuItem
	mCopyToken    *systray.MenuItem
	mCopyHeader   *systray.MenuItem
	mRefresh      *systray.MenuItem
	mDebug        *systray.MenuItem
	mReloadCfg    *systray.MenuItem
	mAbout        *systray.MenuItem
	mQuit         *systray.MenuItem

	// SPN submenu items with their click handlers
	spnMenuItems     []*systray.MenuItem
	secretMenuItems  []*systray.MenuItem
	urlMenuItems     []*systray.MenuItem
	snippetMenuItems []*systray.MenuItem

	// Currently selected secret
	currentSecret *SecretEntry
)

func main() {
	// Initialize the cache
	InitCache()

	systray.Run(onReady, onExit)
}

func onReady() {
	// Set tray icon (no title text, just the icon)
	// Use SetIcon for colored icon (SetTemplateIcon would make it monochrome)
	systray.SetIcon(getIcon())
	systray.SetTitle("") // No text, just the icon
	systray.SetTooltip("Kerberos Service Ticket Tool")

	// Status display (disabled, just for display)
	mStatus = systray.AddMenuItem("No ticket", "Current ticket status")
	mStatus.Disable()

	systray.AddSeparator()

	// SPN submenu - will be populated from config
	mSPNMenu = systray.AddMenuItem("Select SPN", "Choose a service principal")
	loadAndBuildSPNMenu()

	// CSM Secrets submenu
	mSecretsMenu = systray.AddMenuItem("CSM Secrets", "Manage CSM secrets")
	loadAndBuildSecretsMenu()

	// URLs submenu
	mURLsMenu = systray.AddMenuItem("URLs", "Open URLs in browser")
	loadAndBuildURLsMenu()

	// Snippets submenu
	mSnippetsMenu = systray.AddMenuItem("Snippets", "Copy snippets to clipboard")
	loadAndBuildSnippetsMenu()

	systray.AddSeparator()

	// Actions
	mRefresh = systray.AddMenuItem("Refresh Ticket", "Re-request service ticket for current SPN")
	mRefresh.Disable() // Disabled until SPN is selected

	mCopyHeader = systray.AddMenuItem("Copy HTTP Header", "Copy 'Negotiate <token>' to clipboard")
	mCopyHeader.Disable()

	mCopyToken = systray.AddMenuItem("Copy Token", "Copy base64 token to clipboard")
	mCopyToken.Disable()

	systray.AddSeparator()

	// Settings
	mDebug = systray.AddMenuItemCheckbox("Debug Mode", "Enable debug output", false)
	mReloadCfg = systray.AddMenuItem("Reload Config", "Reload configuration from file")

	systray.AddSeparator()

	// About submenu with version info
	mAbout = systray.AddMenuItem("About", "About krb5tray")
	mAboutVersion := mAbout.AddSubMenuItem(fmt.Sprintf("Version: %s", Version), "")
	mAboutVersion.Disable()
	mAboutCommit := mAbout.AddSubMenuItem(fmt.Sprintf("Commit: %s", getShortCommit()), "")
	mAboutCommit.Disable()
	mAboutBuild := mAbout.AddSubMenuItem(fmt.Sprintf("Build: %s", buildDate), "")
	mAboutBuild.Disable()

	systray.AddSeparator()

	// Quit
	mQuit = systray.AddMenuItem("Quit", "Quit the application")

	// Handle menu clicks
	go handleMenuClicks()

	// Check for initial SPN from environment (fallback)
	if spn := os.Getenv("KRB5_SPN"); spn != "" && currentSPN == "" {
		setSPN(spn, "Environment")
	}

	// Show platform info in status
	updatePlatformStatus()

	// Initialize global hotkeys for snippet selection
	InitHotkeys()
}

func loadAndBuildSPNMenu() {
	// Try to load config
	cfg, err := LoadConfig("")
	if err != nil {
		// Config doesn't exist, create default
		if os.IsNotExist(err) {
			if createErr := CreateDefaultConfig(); createErr == nil {
				cfg, _ = LoadConfig("")
			}
		}
	}

	appConfig = cfg

	if cfg == nil || len(cfg.SPNs) == 0 {
		// No SPNs configured
		noSPN := mSPNMenu.AddSubMenuItem("No SPNs configured", "Edit config file to add SPNs")
		noSPN.Disable()

		configPath := mSPNMenu.AddSubMenuItem(fmt.Sprintf("Config: %s", DefaultConfigPath()), "Configuration file location")
		configPath.Disable()
		return
	}

	// Clear existing submenu items tracking
	spnMenuItems = make([]*systray.MenuItem, 0, len(cfg.SPNs))

	// Add each SPN as a submenu item
	for _, entry := range cfg.SPNs {
		item := mSPNMenu.AddSubMenuItem(entry.Name, entry.SPN)
		spnMenuItems = append(spnMenuItems, item)

		// Start a goroutine to handle clicks for this SPN
		go handleSPNClick(item, entry)
	}

	// Add separator and config path info
	mSPNMenu.AddSubMenuItem("", "")
	configInfo := mSPNMenu.AddSubMenuItem(fmt.Sprintf("Config: %s", DefaultConfigPath()), "Configuration file location")
	configInfo.Disable()
}

func handleSPNClick(item *systray.MenuItem, entry SPNEntry) {
	for range item.ClickedCh {
		setSPN(entry.SPN, entry.Name)
	}
}

func loadAndBuildSecretsMenu() {
	if appConfig == nil || len(appConfig.Secrets) == 0 {
		noSecrets := mSecretsMenu.AddSubMenuItem("No secrets configured", "Edit config file to add secrets")
		noSecrets.Disable()
		return
	}

	// Clear existing submenu items tracking
	secretMenuItems = make([]*systray.MenuItem, 0, len(appConfig.Secrets))

	// Add each secret as a submenu item
	for i := range appConfig.Secrets {
		entry := &appConfig.Secrets[i]
		item := mSecretsMenu.AddSubMenuItem(entry.Name, fmt.Sprintf("Role: %s (%s)", entry.RoleName, entry.RoleType))
		secretMenuItems = append(secretMenuItems, item)

		// Start a goroutine to handle clicks for this secret
		go handleSecretClick(item, entry)
	}
}

func handleSecretClick(item *systray.MenuItem, entry *SecretEntry) {
	for range item.ClickedCh {
		setSecret(entry)
	}
}

func setSecret(entry *SecretEntry) {
	stateMutex.Lock()
	currentSecret = entry
	stateMutex.Unlock()

	mSecretsMenu.SetTitle(fmt.Sprintf("Secret: %s", entry.Name))
	mStatus.SetTitle(fmt.Sprintf("Selected: %s", entry.Name))
}

func loadAndBuildURLsMenu() {
	if appConfig == nil || len(appConfig.URLs) == 0 {
		noURLs := mURLsMenu.AddSubMenuItem("No URLs configured", "Edit config file to add URLs")
		noURLs.Disable()
		return
	}

	// Clear existing submenu items tracking
	urlMenuItems = make([]*systray.MenuItem, 0, len(appConfig.URLs))

	// Add each URL as a submenu item
	for _, entry := range appConfig.URLs {
		item := mURLsMenu.AddSubMenuItem(entry.Name, entry.URL)
		urlMenuItems = append(urlMenuItems, item)

		// Start a goroutine to handle clicks for this URL
		go handleURLClick(item, entry)
	}
}

func handleURLClick(item *systray.MenuItem, entry URLEntry) {
	for range item.ClickedCh {
		if err := openBrowser(entry.URL); err != nil {
			mStatus.SetTitle(fmt.Sprintf("Failed to open: %s", entry.Name))
		} else {
			mStatus.SetTitle(fmt.Sprintf("Opened: %s", entry.Name))
		}
	}
}

func loadAndBuildSnippetsMenu() {
	if appConfig == nil || len(appConfig.Snippets) == 0 {
		noSnippets := mSnippetsMenu.AddSubMenuItem("No snippets configured", "Edit config file to add snippets")
		noSnippets.Disable()
		return
	}

	// Clear existing submenu items tracking
	snippetMenuItems = make([]*systray.MenuItem, 0, len(appConfig.Snippets))

	// Group snippets by index ranges (1-10, 11-20, etc.)
	const groupSize = 10

	// Find max index to determine number of groups needed
	maxIndex := 0
	for _, entry := range appConfig.Snippets {
		if entry.Index > maxIndex {
			maxIndex = entry.Index
		}
	}

	// If 10 or fewer snippets, don't group - just list them directly
	if len(appConfig.Snippets) <= groupSize {
		for _, entry := range appConfig.Snippets {
			addSnippetMenuItem(mSnippetsMenu, entry)
		}
		return
	}

	// Create groups based on index ranges
	numGroups := (maxIndex / groupSize) + 1
	groups := make(map[int][]SnippetEntry)

	for _, entry := range appConfig.Snippets {
		groupNum := entry.Index / groupSize
		groups[groupNum] = append(groups[groupNum], entry)
	}

	// Create submenus for each group that has snippets
	for g := 0; g < numGroups; g++ {
		snippets, exists := groups[g]
		if !exists || len(snippets) == 0 {
			continue
		}

		// Calculate range for this group
		startIdx := g * groupSize
		endIdx := startIdx + groupSize - 1
		if g == 0 {
			startIdx = 0 // First group is 0-9
		}

		groupLabel := fmt.Sprintf("Snippets %d-%d", startIdx, endIdx)
		groupMenu := mSnippetsMenu.AddSubMenuItem(groupLabel, "")

		// Add snippets in this group
		for _, entry := range snippets {
			addSnippetSubMenuItem(groupMenu, entry)
		}
	}
}

func addSnippetMenuItem(parent *systray.MenuItem, entry SnippetEntry) {
	displayName := fmt.Sprintf("[%d] %s", entry.Index, entry.Name)
	tooltip := entry.Value
	if len(tooltip) > 50 {
		tooltip = tooltip[:50] + "..."
	}
	item := parent.AddSubMenuItem(displayName, tooltip)
	snippetMenuItems = append(snippetMenuItems, item)
	go handleSnippetClick(item, entry)
}

func addSnippetSubMenuItem(parent *systray.MenuItem, entry SnippetEntry) {
	displayName := fmt.Sprintf("[%d] %s", entry.Index, entry.Name)
	tooltip := entry.Value
	if len(tooltip) > 50 {
		tooltip = tooltip[:50] + "..."
	}
	item := parent.AddSubMenuItem(displayName, tooltip)
	snippetMenuItems = append(snippetMenuItems, item)
	go handleSnippetClick(item, entry)
}

func handleSnippetClick(item *systray.MenuItem, entry SnippetEntry) {
	for range item.ClickedCh {
		if err := copyToClipboard(entry.Value); err != nil {
			mStatus.SetTitle(fmt.Sprintf("Copy failed: %s", entry.Name))
		} else {
			mStatus.SetTitle(fmt.Sprintf("Copied: %s", entry.Name))
		}
	}
}

func onExit() {
	// Cleanup hotkeys
	CleanupHotkeys()
}

func handleMenuClicks() {
	for {
		select {
		case <-mRefresh.ClickedCh:
			refreshToken()

		case <-mCopyHeader.ClickedCh:
			copyHTTPHeader()

		case <-mCopyToken.ClickedCh:
			copyToken()

		case <-mDebug.ClickedCh:
			toggleDebug()

		case <-mReloadCfg.ClickedCh:
			reloadConfig()

		case <-mQuit.ClickedCh:
			systray.Quit()
			return
		}
	}
}

func reloadConfig() {
	cfg, err := LoadConfig("")
	if err != nil {
		mStatus.SetTitle(fmt.Sprintf("Config error: %v", truncateError(err)))
		return
	}
	appConfig = cfg
	mStatus.SetTitle(fmt.Sprintf("Config reloaded (%d SPNs)", len(cfg.SPNs)))

	// Note: We can't dynamically rebuild the menu in systray
	// User needs to restart the app to see new SPNs
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

func setSPN(spn string, displayName string) {
	stateMutex.Lock()
	currentSPN = spn
	stateMutex.Unlock()

	if displayName != "" {
		mSPNMenu.SetTitle(fmt.Sprintf("SPN: %s", displayName))
	} else {
		mSPNMenu.SetTitle(fmt.Sprintf("SPN: %s", spn))
	}
	mRefresh.Enable()

	// Auto-refresh token when SPN is selected
	refreshToken()
}

func refreshToken() {
	stateMutex.RLock()
	spn := currentSPN
	stateMutex.RUnlock()

	if spn == "" {
		mStatus.SetTitle("Error: No SPN selected")
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

	// Store token in memory
	stateMutex.Lock()
	lastToken = base64.StdEncoding.EncodeToString(token)
	lastTokenTime = time.Now()
	stateMutex.Unlock()

	// Cache the token for this SPN
	GetCache().SetToken(spn, lastToken, DefaultTokenExpiration)

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
func getIcon() []byte {
	return defaultIcon
}

// getShortCommit returns the first 8 characters of the commit hash
func getShortCommit() string {
	if len(commit) >= 8 {
		return commit[:8]
	}
	if commit == "" {
		return "dev"
	}
	return commit
}

// Platform-specific clipboard implementation
func copyToClipboard(text string) error {
	return copyToClipboardPlatform(text)
}
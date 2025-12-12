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
	mSSHMenu      *systray.MenuItem
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
	sshMenuItems     []*systray.MenuItem

	// Data bound to menu items (used for click handling after reload)
	spnEntries     []SPNEntry
	secretEntries  []*SecretEntry
	urlEntries     []URLEntry
	snippetEntries []SnippetEntry
	sshEntries     []SSHEntry

	// Currently selected secret
	currentSecret *SecretEntry
)

func main() {
	// Ensure only one instance is running
	if err := EnsureSingleInstance(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Try to load config early for logging settings
	// If config doesn't exist, use defaults
	var logCfg LogConfig
	if cfg, err := LoadConfig(""); err == nil {
		logCfg = cfg.GetLogConfigWithDefaults()
	} else {
		logCfg = DefaultLogConfig()
	}

	// Initialize logger with config
	if err := InitLoggerWithConfig(logCfg); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Warning: Failed to initialize logger: %v\n", err)
	}

	LogStartup()

	// Initialize the cache
	InitCache()

	// Initialize Lua scripting engine
	if err := InitLuaEngine(); err != nil {
		LogWarn("Failed to initialize Lua engine: %v", err)
	}

	systray.Run(onReady, onExit)
}

func onReady() {
	// Set tray icon (no title text, just the icon)
	// Use SetIcon for colored icon (SetTemplateIcon would make it monochrome)
	systray.SetIcon(getIcon())
	systray.SetTitle("") // No text, just the icon
	systray.SetTooltip("Kerberos Service Ticket Tool")

	// Status display as submenu (kept enabled for better contrast)
	mStatusMenu := systray.AddMenuItem("Status", "Current status")
	mStatus = mStatusMenu.AddSubMenuItem("Ready", "")

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

	// SSH submenu
	mSSHMenu = systray.AddMenuItem("SSH", "Open SSH connections in terminal")
	loadAndBuildSSHMenu()

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

	// Hotkeys submenu showing keyboard shortcuts (kept enabled for better contrast)
	mHotkeys := systray.AddMenuItem("Hotkeys", "Keyboard shortcuts")
	_, snippetDesc := getSnippetHotkeyModifiers()
	_, urlDesc := getURLHotkeyModifiers()
	_, sshDesc := getSSHHotkeyModifiers()

	_ = mHotkeys.AddSubMenuItem(fmt.Sprintf("Snippets: %s+[0-9]", snippetDesc), "Copy snippet to clipboard")
	_ = mHotkeys.AddSubMenuItem(fmt.Sprintf("URLs: %s+[0-9]", urlDesc), "Open URL in browser")
	_ = mHotkeys.AddSubMenuItem(fmt.Sprintf("SSH: %s+[0-9]", sshDesc), "Open SSH connection in terminal")
	mHotkeys.AddSubMenuItem("", "")
	_ = mHotkeys.AddSubMenuItem("Hold modifiers, press digits", "Multi-digit: 1 sec timeout")

	systray.AddSeparator()

	// About submenu with version info (kept enabled for better contrast)
	mAbout = systray.AddMenuItem("About", "About krb5tray")
	_ = mAbout.AddSubMenuItem(fmt.Sprintf("Version: %s", Version), "")
	_ = mAbout.AddSubMenuItem(fmt.Sprintf("Commit: %s", getShortCommit()), "")
	_ = mAbout.AddSubMenuItem(fmt.Sprintf("Build: %s", buildDate), "")

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

const maxMenuItems = 50 // Maximum items per menu type

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

	// Pre-allocate menu items pool
	spnMenuItems = make([]*systray.MenuItem, maxMenuItems)
	spnEntries = make([]SPNEntry, maxMenuItems)

	for i := 0; i < maxMenuItems; i++ {
		item := mSPNMenu.AddSubMenuItem("", "")
		item.Hide()
		spnMenuItems[i] = item
		go handleSPNClickByIndex(item, i)
	}

	// Add config path info at the end (always visible)
	mSPNMenu.AddSubMenuItem("", "")
	configInfo := mSPNMenu.AddSubMenuItem(fmt.Sprintf("Config: %s", DefaultConfigPath()), "Configuration file location")
	configInfo.Disable()

	// Now populate with actual data
	updateSPNMenu()
}

func updateSPNMenu() {
	// Hide all items first
	for i := 0; i < maxMenuItems; i++ {
		spnMenuItems[i].Hide()
	}

	if appConfig == nil || len(appConfig.SPNs) == 0 {
		// Show "No SPNs configured" in first slot
		spnMenuItems[0].SetTitle("No SPNs configured")
		spnMenuItems[0].SetTooltip("Edit config file to add SPNs")
		spnMenuItems[0].Disable()
		spnMenuItems[0].Show()
		return
	}

	// Update entries and show items
	for i, entry := range appConfig.SPNs {
		if i >= maxMenuItems {
			break
		}
		spnEntries[i] = entry
		spnMenuItems[i].SetTitle(entry.Name)
		spnMenuItems[i].SetTooltip(entry.SPN)
		spnMenuItems[i].Enable()
		spnMenuItems[i].Show()
	}
}

func handleSPNClickByIndex(item *systray.MenuItem, index int) {
	for range item.ClickedCh {
		stateMutex.RLock()
		entry := spnEntries[index]
		stateMutex.RUnlock()
		if entry.SPN != "" {
			setSPN(entry.SPN, entry.Name)
		}
	}
}

func loadAndBuildSecretsMenu() {
	// Pre-allocate menu items pool
	secretMenuItems = make([]*systray.MenuItem, maxMenuItems)
	secretEntries = make([]*SecretEntry, maxMenuItems)

	for i := 0; i < maxMenuItems; i++ {
		item := mSecretsMenu.AddSubMenuItem("", "")
		item.Hide()
		secretMenuItems[i] = item
		go handleSecretClickByIndex(item, i)
	}

	// Now populate with actual data
	updateSecretsMenu()
}

func updateSecretsMenu() {
	// Hide all items first
	for i := 0; i < maxMenuItems; i++ {
		secretMenuItems[i].Hide()
	}

	if appConfig == nil || len(appConfig.Secrets) == 0 {
		// Show "No secrets configured" in first slot
		secretMenuItems[0].SetTitle("No secrets configured")
		secretMenuItems[0].SetTooltip("Edit config file to add secrets")
		secretMenuItems[0].Disable()
		secretMenuItems[0].Show()
		return
	}

	// Update entries and show items
	for i := range appConfig.Secrets {
		if i >= maxMenuItems {
			break
		}
		entry := &appConfig.Secrets[i]
		secretEntries[i] = entry
		secretMenuItems[i].SetTitle(entry.Name)
		secretMenuItems[i].SetTooltip(fmt.Sprintf("Role: %s (%s)", entry.RoleName, entry.RoleType))
		secretMenuItems[i].Enable()
		secretMenuItems[i].Show()
	}
}

func handleSecretClickByIndex(item *systray.MenuItem, index int) {
	for range item.ClickedCh {
		stateMutex.RLock()
		entry := secretEntries[index]
		stateMutex.RUnlock()
		if entry != nil {
			setSecret(entry)
		}
	}
}

func setSecret(entry *SecretEntry) {
	stateMutex.Lock()
	currentSecret = entry
	stateMutex.Unlock()

	LogSecretSelected(entry.Name)
	mSecretsMenu.SetTitle(fmt.Sprintf("Secret: %s", entry.Name))
	mStatus.SetTitle(fmt.Sprintf("Selected: %s", entry.Name))
}

func loadAndBuildURLsMenu() {
	// Pre-allocate menu items pool
	urlMenuItems = make([]*systray.MenuItem, maxMenuItems)
	urlEntries = make([]URLEntry, maxMenuItems)

	for i := 0; i < maxMenuItems; i++ {
		item := mURLsMenu.AddSubMenuItem("", "")
		item.Hide()
		urlMenuItems[i] = item
		go handleURLClickByIndex(item, i)
	}

	// Now populate with actual data
	updateURLsMenu()
}

func updateURLsMenu() {
	// Hide all items first
	for i := 0; i < maxMenuItems; i++ {
		urlMenuItems[i].Hide()
	}

	if appConfig == nil || len(appConfig.URLs) == 0 {
		// Show "No URLs configured" in first slot
		urlMenuItems[0].SetTitle("No URLs configured")
		urlMenuItems[0].SetTooltip("Edit config file to add URLs")
		urlMenuItems[0].Disable()
		urlMenuItems[0].Show()
		return
	}

	// Update entries and show items
	for i, entry := range appConfig.URLs {
		if i >= maxMenuItems {
			break
		}
		urlEntries[i] = entry
		displayName := fmt.Sprintf("[%d] %s", entry.Index, entry.Name)
		urlMenuItems[i].SetTitle(displayName)
		urlMenuItems[i].SetTooltip(entry.URL)
		urlMenuItems[i].Enable()
		urlMenuItems[i].Show()
	}
}

func handleURLClickByIndex(item *systray.MenuItem, index int) {
	for range item.ClickedCh {
		stateMutex.RLock()
		entry := urlEntries[index]
		stateMutex.RUnlock()
		if entry.URL != "" || entry.Script != "" {
			executeURLEntry(entry)
		}
	}
}

func executeURLEntry(entry URLEntry) {
	// If script is defined, run it instead of opening URL directly
	if entry.Script != "" {
		engine := GetLuaEngine()
		if engine != nil {
			ctx := map[string]string{
				"url":   entry.URL,
				"name":  entry.Name,
				"index": fmt.Sprintf("%d", entry.Index),
			}
			_, err := engine.RunScript(entry.Script, ctx)
			if err != nil {
				LogScriptExecuted(entry.Script, "url", false)
				mStatus.SetTitle(fmt.Sprintf("Script error: %s", truncateError(err)))
			} else {
				LogScriptExecuted(entry.Script, "url", true)
				mStatus.SetTitle(fmt.Sprintf("Script: %s", entry.Name))
			}
			return
		}
	}

	// Default behavior: open URL in browser
	if err := openBrowser(entry.URL); err != nil {
		LogError("Failed to open URL %s: %v", entry.Name, err)
		mStatus.SetTitle(fmt.Sprintf("Failed to open: %s", entry.Name))
	} else {
		LogURLOpened(entry.Name)
		mStatus.SetTitle(fmt.Sprintf("Opened: %s", entry.Name))
	}
}

func loadAndBuildSnippetsMenu() {
	// Pre-allocate menu items pool (flat list, no grouping for simplicity in reload)
	snippetMenuItems = make([]*systray.MenuItem, maxMenuItems)
	snippetEntries = make([]SnippetEntry, maxMenuItems)

	for i := 0; i < maxMenuItems; i++ {
		item := mSnippetsMenu.AddSubMenuItem("", "")
		item.Hide()
		snippetMenuItems[i] = item
		go handleSnippetClickByIndex(item, i)
	}

	// Now populate with actual data
	updateSnippetsMenu()
}

func updateSnippetsMenu() {
	// Hide all items first
	for i := 0; i < maxMenuItems; i++ {
		snippetMenuItems[i].Hide()
	}

	if appConfig == nil || len(appConfig.Snippets) == 0 {
		// Show "No snippets configured" in first slot
		snippetMenuItems[0].SetTitle("No snippets configured")
		snippetMenuItems[0].SetTooltip("Edit config file to add snippets")
		snippetMenuItems[0].Disable()
		snippetMenuItems[0].Show()
		return
	}

	// Update entries and show items
	for i, entry := range appConfig.Snippets {
		if i >= maxMenuItems {
			break
		}
		snippetEntries[i] = entry
		displayName := fmt.Sprintf("[%d] %s", entry.Index, entry.Name)
		tooltip := entry.Value
		if len(tooltip) > 50 {
			tooltip = tooltip[:50] + "..."
		}
		snippetMenuItems[i].SetTitle(displayName)
		snippetMenuItems[i].SetTooltip(tooltip)
		snippetMenuItems[i].Enable()
		snippetMenuItems[i].Show()
	}
}

func handleSnippetClickByIndex(item *systray.MenuItem, index int) {
	for range item.ClickedCh {
		stateMutex.RLock()
		entry := snippetEntries[index]
		stateMutex.RUnlock()
		if entry.Name != "" || entry.Script != "" {
			executeSnippetEntry(entry)
		}
	}
}

func executeSnippetEntry(entry SnippetEntry) {
	// If script is defined, run it instead of copying value directly
	if entry.Script != "" {
		engine := GetLuaEngine()
		if engine != nil {
			ctx := map[string]string{
				"value": entry.Value,
				"name":  entry.Name,
				"index": fmt.Sprintf("%d", entry.Index),
			}
			result, err := engine.RunScript(entry.Script, ctx)
			if err != nil {
				LogScriptExecuted(entry.Script, "snippet", false)
				mStatus.SetTitle(fmt.Sprintf("Script error: %s", truncateError(err)))
			} else if result != "" {
				LogScriptExecuted(entry.Script, "snippet", true)
				// If script returns a result, copy that to clipboard
				if err := copyToClipboard(result); err != nil {
					mStatus.SetTitle(fmt.Sprintf("Copy failed: %s", entry.Name))
				} else {
					LogClipboardCopy("snippet", entry.Name)
					mStatus.SetTitle(fmt.Sprintf("Copied: %s", entry.Name))
				}
			} else {
				LogScriptExecuted(entry.Script, "snippet", true)
				mStatus.SetTitle(fmt.Sprintf("Script: %s", entry.Name))
			}
			return
		}
	}

	// Default behavior: copy value to clipboard
	if err := copyToClipboard(entry.Value); err != nil {
		LogError("Failed to copy snippet %s: %v", entry.Name, err)
		mStatus.SetTitle(fmt.Sprintf("Copy failed: %s", entry.Name))
	} else {
		LogClipboardCopy("snippet", entry.Name)
		mStatus.SetTitle(fmt.Sprintf("Copied: %s", entry.Name))
	}
}

func loadAndBuildSSHMenu() {
	// Pre-allocate menu items pool
	sshMenuItems = make([]*systray.MenuItem, maxMenuItems)
	sshEntries = make([]SSHEntry, maxMenuItems)

	for i := 0; i < maxMenuItems; i++ {
		item := mSSHMenu.AddSubMenuItem("", "")
		item.Hide()
		sshMenuItems[i] = item
		go handleSSHClickByIndex(item, i)
	}

	// Now populate with actual data
	updateSSHMenu()
}

func updateSSHMenu() {
	// Hide all items first
	for i := 0; i < maxMenuItems; i++ {
		sshMenuItems[i].Hide()
	}

	if appConfig == nil || len(appConfig.SSH) == 0 {
		// Show "No SSH configured" in first slot
		sshMenuItems[0].SetTitle("No SSH connections configured")
		sshMenuItems[0].SetTooltip("Edit config file to add SSH connections")
		sshMenuItems[0].Disable()
		sshMenuItems[0].Show()
		return
	}

	// Update entries and show items
	for i, entry := range appConfig.SSH {
		if i >= maxMenuItems {
			break
		}
		sshEntries[i] = entry
		displayName := fmt.Sprintf("[%d] %s", entry.Index, entry.Name)
		sshMenuItems[i].SetTitle(displayName)
		sshMenuItems[i].SetTooltip(entry.Command)
		sshMenuItems[i].Enable()
		sshMenuItems[i].Show()
	}
}

func handleSSHClickByIndex(item *systray.MenuItem, index int) {
	for range item.ClickedCh {
		stateMutex.RLock()
		entry := sshEntries[index]
		stateMutex.RUnlock()
		if entry.Command != "" || entry.Script != "" {
			executeSSHEntry(entry)
		}
	}
}

func executeSSHEntry(entry SSHEntry) {
	// If script is defined, run it instead of/before opening terminal
	if entry.Script != "" {
		engine := GetLuaEngine()
		if engine != nil {
			ctx := map[string]string{
				"command":  entry.Command,
				"terminal": entry.Terminal,
				"name":     entry.Name,
				"index":    fmt.Sprintf("%d", entry.Index),
			}
			_, err := engine.RunScript(entry.Script, ctx)
			if err != nil {
				LogScriptExecuted(entry.Script, "ssh", false)
				mStatus.SetTitle(fmt.Sprintf("Script error: %s", truncateError(err)))
			} else {
				LogScriptExecuted(entry.Script, "ssh", true)
				mStatus.SetTitle(fmt.Sprintf("Script: %s", entry.Name))
			}
			return
		}
	}

	// Default behavior: open terminal with SSH command
	if err := openTerminal(entry); err != nil {
		LogError("Failed to open SSH %s: %v", entry.Name, err)
		mStatus.SetTitle(fmt.Sprintf("SSH failed: %s", entry.Name))
	} else {
		LogSSHOpened(entry.Name)
		mStatus.SetTitle(fmt.Sprintf("SSH: %s", entry.Name))
	}
}

func onExit() {
	LogShutdown()

	// Cleanup hotkeys
	CleanupHotkeys()

	// Release single instance lock
	ReleaseSingleInstance()
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
		LogError("Config reload failed: %v", err)
		mStatus.SetTitle(fmt.Sprintf("Config error: %v", truncateError(err)))
		return
	}
	appConfig = cfg

	// Update all menus with new config data
	updateSPNMenu()
	updateSecretsMenu()
	updateURLsMenu()
	updateSnippetsMenu()
	updateSSHMenu()

	LogConfigLoaded(len(cfg.SPNs), len(cfg.Secrets), len(cfg.URLs), len(cfg.Snippets), len(cfg.SSH))
	mStatus.SetTitle(fmt.Sprintf("Config reloaded (%d SPNs, %d snippets, %d SSH)", len(cfg.SPNs), len(cfg.Snippets), len(cfg.SSH)))
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

	LogSPNSelected(displayName)

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

	LogDebug("Requesting ticket for SPN")
	mStatus.SetTitle("Requesting ticket...")

	// Get the service ticket
	token, err := getServiceTicket(spn)
	if err != nil {
		LogTicketRequested("(current)", false, 0)
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

	LogTicketRequested("(current)", true, len(token))

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
		LogError("Failed to copy HTTP header: %v", err)
		mStatus.SetTitle(fmt.Sprintf("Copy failed: %v", err))
		return
	}
	LogClipboardCopy("http_header", "Negotiate token")
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
		LogError("Failed to copy token: %v", err)
		mStatus.SetTitle(fmt.Sprintf("Copy failed: %v", err))
		return
	}
	LogClipboardCopy("token", "Base64 token")
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
	SetLogLevel(debugMode)
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

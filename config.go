package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// LogConfig represents logging configuration
type LogConfig struct {
	MaxSizeMB  int  `json:"max_size_mb,omitempty"`  // Max log file size in MB before rotation (default: 10)
	MaxBackups int  `json:"max_backups,omitempty"`  // Max number of old log files to keep (default: 7)
	MaxAgeDays int  `json:"max_age_days,omitempty"` // Max days to retain old log files (default: 7)
	Compress   bool `json:"compress,omitempty"`     // Compress rotated log files (default: true)
	ToStdout   bool `json:"to_stdout,omitempty"`    // Also write logs to stdout (default: true)
}

// DefaultLogConfig returns the default logging configuration
func DefaultLogConfig() LogConfig {
	return LogConfig{
		MaxSizeMB:  10,
		MaxBackups: 7,
		MaxAgeDays: 7,
		Compress:   true,
		ToStdout:   true,
	}
}

// Config represents the application configuration
type Config struct {
	SPNs     []SPNEntry     `json:"spns"`
	Secrets  []SecretEntry  `json:"secrets,omitempty"`
	URLs     []URLEntry     `json:"urls,omitempty"`
	Snippets []SnippetEntry `json:"snippets,omitempty"`
	SSH      []SSHEntry     `json:"ssh,omitempty"`
	Logging  *LogConfig     `json:"logging,omitempty"`
}

// GetLogConfig returns the logging config with defaults applied
func (c *Config) GetLogConfig() LogConfig {
	if c == nil || c.Logging == nil {
		return DefaultLogConfig()
	}

	cfg := *c.Logging
	defaults := DefaultLogConfig()

	// Apply defaults for zero values
	if cfg.MaxSizeMB <= 0 {
		cfg.MaxSizeMB = defaults.MaxSizeMB
	}
	if cfg.MaxBackups <= 0 {
		cfg.MaxBackups = defaults.MaxBackups
	}
	if cfg.MaxAgeDays <= 0 {
		cfg.MaxAgeDays = defaults.MaxAgeDays
	}
	// Note: Compress and ToStdout use their zero value (false) if not set,
	// but we want true as default. Handle this with a pointer or explicit check.
	// For simplicity, if Logging section exists but these are false, we respect that.
	// If Logging section doesn't exist at all, we use defaults (true).

	return cfg
}

// GetLogConfigWithDefaults returns log config, using defaults if logging section is absent
func (c *Config) GetLogConfigWithDefaults() LogConfig {
	if c == nil || c.Logging == nil {
		return DefaultLogConfig()
	}

	cfg := DefaultLogConfig()

	// Override with user values if set
	if c.Logging.MaxSizeMB > 0 {
		cfg.MaxSizeMB = c.Logging.MaxSizeMB
	}
	if c.Logging.MaxBackups > 0 {
		cfg.MaxBackups = c.Logging.MaxBackups
	}
	if c.Logging.MaxAgeDays > 0 {
		cfg.MaxAgeDays = c.Logging.MaxAgeDays
	}
	// For booleans, only override if the logging section exists
	// This allows users to explicitly set false
	cfg.Compress = c.Logging.Compress
	cfg.ToStdout = c.Logging.ToStdout

	return cfg
}

// SnippetEntry represents a text snippet that can be copied to clipboard
type SnippetEntry struct {
	Index  int    `json:"index"`            // Numeric index for ordering/reference
	Name   string `json:"name"`             // Display name in menu
	Value  string `json:"value"`            // The value to copy to clipboard
	Script string `json:"script,omitempty"` // Optional Lua script to run (filename in scripts folder)
}

// URLEntry represents a URL bookmark
type URLEntry struct {
	Index  int    `json:"index"`            // Numeric index for hotkey access
	Name   string `json:"name"`             // Display name in menu
	URL    string `json:"url"`              // The URL to open
	Script string `json:"script,omitempty"` // Optional Lua script to run instead of opening URL
}

// SSHEntry represents an SSH connection configuration
type SSHEntry struct {
	Index    int    `json:"index"`            // Numeric index for hotkey access
	Name     string `json:"name"`             // Display name in menu
	Command  string `json:"command"`          // SSH command to execute (e.g., "ssh user@host")
	Terminal string `json:"terminal"`         // Terminal command template with {cmd} placeholder
	Script   string `json:"script,omitempty"` // Optional Lua script to run before/instead of SSH
}

// SecretEntry represents a CSM secret configuration
type SecretEntry struct {
	Name      string `json:"name"`       // Display name in menu
	AuthURL   string `json:"auth_url"`   // Authentication URL
	RoleName  string `json:"role_name"`  // Role name
	RoleType  string `json:"role_type"`  // Role type
	RotateURL string `json:"rotate_url"` // Rotate URL
	SecretURL string `json:"secret_url"` // Secret URL
}

// SPNEntry represents a single SPN configuration
// Supports both simple string format and object format
type SPNEntry struct {
	Name string `json:"name"` // Display name in menu
	SPN  string `json:"spn"`  // The actual SPN value
}

// UnmarshalJSON implements custom unmarshaling to support both string and object formats
func (e *SPNEntry) UnmarshalJSON(data []byte) error {
	// Try as simple string first
	var simpleString string
	if err := json.Unmarshal(data, &simpleString); err == nil {
		e.Name = simpleString
		e.SPN = simpleString
		return nil
	}

	// Try as object
	type spnEntryAlias SPNEntry
	var obj spnEntryAlias
	if err := json.Unmarshal(data, &obj); err != nil {
		return err
	}

	e.Name = obj.Name
	e.SPN = obj.SPN

	// If name is empty, use SPN as name
	if e.Name == "" {
		e.Name = e.SPN
	}

	return nil
}

// ConfigDir returns the configuration directory path
func ConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "ktray")
}

// ScriptsDir returns the Lua scripts directory path
func ScriptsDir() string {
	return filepath.Join(ConfigDir(), "scripts")
}

// DefaultConfigPath returns the default configuration file path
func DefaultConfigPath() string {
	return filepath.Join(ConfigDir(), "ktray.json")
}

// ScriptPath returns the full path for a script filename
func ScriptPath(scriptName string) string {
	return filepath.Join(ScriptsDir(), scriptName)
}

// LoadConfig loads configuration from the specified path
// If path is empty, uses the default path
func LoadConfig(path string) (*Config, error) {
	if path == "" {
		path = DefaultConfigPath()
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// SaveConfig saves configuration to the specified path
// If path is empty, uses the default path
func SaveConfig(cfg *Config, path string) error {
	if path == "" {
		path = DefaultConfigPath()
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

// CreateDefaultConfig creates a default configuration file if it doesn't exist
func CreateDefaultConfig() error {
	path := DefaultConfigPath()
	if _, err := os.Stat(path); err == nil {
		// Config already exists
		return nil
	}

	cfg := &Config{
		SPNs: []SPNEntry{
			{
				Name: "Example Service",
				SPN:  "HTTP/example.com@REALM.COM",
			},
		},
	}

	return SaveConfig(cfg, path)
}

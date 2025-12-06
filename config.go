package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config represents the application configuration
type Config struct {
	SPNs     []SPNEntry     `json:"spns"`
	Secrets  []SecretEntry  `json:"secrets,omitempty"`
	URLs     []URLEntry     `json:"urls,omitempty"`
	Snippets []SnippetEntry `json:"snippets,omitempty"`
}

// SnippetEntry represents a text snippet that can be copied to clipboard
type SnippetEntry struct {
	Index int    `json:"index"` // Numeric index for ordering/reference
	Name  string `json:"name"`  // Display name in menu
	Value string `json:"value"` // The value to copy to clipboard
}

// URLEntry represents a URL bookmark
type URLEntry struct {
	Name string `json:"name"` // Display name in menu
	URL  string `json:"url"`  // The URL to open
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

// DefaultConfigPath returns the default configuration file path
func DefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "krb5tray.json")
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

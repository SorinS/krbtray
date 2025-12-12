package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

var log *logrus.Logger

// InitLogger initializes the logger with file rotation using default config
// Logs are written to ~/.config/ktray/ktray.log
func InitLogger() error {
	return InitLoggerWithConfig(DefaultLogConfig())
}

// InitLoggerWithConfig initializes the logger with the provided configuration
func InitLoggerWithConfig(cfg LogConfig) error {
	log = logrus.New()

	// Create log directory if needed
	logDir := ConfigDir()
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	logPath := filepath.Join(logDir, "ktray.log")

	// Configure lumberjack for log rotation
	lj := &lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    cfg.MaxSizeMB,  // MB - rotate when file reaches this size
		MaxBackups: cfg.MaxBackups, // Number of backup files to keep
		MaxAge:     cfg.MaxAgeDays, // Days to keep old files
		Compress:   cfg.Compress,   // Compress rotated files
		LocalTime:  true,           // Use local time for rotation
	}

	// Write to file, and optionally to stdout
	if cfg.ToStdout {
		multiWriter := io.MultiWriter(lj, os.Stdout)
		log.SetOutput(multiWriter)
	} else {
		log.SetOutput(lj)
	}

	// Set formatter - use text format with timestamps
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		DisableColors:   true, // No colors in log file
	})

	// Default to Info level, Debug mode will change this
	log.SetLevel(logrus.InfoLevel)

	log.WithFields(logrus.Fields{
		"max_size_mb":  cfg.MaxSizeMB,
		"max_backups":  cfg.MaxBackups,
		"max_age_days": cfg.MaxAgeDays,
		"compress":     cfg.Compress,
		"to_stdout":    cfg.ToStdout,
	}).Info("Logger initialized")
	return nil
}

// SetLogLevel sets the logging level based on debug mode
func SetLogLevel(debug bool) {
	if log == nil {
		return
	}
	if debug {
		log.SetLevel(logrus.DebugLevel)
		log.Debug("Debug logging enabled")
	} else {
		log.SetLevel(logrus.InfoLevel)
		log.Info("Debug logging disabled")
	}
}

// LogInfo logs an info level message (always logged)
func LogInfo(format string, args ...interface{}) {
	if log != nil {
		log.Infof(format, args...)
	}
}

// LogDebug logs a debug level message (only when debug mode is on)
func LogDebug(format string, args ...interface{}) {
	if log != nil {
		log.Debugf(format, args...)
	}
}

// LogWarn logs a warning level message
func LogWarn(format string, args ...interface{}) {
	if log != nil {
		log.Warnf(format, args...)
	}
}

// LogError logs an error level message
func LogError(format string, args ...interface{}) {
	if log != nil {
		log.Errorf(format, args...)
	}
}

// LogAction logs a business action (always logged at info level)
// Used for tracking user actions like selecting SPN, copying tokens, etc.
func LogAction(action string, details string) {
	if log != nil {
		log.WithFields(logrus.Fields{
			"action": action,
		}).Info(details)
	}
}

// LogActionWithFields logs a business action with additional fields
func LogActionWithFields(action string, details string, fields map[string]interface{}) {
	if log != nil {
		f := logrus.Fields{"action": action}
		for k, v := range fields {
			f[k] = v
		}
		log.WithFields(f).Info(details)
	}
}

// LogStartup logs application startup information
func LogStartup() {
	if log == nil {
		return
	}
	log.WithFields(logrus.Fields{
		"version":    Version,
		"commit":     getShortCommit(),
		"build_date": buildDate,
		"pid":        os.Getpid(),
	}).Info("krb5tray starting")
}

// LogShutdown logs application shutdown
func LogShutdown() {
	if log != nil {
		log.Info("krb5tray shutting down")
	}
}

// LogConfigLoaded logs when configuration is loaded
func LogConfigLoaded(spnCount, secretCount, urlCount, snippetCount, sshCount int) {
	if log != nil {
		log.WithFields(logrus.Fields{
			"spns":     spnCount,
			"secrets":  secretCount,
			"urls":     urlCount,
			"snippets": snippetCount,
			"ssh":      sshCount,
		}).Info("Configuration loaded")
	}
}

// LogSPNSelected logs when an SPN is selected (without exposing the full SPN)
func LogSPNSelected(displayName string) {
	LogAction("spn_selected", fmt.Sprintf("Selected SPN: %s", displayName))
}

// LogTicketRequested logs when a ticket is requested
func LogTicketRequested(displayName string, success bool, tokenSize int) {
	if success {
		log.WithFields(logrus.Fields{
			"action":     "ticket_request",
			"spn":        displayName,
			"token_size": tokenSize,
		}).Info("Ticket obtained successfully")
	} else {
		log.WithFields(logrus.Fields{
			"action": "ticket_request",
			"spn":    displayName,
		}).Warn("Ticket request failed")
	}
}

// LogClipboardCopy logs clipboard operations (without exposing content)
func LogClipboardCopy(itemType string, itemName string) {
	LogAction("clipboard_copy", fmt.Sprintf("Copied %s: %s", itemType, itemName))
}

// LogURLOpened logs when a URL is opened
func LogURLOpened(name string) {
	LogAction("url_opened", fmt.Sprintf("Opened URL: %s", name))
}

// LogSSHOpened logs when an SSH connection is opened
func LogSSHOpened(name string) {
	LogAction("ssh_opened", fmt.Sprintf("Opened SSH: %s", name))
}

// LogScriptExecuted logs when a Lua script is executed
func LogScriptExecuted(scriptName string, entryType string, err error) {
	status := "success"
	fields := logrus.Fields{
		"action":     "script_executed",
		"script":     scriptName,
		"entry_type": entryType,
	}
	if err != nil {
		status = "failed"
		fields["error"] = err.Error()
	}
	fields["status"] = status
	log.WithFields(fields).Info(fmt.Sprintf("Script executed: %s", scriptName))
}

// LogHotkeyTriggered logs when a hotkey is triggered
func LogHotkeyTriggered(hotkeyType string, index int) {
	LogDebug("Hotkey triggered: %s index=%d", hotkeyType, index)
}

// LogSecretSelected logs when a secret is selected
func LogSecretSelected(name string) {
	LogAction("secret_selected", fmt.Sprintf("Selected secret: %s", name))
}

// DailyRotationCheck checks if we should force a rotation (called periodically)
// This is a workaround since lumberjack doesn't support time-based rotation
func DailyRotationCheck() {
	// lumberjack handles rotation by size, and MaxAge handles cleanup
	// For true daily rotation, we could add a timestamp check here
	// but for simplicity, we rely on size-based rotation + age cleanup
}

// GetLogPath returns the path to the log file
func GetLogPath() string {
	return filepath.Join(ConfigDir(), "ktray.log")
}

// Helper to get current timestamp for logging
func logTimestamp() string {
	return time.Now().Format("2006-01-02 15:04:05")
}
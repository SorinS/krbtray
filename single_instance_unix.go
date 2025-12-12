//go:build !windows

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

var lockFile *os.File

// EnsureSingleInstance ensures only one instance of the application is running.
// Returns an error if another instance is already running.
func EnsureSingleInstance() error {
	lockPath := getLockFilePath()

	// Create directory if it doesn't exist
	dir := filepath.Dir(lockPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create lock directory: %w", err)
	}

	// Try to create/open the lock file
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return fmt.Errorf("failed to open lock file: %w", err)
	}

	// Try to acquire an exclusive lock (non-blocking)
	err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		f.Close()
		return fmt.Errorf("another instance of krb5tray is already running")
	}

	// Write PID to lock file
	f.Truncate(0)
	f.Seek(0, 0)
	fmt.Fprintf(f, "%d\n", os.Getpid())

	// Keep the file open (lock is held as long as file is open)
	lockFile = f

	return nil
}

// ReleaseSingleInstance releases the lock file
func ReleaseSingleInstance() {
	if lockFile != nil {
		syscall.Flock(int(lockFile.Fd()), syscall.LOCK_UN)
		lockFile.Close()
		os.Remove(getLockFilePath())
		lockFile = nil
	}
}

func getLockFilePath() string {
	// Use a standard location for the lock file
	home, err := os.UserHomeDir()
	if err != nil {
		home = os.TempDir()
	}
	return filepath.Join(home, ".config", "krb5tray.lock")
}
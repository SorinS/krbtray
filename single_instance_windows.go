//go:build windows

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/windows"
)

var lockHandle windows.Handle

// EnsureSingleInstance ensures only one instance of the application is running.
// Returns an error if another instance is already running.
// On Windows, uses a named mutex for single instance detection.
func EnsureSingleInstance() error {
	// Use a named mutex for single instance on Windows
	mutexName, err := windows.UTF16PtrFromString("Global\\krb5tray-single-instance")
	if err != nil {
		return fmt.Errorf("failed to create mutex name: %w", err)
	}

	handle, err := windows.CreateMutex(nil, false, mutexName)
	if err != nil {
		if err == windows.ERROR_ALREADY_EXISTS {
			return fmt.Errorf("another instance of krb5tray is already running")
		}
		return fmt.Errorf("failed to create mutex: %w", err)
	}

	// Check if we actually got the mutex or if it already existed
	// WAIT_OBJECT_0 (0) means we got the mutex
	// WAIT_TIMEOUT (0x102) means timeout (mutex held by another)
	event, err := windows.WaitForSingleObject(handle, 0)
	if err != nil || event != windows.WAIT_OBJECT_0 {
		windows.CloseHandle(handle)
		return fmt.Errorf("another instance of krb5tray is already running")
	}

	lockHandle = handle

	// Also write a PID file for compatibility
	writePIDFile()

	return nil
}

// ReleaseSingleInstance releases the mutex
func ReleaseSingleInstance() {
	if lockHandle != 0 {
		windows.ReleaseMutex(lockHandle)
		windows.CloseHandle(lockHandle)
		lockHandle = 0
	}
	removePIDFile()
}

func writePIDFile() {
	pidPath := getLockFilePath()
	dir := filepath.Dir(pidPath)
	os.MkdirAll(dir, 0755)
	f, err := os.Create(pidPath)
	if err == nil {
		fmt.Fprintf(f, "%d\n", os.Getpid())
		f.Close()
	}
}

func removePIDFile() {
	os.Remove(getLockFilePath())
}

func getLockFilePath() string {
	// Use a standard location for the lock file
	home, err := os.UserHomeDir()
	if err != nil {
		home = os.TempDir()
	}
	return filepath.Join(home, ".config", "krb5tray.lock")
}
//go:build windows
// +build windows

package main

import (
	"syscall"
	"time"
	"unsafe"
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	// Clipboard functions
	openClipboard    = user32.NewProc("OpenClipboard")
	closeClipboard   = user32.NewProc("CloseClipboard")
	emptyClipboard   = user32.NewProc("EmptyClipboard")
	setClipboardData = user32.NewProc("SetClipboardData")

	// Memory functions
	globalAlloc  = kernel32.NewProc("GlobalAlloc")
	globalLock   = kernel32.NewProc("GlobalLock")
	globalUnlock = kernel32.NewProc("GlobalUnlock")

	// Input simulation
	sendInput = user32.NewProc("SendInput")
)

const (
	cfUnicodeText = 13
	gmemMoveable  = 0x0002
)

// INPUT structure for SendInput
const (
	inputKeyboard  = 1
	keyEventFKeyUp = 0x0002
)

type keyboardInput struct {
	wVk         uint16
	wScan       uint16
	dwFlags     uint32
	time        uint32
	dwExtraInfo uintptr
}

type input struct {
	inputType uint32
	ki        keyboardInput
	padding   [8]byte // Padding to match C structure size
}

const (
	vkControl = 0x11
	vkV       = 0x56
)

func copyToClipboardPlatform(text string) error {
	// Convert to UTF-16
	utf16, err := syscall.UTF16FromString(text)
	if err != nil {
		return err
	}

	// Open clipboard
	ret, _, _ := openClipboard.Call(0)
	if ret == 0 {
		return syscall.GetLastError()
	}
	defer closeClipboard.Call()

	// Empty clipboard
	emptyClipboard.Call()

	// Allocate global memory for the text
	size := len(utf16) * 2 // UTF-16 = 2 bytes per character
	hMem, _, _ := globalAlloc.Call(gmemMoveable, uintptr(size))
	if hMem == 0 {
		return syscall.GetLastError()
	}

	// Lock memory and copy text
	ptr, _, _ := globalLock.Call(hMem)
	if ptr == 0 {
		return syscall.GetLastError()
	}

	// Copy UTF-16 data
	dst := unsafe.Slice((*uint16)(unsafe.Pointer(ptr)), len(utf16))
	copy(dst, utf16)

	globalUnlock.Call(hMem)

	// Set clipboard data
	ret, _, _ = setClipboardData.Call(cfUnicodeText, hMem)
	if ret == 0 {
		return syscall.GetLastError()
	}

	return nil
}

// pasteFromClipboard simulates Ctrl+V to paste the current clipboard contents
func pasteFromClipboard() {
	// Small delay to ensure clipboard is ready
	time.Sleep(50 * time.Millisecond)

	// Create input events for Ctrl+V
	inputs := []input{
		// Ctrl down
		{
			inputType: inputKeyboard,
			ki: keyboardInput{
				wVk:     vkControl,
				dwFlags: 0,
			},
		},
		// V down
		{
			inputType: inputKeyboard,
			ki: keyboardInput{
				wVk:     vkV,
				dwFlags: 0,
			},
		},
		// V up
		{
			inputType: inputKeyboard,
			ki: keyboardInput{
				wVk:     vkV,
				dwFlags: keyEventFKeyUp,
			},
		},
		// Ctrl up
		{
			inputType: inputKeyboard,
			ki: keyboardInput{
				wVk:     vkControl,
				dwFlags: keyEventFKeyUp,
			},
		},
	}

	// Send the input events
	sendInput.Call(
		uintptr(len(inputs)),
		uintptr(unsafe.Pointer(&inputs[0])),
		uintptr(unsafe.Sizeof(inputs[0])),
	)
}

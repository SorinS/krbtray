//go:build darwin
// +build darwin

package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa -framework CoreGraphics

#import <Cocoa/Cocoa.h>
#include <CoreGraphics/CoreGraphics.h>
#include <unistd.h>

void copyToClipboardNative(const char *text) {
    NSString *str = [NSString stringWithUTF8String:text];
    NSPasteboard *pasteboard = [NSPasteboard generalPasteboard];
    [pasteboard clearContents];
    [pasteboard setString:str forType:NSPasteboardTypeString];
}

// simulatePaste simulates Cmd+V keystroke to paste from clipboard
void simulatePaste(void) {
    CGEventSourceRef source = CGEventSourceCreate(kCGEventSourceStateHIDSystemState);
    if (source == NULL) return;

    // Press Command
    CGEventRef cmdDown = CGEventCreateKeyboardEvent(source, (CGKeyCode)55, true);
    CGEventSetFlags(cmdDown, kCGEventFlagMaskCommand);
    CGEventPost(kCGHIDEventTap, cmdDown);
    CFRelease(cmdDown);

    // Press V
    CGEventRef vDown = CGEventCreateKeyboardEvent(source, (CGKeyCode)9, true);
    CGEventSetFlags(vDown, kCGEventFlagMaskCommand);
    CGEventPost(kCGHIDEventTap, vDown);
    CFRelease(vDown);

    // Release V
    CGEventRef vUp = CGEventCreateKeyboardEvent(source, (CGKeyCode)9, false);
    CGEventPost(kCGHIDEventTap, vUp);
    CFRelease(vUp);

    // Release Command
    CGEventRef cmdUp = CGEventCreateKeyboardEvent(source, (CGKeyCode)55, false);
    CGEventPost(kCGHIDEventTap, cmdUp);
    CFRelease(cmdUp);

    CFRelease(source);
}
*/
import "C"
import "unsafe"

func copyToClipboardPlatform(text string) error {
	cstr := C.CString(text)
	defer C.free(unsafe.Pointer(cstr))
	C.copyToClipboardNative(cstr)
	return nil
}

// pasteFromClipboard simulates Cmd+V to paste the current clipboard contents
func pasteFromClipboard() {
	C.simulatePaste()
}

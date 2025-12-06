//go:build darwin
// +build darwin

package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa

#import <Cocoa/Cocoa.h>

void copyToClipboardNative(const char *text) {
    NSString *str = [NSString stringWithUTF8String:text];
    NSPasteboard *pasteboard = [NSPasteboard generalPasteboard];
    [pasteboard clearContents];
    [pasteboard setString:str forType:NSPasteboardTypeString];
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

//go:build darwin
// +build darwin

package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa

#import <Cocoa/Cocoa.h>

int openURLNative(const char *urlStr) {
    NSString *str = [NSString stringWithUTF8String:urlStr];
    NSURL *url = [NSURL URLWithString:str];
    if (url == nil) {
        return -1;
    }
    return [[NSWorkspace sharedWorkspace] openURL:url] ? 0 : -1;
}
*/
import "C"
import (
	"errors"
	"unsafe"
)

// openBrowser opens the specified URL in the default browser
func openBrowser(url string) error {
	cstr := C.CString(url)
	defer C.free(unsafe.Pointer(cstr))
	if C.openURLNative(cstr) != 0 {
		return errors.New("failed to open URL")
	}
	return nil
}
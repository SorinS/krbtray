//go:build darwin

package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa

#import <Cocoa/Cocoa.h>
#include <stdlib.h>

// showPromptDialog displays an NSAlert with a text input field
// Returns the entered text and 1 if OK was clicked, or empty string and 0 if cancelled
char* showPromptDialog(const char* title, const char* message, const char* defaultValue, int isSecure) {
    @autoreleasepool {
        // Ensure we're on the main thread for UI operations
        __block NSString* result = nil;
        __block BOOL confirmed = NO;

        void (^showAlert)(void) = ^{
            NSAlert *alert = [[NSAlert alloc] init];
            [alert setMessageText:[NSString stringWithUTF8String:title]];
            [alert setInformativeText:[NSString stringWithUTF8String:message]];
            [alert addButtonWithTitle:@"OK"];
            [alert addButtonWithTitle:@"Cancel"];
            [alert setAlertStyle:NSAlertStyleInformational];

            // Create text field for input
            NSTextField *input;
            if (isSecure) {
                input = [[NSSecureTextField alloc] initWithFrame:NSMakeRect(0, 0, 300, 24)];
            } else {
                input = [[NSTextField alloc] initWithFrame:NSMakeRect(0, 0, 300, 24)];
            }

            if (defaultValue != NULL) {
                [input setStringValue:[NSString stringWithUTF8String:defaultValue]];
            }
            [input setPlaceholderString:@"Enter value..."];
            [alert setAccessoryView:input];

            // Make the input field the first responder
            [[alert window] setInitialFirstResponder:input];

            // Run the alert
            NSModalResponse response = [alert runModal];

            if (response == NSAlertFirstButtonReturn) {
                result = [input stringValue];
                confirmed = YES;
            }
        };

        if ([NSThread isMainThread]) {
            showAlert();
        } else {
            dispatch_sync(dispatch_get_main_queue(), showAlert);
        }

        if (confirmed && result != nil) {
            // Return the result with a prefix to indicate success
            NSString *prefixed = [NSString stringWithFormat:@"1:%@", result];
            return strdup([prefixed UTF8String]);
        } else {
            return strdup("0:");
        }
    }
}

// showConfirmDialog displays a simple Yes/No confirmation dialog
int showConfirmDialog(const char* title, const char* message) {
    @autoreleasepool {
        __block BOOL confirmed = NO;

        void (^showAlert)(void) = ^{
            NSAlert *alert = [[NSAlert alloc] init];
            [alert setMessageText:[NSString stringWithUTF8String:title]];
            [alert setInformativeText:[NSString stringWithUTF8String:message]];
            [alert addButtonWithTitle:@"Yes"];
            [alert addButtonWithTitle:@"No"];
            [alert setAlertStyle:NSAlertStyleInformational];

            NSModalResponse response = [alert runModal];
            confirmed = (response == NSAlertFirstButtonReturn);
        };

        if ([NSThread isMainThread]) {
            showAlert();
        } else {
            dispatch_sync(dispatch_get_main_queue(), showAlert);
        }

        return confirmed ? 1 : 0;
    }
}
*/
import "C"
import (
	"strings"
	"unsafe"
)

// PromptForInput shows a dialog asking the user for text input
// Returns the entered text and true if OK was clicked, or empty string and false if cancelled
func PromptForInput(title, message, defaultValue string, secure bool) (string, bool) {
	cTitle := C.CString(title)
	cMessage := C.CString(message)
	cDefault := C.CString(defaultValue)
	defer C.free(unsafe.Pointer(cTitle))
	defer C.free(unsafe.Pointer(cMessage))
	defer C.free(unsafe.Pointer(cDefault))

	isSecure := C.int(0)
	if secure {
		isSecure = C.int(1)
	}

	cResult := C.showPromptDialog(cTitle, cMessage, cDefault, isSecure)
	defer C.free(unsafe.Pointer(cResult))

	result := C.GoString(cResult)

	// Parse result: "1:value" for success, "0:" for cancel
	if strings.HasPrefix(result, "1:") {
		return result[2:], true
	}
	return "", false
}

// ConfirmDialog shows a Yes/No confirmation dialog
// Returns true if Yes was clicked, false otherwise
func ConfirmDialog(title, message string) bool {
	cTitle := C.CString(title)
	cMessage := C.CString(message)
	defer C.free(unsafe.Pointer(cTitle))
	defer C.free(unsafe.Pointer(cMessage))

	result := C.showConfirmDialog(cTitle, cMessage)
	return result == 1
}

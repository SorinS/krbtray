//go:build linux
// +build linux

package main

/*
#cgo LDFLAGS: -lX11 -lXtst

#include <X11/Xlib.h>
#include <X11/Xatom.h>
#include <X11/extensions/XTest.h>
#include <X11/keysym.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <time.h>

// Clipboard data storage
static char* clipboard_data = NULL;
static size_t clipboard_len = 0;
static Display* clip_display = NULL;
static Window clip_window = 0;
static Atom clipboard_atom;
static Atom targets_atom;
static Atom utf8_atom;
static Atom text_atom;

// Initialize clipboard support
int init_clipboard() {
    if (clip_display != NULL) return 1;

    clip_display = XOpenDisplay(NULL);
    if (clip_display == NULL) return 0;

    clipboard_atom = XInternAtom(clip_display, "CLIPBOARD", False);
    targets_atom = XInternAtom(clip_display, "TARGETS", False);
    utf8_atom = XInternAtom(clip_display, "UTF8_STRING", False);
    text_atom = XInternAtom(clip_display, "TEXT", False);

    // Create a hidden window for clipboard ownership
    clip_window = XCreateSimpleWindow(clip_display,
        DefaultRootWindow(clip_display),
        0, 0, 1, 1, 0, 0, 0);

    return 1;
}

// Set clipboard content
int set_clipboard(const char* text, size_t len) {
    if (!init_clipboard()) return 0;

    // Free old data
    if (clipboard_data != NULL) {
        free(clipboard_data);
    }

    // Copy new data
    clipboard_data = (char*)malloc(len + 1);
    if (clipboard_data == NULL) return 0;
    memcpy(clipboard_data, text, len);
    clipboard_data[len] = '\0';
    clipboard_len = len;

    // Claim clipboard ownership
    XSetSelectionOwner(clip_display, clipboard_atom, clip_window, CurrentTime);
    XFlush(clip_display);

    // Check if we got ownership
    if (XGetSelectionOwner(clip_display, clipboard_atom) != clip_window) {
        return 0;
    }

    return 1;
}

// Handle clipboard selection requests (must be called periodically or in event loop)
void handle_clipboard_events() {
    if (clip_display == NULL) return;

    XEvent event;
    while (XPending(clip_display)) {
        XNextEvent(clip_display, &event);

        if (event.type == SelectionRequest) {
            XSelectionRequestEvent* req = &event.xselectionrequest;
            XSelectionEvent response;

            response.type = SelectionNotify;
            response.display = req->display;
            response.requestor = req->requestor;
            response.selection = req->selection;
            response.target = req->target;
            response.time = req->time;
            response.property = None;

            if (req->target == targets_atom) {
                // Return supported targets
                Atom targets[] = {targets_atom, utf8_atom, XA_STRING, text_atom};
                XChangeProperty(clip_display, req->requestor, req->property,
                    XA_ATOM, 32, PropModeReplace,
                    (unsigned char*)targets, 4);
                response.property = req->property;
            } else if (req->target == utf8_atom || req->target == XA_STRING || req->target == text_atom) {
                // Return clipboard data
                if (clipboard_data != NULL) {
                    XChangeProperty(clip_display, req->requestor, req->property,
                        req->target, 8, PropModeReplace,
                        (unsigned char*)clipboard_data, clipboard_len);
                    response.property = req->property;
                }
            }

            XSendEvent(clip_display, req->requestor, False, 0, (XEvent*)&response);
            XFlush(clip_display);
        }
    }
}

// Simulate Ctrl+V keystroke
void simulate_paste() {
    Display* dpy = XOpenDisplay(NULL);
    if (dpy == NULL) return;

    // Small delay to ensure clipboard is ready
    usleep(50000); // 50ms

    // Get keycodes
    KeyCode ctrl = XKeysymToKeycode(dpy, XK_Control_L);
    KeyCode v = XKeysymToKeycode(dpy, XK_v);

    // Press Ctrl
    XTestFakeKeyEvent(dpy, ctrl, True, 0);
    // Press V
    XTestFakeKeyEvent(dpy, v, True, 0);
    // Release V
    XTestFakeKeyEvent(dpy, v, False, 0);
    // Release Ctrl
    XTestFakeKeyEvent(dpy, ctrl, False, 0);

    XFlush(dpy);
    XCloseDisplay(dpy);
}

// Cleanup
void cleanup_clipboard() {
    if (clipboard_data != NULL) {
        free(clipboard_data);
        clipboard_data = NULL;
    }
    if (clip_display != NULL) {
        if (clip_window != 0) {
            XDestroyWindow(clip_display, clip_window);
        }
        XCloseDisplay(clip_display);
        clip_display = NULL;
    }
}
*/
import "C"

import (
	"fmt"
	"sync"
	"time"
	"unsafe"
)

var (
	clipboardMu     sync.Mutex
	clipboardInited bool
	clipboardDone   chan struct{}
)

// startClipboardEventLoop starts a goroutine to handle X11 selection events
func startClipboardEventLoop() {
	if clipboardDone != nil {
		return
	}
	clipboardDone = make(chan struct{})
	go func() {
		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-clipboardDone:
				return
			case <-ticker.C:
				clipboardMu.Lock()
				C.handle_clipboard_events()
				clipboardMu.Unlock()
			}
		}
	}()
}

func copyToClipboardPlatform(text string) error {
	clipboardMu.Lock()
	defer clipboardMu.Unlock()

	// Start event loop if not already running
	if !clipboardInited {
		startClipboardEventLoop()
		clipboardInited = true
	}

	cstr := C.CString(text)
	defer C.free(unsafe.Pointer(cstr))

	ret := C.set_clipboard(cstr, C.size_t(len(text)))
	if ret == 0 {
		return fmt.Errorf("failed to set clipboard")
	}

	return nil
}

// pasteFromClipboard simulates Ctrl+V to paste the current clipboard contents
func pasteFromClipboard() {
	C.simulate_paste()
}

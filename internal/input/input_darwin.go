//go:build darwin

package input

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework ApplicationServices -framework Foundation
#import <ApplicationServices/ApplicationServices.h>
#import <Foundation/Foundation.h>
#include <stdlib.h>

void typeText(const char* text) {
    NSString *str = [NSString stringWithUTF8String:text];

    for (NSUInteger i = 0; i < [str length]; i++) {
        unichar c = [str characterAtIndex:i];

        CGEventRef keyDown = CGEventCreateKeyboardEvent(NULL, 0, true);
        CGEventRef keyUp = CGEventCreateKeyboardEvent(NULL, 0, false);

        CGEventKeyboardSetUnicodeString(keyDown, 1, &c);
        CGEventKeyboardSetUnicodeString(keyUp, 1, &c);

        CGEventPost(kCGHIDEventTap, keyDown);
        CGEventPost(kCGHIDEventTap, keyUp);

        CFRelease(keyDown);
        CFRelease(keyUp);
    }
}
*/
import "C"
import "unsafe"

type darwinTyper struct{}

func newTyper() (Typer, error) {
	return &darwinTyper{}, nil
}

func (t *darwinTyper) Type(text string) error {
	cstr := C.CString(text)
	defer C.free(unsafe.Pointer(cstr))
	C.typeText(cstr)
	return nil
}

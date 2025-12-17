//go:build windows

package input

import (
	"syscall"
	"unicode/utf16"
	"unsafe"
)

var (
	user32           = syscall.NewLazyDLL("user32.dll")
	procSendInput    = user32.NewProc("SendInput")
	procGetKeyState  = user32.NewProc("GetKeyboardState")
)

const (
	inputKeyboard = 1
	keyEventFKeyUp    = 0x0002
	keyEventFUnicode  = 0x0004
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
	padding   uint64
}

type windowsTyper struct{}

func newTyper() (Typer, error) {
	return &windowsTyper{}, nil
}

func (t *windowsTyper) Type(text string) error {
	runes := utf16.Encode([]rune(text))
	inputs := make([]input, 0, len(runes)*2)

	for _, r := range runes {
		// Key down
		inputs = append(inputs, input{
			inputType: inputKeyboard,
			ki: keyboardInput{
				wScan:   r,
				dwFlags: keyEventFUnicode,
			},
		})
		// Key up
		inputs = append(inputs, input{
			inputType: inputKeyboard,
			ki: keyboardInput{
				wScan:   r,
				dwFlags: keyEventFUnicode | keyEventFKeyUp,
			},
		})
	}

	if len(inputs) == 0 {
		return nil
	}

	procSendInput.Call(
		uintptr(len(inputs)),
		uintptr(unsafe.Pointer(&inputs[0])),
		uintptr(unsafe.Sizeof(inputs[0])),
	)

	return nil
}

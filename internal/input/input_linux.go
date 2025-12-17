//go:build linux

package input

import (
	"os"
	"os/exec"
)

type linuxTyper struct {
	useWayland bool
}

func newTyper() (Typer, error) {
	t := &linuxTyper{
		useWayland: os.Getenv("WAYLAND_DISPLAY") != "",
	}
	return t, nil
}

func (t *linuxTyper) Type(text string) error {
	if t.useWayland {
		return t.typeWayland(text)
	}
	return t.typeX11(text)
}

func (t *linuxTyper) typeX11(text string) error {
	cmd := exec.Command("xdotool", "type", "--clearmodifiers", "--", text)
	return cmd.Run()
}

func (t *linuxTyper) typeWayland(text string) error {
	cmd := exec.Command("wtype", text)
	return cmd.Run()
}

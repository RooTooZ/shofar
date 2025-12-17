//go:build linux

package hotkey

import (
	"golang.design/x/hotkey"
	"shofar/internal/config"
)

// modifierMap маппинг config.Modifier -> hotkey.Modifier для Linux
var modifierMap = map[config.Modifier]hotkey.Modifier{
	config.ModCtrl:  hotkey.ModCtrl,
	config.ModShift: hotkey.ModShift,
	config.ModAlt:   hotkey.Mod1,  // Alt = Mod1 на X11
	config.ModSuper: hotkey.Mod4,  // Super/Win = Mod4 на X11
}

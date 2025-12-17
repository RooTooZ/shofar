//go:build darwin

package hotkey

import (
	"golang.design/x/hotkey"
	"shofar/internal/config"
)

// modifierMap маппинг config.Modifier -> hotkey.Modifier для macOS
var modifierMap = map[config.Modifier]hotkey.Modifier{
	config.ModCtrl:  hotkey.ModCtrl,
	config.ModShift: hotkey.ModShift,
	config.ModAlt:   hotkey.ModOption,
	config.ModSuper: hotkey.ModCmd,
}

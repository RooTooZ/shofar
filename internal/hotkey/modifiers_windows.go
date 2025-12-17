//go:build windows

package hotkey

import (
	"golang.design/x/hotkey"
	"whisper-input/internal/config"
)

// modifierMap маппинг config.Modifier -> hotkey.Modifier для Windows
var modifierMap = map[config.Modifier]hotkey.Modifier{
	config.ModCtrl:  hotkey.ModCtrl,
	config.ModShift: hotkey.ModShift,
	config.ModAlt:   hotkey.ModAlt,
	config.ModSuper: hotkey.ModWin,
}

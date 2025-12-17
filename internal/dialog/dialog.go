// Package dialog предоставляет GUI диалоги для настройки приложения.
package dialog

import (
	"fmt"
	"strings"

	"github.com/ncruces/zenity"
	"shofar/internal/config"
)

// SelectHotkey открывает диалог выбора горячей клавиши.
// Возвращает выбранную конфигурацию или ошибку если пользователь отменил.
func SelectHotkey(current config.HotkeyConfig) (config.HotkeyConfig, error) {
	// Шаг 1: Выбор модификаторов
	modOptions := []string{"Ctrl", "Shift", "Alt", "Super (Win/Cmd)"}
	modValues := []config.Modifier{config.ModCtrl, config.ModShift, config.ModAlt, config.ModSuper}

	// Определяем текущие выбранные модификаторы
	currentMods := make([]string, 0)
	for _, m := range current.Modifiers {
		switch m {
		case config.ModCtrl:
			currentMods = append(currentMods, "Ctrl")
		case config.ModShift:
			currentMods = append(currentMods, "Shift")
		case config.ModAlt:
			currentMods = append(currentMods, "Alt")
		case config.ModSuper:
			currentMods = append(currentMods, "Super (Win/Cmd)")
		}
	}

	selectedMods, err := zenity.ListMultiple(
		"Выберите модификаторы:",
		modOptions,
		zenity.Title("Настройка горячей клавиши - Модификаторы"),
		zenity.DefaultItems(currentMods...),
	)
	if err != nil {
		return current, err // Пользователь отменил
	}

	if len(selectedMods) == 0 {
		return current, fmt.Errorf("необходимо выбрать хотя бы один модификатор")
	}

	// Преобразуем выбранные модификаторы
	newMods := make([]config.Modifier, 0, len(selectedMods))
	for _, s := range selectedMods {
		for i, opt := range modOptions {
			if s == opt {
				newMods = append(newMods, modValues[i])
				break
			}
		}
	}

	// Шаг 2: Выбор клавиши
	keyOptions := []string{
		"Space", "Return", "Tab",
		"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M",
		"N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z",
		"F1", "F2", "F3", "F4", "F5", "F6", "F7", "F8", "F9", "F10", "F11", "F12",
	}
	keyValues := []config.Key{
		config.KeySpace, config.KeyReturn, config.KeyTab,
		config.KeyA, config.KeyB, config.KeyC, config.KeyD, config.KeyE,
		config.KeyF, config.KeyG, config.KeyH, config.KeyI, config.KeyJ,
		config.KeyK, config.KeyL, config.KeyM, config.KeyN, config.KeyO,
		config.KeyP, config.KeyQ, config.KeyR, config.KeyS, config.KeyT,
		config.KeyU, config.KeyV, config.KeyW, config.KeyX, config.KeyY, config.KeyZ,
		config.KeyF1, config.KeyF2, config.KeyF3, config.KeyF4,
		config.KeyF5, config.KeyF6, config.KeyF7, config.KeyF8,
		config.KeyF9, config.KeyF10, config.KeyF11, config.KeyF12,
	}

	// Текущая клавиша
	currentKey := strings.ToUpper(string(current.Key))
	if current.Key == config.KeySpace {
		currentKey = "Space"
	} else if current.Key == config.KeyReturn {
		currentKey = "Return"
	} else if current.Key == config.KeyTab {
		currentKey = "Tab"
	}

	selectedKey, err := zenity.List(
		"Выберите клавишу:",
		keyOptions,
		zenity.Title("Настройка горячей клавиши - Клавиша"),
		zenity.DefaultItems(currentKey),
	)
	if err != nil {
		return current, err // Пользователь отменил
	}

	// Преобразуем выбранную клавишу
	var newKey config.Key
	for i, opt := range keyOptions {
		if selectedKey == opt {
			newKey = keyValues[i]
			break
		}
	}

	return config.HotkeyConfig{
		Modifiers: newMods,
		Key:       newKey,
	}, nil
}

// ShowInfo показывает информационное сообщение.
func ShowInfo(title, message string) {
	zenity.Info(message, zenity.Title(title))
}

// ShowError показывает сообщение об ошибке.
func ShowError(title, message string) {
	zenity.Error(message, zenity.Title(title))
}

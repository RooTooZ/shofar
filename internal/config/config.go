// Package config предоставляет конфигурацию приложения с сохранением в файл.
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// Modifier представляет модификатор клавиши.
type Modifier string

const (
	ModCtrl  Modifier = "ctrl"
	ModShift Modifier = "shift"
	ModAlt   Modifier = "alt"
	ModSuper Modifier = "super" // Win/Cmd
)

// Key представляет клавишу.
type Key string

const (
	KeySpace  Key = "space"
	KeyReturn Key = "return"
	KeyTab    Key = "tab"
	KeyA      Key = "a"
	KeyB      Key = "b"
	KeyC      Key = "c"
	KeyD      Key = "d"
	KeyE      Key = "e"
	KeyF      Key = "f"
	KeyG      Key = "g"
	KeyH      Key = "h"
	KeyI      Key = "i"
	KeyJ      Key = "j"
	KeyK      Key = "k"
	KeyL      Key = "l"
	KeyM      Key = "m"
	KeyN      Key = "n"
	KeyO      Key = "o"
	KeyP      Key = "p"
	KeyQ      Key = "q"
	KeyR      Key = "r"
	KeyS      Key = "s"
	KeyT      Key = "t"
	KeyU      Key = "u"
	KeyV      Key = "v"
	KeyW      Key = "w"
	KeyX      Key = "x"
	KeyY      Key = "y"
	KeyZ      Key = "z"
	KeyF1     Key = "f1"
	KeyF2     Key = "f2"
	KeyF3     Key = "f3"
	KeyF4     Key = "f4"
	KeyF5     Key = "f5"
	KeyF6     Key = "f6"
	KeyF7     Key = "f7"
	KeyF8     Key = "f8"
	KeyF9     Key = "f9"
	KeyF10    Key = "f10"
	KeyF11    Key = "f11"
	KeyF12    Key = "f12"
)

// HotkeyConfig хранит настройки горячей клавиши.
type HotkeyConfig struct {
	Modifiers []Modifier `json:"modifiers"`
	Key       Key        `json:"key"`
}

// String возвращает строковое представление горячей клавиши.
func (h HotkeyConfig) String() string {
	result := ""
	for _, m := range h.Modifiers {
		if result != "" {
			result += "+"
		}
		result += string(m)
	}
	if result != "" {
		result += "+"
	}
	result += string(h.Key)
	return result
}

// LLMConfig хранит настройки LLM для исправления текста.
type LLMConfig struct {
	Enabled bool   `json:"enabled"`
	ModelID string `json:"model_id,omitempty"` // ID модели из registry (llm-qwen2.5-0.5b)
}

// configData структура для сериализации.
type configData struct {
	Language      string       `json:"language"`
	UILanguage    string       `json:"ui_language,omitempty"`
	Notifications bool         `json:"notifications"`
	Hotkey        HotkeyConfig `json:"hotkey"`
	ModelID       string       `json:"model_id,omitempty"`
	LLM           LLMConfig    `json:"llm,omitempty"`
}

// Config хранит настройки приложения.
type Config struct {
	mu             sync.RWMutex
	language       string
	uiLanguage     string
	notifications  bool
	hotkey         HotkeyConfig
	modelID        string
	llm            LLMConfig
	configPath     string
	onHotkeyChange func(HotkeyConfig)
}

// New создаёт конфигурацию, загружая из файла или с настройками по умолчанию.
func New() *Config {
	c := &Config{
		language:      "auto", // auto для смешанного русского/английского
		uiLanguage:    "ru",   // По умолчанию русский интерфейс
		notifications: true,
		hotkey: HotkeyConfig{
			Modifiers: []Modifier{ModCtrl, ModShift},
			Key:       KeySpace,
		},
		llm: LLMConfig{
			Enabled: false,
			ModelID: "llm-qwen2.5-0.5b",
		},
	}

	// Определяем путь к файлу конфигурации рядом с бинарником
	execPath, err := os.Executable()
	if err == nil {
		// Резолвим симлинки
		execPath, err = filepath.EvalSymlinks(execPath)
		if err == nil {
			execDir := filepath.Dir(execPath)
			c.configPath = filepath.Join(execDir, "config.json")
		}
	}

	// Пытаемся загрузить конфигурацию
	c.load()

	return c
}

// load загружает конфигурацию из файла.
func (c *Config) load() {
	if c.configPath == "" {
		return
	}

	data, err := os.ReadFile(c.configPath)
	if err != nil {
		return // Файл не существует, используем defaults
	}

	var cfg configData
	if err := json.Unmarshal(data, &cfg); err != nil {
		return
	}

	c.language = cfg.Language
	if cfg.UILanguage != "" {
		c.uiLanguage = cfg.UILanguage
	}
	c.notifications = cfg.Notifications
	if cfg.Hotkey.Key != "" {
		c.hotkey = cfg.Hotkey
	}
	c.modelID = cfg.ModelID
	// LLM config
	c.llm.Enabled = cfg.LLM.Enabled
	if cfg.LLM.ModelID != "" {
		c.llm.ModelID = cfg.LLM.ModelID
	}
}

// save сохраняет конфигурацию в файл.
func (c *Config) save() {
	if c.configPath == "" {
		return
	}

	cfg := configData{
		Language:      c.language,
		UILanguage:    c.uiLanguage,
		Notifications: c.notifications,
		Hotkey:        c.hotkey,
		ModelID:       c.modelID,
		LLM:           c.llm,
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return
	}

	os.WriteFile(c.configPath, data, 0644)
}

// SetLanguage устанавливает язык распознавания.
func (c *Config) SetLanguage(lang string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.language = lang
	c.save()
}

// Language возвращает текущий язык распознавания.
func (c *Config) Language() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.language
}

// SetNotifications включает/выключает уведомления.
func (c *Config) SetNotifications(enabled bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.notifications = enabled
	c.save()
}

// ToggleNotifications переключает состояние уведомлений.
func (c *Config) ToggleNotifications() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.notifications = !c.notifications
	c.save()
	return c.notifications
}

// NotificationsEnabled возвращает true если уведомления включены.
func (c *Config) NotificationsEnabled() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.notifications
}

// Hotkey возвращает текущую горячую клавишу.
func (c *Config) Hotkey() HotkeyConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.hotkey
}

// SetHotkey устанавливает горячую клавишу.
func (c *Config) SetHotkey(hk HotkeyConfig) {
	c.mu.Lock()
	c.hotkey = hk
	callback := c.onHotkeyChange
	c.save()
	c.mu.Unlock()

	if callback != nil {
		callback(hk)
	}
}

// OnHotkeyChange устанавливает callback для изменения горячей клавиши.
func (c *Config) OnHotkeyChange(fn func(HotkeyConfig)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onHotkeyChange = fn
}

// ModelID возвращает ID текущей модели распознавания.
func (c *Config) ModelID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.modelID
}

// SetModelID устанавливает ID модели распознавания.
func (c *Config) SetModelID(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.modelID = id
	c.save()
}

// LLM возвращает текущие настройки LLM.
func (c *Config) LLM() LLMConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.llm
}

// SetLLM устанавливает настройки LLM.
func (c *Config) SetLLM(cfg LLMConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.llm = cfg
	c.save()
}

// SetLLMEnabled включает/выключает LLM коррекцию.
func (c *Config) SetLLMEnabled(enabled bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.llm.Enabled = enabled
	c.save()
}

// LLMEnabled возвращает true если LLM коррекция включена.
func (c *Config) LLMEnabled() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.llm.Enabled
}

// LLMModelID возвращает ID модели LLM.
func (c *Config) LLMModelID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.llm.ModelID
}

// SetLLMModelID устанавливает ID модели LLM.
func (c *Config) SetLLMModelID(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.llm.ModelID = id
	c.save()
}

// AvailableModifiers возвращает список доступных модификаторов.
func AvailableModifiers() []Modifier {
	return []Modifier{ModCtrl, ModShift, ModAlt, ModSuper}
}

// AvailableKeys возвращает список доступных клавиш.
func AvailableKeys() []Key {
	return []Key{
		KeySpace, KeyReturn, KeyTab,
		KeyA, KeyB, KeyC, KeyD, KeyE, KeyF, KeyG, KeyH, KeyI, KeyJ, KeyK, KeyL, KeyM,
		KeyN, KeyO, KeyP, KeyQ, KeyR, KeyS, KeyT, KeyU, KeyV, KeyW, KeyX, KeyY, KeyZ,
		KeyF1, KeyF2, KeyF3, KeyF4, KeyF5, KeyF6, KeyF7, KeyF8, KeyF9, KeyF10, KeyF11, KeyF12,
	}
}

// UILanguage возвращает язык интерфейса.
func (c *Config) UILanguage() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.uiLanguage
}

// SetUILanguage устанавливает язык интерфейса.
func (c *Config) SetUILanguage(lang string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.uiLanguage = lang
	c.save()
}

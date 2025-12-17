// Package settings provides Gio-based settings UI.
package settings

import (
	"context"
	"log"
	"sync"
	"time"

	"gioui.org/app"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"

	"shofar/internal/config"
	"shofar/internal/i18n"
	"shofar/internal/models"
)

// Colors are defined in widgets.go

// Window represents the settings dialog window.
type Window struct {
	mu      sync.Mutex
	manager *models.Manager
	config  *config.Config

	// Window state
	window  *app.Window
	running bool
	stopCh  chan struct{}
	doneCh  chan struct{}

	// UI state - Model
	selectedEngine models.Engine
	selectedModel  string

	// UI state - Hotkey
	hotkeyModifiers map[config.Modifier]bool
	hotkeyKey       config.Key

	// Download state
	downloading    bool
	downloadCtx    context.Context
	downloadCancel context.CancelFunc
	progress       float64
	progressModel  string

	// Model loading state
	loadingModel   bool
	loadingModelID string

	// Widgets - Engine/Model
	engineEnum    widget.Enum
	engineButtons map[models.Engine]*widget.Clickable
	modelButtons  map[string]*widget.Clickable
	downloadBtns  map[string]*widget.Clickable

	// Widgets - Hotkey
	modCtrl       widget.Bool
	modShift      widget.Bool
	modAlt        widget.Bool
	modSuper      widget.Bool
	keyEnum       widget.Enum
	keyButtons    map[config.Key]*widget.Clickable
	keyList       widget.List
	hotkeyEditBtn widget.Clickable
	hotkeyRecordTag int // stable tag for focus during recording
	recordingHotkey bool
	recordedMods    map[config.Modifier]bool
	recordedKey     config.Key
	hotkeyFilters   []event.Filter // cached filters for hotkey recording

	// Widgets - Buttons
	applyBtn  widget.Clickable
	cancelBtn widget.Clickable

	// Widgets - LLM
	llmEnabled widget.Bool

	// Widgets - UI Language
	selectedUILang i18n.Language
	langButtons    map[i18n.Language]*widget.Clickable

	// Scroll state
	modelList   widget.List
	contentList widget.List // Main scrollable content

	// Callbacks
	onApply        func(modelID string)
	onHotkeyChange func(config.HotkeyConfig)
	onLLMChange    func(enabled bool, modelID string)
	onUILangChange func(lang i18n.Language)
}

// New creates a new settings window.
func New(manager *models.Manager, cfg *config.Config) *Window {
	w := &Window{
		manager:         manager,
		config:          cfg,
		selectedEngine:  models.EngineWhisper,
		modelButtons:    make(map[string]*widget.Clickable),
		downloadBtns:    make(map[string]*widget.Clickable),
		hotkeyModifiers: make(map[config.Modifier]bool),
	}

	// Load current model selection from config
	currentModelID := cfg.ModelID()
	if currentModelID != "" {
		if info, ok := models.GetModel(currentModelID); ok {
			w.selectedEngine = info.Engine
			w.selectedModel = currentModelID
		}
	}

	// Load current hotkey from config
	currentHotkey := cfg.Hotkey()
	for _, m := range currentHotkey.Modifiers {
		w.hotkeyModifiers[m] = true
	}
	w.hotkeyKey = currentHotkey.Key

	// Initialize widgets for all models
	for _, m := range models.Registry {
		w.modelButtons[m.ID] = new(widget.Clickable)
		w.downloadBtns[m.ID] = new(widget.Clickable)
	}

	// Set engine enum value
	w.engineEnum.Value = string(w.selectedEngine)

	// Set modifier checkboxes
	w.modCtrl.Value = w.hotkeyModifiers[config.ModCtrl]
	w.modShift.Value = w.hotkeyModifiers[config.ModShift]
	w.modAlt.Value = w.hotkeyModifiers[config.ModAlt]
	w.modSuper.Value = w.hotkeyModifiers[config.ModSuper]

	// Set key enum value
	w.keyEnum.Value = string(w.hotkeyKey)

	// Initialize LLM toggle
	w.llmEnabled.Value = cfg.LLMEnabled()

	// Initialize UI language selector
	w.langButtons = make(map[i18n.Language]*widget.Clickable)
	for _, lang := range i18n.AvailableLanguages() {
		w.langButtons[lang] = new(widget.Clickable)
	}
	w.selectedUILang = i18n.GetLanguage()

	// Initialize lists
	w.modelList.Axis = layout.Vertical
	w.keyList.Axis = layout.Horizontal
	w.contentList.Axis = layout.Vertical

	// Initialize hotkey filters once
	w.initHotkeyFilters()

	return w
}

func (w *Window) initHotkeyFilters() {
	modifiers := key.ModCtrl | key.ModShift | key.ModAlt | key.ModSuper

	filters := []key.Filter{
		{Name: key.NameSpace, Optional: modifiers},
		{Name: key.NameTab, Optional: modifiers},
		{Name: key.NameReturn, Optional: modifiers},
		{Name: key.NameEscape, Optional: modifiers},
		{Name: key.NameF1, Optional: modifiers},
		{Name: key.NameF2, Optional: modifiers},
		{Name: key.NameF3, Optional: modifiers},
		{Name: key.NameF4, Optional: modifiers},
		{Name: key.NameF5, Optional: modifiers},
		{Name: key.NameF6, Optional: modifiers},
		{Name: key.NameF7, Optional: modifiers},
		{Name: key.NameF8, Optional: modifiers},
		{Name: key.NameF9, Optional: modifiers},
		{Name: key.NameF10, Optional: modifiers},
		{Name: key.NameF11, Optional: modifiers},
		{Name: key.NameF12, Optional: modifiers},
	}
	// Add letters A-Z
	for c := 'A'; c <= 'Z'; c++ {
		filters = append(filters, key.Filter{Name: key.Name(string(c)), Optional: modifiers})
	}
	// Also capture modifier-only events
	filters = append(filters, key.Filter{Optional: modifiers})

	w.hotkeyFilters = make([]event.Filter, len(filters))
	for i, f := range filters {
		w.hotkeyFilters[i] = f
	}
}

// OnApply sets the callback for when user applies model changes.
func (w *Window) OnApply(fn func(modelID string)) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.onApply = fn
}

// OnHotkeyChange sets the callback for when user changes hotkey.
func (w *Window) OnHotkeyChange(fn func(config.HotkeyConfig)) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.onHotkeyChange = fn
}

// OnLLMChange sets the callback for when user changes LLM settings.
func (w *Window) OnLLMChange(fn func(enabled bool, modelID string)) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.onLLMChange = fn
}

// OnUILangChange sets the callback for when user changes UI language.
func (w *Window) OnUILangChange(fn func(lang i18n.Language)) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.onUILangChange = fn
}

// Show displays the settings window (non-blocking).
func (w *Window) Show() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.running {
		return
	}

	// Reload current settings
	currentModelID := w.config.ModelID()
	if currentModelID != "" {
		if info, ok := models.GetModel(currentModelID); ok {
			w.selectedEngine = info.Engine
			w.selectedModel = currentModelID
		}
	}

	// Auto-select first downloaded model if none selected
	if w.selectedModel == "" {
		for _, m := range models.GetModelsByEngine(w.selectedEngine) {
			if w.manager.IsDownloaded(m) {
				w.selectedModel = m.ID
				break
			}
		}
	}

	w.engineEnum.Value = string(w.selectedEngine)

	currentHotkey := w.config.Hotkey()
	w.hotkeyModifiers = make(map[config.Modifier]bool)
	for _, m := range currentHotkey.Modifiers {
		w.hotkeyModifiers[m] = true
	}
	w.hotkeyKey = currentHotkey.Key
	w.modCtrl.Value = w.hotkeyModifiers[config.ModCtrl]
	w.modShift.Value = w.hotkeyModifiers[config.ModShift]
	w.modAlt.Value = w.hotkeyModifiers[config.ModAlt]
	w.modSuper.Value = w.hotkeyModifiers[config.ModSuper]
	w.keyEnum.Value = string(w.hotkeyKey)

	// Reload LLM setting
	w.llmEnabled.Value = w.config.LLMEnabled()

	w.running = true
	w.stopCh = make(chan struct{})
	w.doneCh = make(chan struct{})

	go w.runEventLoop()
}

// Hide closes the settings window.
func (w *Window) Hide() {
	w.mu.Lock()
	if !w.running {
		w.mu.Unlock()
		return
	}
	w.running = false
	stopCh := w.stopCh
	doneCh := w.doneCh
	w.stopCh = nil

	// Cancel any ongoing download
	if w.downloadCancel != nil {
		w.downloadCancel()
	}
	w.mu.Unlock()

	if stopCh != nil {
		close(stopCh)
	}

	if doneCh != nil {
		select {
		case <-doneCh:
		case <-time.After(time.Second):
		}
	}
}

// IsVisible returns true if window is currently shown.
func (w *Window) IsVisible() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.running
}

func (w *Window) runEventLoop() {
	defer close(w.doneCh)

	w.window = new(app.Window)
	w.window.Option(
		app.Title("Shofar - "+i18n.T("settings_title")),
		app.Size(unit.Dp(450), unit.Dp(600)),
		app.MinSize(unit.Dp(400), unit.Dp(500)),
	)

	var ops op.Ops

	// Invalidation goroutine
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-w.stopCh:
				if w.window != nil {
					w.window.Perform(system.ActionClose)
				}
				return
			case <-ticker.C:
				if w.window != nil {
					w.window.Invalidate()
				}
			}
		}
	}()

	for {
		switch e := w.window.Event().(type) {
		case app.DestroyEvent:
			return
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			w.handleEvents(gtx)
			w.draw(gtx)
			e.Frame(gtx.Ops)
		}
	}
}

func (w *Window) handleEvents(gtx layout.Context) {
	// Handle hotkey edit button
	if w.hotkeyEditBtn.Clicked(gtx) {
		w.mu.Lock()
		w.recordingHotkey = true
		w.recordedMods = make(map[config.Modifier]bool)
		w.recordedKey = ""
		w.mu.Unlock()
	}

	// Handle hotkey recording
	if w.recordingHotkey {
		w.handleHotkeyRecording(gtx)
	}

	// Handle engine selection change
	if w.engineEnum.Update(gtx) {
		w.mu.Lock()
		newEngine := models.Engine(w.engineEnum.Value)
		if newEngine != w.selectedEngine {
			w.selectedEngine = newEngine
			w.selectedModel = "" // Reset model selection
		}
		w.mu.Unlock()
	}

	// Handle model selection buttons (only speech recognition models, not LLM)
	for id, btn := range w.modelButtons {
		// Skip LLM models - they are handled separately in drawLLMModelItem
		// Check BEFORE calling Clicked() to avoid consuming the event
		if info, ok := models.GetModel(id); ok && info.Engine == models.EngineLLM {
			continue
		}
		if btn.Clicked(gtx) {
			w.mu.Lock()
			w.selectedModel = id
			w.mu.Unlock()
		}
	}

	// Handle download buttons
	for id, btn := range w.downloadBtns {
		if btn.Clicked(gtx) {
			w.startDownload(id)
		}
	}

	// Handle UI language buttons - apply immediately
	for lang, btn := range w.langButtons {
		if btn.Clicked(gtx) {
			w.mu.Lock()
			if w.selectedUILang != lang {
				w.selectedUILang = lang
				i18n.SetLanguage(lang)
				w.config.SetUILanguage(string(lang))
				callback := w.onUILangChange
				w.mu.Unlock()
				if callback != nil {
					callback(lang)
				}
			} else {
				w.mu.Unlock()
			}
		}
	}

	// Handle cancel button
	if w.cancelBtn.Clicked(gtx) {
		w.Hide()
	}

	// Handle apply button
	if w.applyBtn.Clicked(gtx) {
		w.applySettings()
	}
}

func (w *Window) handleHotkeyRecording(gtx layout.Context) {
	// Track modifiers at the time of key press (not release)
	var pressedMods map[config.Modifier]bool

	for {
		event, ok := gtx.Event(w.hotkeyFilters...)
		if !ok {
			break
		}

		switch e := event.(type) {
		case key.Event:
			w.mu.Lock()

			// Map key name to our config key
			if e.State == key.Press {
				// Store modifiers at the time of key press
				pressedMods = map[config.Modifier]bool{
					config.ModCtrl:  e.Modifiers.Contain(key.ModCtrl),
					config.ModShift: e.Modifiers.Contain(key.ModShift),
					config.ModAlt:   e.Modifiers.Contain(key.ModAlt),
					config.ModSuper: e.Modifiers.Contain(key.ModSuper),
				}
				w.recordedMods = pressedMods

				switch e.Name {
				case key.NameSpace:
					w.recordedKey = config.KeySpace
				case key.NameReturn:
					w.recordedKey = config.KeyReturn
				case key.NameTab:
					w.recordedKey = config.KeyTab
				case key.NameEscape:
					// Cancel recording
					w.recordingHotkey = false
					w.mu.Unlock()
					return
				case key.NameF1:
					w.recordedKey = config.KeyF1
				case key.NameF2:
					w.recordedKey = config.KeyF2
				case key.NameF3:
					w.recordedKey = config.KeyF3
				case key.NameF4:
					w.recordedKey = config.KeyF4
				case key.NameF5:
					w.recordedKey = config.KeyF5
				default:
					// Letter keys (A-Z)
					if len(e.Name) == 1 && e.Name >= "A" && e.Name <= "Z" {
						w.recordedKey = config.Key(string(e.Name[0] + 32)) // lowercase
					}
				}
			}

			// Check if we have modifiers + key
			hasModifiers := w.recordedMods[config.ModCtrl] || w.recordedMods[config.ModShift] ||
				w.recordedMods[config.ModAlt] || w.recordedMods[config.ModSuper]
			hasKey := w.recordedKey != ""

			// On key release, if we have modifiers + key, finish recording
			if e.State == key.Release && hasModifiers && hasKey {
				// Apply the recorded hotkey
				w.hotkeyModifiers = make(map[config.Modifier]bool)
				for k, v := range w.recordedMods {
					w.hotkeyModifiers[k] = v
				}
				w.hotkeyKey = w.recordedKey
				w.recordingHotkey = false
			}

			w.mu.Unlock()
		}
	}
}

func (w *Window) applySettings() {
	w.mu.Lock()
	// Prevent double apply
	if w.loadingModel {
		w.mu.Unlock()
		return
	}

	selectedModel := w.selectedModel
	modelCallback := w.onApply
	hotkeyCallback := w.onHotkeyChange
	llmCallback := w.onLLMChange
	llmEnabled := w.llmEnabled.Value
	llmModelID := w.config.LLMModelID()
	if llmModelID == "" {
		llmModelID = models.DefaultLLMModelID()
	}

	// Save LLM setting immediately
	w.config.SetLLMEnabled(llmEnabled)

	// Build hotkey config
	var mods []config.Modifier
	if w.hotkeyModifiers[config.ModCtrl] {
		mods = append(mods, config.ModCtrl)
	}
	if w.hotkeyModifiers[config.ModShift] {
		mods = append(mods, config.ModShift)
	}
	if w.hotkeyModifiers[config.ModAlt] {
		mods = append(mods, config.ModAlt)
	}
	if w.hotkeyModifiers[config.ModSuper] {
		mods = append(mods, config.ModSuper)
	}
	newHotkey := config.HotkeyConfig{
		Modifiers: mods,
		Key:       w.hotkeyKey,
	}
	w.mu.Unlock()

	// Apply hotkey if changed (this is fast, do it synchronously)
	currentHotkey := w.config.Hotkey()
	if newHotkey.String() != currentHotkey.String() {
		if len(mods) > 0 && newHotkey.Key != "" {
			if hotkeyCallback != nil {
				hotkeyCallback(newHotkey)
			}
		}
	}

	// Apply LLM settings change
	if llmCallback != nil {
		llmCallback(llmEnabled, llmModelID)
	}

	// Check if we need to load a speech recognition model (not LLM)
	needModelLoad := false
	if selectedModel != "" && modelCallback != nil {
		info, ok := models.GetModel(selectedModel)
		// Only load speech recognition models (Whisper/Vosk), not LLM
		if ok && w.manager.IsDownloaded(info) && info.Engine != models.EngineLLM {
			needModelLoad = true
		}
	}

	if !needModelLoad {
		// No model to load, just close
		w.Hide()
		return
	}

	// Load model in background with spinner
	w.mu.Lock()
	w.loadingModel = true
	w.loadingModelID = selectedModel
	w.mu.Unlock()

	go func() {
		// Call the callback (this is the slow part - loading model into memory)
		modelCallback(selectedModel)

		w.mu.Lock()
		w.loadingModel = false
		w.loadingModelID = ""
		w.mu.Unlock()

		// Hide window after loading is complete
		w.Hide()
	}()
}

func (w *Window) startDownload(modelID string) {
	w.mu.Lock()
	if w.downloading {
		w.mu.Unlock()
		return
	}

	info, ok := models.GetModel(modelID)
	if !ok {
		w.mu.Unlock()
		return
	}

	w.downloading = true
	w.progressModel = modelID
	w.progress = 0
	w.downloadCtx, w.downloadCancel = context.WithCancel(context.Background())
	ctx := w.downloadCtx
	w.mu.Unlock()

	go func() {
		progressCh := make(chan models.Progress, 10)

		go func() {
			for p := range progressCh {
				w.mu.Lock()
				if p.Total > 0 {
					w.progress = float64(p.Downloaded) / float64(p.Total)
				}
				w.mu.Unlock()
			}
		}()

		err := w.manager.Download(ctx, info, progressCh)
		close(progressCh)

		w.mu.Lock()
		w.downloading = false
		w.downloadCancel = nil
		if err == nil {
			w.selectedModel = modelID
		} else if err != context.Canceled {
			log.Printf("Settings: download error: %v", err)
		}
		w.mu.Unlock()
	}()
}

func (w *Window) getState() (engine models.Engine, selectedModel string, downloading bool, progress float64, progressModel string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.selectedEngine, w.selectedModel, w.downloading, w.progress, w.progressModel
}

func (w *Window) getLoadingState() (loading bool, modelID string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.loadingModel, w.loadingModelID
}

func (w *Window) getHotkeyState() (mods map[config.Modifier]bool, key config.Key) {
	w.mu.Lock()
	defer w.mu.Unlock()
	// Return a copy
	modsCopy := make(map[config.Modifier]bool)
	for k, v := range w.hotkeyModifiers {
		modsCopy[k] = v
	}
	return modsCopy, w.hotkeyKey
}

func (w *Window) isRecordingHotkey() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.recordingHotkey
}

func (w *Window) getRecordingState() (mods map[config.Modifier]bool, key config.Key) {
	w.mu.Lock()
	defer w.mu.Unlock()
	modsCopy := make(map[config.Modifier]bool)
	for k, v := range w.recordedMods {
		modsCopy[k] = v
	}
	return modsCopy, w.recordedKey
}

func (w *Window) getSelectedUILang() i18n.Language {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.selectedUILang
}

func (w *Window) getLangButton(lang i18n.Language) *widget.Clickable {
	if w.langButtons == nil {
		w.langButtons = make(map[i18n.Language]*widget.Clickable)
	}
	if w.langButtons[lang] == nil {
		w.langButtons[lang] = new(widget.Clickable)
	}
	return w.langButtons[lang]
}

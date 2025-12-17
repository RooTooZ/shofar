// Package tray предоставляет системный трей с меню.
package tray

import (
	"github.com/getlantern/systray"
	"whisper-input/embedded"
	"whisper-input/internal/i18n"
)

// State представляет состояние приложения для отображения в трее.
type State int

const (
	StateIdle State = iota
	StateRecording
	StateProcessing
)

// Callbacks содержит обработчики событий меню.
type Callbacks struct {
	OnNotificationsToggle func() bool
	OnSettingsClick       func()
	OnQuit                func()
}

// Tray управляет иконкой в системном трее.
type Tray struct {
	callbacks   Callbacks
	notifyOn    *systray.MenuItem
	status      *systray.MenuItem
	settingsBtn *systray.MenuItem
	quitBtn     *systray.MenuItem
}

// New создаёт новый Tray.
func New(callbacks Callbacks) *Tray {
	return &Tray{
		callbacks: callbacks,
	}
}

// Run запускает системный трей. Блокирующая функция.
func (t *Tray) Run(onReady func()) {
	systray.Run(func() {
		t.onReady()
		if onReady != nil {
			onReady()
		}
	}, t.onExit)
}

func (t *Tray) onReady() {
	systray.SetIcon(embedded.IconIdle)
	systray.SetTitle("Shofar")
	systray.SetTooltip(i18n.T("app_tooltip"))

	// Статус
	t.status = systray.AddMenuItem(i18n.T("tray_ready"), "")
	t.status.Disable()

	systray.AddSeparator()

	// Уведомления
	t.notifyOn = systray.AddMenuItemCheckbox(i18n.T("tray_notifications"), i18n.T("tray_notifications_hint"), true)

	// Настройки
	t.settingsBtn = systray.AddMenuItem(i18n.T("tray_settings"), i18n.T("tray_settings_hint"))

	systray.AddSeparator()

	// Выход
	t.quitBtn = systray.AddMenuItem(i18n.T("tray_quit"), i18n.T("tray_quit_hint"))

	// Обработка событий меню
	go t.handleMenuEvents()
}

func (t *Tray) handleMenuEvents() {
	for {
		select {
		// Уведомления
		case <-t.notifyOn.ClickedCh:
			if t.callbacks.OnNotificationsToggle != nil {
				enabled := t.callbacks.OnNotificationsToggle()
				if enabled {
					t.notifyOn.Check()
				} else {
					t.notifyOn.Uncheck()
				}
			}

		// Настройки
		case <-t.settingsBtn.ClickedCh:
			if t.callbacks.OnSettingsClick != nil {
				t.callbacks.OnSettingsClick()
			}

		// Выход
		case <-t.quitBtn.ClickedCh:
			if t.callbacks.OnQuit != nil {
				t.callbacks.OnQuit()
			}
			systray.Quit()
		}
	}
}


// SetState устанавливает состояние приложения и обновляет иконку.
func (t *Tray) SetState(state State) {
	switch state {
	case StateIdle:
		systray.SetIcon(embedded.IconIdle)
		systray.SetTooltip("Shofar - " + i18n.T("tray_ready"))
		if t.status != nil {
			t.status.SetTitle(i18n.T("tray_ready"))
		}
	case StateRecording:
		systray.SetIcon(embedded.IconRecording)
		systray.SetTooltip("Shofar - " + i18n.T("tray_recording"))
		if t.status != nil {
			t.status.SetTitle(i18n.T("tray_recording"))
		}
	case StateProcessing:
		systray.SetIcon(embedded.IconProcessing)
		systray.SetTooltip("Shofar - " + i18n.T("tray_processing"))
		if t.status != nil {
			t.status.SetTitle(i18n.T("tray_processing"))
		}
	}
}

func (t *Tray) onExit() {
	// Cleanup при выходе
}

// Quit закрывает системный трей.
func (t *Tray) Quit() {
	systray.Quit()
}

// RefreshUI обновляет все тексты меню на текущем языке.
func (t *Tray) RefreshUI() {
	systray.SetTooltip(i18n.T("app_tooltip"))

	if t.status != nil {
		t.status.SetTitle(i18n.T("tray_ready"))
	}
	if t.notifyOn != nil {
		t.notifyOn.SetTitle(i18n.T("tray_notifications"))
		t.notifyOn.SetTooltip(i18n.T("tray_notifications_hint"))
	}
	if t.settingsBtn != nil {
		t.settingsBtn.SetTitle(i18n.T("tray_settings"))
		t.settingsBtn.SetTooltip(i18n.T("tray_settings_hint"))
	}
	if t.quitBtn != nil {
		t.quitBtn.SetTitle(i18n.T("tray_quit"))
		t.quitBtn.SetTooltip(i18n.T("tray_quit_hint"))
	}
}

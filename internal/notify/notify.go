// Package notify предоставляет системные уведомления.
package notify

import (
	"github.com/gen2brain/beeep"
	"shofar/internal/i18n"
)

const appName = "Shofar"

// Notifier отправляет системные уведомления.
type Notifier struct {
	enabled bool
}

// New создаёт новый Notifier.
func New(enabled bool) *Notifier {
	return &Notifier{enabled: enabled}
}

// SetEnabled включает/выключает уведомления.
func (n *Notifier) SetEnabled(enabled bool) {
	n.enabled = enabled
}

// Recording показывает уведомление о начале записи.
func (n *Notifier) Recording() {
	n.notify(i18n.T("notify_recording"), i18n.T("notify_recording_hint"))
}

// Processing показывает уведомление об обработке.
func (n *Notifier) Processing() {
	n.notify(i18n.T("notify_processing"), i18n.T("notify_processing_hint"))
}

// Success показывает уведомление об успешном распознавании.
func (n *Notifier) Success(text string) {
	if len(text) > 100 {
		text = text[:100] + "..."
	}
	n.notify(i18n.T("notify_done"), text)
}

// Empty показывает уведомление о пустом результате.
func (n *Notifier) Empty() {
	n.notify(i18n.T("notify_empty"), i18n.T("notify_empty_hint"))
}

// Error показывает уведомление об ошибке.
func (n *Notifier) Error(msg string) {
	n.notify(i18n.T("notify_error"), msg)
}

// Info показывает информационное уведомление (для streaming).
func (n *Notifier) Info(msg string) {
	if len(msg) > 100 {
		msg = msg[:100] + "..."
	}
	n.notify("", msg)
}

func (n *Notifier) notify(title, message string) {
	if !n.enabled {
		return
	}
	// Игнорируем ошибки уведомлений - они не критичны
	if title != "" {
		_ = beeep.Notify(appName+": "+title, message, "")
	} else {
		_ = beeep.Notify(appName, message, "")
	}
}

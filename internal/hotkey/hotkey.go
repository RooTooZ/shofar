// Package hotkey предоставляет глобальные горячие клавиши.
package hotkey

import (
	"log"
	"sync"
	"time"

	"golang.design/x/hotkey"
	"golang.design/x/hotkey/mainthread"
	"whisper-input/internal/config"
)

// Handler обрабатывает события горячих клавиш.
type Handler struct {
	mu        sync.Mutex
	hk        *hotkey.Hotkey
	onPress   func()
	onRelease func()
	current   config.HotkeyConfig
	stopCh    chan struct{}
}

// New создаёт обработчик горячей клавиши.
func New(onPress, onRelease func()) *Handler {
	return &Handler{
		onPress:   onPress,
		onRelease: onRelease,
	}
}

// Register регистрирует горячую клавишу.
func (h *Handler) Register(cfg config.HotkeyConfig) error {
	log.Printf("Регистрация горячей клавиши: %s", cfg.String())

	h.mu.Lock()

	// Останавливаем предыдущий listener
	if h.stopCh != nil {
		close(h.stopCh)
		h.stopCh = nil
	}

	// Даём время listener'у завершиться
	oldHk := h.hk
	h.hk = nil
	h.mu.Unlock()

	// Небольшая задержка чтобы listener завершился
	time.Sleep(50 * time.Millisecond)

	// Отменяем предыдущую регистрацию в горутине с таймаутом
	if oldHk != nil {
		done := make(chan struct{})
		go func() {
			oldHk.Unregister()
			close(done)
		}()
		select {
		case <-done:
		case <-time.After(500 * time.Millisecond):
			log.Printf("Hotkey unregister timeout")
		}
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	// Конвертируем модификаторы
	mods := make([]hotkey.Modifier, 0, len(cfg.Modifiers))
	for _, m := range cfg.Modifiers {
		if mod, ok := modifierMap[m]; ok {
			mods = append(mods, mod)
		}
	}

	// Конвертируем клавишу
	key, ok := keyMap[cfg.Key]
	if !ok {
		key = hotkey.KeySpace // fallback
	}

	h.hk = hotkey.New(mods, key)
	h.current = cfg
	h.stopCh = make(chan struct{})

	if err := h.hk.Register(); err != nil {
		log.Printf("Ошибка регистрации: %v", err)
		h.hk = nil
		h.stopCh = nil
		return err
	}

	log.Printf("Горячая клавиша успешно зарегистрирована: %s", cfg.String())
	go h.listen(h.stopCh)
	return nil
}

func (h *Handler) listen(stopCh chan struct{}) {
	h.mu.Lock()
	hk := h.hk
	h.mu.Unlock()

	if hk == nil {
		return
	}

	var lastKeydown time.Time
	const debounceInterval = 300 * time.Millisecond // Защита от key repeat

	for {
		select {
		case <-stopCh:
			return
		case _, ok := <-hk.Keydown():
			if !ok {
				return
			}
			// Debounce: игнорируем повторные keydown от key repeat
			now := time.Now()
			if now.Sub(lastKeydown) < debounceInterval {
				continue
			}
			lastKeydown = now
			if h.onPress != nil {
				h.onPress()
			}
		case _, ok := <-hk.Keyup():
			if !ok {
				return
			}
			// В toggle режиме игнорируем keyup
		}
	}
}

// Unregister отменяет регистрацию горячей клавиши.
func (h *Handler) Unregister() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.stopCh != nil {
		close(h.stopCh)
		h.stopCh = nil
	}

	if h.hk != nil {
		err := h.hk.Unregister()
		h.hk = nil
		return err
	}
	return nil
}

// Current возвращает текущую зарегистрированную горячую клавишу.
func (h *Handler) Current() config.HotkeyConfig {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.current
}

// RunOnMainThread запускает функцию в главном потоке (требование для macOS).
func RunOnMainThread(fn func()) {
	mainthread.Init(fn)
}

// modifierMap определён в platform-specific файлах:
// - modifiers_linux.go
// - modifiers_darwin.go
// - modifiers_windows.go

// keyMap маппинг config.Key -> hotkey.Key
var keyMap = map[config.Key]hotkey.Key{
	config.KeySpace:  hotkey.KeySpace,
	config.KeyReturn: hotkey.KeyReturn,
	config.KeyTab:    hotkey.KeyTab,
	config.KeyA:      hotkey.KeyA,
	config.KeyB:      hotkey.KeyB,
	config.KeyC:      hotkey.KeyC,
	config.KeyD:      hotkey.KeyD,
	config.KeyE:      hotkey.KeyE,
	config.KeyF:      hotkey.KeyF,
	config.KeyG:      hotkey.KeyG,
	config.KeyH:      hotkey.KeyH,
	config.KeyI:      hotkey.KeyI,
	config.KeyJ:      hotkey.KeyJ,
	config.KeyK:      hotkey.KeyK,
	config.KeyL:      hotkey.KeyL,
	config.KeyM:      hotkey.KeyM,
	config.KeyN:      hotkey.KeyN,
	config.KeyO:      hotkey.KeyO,
	config.KeyP:      hotkey.KeyP,
	config.KeyQ:      hotkey.KeyQ,
	config.KeyR:      hotkey.KeyR,
	config.KeyS:      hotkey.KeyS,
	config.KeyT:      hotkey.KeyT,
	config.KeyU:      hotkey.KeyU,
	config.KeyV:      hotkey.KeyV,
	config.KeyW:      hotkey.KeyW,
	config.KeyX:      hotkey.KeyX,
	config.KeyY:      hotkey.KeyY,
	config.KeyZ:      hotkey.KeyZ,
	config.KeyF1:     hotkey.KeyF1,
	config.KeyF2:     hotkey.KeyF2,
	config.KeyF3:     hotkey.KeyF3,
	config.KeyF4:     hotkey.KeyF4,
	config.KeyF5:     hotkey.KeyF5,
	config.KeyF6:     hotkey.KeyF6,
	config.KeyF7:     hotkey.KeyF7,
	config.KeyF8:     hotkey.KeyF8,
	config.KeyF9:     hotkey.KeyF9,
	config.KeyF10:    hotkey.KeyF10,
	config.KeyF11:    hotkey.KeyF11,
	config.KeyF12:    hotkey.KeyF12,
}

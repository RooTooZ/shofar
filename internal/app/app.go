// Package app содержит основную логику приложения.
package app

import (
	"context"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"shofar/internal/audio"
	"shofar/internal/config"
	"shofar/internal/hotkey"
	"shofar/internal/i18n"
	"shofar/internal/input"
	"shofar/internal/llm"
	"shofar/internal/models"
	"shofar/internal/notify"
	"shofar/internal/settings"
	"shofar/internal/speech"
	"shofar/internal/startup"
	"shofar/internal/tray"
	"shofar/internal/waveform"
)

const (
	// MinRecordingDuration - минимальная длительность записи для распознавания
	MinRecordingDuration = 500 * time.Millisecond
)

// App представляет главное приложение.
type App struct {
	mu             sync.Mutex
	config         *config.Config
	recorder       *audio.Recorder
	modelManager   *models.Manager
	speechFactory  *speech.Factory
	llmModel       *llm.LlamaModel
	llmModelID     string // ID текущей загруженной LLM модели
	typer          input.Typer
	notifier       *notify.Notifier
	tray           *tray.Tray
	hotkey         *hotkey.Handler
	waveformWin    *waveform.Window
	settingsWin    *settings.Window
	startupWin     *startup.Window
	recordingStart time.Time
	processing     bool // защита от множественных событий
}

// New создаёт новое приложение.
func New() (*App, error) {
	cfg := config.New()

	// Инициализируем язык интерфейса из конфига
	if uiLang := cfg.UILanguage(); uiLang != "" {
		i18n.SetLanguage(i18n.Language(uiLang))
	}

	recorder, err := audio.New()
	if err != nil {
		return nil, err
	}

	typer, err := input.New()
	if err != nil {
		recorder.Close()
		return nil, err
	}

	// Создаём менеджер моделей
	modelManager, err := models.NewManager()
	if err != nil {
		recorder.Close()
		return nil, err
	}

	// Создаём фабрику распознавателей
	speechFactory := speech.NewFactory(modelManager)

	notifier := notify.New(cfg.NotificationsEnabled())

	app := &App{
		config:        cfg,
		recorder:      recorder,
		modelManager:  modelManager,
		speechFactory: speechFactory,
		typer:         typer,
		notifier:      notifier,
	}

	// Создаём окно визуализации (recorder реализует SampleProvider)
	app.waveformWin = waveform.New(recorder, waveform.DefaultConfig())

	// Callback для вставки текста (Enter или кнопка "Вставить")
	app.waveformWin.OnInsert(func(text string) {
		// Даём время на закрытие окна и переключение фокуса
		time.Sleep(150 * time.Millisecond)
		if err := app.typer.Type(text); err != nil {
			log.Printf("Ошибка ввода текста: %v", err)
			app.notifier.Error(i18n.T("error_input") + ": " + err.Error())
		} else {
			app.notifier.Success(text)
		}
		app.tray.SetState(tray.StateIdle)
	})

	// Callback для копирования в буфер обмена
	app.waveformWin.OnCopy(func(text string) {
		if err := copyToClipboard(text); err != nil {
			log.Printf("Ошибка копирования в буфер: %v", err)
			app.notifier.Error(i18n.T("error_clipboard"))
		} else {
			app.notifier.Success(text)
		}
		app.tray.SetState(tray.StateIdle)
	})

	// Callback для отмены (ESC или кнопка закрытия)
	app.waveformWin.OnCancel(func() {
		// Останавливаем запись если она идёт
		if app.recorder.IsRecording() {
			app.recorder.Stop()
		}
		app.tray.SetState(tray.StateIdle)
		app.mu.Lock()
		app.processing = false
		app.mu.Unlock()
	})

	// Создаём обработчик горячих клавиш
	app.hotkey = hotkey.New(app.onHotkeyPress, app.onHotkeyRelease)

	// Создаём окно настроек
	app.settingsWin = settings.New(modelManager, cfg)
	app.settingsWin.OnApply(func(modelID string) {
		if err := app.speechFactory.Swap(modelID); err != nil {
			log.Printf("Ошибка смены модели: %v", err)
			app.notifier.Error(i18n.T("error_model_load"))
			return
		}
		app.config.SetModelID(modelID)
		app.notifier.Info(i18n.T("success_model_loaded"))
	})
	app.settingsWin.OnHotkeyChange(func(hk config.HotkeyConfig) {
		app.config.SetHotkey(hk)
		// Перерегистрируем горячую клавишу
		if err := app.hotkey.Register(hk); err != nil {
			log.Printf("Ошибка регистрации горячей клавиши: %v", err)
			app.notifier.Error(i18n.T("error_hotkey_register"))
		}
	})
	app.settingsWin.OnLLMChange(func(enabled bool, modelID string) {
		if enabled {
			// Проверяем нужно ли загрузить новую модель или сменить текущую
			app.mu.Lock()
			needLoad := app.llmModel == nil
			needSwap := app.llmModel != nil && app.llmModelID != modelID
			app.mu.Unlock()

			if needSwap {
				// Сначала выгружаем старую модель
				app.mu.Lock()
				if app.llmModel != nil {
					app.llmModel.Close()
					app.llmModel = nil
					app.llmModelID = ""
				}
				app.mu.Unlock()
				needLoad = true
			}

			if needLoad {
				go app.loadLLMModel()
			}
		} else {
			// Выгружаем модель при отключении
			app.mu.Lock()
			if app.llmModel != nil {
				app.llmModel.Close()
				app.llmModel = nil
				app.llmModelID = ""
			}
			app.mu.Unlock()
		}
	})

	// Создаём системный трей с обработчиками
	app.tray = tray.New(tray.Callbacks{
		OnNotificationsToggle: func() bool {
			enabled := app.config.ToggleNotifications()
			app.notifier.SetEnabled(enabled)
			return enabled
		},
		OnSettingsClick: func() {
			app.settingsWin.Show()
		},
		OnQuit: func() {
			app.Close()
		},
	})

	// Callback для смены языка UI - обновляем трей
	app.settingsWin.OnUILangChange(func(lang i18n.Language) {
		app.tray.RefreshUI()
	})

	return app, nil
}

// Run запускает приложение.
func (a *App) Run() {
	a.tray.Run(func() {
		// Регистрируем горячую клавишу после инициализации трея
		hk := a.config.Hotkey()
		if err := a.hotkey.Register(hk); err != nil {
			log.Printf("Ошибка регистрации горячей клавиши: %v", err)
		}

		// Ленивая загрузка распознавателя в фоне
		go a.loadRecognizer()
	})
}

func (a *App) loadRecognizer() {
	// Определяем какую модель загружать
	modelID := a.config.ModelID()
	if modelID == "" {
		modelID = models.DefaultModelID()
	}

	info, ok := models.GetModel(modelID)
	if !ok {
		modelID = models.DefaultModelID()
		info, _ = models.GetModel(modelID)
	}

	// Проверяем скачана ли модель
	if !a.modelManager.IsDownloaded(info) {
		a.notifier.Info(i18n.T("error_model_not_downloaded"))
		return
	}

	// Показываем окно загрузки
	a.startupWin = startup.New()
	a.startupWin.SetStatus(i18n.T("startup_loading"), info.Name)
	a.startupWin.Show()

	// Загружаем модель
	if err := a.speechFactory.Load(modelID); err != nil {
		log.Printf("Ошибка загрузки модели: %v", err)
		a.startupWin.Hide()
		a.notifier.Error(i18n.T("error_model_load"))
		return
	}

	a.config.SetModelID(modelID)

	// Загружаем LLM модель если коррекция включена
	if a.config.LLMEnabled() {
		a.loadLLMModelWithStatus()
	}

	// Скрываем окно загрузки и показываем уведомление
	a.startupWin.Hide()
	a.notifier.Info(i18n.T("notify_ready"))
}

func (a *App) loadLLMModel() {
	a.loadLLMModelInternal(false)
}

func (a *App) loadLLMModelWithStatus() {
	a.loadLLMModelInternal(true)
}

func (a *App) loadLLMModelInternal(updateStatus bool) {
	modelID := a.config.LLMModelID()
	if modelID == "" {
		modelID = models.DefaultLLMModelID()
	}

	info, ok := models.GetModel(modelID)
	if !ok {
		return
	}

	if !a.modelManager.IsDownloaded(info) {
		if !updateStatus {
			a.notifier.Info(i18n.T("error_llm_not_downloaded"))
		}
		return
	}

	// Обновляем статус в окне загрузки
	if updateStatus && a.startupWin != nil {
		a.startupWin.SetStatus(i18n.T("startup_loading_llm"), info.Name)
	}

	modelPath := a.modelManager.GetModelPath(info)
	model, err := llm.NewLlamaModel(modelPath, 2048)
	if err != nil {
		log.Printf("Ошибка загрузки LLM модели: %v", err)
		if !updateStatus {
			a.notifier.Error(i18n.T("error_llm_load"))
		}
		return
	}

	a.mu.Lock()
	// Закрываем старую модель если была
	if a.llmModel != nil {
		a.llmModel.Close()
	}
	a.llmModel = model
	a.llmModelID = modelID
	a.mu.Unlock()
}

func (a *App) onHotkeyPress() {
	a.mu.Lock()

	// Toggle режим: если идёт запись - останавливаем
	if a.recorder.IsRecording() {
		a.mu.Unlock()
		a.stopRecording()
		return
	}

	if a.processing {
		a.mu.Unlock()
		return
	}

	// Проверяем что модель загружена
	if !a.speechFactory.IsLoaded() {
		a.mu.Unlock()
		a.notifier.Error(i18n.T("error_model_loading"))
		return
	}
	a.recordingStart = time.Now()
	a.tray.SetState(tray.StateRecording)
	a.notifier.Recording()

	// Очищаем предыдущий результат
	a.waveformWin.ClearResult()

	if err := a.recorder.Start(); err != nil {
		log.Printf("Ошибка начала записи: %v", err)
		a.notifier.Error(i18n.T("error_recording") + ": " + err.Error())
		a.tray.SetState(tray.StateIdle)
		a.mu.Unlock()
		return
	}

	// Показываем окно визуализации
	a.waveformWin.SetStartTime(a.recordingStart)
	a.waveformWin.Show()

	a.mu.Unlock()
}

func (a *App) onHotkeyRelease() {
	// В toggle режиме игнорируем keyup события
}

func (a *App) stopRecording() {
	a.mu.Lock()

	if !a.recorder.IsRecording() || a.processing {
		a.mu.Unlock()
		return
	}

	a.processing = true
	elapsed := time.Since(a.recordingStart)
	recognizer := a.speechFactory.Current()
	a.mu.Unlock()

	// Переключаем окно в режим распознавания речи
	a.waveformWin.SetState(waveform.StateSpeechProcess)

	// Теперь безопасно останавливаем запись
	samples := a.recorder.Stop()

	// Проверяем минимальную длительность записи
	if elapsed < MinRecordingDuration {
		a.waveformWin.Hide()
		a.tray.SetState(tray.StateIdle)
		a.mu.Lock()
		a.processing = false
		a.mu.Unlock()
		return
	}

	a.tray.SetState(tray.StateProcessing)
	a.notifier.Processing()

	if recognizer == nil {
		a.notifier.Error(i18n.T("error_model_not_loaded"))
		a.waveformWin.Hide()
		a.tray.SetState(tray.StateIdle)
		a.mu.Lock()
		a.processing = false
		a.mu.Unlock()
		return
	}

	if len(samples) == 0 {
		a.notifier.Empty()
		a.waveformWin.Hide()
		a.tray.SetState(tray.StateIdle)
		a.mu.Lock()
		a.processing = false
		a.mu.Unlock()
		return
	}

	// Распознаём в отдельной горутине
	go func() {
		defer func() {
			a.mu.Lock()
			a.processing = false
			a.mu.Unlock()
		}()

		lang := a.config.Language()
		originalText, err := recognizer.Transcribe(samples, lang)

		if err != nil {
			a.notifier.Error(i18n.T("error_recognition"))
			a.waveformWin.Hide()
			a.tray.SetState(tray.StateIdle)
			return
		}

		if originalText == "" {
			a.notifier.Empty()
			a.waveformWin.Hide()
			a.tray.SetState(tray.StateIdle)
			return
		}

		correctedText := ""

		// Коррекция текста через LLM (если включена и модель загружена)
		if a.config.LLMEnabled() && a.llmModel != nil {
			// Переключаем окно в режим LLM обработки
			a.waveformWin.SetState(waveform.StateLLMProcess)

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			corrected, err := a.llmModel.CorrectText(ctx, originalText)
			cancel()
			if err == nil && corrected != "" {
				correctedText = corrected
			}
		}

		a.waveformWin.SetResult(originalText, correctedText)
		a.tray.SetState(tray.StateIdle)
		// Окно остаётся открытым - пользователь закроет его сам или нажмёт копировать
	}()
}

// Close освобождает ресурсы приложения.
func (a *App) Close() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.hotkey != nil {
		a.hotkey.Unregister()
	}

	if a.recorder != nil {
		a.recorder.Close()
	}

	if a.speechFactory != nil {
		a.speechFactory.Close()
	}

	if a.llmModel != nil {
		a.llmModel.Close()
		a.llmModel = nil
		a.llmModelID = ""
	}

	if a.settingsWin != nil {
		a.settingsWin.Hide()
	}
}

// copyToClipboard copies text to system clipboard.
func copyToClipboard(text string) error {
	// Detect Wayland vs X11
	if os.Getenv("WAYLAND_DISPLAY") != "" {
		// Wayland: use wl-copy
		cmd := exec.Command("wl-copy")
		cmd.Stdin = strings.NewReader(text)
		return cmd.Run()
	}

	// X11: use xclip
	cmd := exec.Command("xclip", "-selection", "clipboard")
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

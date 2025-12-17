// Package i18n provides internationalization support.
package i18n

import "sync"

// Language represents a UI language.
type Language string

const (
	RU Language = "ru"
	EN Language = "en"
)

var (
	mu      sync.RWMutex
	current = RU // Default language
)

// Translations for all supported languages.
var translations = map[Language]map[string]string{
	RU: {
		// App
		"app_name":    "Shofar",
		"app_tooltip": "Shofar - голосовой ввод",

		// Tray menu
		"tray_ready":              "Готов к работе",
		"tray_recording":          "Запись...",
		"tray_processing":         "Распознавание...",
		"tray_language":           "Язык",
		"tray_lang_select":        "Выбор языка распознавания",
		"tray_lang_ru":            "Русский",
		"tray_lang_ru_hint":       "Распознавание на русском (рекомендуется для смешанной речи)",
		"tray_lang_en":            "English",
		"tray_lang_en_hint":       "Распознавание на английском",
		"tray_lang_auto":          "Авто",
		"tray_lang_auto_hint":     "Автоопределение (не рекомендуется для смешанной речи)",
		"tray_notifications":      "Уведомления",
		"tray_notifications_hint": "Показывать уведомления",
		"tray_settings":           "Настройки...",
		"tray_settings_hint":      "Горячая клавиша, движок, модель",
		"tray_quit":               "Выход",
		"tray_quit_hint":          "Закрыть приложение",

		// Notifications
		"notify_recording":       "Запись...",
		"notify_recording_hint":  "Говорите в микрофон",
		"notify_processing":      "Распознаю...",
		"notify_processing_hint": "Пожалуйста, подождите",
		"notify_done":            "Готово",
		"notify_empty":           "Не удалось распознать",
		"notify_empty_hint":      "Попробуйте ещё раз",
		"notify_error":           "Ошибка",
		"notify_ready":           "Shofar готов к работе",

		// Waveform window
		"waveform_recording":         "Запись",
		"waveform_speech_processing": "Распознавание речи...",
		"waveform_speech_hint":       "Преобразование аудио в текст",
		"waveform_llm_processing":    "Коррекция текста...",
		"waveform_llm_hint":          "LLM обрабатывает результат",
		"waveform_result":            "Результат",
		"waveform_original":          "Исходный",
		"waveform_corrected":         "Исправлено",
		"waveform_insert":            "Вставить",
		"waveform_copy":              "Скопировать",

		// Startup window
		"startup_loading":     "Загрузка модели распознавания...",
		"startup_loading_llm": "Загрузка LLM модели...",
		"startup_status":      "Запуск...",

		// Settings window
		"settings_title":          "Настройки",
		"settings_hotkey":         "Горячая клавиша",
		"settings_hotkey_edit":    "Изменить",
		"settings_hotkey_cancel":  "Отмена",
		"settings_hotkey_not_set": "Не задана",
		"settings_hotkey_prompt":  "Нажмите комбинацию...",
		"settings_llm":            "Коррекция текста (LLM)",
		"settings_llm_enable":     "Исправлять ошибки распознавания",
		"settings_llm_hint":       "Встроенная модель для коррекции текста",
		"settings_recognition":    "Распознавание",
		"settings_engine":         "Движок:",
		"settings_apply":          "Применить",
		"settings_cancel":         "Отмена",
		"settings_downloading":    "Загрузка",
		"settings_loading_model":  "Загрузка модели",
		"settings_loading_hint":   "Это может занять некоторое время",
		"settings_ui_language":    "Язык интерфейса",
		"settings_key":            "Клавиша:",

		// Errors
		"error_model_loading":        "Модель ещё загружается...",
		"error_model_not_loaded":     "Модель ещё не загружена",
		"error_model_not_downloaded": "Модель не скачана. Откройте настройки для загрузки.",
		"error_llm_not_downloaded":   "LLM модель не скачана. Скачайте в настройках.",
		"error_recording":            "Ошибка записи",
		"error_recognition":          "Ошибка распознавания",
		"error_input":                "Ошибка ввода",
		"error_hotkey_register":      "Не удалось зарегистрировать горячую клавишу",
		"error_model_load":           "Не удалось загрузить модель",
		"error_llm_load":             "Не удалось загрузить LLM модель",
		"error_clipboard":            "Ошибка копирования в буфер обмена",

		// Success messages
		"success_model_loaded": "Модель загружена",
	},

	EN: {
		// App
		"app_name":    "Shofar",
		"app_tooltip": "Shofar - voice input",

		// Tray menu
		"tray_ready":              "Ready",
		"tray_recording":          "Recording...",
		"tray_processing":         "Processing...",
		"tray_language":           "Language",
		"tray_lang_select":        "Select recognition language",
		"tray_lang_ru":            "Русский",
		"tray_lang_ru_hint":       "Russian recognition (recommended for mixed speech)",
		"tray_lang_en":            "English",
		"tray_lang_en_hint":       "English recognition",
		"tray_lang_auto":          "Auto",
		"tray_lang_auto_hint":     "Auto-detect (not recommended for mixed speech)",
		"tray_notifications":      "Notifications",
		"tray_notifications_hint": "Show notifications",
		"tray_settings":           "Settings...",
		"tray_settings_hint":      "Hotkey, engine, model",
		"tray_quit":               "Quit",
		"tray_quit_hint":          "Close application",

		// Notifications
		"notify_recording":       "Recording...",
		"notify_recording_hint":  "Speak into the microphone",
		"notify_processing":      "Processing...",
		"notify_processing_hint": "Please wait",
		"notify_done":            "Done",
		"notify_empty":           "Could not recognize",
		"notify_empty_hint":      "Please try again",
		"notify_error":           "Error",
		"notify_ready":           "Shofar is ready",

		// Waveform window
		"waveform_recording":         "Recording",
		"waveform_speech_processing": "Speech recognition...",
		"waveform_speech_hint":       "Converting audio to text",
		"waveform_llm_processing":    "Text correction...",
		"waveform_llm_hint":          "LLM processing result",
		"waveform_result":            "Result",
		"waveform_original":          "Original",
		"waveform_corrected":         "Corrected",
		"waveform_insert":            "Insert",
		"waveform_copy":              "Copy",

		// Startup window
		"startup_loading":     "Loading recognition model...",
		"startup_loading_llm": "Loading LLM model...",
		"startup_status":      "Starting...",

		// Settings window
		"settings_title":          "Settings",
		"settings_hotkey":         "Hotkey",
		"settings_hotkey_edit":    "Edit",
		"settings_hotkey_cancel":  "Cancel",
		"settings_hotkey_not_set": "Not set",
		"settings_hotkey_prompt":  "Press key combination...",
		"settings_llm":            "Text correction (LLM)",
		"settings_llm_enable":     "Fix recognition errors",
		"settings_llm_hint":       "Built-in model for text correction",
		"settings_recognition":    "Recognition",
		"settings_engine":         "Engine:",
		"settings_apply":          "Apply",
		"settings_cancel":         "Cancel",
		"settings_downloading":    "Downloading",
		"settings_loading_model":  "Loading model",
		"settings_loading_hint":   "This may take a while",
		"settings_ui_language":    "Interface language",
		"settings_key":            "Key:",

		// Errors
		"error_model_loading":        "Model is still loading...",
		"error_model_not_loaded":     "Model not loaded yet",
		"error_model_not_downloaded": "Model not downloaded. Open settings to download.",
		"error_llm_not_downloaded":   "LLM model not downloaded. Download in settings.",
		"error_recording":            "Recording error",
		"error_recognition":          "Recognition error",
		"error_input":                "Input error",
		"error_hotkey_register":      "Could not register hotkey",
		"error_model_load":           "Could not load model",
		"error_llm_load":             "Could not load LLM model",
		"error_clipboard":            "Clipboard copy error",

		// Success messages
		"success_model_loaded": "Model loaded",
	},
}

// T returns the translation for the given key.
func T(key string) string {
	mu.RLock()
	defer mu.RUnlock()

	if strings, ok := translations[current]; ok {
		if s, ok := strings[key]; ok {
			return s
		}
	}
	// Fallback to key itself
	return key
}

// SetLanguage sets the current UI language.
func SetLanguage(lang Language) {
	mu.Lock()
	defer mu.Unlock()
	current = lang
}

// GetLanguage returns the current UI language.
func GetLanguage() Language {
	mu.RLock()
	defer mu.RUnlock()
	return current
}

// AvailableLanguages returns list of supported languages.
func AvailableLanguages() []Language {
	return []Language{RU, EN}
}

// LanguageName returns display name for a language.
func LanguageName(lang Language) string {
	switch lang {
	case RU:
		return "Русский"
	case EN:
		return "English"
	default:
		return string(lang)
	}
}

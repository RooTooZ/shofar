// Package speech предоставляет абстракцию для движков распознавания речи.
package speech

// Engine тип движка распознавания.
type Engine string

const (
	// EngineWhisper - whisper.cpp движок.
	EngineWhisper Engine = "whisper"
	// EngineVosk - Vosk движок.
	EngineVosk Engine = "vosk"
)

// Recognizer - интерфейс для движков распознавания речи.
type Recognizer interface {
	// Transcribe распознаёт речь из аудио сэмплов.
	// samples - аудио данные в формате float32, 16kHz, mono.
	// lang - язык распознавания ("ru", "en", "auto" для автоопределения).
	// Возвращает распознанный текст или ошибку.
	Transcribe(samples []float32, lang string) (string, error)

	// Close освобождает ресурсы движка.
	Close()

	// Name возвращает название движка (для логирования).
	Name() string
}

// Config содержит общие настройки для создания распознавателя.
type Config struct {
	// Engine - тип движка (whisper, vosk).
	Engine Engine

	// ModelPath - путь к модели (для Vosk или внешней Whisper модели).
	ModelPath string

	// Language - язык по умолчанию.
	Language string
}

// Package models управляет моделями распознавания речи.
package models

// Engine тип движка распознавания.
type Engine string

const (
	EngineWhisper Engine = "whisper"
	EngineVosk    Engine = "vosk"
	EngineLLM     Engine = "llm"
)

// ModelInfo информация о модели.
type ModelInfo struct {
	ID       string // Уникальный идентификатор: "whisper-tiny-q5"
	Engine   Engine // Движок: whisper или vosk
	Name     string // Отображаемое имя: "Tiny Q5 (32MB)"
	Filename string // Имя файла/директории: "ggml-tiny-q5_1.bin"
	URL      string // URL для скачивания
	Size     int64  // Размер в байтах (для прогресса)
	IsZip    bool   // Нужно ли распаковывать
}

// Registry все доступные модели.
var Registry = []ModelInfo{
	// Whisper - квантизированные модели (рекомендуется для CPU)
	{
		ID:       "whisper-tiny-q5",
		Engine:   EngineWhisper,
		Name:     "Tiny Q5",
		Filename: "ggml-tiny-q5_1.bin",
		URL:      "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-tiny-q5_1.bin",
		Size:     32 * 1024 * 1024,
		IsZip:    false,
	},
	{
		ID:       "whisper-base-q5",
		Engine:   EngineWhisper,
		Name:     "Base Q5",
		Filename: "ggml-base-q5_1.bin",
		URL:      "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-base-q5_1.bin",
		Size:     60 * 1024 * 1024,
		IsZip:    false,
	},
	{
		ID:       "whisper-small-q5",
		Engine:   EngineWhisper,
		Name:     "Small Q5",
		Filename: "ggml-small-q5_1.bin",
		URL:      "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-small-q5_1.bin",
		Size:     190 * 1024 * 1024,
		IsZip:    false,
	},
	{
		ID:       "whisper-turbo",
		Engine:   EngineWhisper,
		Name:     "Large v3 Turbo",
		Filename: "ggml-large-v3-turbo-q5_0.bin",
		URL:      "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-large-v3-turbo-q5_0.bin",
		Size:     574 * 1024 * 1024,
		IsZip:    false,
	},
	// Whisper - оригинальные модели (больше размер, чуть лучше качество)
	{
		ID:       "whisper-tiny",
		Engine:   EngineWhisper,
		Name:     "Tiny",
		Filename: "ggml-tiny.bin",
		URL:      "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-tiny.bin",
		Size:     75 * 1024 * 1024,
		IsZip:    false,
	},
	{
		ID:       "whisper-base",
		Engine:   EngineWhisper,
		Name:     "Base",
		Filename: "ggml-base.bin",
		URL:      "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-base.bin",
		Size:     142 * 1024 * 1024,
		IsZip:    false,
	},
	{
		ID:       "whisper-small",
		Engine:   EngineWhisper,
		Name:     "Small",
		Filename: "ggml-small.bin",
		URL:      "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-small.bin",
		Size:     466 * 1024 * 1024,
		IsZip:    false,
	},
	// Vosk
	{
		ID:       "vosk-ru-small",
		Engine:   EngineVosk,
		Name:     "Russian Small",
		Filename: "vosk-model-small-ru-0.22",
		URL:      "https://alphacephei.com/vosk/models/vosk-model-small-ru-0.22.zip",
		Size:     45 * 1024 * 1024,
		IsZip:    true,
	},
	{
		ID:       "vosk-ru",
		Engine:   EngineVosk,
		Name:     "Russian Large",
		Filename: "vosk-model-ru-0.42",
		URL:      "https://alphacephei.com/vosk/models/vosk-model-ru-0.42.zip",
		Size:     1800 * 1024 * 1024,
		IsZip:    true,
	},
	// LLM для коррекции текста
	{
		ID:       "llm-qwen2.5-0.5b",
		Engine:   EngineLLM,
		Name:     "Qwen2.5 0.5B",
		Filename: "qwen2.5-0.5b-instruct-q4_k_m.gguf",
		URL:      "https://huggingface.co/Qwen/Qwen2.5-0.5B-Instruct-GGUF/resolve/main/qwen2.5-0.5b-instruct-q4_k_m.gguf",
		Size:     386 * 1024 * 1024,
		IsZip:    false,
	},
	{
		ID:       "llm-qwen2.5-1.5b",
		Engine:   EngineLLM,
		Name:     "Qwen2.5 1.5B",
		Filename: "qwen2.5-1.5b-instruct-q4_k_m.gguf",
		URL:      "https://huggingface.co/Qwen/Qwen2.5-1.5B-Instruct-GGUF/resolve/main/qwen2.5-1.5b-instruct-q4_k_m.gguf",
		Size:     987 * 1024 * 1024,
		IsZip:    false,
	},
	{
		ID:       "llm-qwen2.5-3b",
		Engine:   EngineLLM,
		Name:     "Qwen2.5 3B",
		Filename: "qwen2.5-3b-instruct-q4_k_m.gguf",
		URL:      "https://huggingface.co/Qwen/Qwen2.5-3B-Instruct-GGUF/resolve/main/qwen2.5-3b-instruct-q4_k_m.gguf",
		Size:     1900 * 1024 * 1024,
		IsZip:    false,
	},
}

// DefaultModelID модель по умолчанию.
func DefaultModelID() string {
	return "whisper-tiny-q5"
}

// GetModel возвращает модель по ID.
func GetModel(id string) (ModelInfo, bool) {
	for _, m := range Registry {
		if m.ID == id {
			return m, true
		}
	}
	return ModelInfo{}, false
}

// GetModelsByEngine возвращает модели для указанного движка.
func GetModelsByEngine(engine Engine) []ModelInfo {
	var result []ModelInfo
	for _, m := range Registry {
		if m.Engine == engine {
			result = append(result, m)
		}
	}
	return result
}

// AllEngines возвращает все доступные движки (без LLM - он отдельно).
func AllEngines() []Engine {
	return []Engine{EngineWhisper, EngineVosk}
}

// EngineName возвращает отображаемое имя движка.
func EngineName(e Engine) string {
	switch e {
	case EngineWhisper:
		return "Whisper"
	case EngineVosk:
		return "Vosk"
	case EngineLLM:
		return "LLM"
	default:
		return string(e)
	}
}

// DefaultLLMModelID модель LLM по умолчанию.
func DefaultLLMModelID() string {
	return "llm-qwen2.5-0.5b"
}

// GetLLMModels возвращает все LLM модели.
func GetLLMModels() []ModelInfo {
	return GetModelsByEngine(EngineLLM)
}

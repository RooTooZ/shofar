package speech

import (
	"io"
	"strings"
	"sync"

	whisper "github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
)

// WhisperRecognizer реализует Recognizer через whisper.cpp.
type WhisperRecognizer struct {
	mu    sync.Mutex
	model whisper.Model
}

// NewWhisperFromFile создаёт WhisperRecognizer из файла модели.
func NewWhisperFromFile(modelPath string) (*WhisperRecognizer, error) {
	model, err := whisper.New(modelPath)
	if err != nil {
		return nil, err
	}

	return &WhisperRecognizer{
		model: model,
	}, nil
}

// Name возвращает название движка.
func (w *WhisperRecognizer) Name() string {
	return "whisper"
}

// Transcribe распознаёт речь из аудио сэмплов.
func (w *WhisperRecognizer) Transcribe(samples []float32, lang string) (string, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	ctx, err := w.model.NewContext()
	if err != nil {
		return "", err
	}

	// Отключаем перевод - только транскрипция
	ctx.SetTranslate(false)

	// Устанавливаем язык (для "auto" включится автодетект)
	if lang != "" {
		ctx.SetLanguage(lang)
	}

	// Обрабатываем аудио
	if err := ctx.Process(samples, nil, nil, nil); err != nil {
		return "", err
	}

	// Собираем результат из сегментов
	var result strings.Builder
	for {
		segment, err := ctx.NextSegment()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		result.WriteString(segment.Text)
	}

	return strings.TrimSpace(result.String()), nil
}

// Close освобождает ресурсы.
func (w *WhisperRecognizer) Close() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.model != nil {
		w.model.Close()
		w.model = nil
	}
}

package speech

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sync"

	vosk "github.com/alphacep/vosk-api/go"
)

// VoskRecognizer реализует Recognizer через Vosk.
type VoskRecognizer struct {
	mu         sync.Mutex
	model      *vosk.VoskModel
	recognizer *vosk.VoskRecognizer
	sampleRate float64
}

// voskResult структура для парсинга JSON результата от Vosk.
type voskResult struct {
	Text string `json:"text"`
}

// NewVosk создаёт VoskRecognizer из пути к модели.
func NewVosk(modelPath string) (*VoskRecognizer, error) {
	// Проверяем существование директории модели
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("модель Vosk не найдена: %s", modelPath)
	}

	model, err := vosk.NewModel(modelPath)
	if err != nil {
		return nil, fmt.Errorf("ошибка загрузки модели Vosk: %w", err)
	}

	// 16000 Hz - стандартная частота для speech recognition
	sampleRate := 16000.0
	rec, err := vosk.NewRecognizer(model, sampleRate)
	if err != nil {
		model.Free()
		return nil, err
	}

	return &VoskRecognizer{
		model:      model,
		recognizer: rec,
		sampleRate: sampleRate,
	}, nil
}

// Name возвращает название движка.
func (v *VoskRecognizer) Name() string {
	return "vosk"
}

// Transcribe распознаёт речь из аудио сэмплов.
// Vosk принимает PCM16 данные, поэтому конвертируем float32 -> int16.
func (v *VoskRecognizer) Transcribe(samples []float32, lang string) (string, error) {
	v.mu.Lock()
	defer v.mu.Unlock()

	// Конвертируем float32 [-1, 1] в int16 [-32768, 32767]
	pcm16 := make([]byte, len(samples)*2)
	for i, sample := range samples {
		if sample > 1.0 {
			sample = 1.0
		} else if sample < -1.0 {
			sample = -1.0
		}
		val := int16(sample * math.MaxInt16)
		binary.LittleEndian.PutUint16(pcm16[i*2:], uint16(val))
	}

	// Обрабатываем аудио
	v.recognizer.AcceptWaveform(pcm16)

	// Получаем финальный результат
	resultJSON := v.recognizer.FinalResult()

	// Сбрасываем распознаватель для следующего использования
	v.recognizer.Reset()

	// Парсим JSON результат
	var result voskResult
	if err := json.Unmarshal([]byte(resultJSON), &result); err != nil {
		return "", err
	}

	return result.Text, nil
}

// Close освобождает ресурсы.
func (v *VoskRecognizer) Close() {
	v.mu.Lock()
	defer v.mu.Unlock()

	if v.recognizer != nil {
		v.recognizer.Free()
		v.recognizer = nil
	}

	if v.model != nil {
		v.model.Free()
		v.model = nil
	}
}

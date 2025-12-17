package speech

import (
	"fmt"
	"sync"

	"whisper-input/internal/models"
)

// Factory управляет созданием и переключением распознавателей.
type Factory struct {
	manager *models.Manager
	current Recognizer
	modelID string
	mu      sync.RWMutex
}

// NewFactory создаёт фабрику распознавателей.
func NewFactory(manager *models.Manager) *Factory {
	return &Factory{
		manager: manager,
	}
}

// Create создаёт распознаватель для указанной модели.
func (f *Factory) Create(modelID string) (Recognizer, error) {
	info, ok := models.GetModel(modelID)
	if !ok {
		return nil, fmt.Errorf("модель не найдена: %s", modelID)
	}

	modelPath := f.manager.GetModelPath(info)

	// Проверяем что модель скачана
	if !f.manager.IsDownloaded(info) {
		return nil, fmt.Errorf("модель не скачана: %s", info.Name)
	}

	var rec Recognizer
	var err error

	switch info.Engine {
	case models.EngineWhisper:
		rec, err = NewWhisperFromFile(modelPath)
	case models.EngineVosk:
		rec, err = NewVosk(modelPath)
	default:
		return nil, fmt.Errorf("неизвестный движок: %s", info.Engine)
	}

	if err != nil {
		return nil, fmt.Errorf("ошибка создания распознавателя: %w", err)
	}

	return rec, nil
}

// Load загружает модель и устанавливает её как текущую.
func (f *Factory) Load(modelID string) error {
	rec, err := f.Create(modelID)
	if err != nil {
		return err
	}

	f.mu.Lock()
	old := f.current
	f.current = rec
	f.modelID = modelID
	f.mu.Unlock()

	// Закрываем старый распознаватель
	if old != nil {
		old.Close()
	}

	return nil
}

// Swap атомарно меняет текущий распознаватель на новый (hot-swap).
func (f *Factory) Swap(modelID string) error {
	// Создаём новый распознаватель
	rec, err := f.Create(modelID)
	if err != nil {
		return err
	}

	f.mu.Lock()
	old := f.current
	f.current = rec
	f.modelID = modelID
	f.mu.Unlock()

	// Закрываем старый распознаватель в фоне
	if old != nil {
		go old.Close()
	}

	return nil
}

// Current возвращает текущий распознаватель (thread-safe).
func (f *Factory) Current() Recognizer {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.current
}

// CurrentModelID возвращает ID текущей модели.
func (f *Factory) CurrentModelID() string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.modelID
}

// IsLoaded проверяет, загружена ли модель.
func (f *Factory) IsLoaded() bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.current != nil
}

// Close закрывает текущий распознаватель.
func (f *Factory) Close() {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.current != nil {
		f.current.Close()
		f.current = nil
	}
}

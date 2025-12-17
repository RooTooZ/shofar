package models

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

// Progress информация о прогрессе загрузки.
type Progress struct {
	ModelID    string
	Downloaded int64
	Total      int64
	Done       bool
	Error      error
}

// Manager управляет моделями.
type Manager struct {
	modelsDir string
	mu        sync.RWMutex
}

// NewManager создаёт менеджер моделей.
// Модели хранятся в директории models/ рядом с бинарником.
func NewManager() (*Manager, error) {
	execPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("не удалось определить путь к бинарнику: %w", err)
	}

	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return nil, fmt.Errorf("не удалось разрешить симлинки: %w", err)
	}

	modelsDir := filepath.Join(filepath.Dir(execPath), "models")

	// Создаём директории для моделей
	whisperDir := filepath.Join(modelsDir, "whisper")
	voskDir := filepath.Join(modelsDir, "vosk")
	llmDir := filepath.Join(modelsDir, "llm")

	if err := os.MkdirAll(whisperDir, 0755); err != nil {
		return nil, fmt.Errorf("не удалось создать директорию whisper: %w", err)
	}
	if err := os.MkdirAll(voskDir, 0755); err != nil {
		return nil, fmt.Errorf("не удалось создать директорию vosk: %w", err)
	}
	if err := os.MkdirAll(llmDir, 0755); err != nil {
		return nil, fmt.Errorf("не удалось создать директорию llm: %w", err)
	}

	return &Manager{modelsDir: modelsDir}, nil
}

// ModelsDir возвращает путь к директории моделей.
func (m *Manager) ModelsDir() string {
	return m.modelsDir
}

// GetModelPath возвращает полный путь к модели.
func (m *Manager) GetModelPath(info ModelInfo) string {
	switch info.Engine {
	case EngineWhisper:
		return filepath.Join(m.modelsDir, "whisper", info.Filename)
	case EngineVosk:
		return filepath.Join(m.modelsDir, "vosk", info.Filename)
	case EngineLLM:
		return filepath.Join(m.modelsDir, "llm", info.Filename)
	default:
		return filepath.Join(m.modelsDir, info.Filename)
	}
}

// IsDownloaded проверяет, скачана ли модель.
func (m *Manager) IsDownloaded(info ModelInfo) bool {
	path := m.GetModelPath(info)
	stat, err := os.Stat(path)
	if err != nil {
		return false
	}

	// Для Vosk проверяем что это директория
	if info.IsZip {
		return stat.IsDir()
	}

	// Для Whisper проверяем что файл не пустой
	return stat.Size() > 0
}

// ListDownloaded возвращает список скачанных моделей.
func (m *Manager) ListDownloaded() []ModelInfo {
	var downloaded []ModelInfo
	for _, model := range Registry {
		if m.IsDownloaded(model) {
			downloaded = append(downloaded, model)
		}
	}
	return downloaded
}

// Download скачивает модель.
// progress канал получает обновления о прогрессе (можно nil).
func (m *Manager) Download(ctx context.Context, info ModelInfo, progress chan<- Progress) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.IsDownloaded(info) {
		if progress != nil {
			progress <- Progress{ModelID: info.ID, Downloaded: info.Size, Total: info.Size, Done: true}
		}
		return nil
	}

	if info.IsZip {
		return m.downloadAndUnzip(ctx, info, progress)
	}
	return m.downloadFile(ctx, info, progress)
}

func (m *Manager) downloadFile(ctx context.Context, info ModelInfo, progress chan<- Progress) error {
	destPath := m.GetModelPath(info)

	// Создаём временный файл
	tmpPath := destPath + ".tmp"
	defer os.Remove(tmpPath)

	req, err := http.NewRequestWithContext(ctx, "GET", info.URL, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("ошибка скачивания: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP ошибка: %s", resp.Status)
	}

	total := resp.ContentLength
	if total <= 0 {
		total = info.Size
	}

	file, err := os.Create(tmpPath)
	if err != nil {
		return err
	}
	defer file.Close()

	var downloaded int64
	buf := make([]byte, 32*1024)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		n, err := resp.Body.Read(buf)
		if n > 0 {
			if _, werr := file.Write(buf[:n]); werr != nil {
				return werr
			}
			downloaded += int64(n)

			if progress != nil {
				select {
				case progress <- Progress{ModelID: info.ID, Downloaded: downloaded, Total: total}:
				default:
				}
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}

	file.Close()

	// Переименовываем в финальное имя
	if err := os.Rename(tmpPath, destPath); err != nil {
		return err
	}

	if progress != nil {
		progress <- Progress{ModelID: info.ID, Downloaded: total, Total: total, Done: true}
	}

	return nil
}

func (m *Manager) downloadAndUnzip(ctx context.Context, info ModelInfo, progress chan<- Progress) error {
	destDir := m.GetModelPath(info)

	// Скачиваем во временный файл
	tmpZip, err := os.CreateTemp("", "model-*.zip")
	if err != nil {
		return err
	}
	tmpPath := tmpZip.Name()
	defer os.Remove(tmpPath)

	req, err := http.NewRequestWithContext(ctx, "GET", info.URL, nil)
	if err != nil {
		tmpZip.Close()
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		tmpZip.Close()
		return fmt.Errorf("ошибка скачивания: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		tmpZip.Close()
		return fmt.Errorf("HTTP ошибка: %s", resp.Status)
	}

	total := resp.ContentLength
	if total <= 0 {
		total = info.Size
	}

	var downloaded int64
	buf := make([]byte, 32*1024)

	for {
		select {
		case <-ctx.Done():
			tmpZip.Close()
			return ctx.Err()
		default:
		}

		n, err := resp.Body.Read(buf)
		if n > 0 {
			if _, werr := tmpZip.Write(buf[:n]); werr != nil {
				tmpZip.Close()
				return werr
			}
			downloaded += int64(n)

			if progress != nil {
				select {
				case progress <- Progress{ModelID: info.ID, Downloaded: downloaded, Total: total}:
				default:
				}
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			tmpZip.Close()
			return err
		}
	}

	tmpZip.Close()

	// Распаковываем
	parentDir := filepath.Dir(destDir)
	if err := unzip(tmpPath, parentDir); err != nil {
		return fmt.Errorf("ошибка распаковки: %w", err)
	}

	if progress != nil {
		progress <- Progress{ModelID: info.ID, Downloaded: total, Total: total, Done: true}
	}

	return nil
}

func unzip(src, destDir string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(destDir, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, 0755)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), 0755); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}

	return nil
}

// Delete удаляет модель.
func (m *Manager) Delete(info ModelInfo) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	path := m.GetModelPath(info)
	return os.RemoveAll(path)
}

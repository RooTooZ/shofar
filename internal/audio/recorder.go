// Package audio предоставляет запись аудио с микрофона.
package audio

import (
	"sync"
	"time"

	"github.com/gordonklaus/portaudio"
)

const (
	// SampleRate - частота дискретизации (требование Whisper).
	SampleRate = 16000
	// Channels - количество каналов (mono).
	Channels = 1
	// FramesPerBuffer - размер буфера.
	FramesPerBuffer = 1024
	// MinSamples - минимальное количество сэмплов (200ms при 16kHz).
	// Whisper требует минимум 100ms, добавляем запас.
	MinSamples = SampleRate / 5 // 3200 samples = 200ms
)

// Recorder записывает аудио с микрофона.
type Recorder struct {
	mu       sync.Mutex
	stream   *portaudio.Stream
	buffer   []float32
	samples  []float32
	running  bool
	done     chan struct{}
}

// New создаёт новый Recorder.
func New() (*Recorder, error) {
	if err := portaudio.Initialize(); err != nil {
		return nil, err
	}

	r := &Recorder{
		buffer: make([]float32, FramesPerBuffer),
	}

	return r, nil
}

// Start начинает запись аудио.
func (r *Recorder) Start() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.running {
		return nil
	}

	r.samples = make([]float32, 0, SampleRate*30) // Буфер на 30 сек
	r.done = make(chan struct{})

	stream, err := portaudio.OpenDefaultStream(
		Channels,        // input channels
		0,               // output channels
		SampleRate,      // sample rate
		FramesPerBuffer, // frames per buffer
		r.buffer,        // buffer
	)
	if err != nil {
		return err
	}

	r.stream = stream
	r.running = true

	if err := stream.Start(); err != nil {
		r.stream.Close()
		r.running = false
		return err
	}

	go r.recordLoop()

	return nil
}

func (r *Recorder) recordLoop() {
	defer func() {
		close(r.done)
	}()

	for {
		r.mu.Lock()
		if !r.running {
			r.mu.Unlock()
			return
		}
		stream := r.stream
		r.mu.Unlock()

		if stream == nil {
			return
		}

		// Проверяем доступность данных с таймаутом
		available, err := stream.AvailableToRead()
		if err != nil {
			r.mu.Lock()
			running := r.running
			r.mu.Unlock()
			if !running {
				return
			}
			time.Sleep(10 * time.Millisecond)
			continue
		}

		if available == 0 {
			// Нет данных - проверяем running и ждём
			r.mu.Lock()
			running := r.running
			r.mu.Unlock()
			if !running {
				return
			}
			time.Sleep(10 * time.Millisecond)
			continue
		}

		if err := stream.Read(); err != nil {
			r.mu.Lock()
			running := r.running
			r.mu.Unlock()
			if !running {
				return
			}
			time.Sleep(10 * time.Millisecond)
			continue
		}

		r.mu.Lock()
		if r.running {
			bufCopy := make([]float32, len(r.buffer))
			copy(bufCopy, r.buffer)
			r.samples = append(r.samples, bufCopy...)
		}
		r.mu.Unlock()
	}
}

// Stop останавливает запись и возвращает записанные сэмплы.
// Если запись слишком короткая, добавляет тишину для Whisper.
func (r *Recorder) Stop() []float32 {
	r.mu.Lock()
	if !r.running {
		r.mu.Unlock()
		return nil
	}

	r.running = false
	stream := r.stream
	r.stream = nil
	samples := r.samples
	r.samples = nil
	done := r.done
	r.mu.Unlock()

	// Ждём завершения recordLoop (максимум 100ms - он проверяет running каждые 10ms)
	if done != nil {
		select {
		case <-done:
		case <-time.After(100 * time.Millisecond):
		}
	}

	// Закрываем stream после завершения recordLoop
	if stream != nil {
		stream.Stop()
		stream.Close()
	}

	// Добавляем тишину если запись слишком короткая
	if len(samples) < MinSamples {
		padding := make([]float32, MinSamples-len(samples))
		samples = append(samples, padding...)
	}

	return samples
}

// Close освобождает ресурсы.
func (r *Recorder) Close() {
	r.Stop()
	portaudio.Terminate()
}

// IsRecording возвращает true если идёт запись.
func (r *Recorder) IsRecording() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.running
}

// GetSamples возвращает копию текущих записанных сэмплов без остановки записи.
// Используется для streaming распознавания.
func (r *Recorder) GetSamples() []float32 {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.running || len(r.samples) == 0 {
		return nil
	}

	// Возвращаем копию чтобы не было race condition
	samples := make([]float32, len(r.samples))
	copy(samples, r.samples)
	return samples
}

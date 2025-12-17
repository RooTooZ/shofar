//go:build ignore

// Скрипт для генерации иконок трея.
// Запуск: go run scripts/generate_icons.go
package main

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"path/filepath"
)

func main() {
	dir := "embedded"
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatalf("Не удалось создать директорию %s: %v", dir, err)
	}

	icons := []struct {
		name  string
		color color.RGBA
	}{
		{"icon_idle.png", color.RGBA{128, 128, 128, 255}},      // Серый
		{"icon_recording.png", color.RGBA{220, 50, 50, 255}},   // Красный
		{"icon_processing.png", color.RGBA{230, 160, 50, 255}}, // Оранжевый
	}

	for _, icon := range icons {
		path := filepath.Join(dir, icon.name)
		if err := generateIcon(path, icon.color); err != nil {
			log.Fatalf("Ошибка генерации %s: %v", icon.name, err)
		}
		log.Printf("Создан: %s", path)
	}
}

func generateIcon(path string, c color.RGBA) error {
	const size = 64
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	// Рисуем круг (микрофон упрощённо)
	centerX, centerY := size/2, size/2
	radius := 20.0

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x - centerX)
			dy := float64(y - centerY)
			if dx*dx+dy*dy <= radius*radius {
				img.Set(x, y, c)
			}
		}
	}

	// Рисуем ножку микрофона
	for y := centerY + int(radius); y < centerY+int(radius)+10; y++ {
		for x := centerX - 3; x <= centerX+3; x++ {
			if y < size {
				img.Set(x, y, c)
			}
		}
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return png.Encode(f, img)
}

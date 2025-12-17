// Shofar - кроссплатформенное приложение для голосового ввода текста.
//
// Работает в системном трее, слушает Ctrl+Shift+Space для push-to-talk.
// Поддерживает Whisper и Vosk для распознавания речи.
package main

import (
	"log"
	"os"

	"shofar/internal/app"
	"shofar/internal/hotkey"
)

// Version устанавливается при сборке через -ldflags.
var Version = "dev"

func main() {
	log.SetFlags(log.Ltime | log.Lshortfile)
	log.Printf("Shofar %s запускается...", Version)

	// Запускаем в главном потоке (требование для macOS и некоторых GUI)
	hotkey.RunOnMainThread(run)
}

func run() {
	application, err := app.New()
	if err != nil {
		log.Printf("Ошибка инициализации: %v", err)
		os.Exit(1)
	}

	log.Println("Приложение запущено. Нажмите Ctrl+Shift+Space для записи.")
	application.Run()
}

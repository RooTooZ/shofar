// Package embedded содержит встроенные ресурсы приложения.
package embedded

import (
	_ "embed"
)

// IconIdle - иконка в состоянии ожидания (серая).
//
//go:embed icon_idle.png
var IconIdle []byte

// IconRecording - иконка во время записи (красная).
//
//go:embed icon_recording.png
var IconRecording []byte

// IconProcessing - иконка во время обработки (оранжевая).
//
//go:embed icon_processing.png
var IconProcessing []byte

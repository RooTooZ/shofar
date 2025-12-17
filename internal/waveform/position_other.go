//go:build !linux

package waveform

// positionWindow is a stub for non-Linux platforms.
// Window positioning is platform-specific and not yet implemented for this OS.
func positionWindow(windowTitle string, width, height int) {
	// TODO: Implement for Windows/macOS
}

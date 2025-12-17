// Package waveform provides a floating window with audio visualization.
package waveform

import (
	"image"
	"image/color"
	"sync"
	"time"

	"gioui.org/app"
	"gioui.org/io/key"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"

	"shofar/internal/i18n"
)

// State represents the window display state.
type State int

const (
	StateRecording     State = iota // Show waveform
	StateSpeechProcess              // Speech-to-text processing
	StateLLMProcess                 // LLM text correction
	StateResult                     // Show recognition result
)

// SampleProvider provides audio samples for visualization.
type SampleProvider interface {
	GetSamples() []float32
	IsRecording() bool
}

// Config holds window configuration.
type Config struct {
	Width        int           // Window width in pixels
	Height       int           // Window height in pixels
	RefreshRate  time.Duration // Refresh interval
	BGColor      color.NRGBA   // Background color
	WaveColor    color.NRGBA   // Waveform color
	VolumeColor  color.NRGBA   // Volume bar color
	TextColor    color.NRGBA   // Text color
	TextDimColor color.NRGBA   // Dim text color
	AccentColor  color.NRGBA   // Accent color (for spinners)
	PanelColor   color.NRGBA   // Panel background
}

// DefaultConfig returns default configuration.
func DefaultConfig() Config {
	return Config{
		Width:        360,
		Height:       100,
		RefreshRate:  33 * time.Millisecond, // ~30fps
		BGColor:      color.NRGBA{R: 30, G: 30, B: 34, A: 245},
		WaveColor:    color.NRGBA{R: 80, G: 200, B: 120, A: 255},
		VolumeColor:  color.NRGBA{R: 255, G: 100, B: 100, A: 255},
		TextColor:    color.NRGBA{R: 240, G: 240, B: 245, A: 255},
		TextDimColor: color.NRGBA{R: 140, G: 140, B: 150, A: 255},
		AccentColor:  color.NRGBA{R: 88, G: 166, B: 255, A: 255},
		PanelColor:   color.NRGBA{R: 45, G: 45, B: 50, A: 255},
	}
}

// Window manages the floating waveform visualization.
type Window struct {
	mu        sync.Mutex
	provider  SampleProvider
	config    Config
	startTime time.Time
	state     State

	// Result display
	resultText string
	editor     widget.Editor
	insertBtn  widget.Clickable
	copyBtn    widget.Clickable
	closeBtn   widget.Clickable
	onInsert   func(text string) // callback when insert is clicked (or Enter)
	onCopy     func(text string) // callback when copy is clicked
	onCancel   func()            // callback when cancelled (ESC or close button)

	window  *app.Window
	running bool
	stopCh  chan struct{}
	doneCh  chan struct{}
}

// New creates a waveform window with the given sample provider.
func New(provider SampleProvider, cfg Config) *Window {
	return &Window{
		provider: provider,
		config:   cfg,
	}
}

// Show displays the waveform window (non-blocking).
func (w *Window) Show() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.running {
		// Window already visible - reset to recording state
		w.state = StateRecording
		w.startTime = time.Now()
		if w.window != nil {
			// Reset window size to normal recording size
			w.window.Option(app.Size(unit.Dp(w.config.Width), unit.Dp(w.config.Height)))
			w.window.Invalidate()
		}
		return
	}

	w.running = true
	w.stopCh = make(chan struct{})
	w.doneCh = make(chan struct{})
	w.startTime = time.Now()
	w.state = StateRecording

	go w.runEventLoop()
}

// Hide closes the waveform window.
func (w *Window) Hide() {
	w.mu.Lock()
	if !w.running {
		w.mu.Unlock()
		return
	}
	w.running = false
	stopCh := w.stopCh
	doneCh := w.doneCh
	w.stopCh = nil
	w.mu.Unlock()

	if stopCh != nil {
		close(stopCh)
	}

	// Wait for window to close
	if doneCh != nil {
		select {
		case <-doneCh:
		case <-time.After(time.Second):
		}
	}
}

// SetStartTime updates the recording start time for the timer display.
func (w *Window) SetStartTime(t time.Time) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.startTime = t
}

// SetState changes the window display state.
func (w *Window) SetState(state State) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.state = state
	if w.window != nil {
		w.window.Invalidate()
	}
}

// SetResult sets the recognition result and switches to result state.
func (w *Window) SetResult(original, corrected string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Use corrected text if available, otherwise original
	result := corrected
	if result == "" {
		result = original
	}
	w.resultText = result

	// Initialize editor with result text
	w.editor = widget.Editor{
		SingleLine: false,
		Submit:     false,
	}
	w.editor.SetText(result)

	w.state = StateResult
	if w.window != nil {
		w.window.Option(app.Size(unit.Dp(450), unit.Dp(220)))
		w.window.Invalidate()
	}
}

// ClearResult clears the stored result text.
func (w *Window) ClearResult() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.resultText = ""
	w.editor.SetText("")
}

// OnInsert sets the callback for when insert button is clicked (or Enter pressed).
func (w *Window) OnInsert(fn func(text string)) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.onInsert = fn
}

// OnCopy sets the callback for when copy button is clicked.
func (w *Window) OnCopy(fn func(text string)) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.onCopy = fn
}

// OnCancel sets the callback for when window is cancelled (ESC or close button).
func (w *Window) OnCancel(fn func()) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.onCancel = fn
}

// IsVisible returns true if window is currently shown.
func (w *Window) IsVisible() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.running
}

const windowTitle = "Shofar - Запись"

func (w *Window) runEventLoop() {
	defer close(w.doneCh)

	// Create window with options
	w.window = new(app.Window)
	w.window.Option(
		app.Title(windowTitle),
		app.Size(unit.Dp(w.config.Width), unit.Dp(w.config.Height)),
		app.Decorated(false), // Borderless
	)

	var ops op.Ops

	// Position window after it appears
	go positionWindow(windowTitle, w.config.Width, w.config.Height)

	// Timer for periodic redraws
	ticker := time.NewTicker(w.config.RefreshRate)
	defer ticker.Stop()

	// Invalidation and close goroutine
	go func() {
		for {
			select {
			case <-w.stopCh:
				// Close the window properly
				if w.window != nil {
					w.window.Perform(system.ActionClose)
				}
				return
			case <-ticker.C:
				if w.window != nil {
					w.window.Invalidate()
				}
			}
		}
	}()

	for {
		switch e := w.window.Event().(type) {
		case app.DestroyEvent:
			return
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)

			// Get current state
			w.mu.Lock()
			startTime := w.startTime
			state := w.state
			w.mu.Unlock()

			// Draw the visualization
			w.draw(gtx, startTime, state)
			e.Frame(gtx.Ops)
		}
	}
}

func (w *Window) draw(gtx layout.Context, startTime time.Time, state State) image.Point {
	// Handle ESC key to cancel and close window
	for {
		event, ok := gtx.Event(key.Filter{Name: key.NameEscape})
		if !ok {
			break
		}
		if e, ok := event.(key.Event); ok && e.State == key.Press {
			w.mu.Lock()
			cancelFn := w.onCancel
			w.mu.Unlock()
			if cancelFn != nil {
				go cancelFn()
			}
			go w.Hide()
			return gtx.Constraints.Max
		}
	}

	elapsed := time.Since(startTime)

	switch state {
	case StateSpeechProcess:
		return drawProcessingStage(gtx, elapsed, w.config, i18n.T("waveform_speech_processing"), i18n.T("waveform_speech_hint"))
	case StateLLMProcess:
		return drawProcessingStage(gtx, elapsed, w.config, i18n.T("waveform_llm_processing"), i18n.T("waveform_llm_hint"))
	case StateResult:
		w.mu.Lock()
		insertCallback := w.onInsert
		copyCallback := w.onCopy
		cancelCallback := w.onCancel
		w.mu.Unlock()

		// Handle Enter key for insert
		for {
			event, ok := gtx.Event(key.Filter{Name: key.NameReturn})
			if !ok {
				break
			}
			if e, ok := event.(key.Event); ok && e.State == key.Press {
				if insertCallback != nil {
					go func() {
						insertCallback(w.editor.Text())
					}()
				}
				go w.Hide()
				return gtx.Constraints.Max
			}
		}

		// Handle button clicks
		if w.insertBtn.Clicked(gtx) && insertCallback != nil {
			text := w.editor.Text()
			go func() {
				insertCallback(text)
			}()
			go w.Hide()
		}
		if w.copyBtn.Clicked(gtx) && copyCallback != nil {
			copyCallback(w.editor.Text())
			go w.Hide()
		}
		if w.closeBtn.Clicked(gtx) {
			if cancelCallback != nil {
				go cancelCallback()
			}
			go w.Hide()
		}

		return drawResultView(gtx, w.config, &w.editor, &w.insertBtn, &w.copyBtn, &w.closeBtn)
	default:
		// Get samples from provider
		var samples []float32
		if w.provider != nil {
			samples = w.provider.GetSamples()
		}
		// Draw recording visualization
		return drawVisualization(gtx, samples, elapsed, w.config)
	}
}

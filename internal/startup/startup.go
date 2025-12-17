// Package startup provides a loading indicator window for app startup.
package startup

import (
	"image"
	"image/color"
	"math"
	"sync"
	"time"

	"gioui.org/app"
	"gioui.org/font"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"

	"whisper-input/internal/i18n"
)

var (
	colorBG     = color.NRGBA{R: 30, G: 30, B: 34, A: 255}
	colorText   = color.NRGBA{R: 240, G: 240, B: 245, A: 255}
	colorDim    = color.NRGBA{R: 140, G: 140, B: 150, A: 255}
	colorAccent = color.NRGBA{R: 88, G: 166, B: 255, A: 255}
)

// Window represents the startup loading window.
type Window struct {
	mu      sync.Mutex
	window  *app.Window
	running bool
	stopCh  chan struct{}
	doneCh  chan struct{}

	// Loading state
	status    string
	substatus string
}

// New creates a new startup window.
func New() *Window {
	return &Window{
		status: i18n.T("startup_status"),
	}
}

// Show displays the loading window.
func (w *Window) Show() {
	w.mu.Lock()
	if w.running {
		w.mu.Unlock()
		return
	}
	w.running = true
	w.stopCh = make(chan struct{})
	w.doneCh = make(chan struct{})
	w.mu.Unlock()

	go w.runEventLoop()
}

// Hide closes the loading window.
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

	if doneCh != nil {
		select {
		case <-doneCh:
		case <-time.After(time.Second):
		}
	}
}

// SetStatus updates the loading status text.
func (w *Window) SetStatus(status, substatus string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.status = status
	w.substatus = substatus
}

func (w *Window) getStatus() (string, string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.status, w.substatus
}

func (w *Window) runEventLoop() {
	defer close(w.doneCh)

	w.window = new(app.Window)
	w.window.Option(
		app.Title("Shofar"),
		app.Size(unit.Dp(300), unit.Dp(150)),
		app.MinSize(unit.Dp(300), unit.Dp(150)),
		app.MaxSize(unit.Dp(300), unit.Dp(150)),
	)

	var ops op.Ops

	// Invalidation goroutine
	go func() {
		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-w.stopCh:
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
			w.draw(gtx)
			e.Frame(gtx.Ops)
		}
	}
}

func (w *Window) draw(gtx layout.Context) layout.Dimensions {
	// Fill background
	rect := clip.Rect{Max: gtx.Constraints.Max}
	paint.FillShape(gtx.Ops, colorBG, rect.Op())

	status, substatus := w.getStatus()

	// Center content
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
			// Spinner
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return w.drawSpinner(gtx)
			}),

			layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),

			// Status text
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				th := material.NewTheme()
				th.Palette.Fg = colorText
				lbl := material.Label(th, unit.Sp(14), status)
				lbl.Font.Weight = font.Medium
				lbl.Alignment = 1 // Center
				return lbl.Layout(gtx)
			}),

			// Substatus text
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if substatus == "" {
					return layout.Dimensions{}
				}
				return layout.Inset{Top: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					th := material.NewTheme()
					th.Palette.Fg = colorDim
					lbl := material.Label(th, unit.Sp(11), substatus)
					lbl.Alignment = 1 // Center
					return lbl.Layout(gtx)
				})
			}),
		)
	})
}

func (w *Window) drawSpinner(gtx layout.Context) layout.Dimensions {
	size := gtx.Dp(unit.Dp(40))
	thickness := gtx.Dp(unit.Dp(3))

	// Animated rotation based on time
	now := time.Now()
	angle := float64(now.UnixMilli()%1000) / 1000.0 * 2 * math.Pi

	center := image.Pt(size/2, size/2)
	radius := size/2 - thickness

	// Draw spinner arc segments
	numSegments := 12
	for i := 0; i < numSegments; i++ {
		segmentAngle := angle + float64(i)*2*math.Pi/float64(numSegments)
		alpha := uint8(255 - i*20)

		x := center.X + int(float64(radius)*math.Cos(segmentAngle))
		y := center.Y + int(float64(radius)*math.Sin(segmentAngle))

		dotRadius := thickness / 2
		dot := clip.Ellipse{
			Min: image.Pt(x-dotRadius, y-dotRadius),
			Max: image.Pt(x+dotRadius, y+dotRadius),
		}
		col := color.NRGBA{R: colorAccent.R, G: colorAccent.G, B: colorAccent.B, A: alpha}
		paint.FillShape(gtx.Ops, col, dot.Op(gtx.Ops))
	}

	return layout.Dimensions{Size: image.Pt(size, size)}
}

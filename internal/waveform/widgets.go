package waveform

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"time"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"shofar/internal/i18n"
)

// drawVisualization draws the complete visualization during recording.
func drawVisualization(gtx layout.Context, samples []float32, elapsed time.Duration, cfg Config) image.Point {
	// Fill background
	drawBackground(gtx, cfg.BGColor)

	// Main content with padding
	layout.UniformInset(unit.Dp(12)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			// Top row: Recording indicator + Timer
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
					// Recording dot
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return drawRecordingDot(gtx, elapsed, cfg.VolumeColor)
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
					// Recording text
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						th := material.NewTheme()
						th.Palette.Fg = cfg.TextColor
						lbl := material.Label(th, unit.Sp(14), i18n.T("waveform_recording"))
						lbl.Font.Weight = font.Medium
						return lbl.Layout(gtx)
					}),
					// Spacer
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						return layout.Dimensions{}
					}),
					// Timer
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return drawTimerBadge(gtx, elapsed, cfg)
					}),
				)
			}),

			layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),

			// Waveform area
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				return drawWaveformPanel(gtx, samples, cfg)
			}),
		)
	})

	return gtx.Constraints.Max
}

// drawBackground draws a rectangle background.
func drawBackground(gtx layout.Context, col color.NRGBA) {
	rect := clip.Rect{Max: gtx.Constraints.Max}
	paint.FillShape(gtx.Ops, col, rect.Op())
}

// drawRecordingDot draws a pulsing recording indicator.
func drawRecordingDot(gtx layout.Context, elapsed time.Duration, col color.NRGBA) layout.Dimensions {
	size := gtx.Dp(unit.Dp(10))

	// Pulsing effect
	pulse := float32(math.Sin(float64(elapsed.Milliseconds())/200.0)*0.3 + 0.7)
	alpha := uint8(float32(col.A) * pulse)
	pulseCol := color.NRGBA{R: col.R, G: col.G, B: col.B, A: alpha}

	// Draw dot
	center := size / 2
	circle := clip.Ellipse{
		Min: image.Pt(0, 0),
		Max: image.Pt(size, size),
	}
	paint.FillShape(gtx.Ops, pulseCol, circle.Op(gtx.Ops))

	return layout.Dimensions{Size: image.Pt(size, size+center/2)}
}

// drawTimerBadge draws the elapsed time in a badge.
func drawTimerBadge(gtx layout.Context, elapsed time.Duration, cfg Config) layout.Dimensions {
	seconds := int(elapsed.Seconds())
	minutes := seconds / 60
	secs := seconds % 60
	timeText := fmt.Sprintf("%d:%02d", minutes, secs)

	// Record content to measure
	macro := op.Record(gtx.Ops)
	dims := layout.Inset{
		Top: unit.Dp(4), Bottom: unit.Dp(4),
		Left: unit.Dp(10), Right: unit.Dp(10),
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		th := material.NewTheme()
		th.Palette.Fg = cfg.TextColor
		lbl := material.Label(th, unit.Sp(13), timeText)
		lbl.Font.Weight = font.Bold
		return lbl.Layout(gtx)
	})
	call := macro.Stop()

	// Draw background
	rr := gtx.Dp(unit.Dp(6))
	rect := clip.RRect{
		Rect: image.Rectangle{Max: dims.Size},
		NE:   rr, NW: rr, SE: rr, SW: rr,
	}
	paint.FillShape(gtx.Ops, cfg.PanelColor, rect.Op(gtx.Ops))

	call.Add(gtx.Ops)
	return dims
}

// drawWaveformPanel draws the waveform in a panel.
func drawWaveformPanel(gtx layout.Context, samples []float32, cfg Config) layout.Dimensions {
	// Draw panel background
	rr := gtx.Dp(unit.Dp(8))
	rect := clip.RRect{
		Rect: image.Rectangle{Max: image.Pt(gtx.Constraints.Max.X, gtx.Constraints.Max.Y)},
		NE:   rr, NW: rr, SE: rr, SW: rr,
	}
	paint.FillShape(gtx.Ops, cfg.PanelColor, rect.Op(gtx.Ops))

	// Draw waveform inside panel with padding
	return layout.UniformInset(unit.Dp(6)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
			// Volume bar
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Max.X = gtx.Dp(unit.Dp(20))
				gtx.Constraints.Min.X = gtx.Constraints.Max.X
				return drawVolumeBar(gtx, samples, cfg)
			}),
			layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
			// Waveform
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				return drawWaveform(gtx, samples, cfg.WaveColor)
			}),
		)
	})
}

// calculateRMS computes root mean square of samples for volume level.
func calculateRMS(samples []float32) float32 {
	if len(samples) == 0 {
		return 0
	}

	// Use only last 1024 samples for responsiveness
	start := 0
	if len(samples) > 1024 {
		start = len(samples) - 1024
	}
	subset := samples[start:]

	var sum float64
	for _, s := range subset {
		sum += float64(s) * float64(s)
	}

	rms := float32(math.Sqrt(sum / float64(len(subset))))

	// Normalize to 0-1 range (typical speech is around 0.1-0.3 RMS)
	level := rms * 3
	if level > 1 {
		level = 1
	}
	return level
}

// drawVolumeBar renders vertical volume indicator.
func drawVolumeBar(gtx layout.Context, samples []float32, cfg Config) layout.Dimensions {
	level := calculateRMS(samples)
	width := gtx.Constraints.Max.X
	height := gtx.Constraints.Max.Y

	// Background
	rr := gtx.Dp(unit.Dp(4))
	bgRect := clip.RRect{
		Rect: image.Rectangle{Max: image.Pt(width, height)},
		NE:   rr, NW: rr, SE: rr, SW: rr,
	}
	paint.FillShape(gtx.Ops, color.NRGBA{R: 35, G: 35, B: 40, A: 255}, bgRect.Op(gtx.Ops))

	// Active bar (from bottom)
	barHeight := int(level * float32(height))
	if barHeight > 0 {
		barRect := clip.RRect{
			Rect: image.Rectangle{
				Min: image.Pt(2, height-barHeight),
				Max: image.Pt(width-2, height-2),
			},
			NE: rr - 1, NW: rr - 1, SE: rr - 1, SW: rr - 1,
		}
		// Gradient effect - brighter at top
		barColor := cfg.VolumeColor
		if level > 0.7 {
			barColor = color.NRGBA{R: 255, G: 80, B: 80, A: 255} // Red for high volume
		} else if level > 0.4 {
			barColor = color.NRGBA{R: 255, G: 180, B: 0, A: 255} // Yellow for medium
		} else {
			barColor = cfg.WaveColor // Green for normal
		}
		paint.FillShape(gtx.Ops, barColor, barRect.Op(gtx.Ops))
	}

	return layout.Dimensions{Size: image.Pt(width, height)}
}

// drawWaveform renders oscilloscope-style waveform.
func drawWaveform(gtx layout.Context, samples []float32, col color.NRGBA) layout.Dimensions {
	width := float32(gtx.Constraints.Max.X)
	height := float32(gtx.Constraints.Max.Y)
	centerY := height / 2

	// Draw center line (dim)
	centerLine := clip.Rect{
		Min: image.Pt(0, int(centerY)),
		Max: image.Pt(int(width), int(centerY)+1),
	}
	paint.FillShape(gtx.Ops, color.NRGBA{R: 60, G: 60, B: 65, A: 255}, centerLine.Op())

	if len(samples) < 2 {
		return layout.Dimensions{Size: image.Pt(int(width), int(height))}
	}

	// Use only last N samples that fit the width
	displaySamples := samples
	maxSamples := int(width)
	if len(samples) > maxSamples {
		displaySamples = samples[len(samples)-maxSamples:]
	}

	// Build path for waveform
	var path clip.Path
	path.Begin(gtx.Ops)

	step := width / float32(len(displaySamples))
	for i, sample := range displaySamples {
		x := float32(i) * step
		y := centerY - (sample * centerY * 0.85)

		if i == 0 {
			path.MoveTo(f32.Pt(x, y))
		} else {
			path.LineTo(f32.Pt(x, y))
		}
	}

	// Stroke the path
	paint.FillShape(gtx.Ops, col, clip.Stroke{
		Path:  path.End(),
		Width: 2,
	}.Op())

	return layout.Dimensions{Size: image.Pt(int(width), int(height))}
}

// drawProcessingStage draws a processing stage with spinner and status.
func drawProcessingStage(gtx layout.Context, elapsed time.Duration, cfg Config, title, subtitle string) image.Point {
	// Fill background
	drawBackground(gtx, cfg.BGColor)

	// Center content
	layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
			// Spinner
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return drawModernSpinner(gtx, elapsed, cfg.AccentColor)
			}),

			layout.Rigid(layout.Spacer{Width: unit.Dp(16)}.Layout),

			// Text content
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					// Title
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						th := material.NewTheme()
						th.Palette.Fg = cfg.TextColor
						lbl := material.Label(th, unit.Sp(15), title)
						lbl.Font.Weight = font.Medium
						return lbl.Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(2)}.Layout),
					// Subtitle
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						th := material.NewTheme()
						th.Palette.Fg = cfg.TextDimColor
						lbl := material.Label(th, unit.Sp(11), subtitle)
						return lbl.Layout(gtx)
					}),
				)
			}),
		)
	})

	return gtx.Constraints.Max
}

// drawModernSpinner draws a modern circular spinner.
func drawModernSpinner(gtx layout.Context, elapsed time.Duration, col color.NRGBA) layout.Dimensions {
	size := gtx.Dp(unit.Dp(36))
	thickness := gtx.Dp(unit.Dp(3))

	// Rotation based on time
	rotation := float64(elapsed.Milliseconds()) / 800.0 * 2 * math.Pi

	center := image.Pt(size/2, size/2)
	radius := size/2 - thickness

	// Draw spinner dots
	numDots := 12
	for i := 0; i < numDots; i++ {
		angle := rotation + float64(i)*2*math.Pi/float64(numDots)
		x := center.X + int(float64(radius)*math.Cos(angle))
		y := center.Y + int(float64(radius)*math.Sin(angle))

		// Fade based on position
		alpha := uint8(255 - i*20)
		if alpha < 40 {
			alpha = 40
		}
		dotColor := color.NRGBA{R: col.R, G: col.G, B: col.B, A: alpha}

		// Draw dot
		dotRadius := thickness / 2
		dot := clip.Ellipse{
			Min: image.Pt(x-dotRadius, y-dotRadius),
			Max: image.Pt(x+dotRadius, y+dotRadius),
		}
		paint.FillShape(gtx.Ops, dotColor, dot.Op(gtx.Ops))
	}

	return layout.Dimensions{Size: image.Pt(size, size)}
}

// drawResultView draws the recognition result with editable text and action buttons.
func drawResultView(gtx layout.Context, cfg Config, editor *widget.Editor, insertBtn, copyBtn, closeBtn *widget.Clickable) image.Point {
	// Fill background
	drawBackground(gtx, cfg.BGColor)

	// Colors
	successColor := color.NRGBA{R: 80, G: 200, B: 120, A: 255}
	secondaryColor := cfg.AccentColor

	// Main content with padding
	layout.UniformInset(unit.Dp(16)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			// Top row: Title + Close button
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
					// Success indicator
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return drawSuccessIcon(gtx, successColor)
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(10)}.Layout),
					// Title
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						th := material.NewTheme()
						th.Palette.Fg = cfg.TextColor
						lbl := material.Label(th, unit.Sp(18), i18n.T("waveform_result"))
						lbl.Font.Weight = font.Medium
						return lbl.Layout(gtx)
					}),
					// Spacer
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						return layout.Dimensions{}
					}),
					// Close button
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return drawCloseButton(gtx, closeBtn, cfg.TextDimColor)
					}),
				)
			}),

			layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),

			// Editable text area
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				return drawEditorPanel(gtx, cfg, editor)
			}),

			layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),

			// Two buttons row
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal, Spacing: layout.SpaceEvenly}.Layout(gtx,
					// Insert button (primary)
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						return drawActionButton(gtx, insertBtn, cfg, successColor, i18n.T("waveform_insert"), true)
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(10)}.Layout),
					// Copy button (secondary)
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						return drawActionButton(gtx, copyBtn, cfg, secondaryColor, i18n.T("waveform_copy"), false)
					}),
				)
			}),
		)
	})

	return gtx.Constraints.Max
}

// drawSuccessIcon draws a checkmark icon.
func drawSuccessIcon(gtx layout.Context, col color.NRGBA) layout.Dimensions {
	size := gtx.Dp(unit.Dp(20))

	// Draw circle background
	circle := clip.Ellipse{
		Min: image.Pt(0, 0),
		Max: image.Pt(size, size),
	}
	paint.FillShape(gtx.Ops, col, circle.Op(gtx.Ops))

	// Draw checkmark
	var path clip.Path
	path.Begin(gtx.Ops)
	// Checkmark coordinates (scaled)
	s := float32(size)
	path.MoveTo(f32.Pt(s*0.25, s*0.5))
	path.LineTo(f32.Pt(s*0.4, s*0.7))
	path.LineTo(f32.Pt(s*0.75, s*0.3))

	paint.FillShape(gtx.Ops, color.NRGBA{R: 255, G: 255, B: 255, A: 255}, clip.Stroke{
		Path:  path.End(),
		Width: float32(gtx.Dp(unit.Dp(2))),
	}.Op())

	return layout.Dimensions{Size: image.Pt(size, size)}
}

// drawCloseButton draws an X button.
func drawCloseButton(gtx layout.Context, btn *widget.Clickable, col color.NRGBA) layout.Dimensions {
	return btn.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		size := gtx.Dp(unit.Dp(24))

		// Hover effect
		if btn.Hovered() {
			col = color.NRGBA{R: 255, G: 100, B: 100, A: 255}
		}

		// Draw X
		var path clip.Path
		path.Begin(gtx.Ops)
		s := float32(size)
		margin := s * 0.25
		// First line of X
		path.MoveTo(f32.Pt(margin, margin))
		path.LineTo(f32.Pt(s-margin, s-margin))

		paint.FillShape(gtx.Ops, col, clip.Stroke{
			Path:  path.End(),
			Width: float32(gtx.Dp(unit.Dp(2))),
		}.Op())

		// Second line of X
		var path2 clip.Path
		path2.Begin(gtx.Ops)
		path2.MoveTo(f32.Pt(s-margin, margin))
		path2.LineTo(f32.Pt(margin, s-margin))

		paint.FillShape(gtx.Ops, col, clip.Stroke{
			Path:  path2.End(),
			Width: float32(gtx.Dp(unit.Dp(2))),
		}.Op())

		return layout.Dimensions{Size: image.Pt(size, size)}
	})
}

// drawEditorPanel draws the panel with editable text.
func drawEditorPanel(gtx layout.Context, cfg Config, editor *widget.Editor) layout.Dimensions {
	// Draw panel background
	rr := gtx.Dp(unit.Dp(10))
	rect := clip.RRect{
		Rect: image.Rectangle{Max: image.Pt(gtx.Constraints.Max.X, gtx.Constraints.Max.Y)},
		NE:   rr, NW: rr, SE: rr, SW: rr,
	}
	paint.FillShape(gtx.Ops, cfg.PanelColor, rect.Op(gtx.Ops))

	// Draw editor with padding
	return layout.UniformInset(unit.Dp(12)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		th := material.NewTheme()
		th.Palette.Fg = cfg.TextColor

		// Style the editor
		ed := material.Editor(th, editor, "")
		ed.TextSize = unit.Sp(16)
		ed.Color = cfg.TextColor
		ed.HintColor = cfg.TextDimColor

		return ed.Layout(gtx)
	})
}

// drawActionButton draws an action button with text.
func drawActionButton(gtx layout.Context, btn *widget.Clickable, cfg Config, bgColor color.NRGBA, text string, primary bool) layout.Dimensions {
	return btn.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		// Hover effect
		currentBg := bgColor
		if btn.Hovered() {
			// Darken on hover
			currentBg = color.NRGBA{
				R: uint8(float32(bgColor.R) * 0.85),
				G: uint8(float32(bgColor.G) * 0.85),
				B: uint8(float32(bgColor.B) * 0.85),
				A: bgColor.A,
			}
		}

		// Record content to measure
		macro := op.Record(gtx.Ops)
		dims := layout.Inset{
			Top: unit.Dp(10), Bottom: unit.Dp(10),
			Left: unit.Dp(12), Right: unit.Dp(12),
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				th := material.NewTheme()
				th.Palette.Fg = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
				lbl := material.Label(th, unit.Sp(14), text)
				lbl.Font.Weight = font.Medium
				return lbl.Layout(gtx)
			})
		})
		call := macro.Stop()

		// Draw button background
		rr := gtx.Dp(unit.Dp(8))
		btnRect := clip.RRect{
			Rect: image.Rectangle{Max: image.Pt(gtx.Constraints.Max.X, dims.Size.Y)},
			NE:   rr, NW: rr, SE: rr, SW: rr,
		}
		paint.FillShape(gtx.Ops, currentBg, btnRect.Op(gtx.Ops))

		call.Add(gtx.Ops)
		return layout.Dimensions{Size: image.Pt(gtx.Constraints.Max.X, dims.Size.Y)}
	})
}

// drawCopyIcon draws a small copy icon.
func drawCopyIcon(gtx layout.Context, col color.NRGBA) layout.Dimensions {
	size := gtx.Dp(unit.Dp(18))
	s := float32(size)

	// Back rectangle
	var path1 clip.Path
	path1.Begin(gtx.Ops)
	path1.MoveTo(f32.Pt(s*0.3, s*0.1))
	path1.LineTo(f32.Pt(s*0.9, s*0.1))
	path1.LineTo(f32.Pt(s*0.9, s*0.7))
	path1.LineTo(f32.Pt(s*0.3, s*0.7))
	path1.Close()

	paint.FillShape(gtx.Ops, col, clip.Stroke{
		Path:  path1.End(),
		Width: float32(gtx.Dp(unit.Dp(1.5))),
	}.Op())

	// Front rectangle
	var path2 clip.Path
	path2.Begin(gtx.Ops)
	path2.MoveTo(f32.Pt(s*0.1, s*0.3))
	path2.LineTo(f32.Pt(s*0.7, s*0.3))
	path2.LineTo(f32.Pt(s*0.7, s*0.9))
	path2.LineTo(f32.Pt(s*0.1, s*0.9))
	path2.Close()

	paint.FillShape(gtx.Ops, col, clip.Stroke{
		Path:  path2.End(),
		Width: float32(gtx.Dp(unit.Dp(1.5))),
	}.Op())

	return layout.Dimensions{Size: image.Pt(size, size)}
}

package settings

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"time"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"whisper-input/internal/config"
	"whisper-input/internal/i18n"
	"whisper-input/internal/models"
)

// Color palette - modern dark theme
var (
	colorBG         = color.NRGBA{R: 30, G: 30, B: 34, A: 255}
	colorPanel      = color.NRGBA{R: 45, G: 45, B: 50, A: 255}
	colorPanelLight = color.NRGBA{R: 55, G: 55, B: 62, A: 255}
	colorText       = color.NRGBA{R: 240, G: 240, B: 245, A: 255}
	colorTextDim    = color.NRGBA{R: 140, G: 140, B: 150, A: 255}
	colorAccent     = color.NRGBA{R: 88, G: 166, B: 255, A: 255}
	colorSuccess    = color.NRGBA{R: 80, G: 200, B: 120, A: 255}
	colorWarning    = color.NRGBA{R: 255, G: 180, B: 0, A: 255}
	colorSelected   = color.NRGBA{R: 60, G: 100, B: 160, A: 255}
)

func (w *Window) draw(gtx layout.Context) layout.Dimensions {
	// Fill background
	rect := clip.Rect{Max: gtx.Constraints.Max}
	paint.FillShape(gtx.Ops, colorBG, rect.Op())

	engine, selectedModel, downloading, progress, progressModel := w.getState()
	loadingModel, loadingModelID := w.getLoadingState()

	// Main layout with padding
	dims := layout.UniformInset(unit.Dp(20)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			// Title (fixed)
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return w.drawTitle(gtx)
			}),

			layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),

			// Scrollable content area
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				th := material.NewTheme()
				return material.List(th, &w.contentList).Layout(gtx, 1, func(gtx layout.Context, _ int) layout.Dimensions {
					return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
						// UI Language section
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return w.drawUILanguageSection(gtx)
						}),

						layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),

						// Hotkey section
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return w.drawHotkeySection(gtx)
						}),

						layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),

						// LLM correction section
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return w.drawLLMSection(gtx)
						}),

						layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),

						// Recognition section (Engine + Model)
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return w.drawSectionHeader(gtx, i18n.T("settings_recognition"))
						}),

						layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),

						// Engine selector
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return w.drawEngineSelector(gtx, engine)
						}),

						layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),

						// Model list (all models shown)
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return w.drawModelListInline(gtx, engine, selectedModel)
						}),
					)
				})
			}),

			// Progress bar (fixed, always visible when downloading)
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if !downloading {
					return layout.Dimensions{}
				}
				return layout.Inset{Top: unit.Dp(12), Bottom: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return w.drawProgressBar(gtx, progress, progressModel)
				})
			}),

			layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),

			// Buttons (fixed at bottom)
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return w.drawButtons(gtx, selectedModel, downloading || loadingModel)
			}),
		)
	})

	// Draw loading overlay if model is being loaded
	if loadingModel {
		w.drawLoadingOverlay(gtx, loadingModelID)
	}

	return dims
}

func (w *Window) drawLoadingOverlay(gtx layout.Context, modelID string) {
	// Semi-transparent overlay
	rect := clip.Rect{Max: gtx.Constraints.Max}
	paint.FillShape(gtx.Ops, color.NRGBA{R: 20, G: 20, B: 24, A: 220}, rect.Op())

	// Center content
	layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
			// Animated spinner
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return w.drawSpinner(gtx)
			}),

			layout.Rigid(layout.Spacer{Height: unit.Dp(20)}.Layout),

			// Loading text
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				th := material.NewTheme()
				th.Palette.Fg = colorText
				info, _ := models.GetModel(modelID)
				text := fmt.Sprintf("%s %s...", i18n.T("settings_loading_model"), info.Name)
				lbl := material.Label(th, unit.Sp(16), text)
				lbl.Font.Weight = font.Medium
				return lbl.Layout(gtx)
			}),

			layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),

			// Hint text
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				th := material.NewTheme()
				th.Palette.Fg = colorTextDim
				lbl := material.Label(th, unit.Sp(12), i18n.T("settings_loading_hint"))
				return lbl.Layout(gtx)
			}),
		)
	})
}

func (w *Window) drawSpinner(gtx layout.Context) layout.Dimensions {
	size := gtx.Dp(unit.Dp(48))
	thickness := gtx.Dp(unit.Dp(4))

	// Animated rotation based on time
	now := time.Now()
	angle := float64(now.UnixMilli()%1000) / 1000.0 * 2 * math.Pi

	center := image.Pt(size/2, size/2)
	radius := size/2 - thickness

	// Draw spinner arc segments
	numSegments := 12
	for i := 0; i < numSegments; i++ {
		segmentAngle := angle + float64(i)*2*math.Pi/float64(numSegments)
		alpha := uint8(255 - i*20) // Fade out segments

		x := center.X + int(float64(radius)*math.Cos(segmentAngle))
		y := center.Y + int(float64(radius)*math.Sin(segmentAngle))

		// Draw dot
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

func (w *Window) drawTitle(gtx layout.Context) layout.Dimensions {
	th := material.NewTheme()
	th.Palette.Fg = colorText

	label := material.Label(th, unit.Sp(22), i18n.T("settings_title"))
	label.Font.Weight = font.Bold
	return label.Layout(gtx)
}

func (w *Window) drawSectionHeader(gtx layout.Context, text string) layout.Dimensions {
	th := material.NewTheme()
	th.Palette.Fg = colorTextDim

	label := material.Label(th, unit.Sp(12), text)
	label.Font.Weight = font.Medium
	return label.Layout(gtx)
}

func (w *Window) drawHotkeySection(gtx layout.Context) layout.Dimensions {
	isRecording := w.isRecordingHotkey()

	return w.drawPanel(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			// Section header
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return w.drawSectionHeader(gtx, i18n.T("settings_hotkey"))
			}),

			layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),

			// Hotkey display and edit button
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
					// Current hotkey preview
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						return w.drawHotkeyPreview(gtx, isRecording)
					}),

					layout.Rigid(layout.Spacer{Width: unit.Dp(12)}.Layout),

					// Edit button
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						if isRecording {
							return w.drawButton(gtx, &w.hotkeyEditBtn, i18n.T("settings_hotkey_cancel"), colorWarning, colorText, true)
						}
						return w.drawButton(gtx, &w.hotkeyEditBtn, i18n.T("settings_hotkey_edit"), colorAccent, colorText, true)
					}),
				)
			}),
		)
	})
}

func (w *Window) drawUILanguageSection(gtx layout.Context) layout.Dimensions {
	selectedLang := w.getSelectedUILang()

	return w.drawPanel(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			// Section header
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return w.drawSectionHeader(gtx, i18n.T("settings_ui_language"))
			}),

			layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),

			// Language buttons
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return w.drawLangButton(gtx, i18n.RU, "Русский", selectedLang == i18n.RU)
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return w.drawLangButton(gtx, i18n.EN, "English", selectedLang == i18n.EN)
					}),
				)
			}),
		)
	})
}

func (w *Window) drawLangButton(gtx layout.Context, lang i18n.Language, label string, selected bool) layout.Dimensions {
	btn := w.getLangButton(lang)

	bgColor := colorPanel
	textColor := colorTextDim
	if selected {
		bgColor = colorAccent
		textColor = colorText
	}

	// Record content to measure size
	macro := op.Record(gtx.Ops)
	dims := material.Clickable(gtx, btn, func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{
			Top: unit.Dp(8), Bottom: unit.Dp(8),
			Left: unit.Dp(16), Right: unit.Dp(16),
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			th := material.NewTheme()
			th.Palette.Fg = textColor
			lbl := material.Label(th, unit.Sp(14), label)
			lbl.Font.Weight = font.Medium
			return lbl.Layout(gtx)
		})
	})
	call := macro.Stop()

	// Draw background
	rr := gtx.Dp(unit.Dp(6))
	rect := clip.RRect{
		Rect: image.Rectangle{Max: dims.Size},
		NE:   rr, NW: rr, SE: rr, SW: rr,
	}
	paint.FillShape(gtx.Ops, bgColor, rect.Op(gtx.Ops))

	// Replay content
	call.Add(gtx.Ops)

	return dims
}

func (w *Window) drawLLMSection(gtx layout.Context) layout.Dimensions {
	return w.drawPanel(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			// Section header
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return w.drawSectionHeader(gtx, i18n.T("settings_llm"))
			}),

			layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),

			// Toggle and description
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
					// Toggle
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return w.drawToggle(gtx, &w.llmEnabled)
					}),

					layout.Rigid(layout.Spacer{Width: unit.Dp(12)}.Layout),

					// Description
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								th := material.NewTheme()
								th.Palette.Fg = colorText
								lbl := material.Label(th, unit.Sp(14), i18n.T("settings_llm_enable"))
								lbl.Font.Weight = font.Medium
								return lbl.Layout(gtx)
							}),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								th := material.NewTheme()
								th.Palette.Fg = colorTextDim
								lbl := material.Label(th, unit.Sp(11), i18n.T("settings_llm_hint"))
								return lbl.Layout(gtx)
							}),
						)
					}),
				)
			}),

			// LLM model list (if LLM enabled)
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if !w.llmEnabled.Value {
					return layout.Dimensions{}
				}
				return layout.Inset{Top: unit.Dp(12)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return w.drawLLMModelList(gtx)
				})
			}),
		)
	})
}

func (w *Window) drawLLMModelList(gtx layout.Context) layout.Dimensions {
	llmModels := models.GetLLMModels()
	selectedLLM := w.config.LLMModelID()
	if selectedLLM == "" {
		selectedLLM = models.DefaultLLMModelID()
	}

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		// LLM models
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			var items []layout.FlexChild
			for _, m := range llmModels {
				model := m // capture
				items = append(items,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Inset{Bottom: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return w.drawLLMModelItem(gtx, model, selectedLLM == model.ID)
						})
					}),
				)
			}
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx, items...)
		}),
	)
}

func (w *Window) drawLLMModelItem(gtx layout.Context, m models.ModelInfo, selected bool) layout.Dimensions {
	isDownloaded := w.manager.IsDownloaded(m)
	btn := w.modelButtons[m.ID]
	downloadBtn := w.downloadBtns[m.ID]

	// Handle click - select this LLM model
	if btn.Clicked(gtx) && isDownloaded {
		w.config.SetLLMModelID(m.ID)
	}

	// Item background
	bgColor := colorPanelLight
	if selected {
		bgColor = colorSelected
	}

	// Record content to measure size
	macro := op.Record(gtx.Ops)
	dims := material.Clickable(gtx, btn, func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{
			Top: unit.Dp(8), Bottom: unit.Dp(8),
			Left: unit.Dp(10), Right: unit.Dp(10),
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
				// Radio indicator
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return w.drawRadioIndicator(gtx, selected)
				}),

				layout.Rigid(layout.Spacer{Width: unit.Dp(10)}.Layout),

				// Model info
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							th := material.NewTheme()
							th.Palette.Fg = colorText
							lbl := material.Label(th, unit.Sp(13), m.Name)
							lbl.Font.Weight = font.Medium
							return lbl.Layout(gtx)
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							th := material.NewTheme()
							th.Palette.Fg = colorTextDim
							size := formatSize(m.Size)
							lbl := material.Label(th, unit.Sp(10), size)
							return lbl.Layout(gtx)
						}),
					)
				}),

				// Status badge or download button
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if isDownloaded {
						return w.drawStatusBadge(gtx, "✓", colorSuccess)
					}
					return w.drawDownloadButton(gtx, downloadBtn)
				}),
			)
		})
	})
	call := macro.Stop()

	// Draw background
	rr := gtx.Dp(unit.Dp(6))
	rect := clip.RRect{
		Rect: image.Rectangle{Max: dims.Size},
		NE:   rr, NW: rr, SE: rr, SW: rr,
	}
	paint.FillShape(gtx.Ops, bgColor, rect.Op(gtx.Ops))

	// Replay content
	call.Add(gtx.Ops)

	return dims
}

func (w *Window) drawToggle(gtx layout.Context, toggle *widget.Bool) layout.Dimensions {
	th := material.NewTheme()

	// Use material Switch
	sw := material.Switch(th, toggle, "")
	sw.Color.Enabled = colorAccent
	sw.Color.Disabled = colorPanel

	return sw.Layout(gtx)
}

func (w *Window) drawHotkeyPreview(gtx layout.Context, isRecording bool) layout.Dimensions {
	var hotkeyStr string
	var textColor color.NRGBA
	var bgColor color.NRGBA

	if isRecording {
		// Show recording state
		mods, key := w.getRecordingState()
		parts := buildHotkeyParts(mods, key)

		if len(parts) > 0 {
			hotkeyStr = ""
			for i, p := range parts {
				if i > 0 {
					hotkeyStr += " + "
				}
				hotkeyStr += p
			}
		} else {
			hotkeyStr = i18n.T("settings_hotkey_prompt")
		}
		textColor = colorWarning
		bgColor = color.NRGBA{R: 80, G: 60, B: 20, A: 255}
	} else {
		// Show current hotkey
		mods, key := w.getHotkeyState()
		parts := buildHotkeyParts(mods, key)

		if len(parts) > 0 {
			hotkeyStr = ""
			for i, p := range parts {
				if i > 0 {
					hotkeyStr += " + "
				}
				hotkeyStr += p
			}
		} else {
			hotkeyStr = i18n.T("settings_hotkey_not_set")
		}
		textColor = colorAccent
		bgColor = colorPanelLight
	}

	// Record content to measure size
	macro := op.Record(gtx.Ops)
	dims := layout.UniformInset(unit.Dp(12)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		th := material.NewTheme()
		th.Palette.Fg = textColor
		label := material.Label(th, unit.Sp(16), "⌨  "+hotkeyStr)
		label.Font.Weight = font.Medium
		return label.Layout(gtx)
	})
	call := macro.Stop()

	// Draw background with measured size
	rr := gtx.Dp(unit.Dp(8))
	rect := clip.RRect{
		Rect: image.Rectangle{Max: dims.Size},
		NE:   rr, NW: rr, SE: rr, SW: rr,
	}
	paint.FillShape(gtx.Ops, bgColor, rect.Op(gtx.Ops))

	// Replay content
	call.Add(gtx.Ops)

	return dims
}

func buildHotkeyParts(mods map[config.Modifier]bool, key config.Key) []string {
	parts := []string{}

	if mods[config.ModCtrl] {
		parts = append(parts, "Ctrl")
	}
	if mods[config.ModShift] {
		parts = append(parts, "Shift")
	}
	if mods[config.ModAlt] {
		parts = append(parts, "Alt")
	}
	if mods[config.ModSuper] {
		parts = append(parts, "Super")
	}

	keyName := keyDisplayName(key)
	if keyName != "" {
		parts = append(parts, keyName)
	}

	return parts
}

func keyDisplayName(key config.Key) string {
	switch key {
	case config.KeySpace:
		return "Space"
	case config.KeyReturn:
		return "Enter"
	case config.KeyTab:
		return "Tab"
	case config.KeyF1:
		return "F1"
	case config.KeyF2:
		return "F2"
	case config.KeyF3:
		return "F3"
	case config.KeyF4:
		return "F4"
	case config.KeyF5:
		return "F5"
	default:
		if key != "" {
			return string(key)
		}
		return ""
	}
}

func (w *Window) drawKeySelector(gtx layout.Context) layout.Dimensions {
	th := material.NewTheme()
	th.Palette.Fg = colorText

	// Available keys
	keys := []struct {
		key   config.Key
		label string
	}{
		{config.KeySpace, "Space"},
		{config.KeyReturn, "Enter"},
		{config.KeyTab, "Tab"},
		{config.KeyF1, "F1"},
		{config.KeyF2, "F2"},
		{config.KeyF3, "F3"},
		{config.KeyF4, "F4"},
		{config.KeyF5, "F5"},
	}

	// Add letter keys A-Z
	for c := 'a'; c <= 'z'; c++ {
		keys = append(keys, struct {
			key   config.Key
			label string
		}{config.Key(c), string(c - 32)}) // uppercase display
	}

	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.Label(th, unit.Sp(14), i18n.T("settings_key"))
			lbl.Color = colorTextDim
			return lbl.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Width: unit.Dp(12)}.Layout),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			// Constrain height for horizontal list
			gtx.Constraints.Max.Y = gtx.Dp(unit.Dp(32))
			gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
			return material.List(th, &w.keyList).Layout(gtx, len(keys), func(gtx layout.Context, i int) layout.Dimensions {
				k := keys[i]
				return layout.Inset{Right: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return w.drawKeyButton(gtx, k.key, k.label)
				})
			})
		}),
	)
}

func (w *Window) drawKeyButton(gtx layout.Context, key config.Key, label string) layout.Dimensions {
	isSelected := w.keyEnum.Value == string(key)

	// Handle click
	btn := w.getKeyButton(key)
	if btn.Clicked(gtx) {
		w.keyEnum.Value = string(key)
		w.mu.Lock()
		w.hotkeyKey = key
		w.mu.Unlock()
	}

	// Button colors - improved contrast
	bgColor := color.NRGBA{R: 70, G: 70, B: 78, A: 255}
	textColor := colorText
	if isSelected {
		bgColor = colorAccent
		textColor = colorText
	}

	// Record content to measure size
	macro := op.Record(gtx.Ops)
	dims := material.Clickable(gtx, btn, func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{
			Top: unit.Dp(6), Bottom: unit.Dp(6),
			Left: unit.Dp(10), Right: unit.Dp(10),
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			th := material.NewTheme()
			th.Palette.Fg = textColor
			lbl := material.Label(th, unit.Sp(12), label)
			lbl.Font.Weight = font.Medium
			return lbl.Layout(gtx)
		})
	})
	call := macro.Stop()

	// Draw background
	rr := gtx.Dp(unit.Dp(4))
	rect := clip.RRect{
		Rect: image.Rectangle{Max: dims.Size},
		NE:   rr, NW: rr, SE: rr, SW: rr,
	}
	paint.FillShape(gtx.Ops, bgColor, rect.Op(gtx.Ops))

	// Replay content
	call.Add(gtx.Ops)

	return dims
}

func (w *Window) getKeyButton(key config.Key) *widget.Clickable {
	if w.keyButtons == nil {
		w.keyButtons = make(map[config.Key]*widget.Clickable)
	}
	if w.keyButtons[key] == nil {
		w.keyButtons[key] = new(widget.Clickable)
	}
	return w.keyButtons[key]
}

func (w *Window) drawEngineSelector(gtx layout.Context, currentEngine models.Engine) layout.Dimensions {
	th := material.NewTheme()
	th.Palette.Fg = colorText

	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.Label(th, unit.Sp(14), i18n.T("settings_engine"))
			lbl.Color = colorTextDim
			return lbl.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Width: unit.Dp(12)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return w.drawEngineButton(gtx, models.EngineWhisper, "Whisper", currentEngine == models.EngineWhisper)
		}),
		layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return w.drawEngineButton(gtx, models.EngineVosk, "Vosk", currentEngine == models.EngineVosk)
		}),
	)
}

func (w *Window) drawEngineButton(gtx layout.Context, engine models.Engine, label string, selected bool) layout.Dimensions {
	btn := w.getEngineButton(engine)
	if btn.Clicked(gtx) {
		w.engineEnum.Value = string(engine)
		w.mu.Lock()
		if w.selectedEngine != engine {
			w.selectedEngine = engine
			w.selectedModel = ""
		}
		w.mu.Unlock()
	}

	bgColor := colorPanel
	textColor := colorTextDim
	if selected {
		bgColor = colorAccent
		textColor = colorText
	}

	// Record content to measure size
	macro := op.Record(gtx.Ops)
	dims := material.Clickable(gtx, btn, func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{
			Top: unit.Dp(8), Bottom: unit.Dp(8),
			Left: unit.Dp(16), Right: unit.Dp(16),
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			th := material.NewTheme()
			th.Palette.Fg = textColor
			lbl := material.Label(th, unit.Sp(14), label)
			lbl.Font.Weight = font.Medium
			return lbl.Layout(gtx)
		})
	})
	call := macro.Stop()

	// Draw background
	rr := gtx.Dp(unit.Dp(6))
	rect := clip.RRect{
		Rect: image.Rectangle{Max: dims.Size},
		NE:   rr, NW: rr, SE: rr, SW: rr,
	}
	paint.FillShape(gtx.Ops, bgColor, rect.Op(gtx.Ops))

	// Replay content
	call.Add(gtx.Ops)

	return dims
}

func (w *Window) getEngineButton(engine models.Engine) *widget.Clickable {
	if w.engineButtons == nil {
		w.engineButtons = make(map[models.Engine]*widget.Clickable)
	}
	if w.engineButtons[engine] == nil {
		w.engineButtons[engine] = new(widget.Clickable)
	}
	return w.engineButtons[engine]
}

func (w *Window) drawPanel(gtx layout.Context, content layout.Widget) layout.Dimensions {
	// First layout content to get its size
	macro := op.Record(gtx.Ops)
	dims := layout.UniformInset(unit.Dp(16)).Layout(gtx, content)
	call := macro.Stop()

	// Draw background with content size
	rr := gtx.Dp(unit.Dp(12))
	rect := clip.RRect{
		Rect: image.Rectangle{Max: dims.Size},
		NE:   rr, NW: rr, SE: rr, SW: rr,
	}
	paint.FillShape(gtx.Ops, colorPanel, rect.Op(gtx.Ops))

	// Replay content drawing
	call.Add(gtx.Ops)

	return dims
}

func (w *Window) drawModelList(gtx layout.Context, engine models.Engine, selectedModel string) layout.Dimensions {
	modelList := models.GetModelsByEngine(engine)

	// Draw panel background
	rr := gtx.Dp(unit.Dp(12))
	rect := clip.RRect{
		Rect: image.Rectangle{Max: gtx.Constraints.Max},
		NE:   rr, NW: rr, SE: rr, SW: rr,
	}
	paint.FillShape(gtx.Ops, colorPanel, rect.Op(gtx.Ops))

	return layout.UniformInset(unit.Dp(8)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		th := material.NewTheme()
		return material.List(th, &w.modelList).Layout(gtx, len(modelList), func(gtx layout.Context, i int) layout.Dimensions {
			m := modelList[i]
			return layout.Inset{Bottom: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return w.drawModelItem(gtx, m, selectedModel == m.ID)
			})
		})
	})
}

// drawModelListInline renders models inline (used in scrollable parent)
func (w *Window) drawModelListInline(gtx layout.Context, engine models.Engine, selectedModel string) layout.Dimensions {
	modelList := models.GetModelsByEngine(engine)

	// Record content to measure size
	macro := op.Record(gtx.Ops)
	dims := layout.UniformInset(unit.Dp(8)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		var items []layout.FlexChild
		for _, m := range modelList {
			model := m // capture
			items = append(items,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Bottom: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return w.drawModelItem(gtx, model, selectedModel == model.ID)
					})
				}),
			)
		}
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx, items...)
	})
	call := macro.Stop()

	// Draw panel background
	rr := gtx.Dp(unit.Dp(12))
	rect := clip.RRect{
		Rect: image.Rectangle{Max: dims.Size},
		NE:   rr, NW: rr, SE: rr, SW: rr,
	}
	paint.FillShape(gtx.Ops, colorPanel, rect.Op(gtx.Ops))

	// Replay content
	call.Add(gtx.Ops)

	return dims
}

func (w *Window) drawModelItem(gtx layout.Context, m models.ModelInfo, selected bool) layout.Dimensions {
	isDownloaded := w.manager.IsDownloaded(m)
	btn := w.modelButtons[m.ID]
	downloadBtn := w.downloadBtns[m.ID]

	// Item background
	bgColor := colorPanelLight
	if selected {
		bgColor = colorSelected
	}

	// Record content to measure size
	macro := op.Record(gtx.Ops)
	dims := material.Clickable(gtx, btn, func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{
			Top: unit.Dp(10), Bottom: unit.Dp(10),
			Left: unit.Dp(12), Right: unit.Dp(12),
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
				// Radio indicator
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return w.drawRadioIndicator(gtx, selected)
				}),

				layout.Rigid(layout.Spacer{Width: unit.Dp(12)}.Layout),

				// Model info
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							th := material.NewTheme()
							th.Palette.Fg = colorText
							lbl := material.Label(th, unit.Sp(14), m.Name)
							lbl.Font.Weight = font.Medium
							return lbl.Layout(gtx)
						}),
						layout.Rigid(layout.Spacer{Height: unit.Dp(2)}.Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							th := material.NewTheme()
							th.Palette.Fg = colorTextDim
							size := formatSize(m.Size)
							lbl := material.Label(th, unit.Sp(11), size)
							return lbl.Layout(gtx)
						}),
					)
				}),

				// Status badge or download button
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if isDownloaded {
						return w.drawStatusBadge(gtx, "✓", colorSuccess)
					}
					return w.drawDownloadButton(gtx, downloadBtn)
				}),
			)
		})
	})
	call := macro.Stop()

	// Draw background
	rr := gtx.Dp(unit.Dp(8))
	rect := clip.RRect{
		Rect: image.Rectangle{Max: dims.Size},
		NE:   rr, NW: rr, SE: rr, SW: rr,
	}
	paint.FillShape(gtx.Ops, bgColor, rect.Op(gtx.Ops))

	// Replay content
	call.Add(gtx.Ops)

	return dims
}

func (w *Window) drawRadioIndicator(gtx layout.Context, selected bool) layout.Dimensions {
	size := gtx.Dp(unit.Dp(18))
	borderWidth := gtx.Dp(unit.Dp(2))

	// Outer circle
	center := image.Pt(size/2, size/2)
	outerRadius := size / 2

	if selected {
		// Filled circle for selected
		circle := clip.Ellipse{
			Min: image.Pt(center.X-outerRadius, center.Y-outerRadius),
			Max: image.Pt(center.X+outerRadius, center.Y+outerRadius),
		}
		paint.FillShape(gtx.Ops, colorAccent, circle.Op(gtx.Ops))

		// Inner dot
		innerRadius := outerRadius - borderWidth*2
		innerCircle := clip.Ellipse{
			Min: image.Pt(center.X-innerRadius, center.Y-innerRadius),
			Max: image.Pt(center.X+innerRadius, center.Y+innerRadius),
		}
		paint.FillShape(gtx.Ops, colorText, innerCircle.Op(gtx.Ops))
	} else {
		// Just border for unselected
		circle := clip.Ellipse{
			Min: image.Pt(center.X-outerRadius, center.Y-outerRadius),
			Max: image.Pt(center.X+outerRadius, center.Y+outerRadius),
		}
		paint.FillShape(gtx.Ops, colorTextDim, circle.Op(gtx.Ops))

		innerRadius := outerRadius - borderWidth
		innerCircle := clip.Ellipse{
			Min: image.Pt(center.X-innerRadius, center.Y-innerRadius),
			Max: image.Pt(center.X+innerRadius, center.Y+innerRadius),
		}
		paint.FillShape(gtx.Ops, colorPanelLight, innerCircle.Op(gtx.Ops))
	}

	return layout.Dimensions{Size: image.Pt(size, size)}
}

func (w *Window) drawStatusBadge(gtx layout.Context, text string, col color.NRGBA) layout.Dimensions {
	th := material.NewTheme()
	th.Palette.Fg = col
	lbl := material.Label(th, unit.Sp(16), text)
	lbl.Font.Weight = font.Bold
	return lbl.Layout(gtx)
}

func (w *Window) drawDownloadButton(gtx layout.Context, btn *widget.Clickable) layout.Dimensions {
	macro := op.Record(gtx.Ops)
	dims := material.Clickable(gtx, btn, func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{
			Top: unit.Dp(4), Bottom: unit.Dp(4),
			Left: unit.Dp(8), Right: unit.Dp(8),
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			th := material.NewTheme()
			th.Palette.Fg = colorText
			lbl := material.Label(th, unit.Sp(11), "↓")
			lbl.Font.Weight = font.Bold
			return lbl.Layout(gtx)
		})
	})
	call := macro.Stop()

	rr := gtx.Dp(unit.Dp(4))
	rect := clip.RRect{
		Rect: image.Rectangle{Max: dims.Size},
		NE:   rr, NW: rr, SE: rr, SW: rr,
	}
	paint.FillShape(gtx.Ops, colorAccent, rect.Op(gtx.Ops))

	call.Add(gtx.Ops)
	return dims
}

func (w *Window) drawProgressBar(gtx layout.Context, progress float64, modelID string) layout.Dimensions {
	info, _ := models.GetModel(modelID)

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		// Progress bar
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			height := gtx.Dp(unit.Dp(6))
			width := gtx.Constraints.Max.X

			rr := height / 2
			bgRect := clip.RRect{
				Rect: image.Rectangle{Max: image.Pt(width, height)},
				NE:   rr, NW: rr, SE: rr, SW: rr,
			}
			paint.FillShape(gtx.Ops, colorPanel, bgRect.Op(gtx.Ops))

			fillWidth := int(float64(width) * progress)
			if fillWidth > 0 {
				fillRect := clip.RRect{
					Rect: image.Rectangle{Max: image.Pt(fillWidth, height)},
					NE:   rr, NW: rr, SE: rr, SW: rr,
				}
				paint.FillShape(gtx.Ops, colorWarning, fillRect.Op(gtx.Ops))
			}

			return layout.Dimensions{Size: image.Pt(width, height)}
		}),

		layout.Rigid(layout.Spacer{Height: unit.Dp(4)}.Layout),

		// Progress text
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			th := material.NewTheme()
			th.Palette.Fg = colorTextDim
			text := fmt.Sprintf("%s %s... %.0f%%", i18n.T("settings_downloading"), info.Name, progress*100)
			lbl := material.Label(th, unit.Sp(11), text)
			return lbl.Layout(gtx)
		}),
	)
}

func (w *Window) drawButtons(gtx layout.Context, selectedModel string, downloading bool) layout.Dimensions {
	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
	}.Layout(gtx,
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return layout.Dimensions{}
		}),

		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return w.drawButton(gtx, &w.cancelBtn, i18n.T("settings_cancel"), colorPanel, colorText, true)
		}),

		layout.Rigid(layout.Spacer{Width: unit.Dp(12)}.Layout),

		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			canApply := !downloading
			bgColor := colorAccent
			if !canApply {
				bgColor = colorPanel
			}
			return w.drawButton(gtx, &w.applyBtn, i18n.T("settings_apply"), bgColor, colorText, canApply)
		}),
	)
}

func (w *Window) drawButton(gtx layout.Context, btn *widget.Clickable, label string, bgColor, textColor color.NRGBA, enabled bool) layout.Dimensions {
	if !enabled {
		textColor = colorTextDim
	}

	macro := op.Record(gtx.Ops)
	dims := material.Clickable(gtx, btn, func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{
			Top: unit.Dp(10), Bottom: unit.Dp(10),
			Left: unit.Dp(20), Right: unit.Dp(20),
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			th := material.NewTheme()
			th.Palette.Fg = textColor
			lbl := material.Label(th, unit.Sp(14), label)
			lbl.Font.Weight = font.Medium
			return lbl.Layout(gtx)
		})
	})
	call := macro.Stop()

	rr := gtx.Dp(unit.Dp(8))
	rect := clip.RRect{
		Rect: image.Rectangle{Max: dims.Size},
		NE:   rr, NW: rr, SE: rr, SW: rr,
	}
	paint.FillShape(gtx.Ops, bgColor, rect.Op(gtx.Ops))

	call.Add(gtx.Ops)
	return dims
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.0f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

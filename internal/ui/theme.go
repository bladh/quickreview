package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// ReviewTheme is a compact theme suited to code review.
// Light selects the bright color scheme; dark is the default.
type ReviewTheme struct {
	Light bool
}

var _ fyne.Theme = (*ReviewTheme)(nil)

func (t *ReviewTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if t.Light {
		return t.lightColor(name)
	}
	return t.darkColor(name)
}

func (t *ReviewTheme) darkColor(name fyne.ThemeColorName) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return color.NRGBA{R: 0x1e, G: 0x1f, B: 0x22, A: 0xff}
	case theme.ColorNameButton:
		return color.NRGBA{R: 0x2d, G: 0x2e, B: 0x33, A: 0xff}
	case theme.ColorNameDisabledButton:
		return color.NRGBA{R: 0x28, G: 0x29, B: 0x2e, A: 0xff}
	case theme.ColorNameForeground:
		return color.NRGBA{R: 0xe8, G: 0xea, B: 0xed, A: 0xff}
	case theme.ColorNameDisabled:
		return color.NRGBA{R: 0x8a, G: 0x8d, B: 0x93, A: 0xff}
	case theme.ColorNamePlaceHolder:
		return color.NRGBA{R: 0x8a, G: 0x8d, B: 0x93, A: 0xff}
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 0x3b, G: 0x82, B: 0xf6, A: 0xff}
	case theme.ColorNameHover:
		return color.NRGBA{R: 0x35, G: 0x36, B: 0x3c, A: 0xff}
	case theme.ColorNameFocus:
		return color.NRGBA{R: 0x3b, G: 0x82, B: 0xf6, A: 0x66}
	case theme.ColorNameSelection:
		return color.NRGBA{R: 0x3b, G: 0x82, B: 0xf6, A: 0x44}
	case theme.ColorNameInputBackground:
		return color.NRGBA{R: 0x25, G: 0x26, B: 0x2b, A: 0xff}
	case theme.ColorNameInputBorder:
		return color.NRGBA{R: 0x3a, G: 0x3b, B: 0x42, A: 0xff}
	case theme.ColorNameMenuBackground:
		return color.NRGBA{R: 0x25, G: 0x26, B: 0x2b, A: 0xff}
	case theme.ColorNameOverlayBackground:
		return color.NRGBA{R: 0x1a, G: 0x1b, B: 0x1e, A: 0xf2}
	case theme.ColorNameSeparator:
		return color.NRGBA{R: 0x3a, G: 0x3b, B: 0x42, A: 0xff}
	case theme.ColorNameShadow:
		return color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x66}
	case theme.ColorNameHeaderBackground:
		return color.NRGBA{R: 0x25, G: 0x26, B: 0x2b, A: 0xff}
	default:
		return theme.DefaultTheme().Color(name, theme.VariantDark)
	}
}

func (t *ReviewTheme) lightColor(name fyne.ThemeColorName) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
	case theme.ColorNameButton:
		return color.NRGBA{R: 0xf6, G: 0xf8, B: 0xfa, A: 0xff}
	case theme.ColorNameDisabledButton:
		return color.NRGBA{R: 0xee, G: 0xf1, B: 0xf4, A: 0xff}
	case theme.ColorNameForeground:
		return color.NRGBA{R: 0x1f, G: 0x23, B: 0x28, A: 0xff}
	case theme.ColorNameDisabled:
		return color.NRGBA{R: 0x65, G: 0x6d, B: 0x76, A: 0xff}
	case theme.ColorNamePlaceHolder:
		return color.NRGBA{R: 0x65, G: 0x6d, B: 0x76, A: 0xff}
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 0x09, G: 0x69, B: 0xda, A: 0xff}
	case theme.ColorNameHover:
		return color.NRGBA{R: 0xea, G: 0xee, B: 0xf2, A: 0xff}
	case theme.ColorNameFocus:
		return color.NRGBA{R: 0x09, G: 0x69, B: 0xda, A: 0x44}
	case theme.ColorNameSelection:
		return color.NRGBA{R: 0x09, G: 0x69, B: 0xda, A: 0x33}
	case theme.ColorNameInputBackground:
		return color.NRGBA{R: 0xf6, G: 0xf8, B: 0xfa, A: 0xff}
	case theme.ColorNameInputBorder:
		return color.NRGBA{R: 0xd0, G: 0xd7, B: 0xde, A: 0xff}
	case theme.ColorNameMenuBackground:
		return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
	case theme.ColorNameOverlayBackground:
		return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xf2}
	case theme.ColorNameSeparator:
		return color.NRGBA{R: 0xd0, G: 0xd7, B: 0xde, A: 0xff}
	case theme.ColorNameShadow:
		return color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x22}
	case theme.ColorNameHeaderBackground:
		return color.NRGBA{R: 0xf6, G: 0xf8, B: 0xfa, A: 0xff}
	default:
		return theme.DefaultTheme().Color(name, theme.VariantLight)
	}
}

func (t *ReviewTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (t *ReviewTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t *ReviewTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		// Keep list rows dense — Fyne List adds this between every item.
		return 1
	case theme.SizeNameInnerPadding:
		return 2
	case theme.SizeNameSeparatorThickness:
		return 1
	case theme.SizeNameText:
		return 13
	case theme.SizeNameCaptionText:
		return 11
	default:
		return theme.DefaultTheme().Size(name)
	}
}

// Diff palette — updated by applyDiffPalette when the color scheme changes.
var (
	colorAddBg    color.Color
	colorAddFg    color.Color
	colorDelBg    color.Color
	colorDelFg    color.Color
	colorHunkBg   color.Color
	colorHunkFg   color.Color
	colorCtxBg    color.Color
	colorLineNoBg color.Color
	colorLineNoFg color.Color
)

func applyDiffPalette(light bool) {
	if light {
		colorAddBg = color.NRGBA{R: 0xda, G: 0xfb, B: 0xe1, A: 0xff}
		colorAddFg = color.NRGBA{R: 0x1a, G: 0x7f, B: 0x37, A: 0xff}
		colorDelBg = color.NRGBA{R: 0xff, G: 0xeb, B: 0xe9, A: 0xff}
		colorDelFg = color.NRGBA{R: 0xcf, G: 0x22, B: 0x2e, A: 0xff}
		colorHunkBg = color.NRGBA{R: 0xdd, G: 0xf4, B: 0xff, A: 0xff}
		colorHunkFg = color.NRGBA{R: 0x09, G: 0x69, B: 0xda, A: 0xff}
		colorCtxBg = color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
		colorLineNoBg = color.NRGBA{R: 0xf6, G: 0xf8, B: 0xfa, A: 0xff}
		colorLineNoFg = color.NRGBA{R: 0x65, G: 0x6d, B: 0x76, A: 0xff}
		return
	}
	colorAddBg = color.NRGBA{R: 0x1a, G: 0x3d, B: 0x2a, A: 0xff}
	colorAddFg = color.NRGBA{R: 0x7e, G: 0xe7, B: 0xa3, A: 0xff}
	colorDelBg = color.NRGBA{R: 0x4a, G: 0x22, B: 0x28, A: 0xff}
	colorDelFg = color.NRGBA{R: 0xff, G: 0xa1, B: 0xa8, A: 0xff}
	colorHunkBg = color.NRGBA{R: 0x1c, G: 0x2d, B: 0x4a, A: 0xff}
	colorHunkFg = color.NRGBA{R: 0x8b, G: 0xb4, B: 0xf8, A: 0xff}
	colorCtxBg = color.NRGBA{R: 0x1e, G: 0x1f, B: 0x22, A: 0xff}
	colorLineNoBg = color.NRGBA{R: 0x25, G: 0x26, B: 0x2b, A: 0xff}
	colorLineNoFg = color.NRGBA{R: 0x6b, G: 0x70, B: 0x7a, A: 0xff}
}

func init() {
	applyDiffPalette(false)
}

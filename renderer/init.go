package renderer

import (
	"golang.org/x/image/font"
)

var MonoFontFace font.Face

func init() {
	// initialize the font face to 96 DPI as a default. RecalculateFontFace should be called
	// once we know the real screen DPI, but that's not known until a size.Event comes in
	// from shiny.
	RecalculateFontFace(96)

}

package renderer

import (
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

// These variables represent a default monospaced font for renderers to use
// and caches of properties that are useful for a renderer. Renderers don't
// *have* to use them, but doing so avoids race conditions with freetype and
// makes it easier to use default implementations built around them, as well
// as ensures consistency between renderers of different languages.
var (
	MonoFontFace, MonoFontFaceBold                                      font.Face
	MonoFontHeight, MonoFontAdvance, MonoFontGlyphWidth, MonoFontAscent fixed.Int26_6
)

func init() {
	// initialize the font face to 96 DPI as a default. RecalculateFontFace should be called
	// once we know the real screen DPI, but that's not known until a size.Event comes in
	// from shiny.
	RecalculateFontFace(96)

}

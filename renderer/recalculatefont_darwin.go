package renderer

import (
	"fmt"

	"github.com/driusan/fonts"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	//	"golang.org/x/image/font/basicfont"
	"os"
)

func RecalculateFontFace(dpi float64) {
	// While shiny correctly reports the PixelsPerPt in Shiny, OS X clearly
	// doesn't expect you to use it, because multiplying it through while
	// parsing fonts results in a constant physical size font on the screen
	// (as it should.) Meanwhile, OS X system preferences don't even refer
	// to resolution, they refer to "more text" and "less text" on the screen.
	//
	// Unfortunately, the DPI that it expects you to parse fonts at is different
	// based on the screen. Retina displays expect it to be parsed at 144 DPI
	// to render correctly (where correctly is defined as "the same as other
	// non-shiny applications"), while normal displays expect it to be parsed at a
	// more standard 96 DPI.
	//
	// Since we don't know what kind of display is being used, as a hack if it's
	// > 120 DPI (half way between 96 and 144), we assume it's a retina display
	// and parse the font at 144 regardless of the reported resolution, otherwise
	// it's parsed at 96 DPI.
	if dpi > 120 {
		dpi = 144
	} else {
		dpi = 96
	}
	ff, err := fonts.Asset("DejaVuSansMono.ttf")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not retrieve font: %s\n", err)
		os.Exit(2)
		return
	}
	ft, err := truetype.Parse(ff)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not parse font: %s\n", err)
		os.Exit(3)
	}

	MonoFontFace = truetype.NewFace(ft,
		&truetype.Options{
			Size:    12,
			DPI:     dpi,
			Hinting: font.HintingNone})
	if MonoFontFace == nil {
		panic("Could not get font face.")
	}
	metrics := MonoFontFace.Metrics()
	MonoFontHeight = metrics.Height
	MonoFontAscent = metrics.Ascent
	MonoFontAdvance, _ = MonoFontFace.GlyphAdvance('M')
	_, MonoFontGlyphWidth, _ = MonoFontFace.GlyphBounds('a')

	// Assume bold has the same metrics and just calculate the face.
	ff, err = fonts.Asset("DejaVuSansMono-Bold.ttf")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not retrieve font: %s\n", err)
		os.Exit(2)
		return
	}
	ft, err = truetype.Parse(ff)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not parse font: %s\n", err)
		os.Exit(3)
	}
	MonoFontFaceBold = truetype.NewFace(ft,
		&truetype.Options{
			Size:    12,
			DPI:     dpi,
			Hinting: font.HintingNone})
	if MonoFontFaceBold == nil {
		panic("Could not get bold font face.")
	}

}

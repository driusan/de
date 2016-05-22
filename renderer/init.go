package renderer

import (
	"fmt"
	"github.com/driusan/fonts"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	//	"golang.org/x/image/font/basicfont"
	"os"
)

var MonoFontFace font.Face

func RecalculateFontFace(dpi float32) {
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
			Size:    float64(16),
			DPI:     72,
			Hinting: font.HintingNone})
	if MonoFontFace == nil {
		panic("Could not get font face.")
	}
}
func init() {
	// initialize the font face to 72 DPI as a default. RecalculateFontFace should be called
	// once we know the real screen DPI, but that's not known until a size.Event comes in
	// from shiny.
	RecalculateFontFace(72.0)

}

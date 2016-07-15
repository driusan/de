// +build !darwin

package renderer

import (
	"fmt"
	"os"
	//"io/ioutil"

	"github.com/driusan/fonts"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	//	"golang.org/x/image/font/basicfont"
)

func RecalculateFontFace(dpi float64) {
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

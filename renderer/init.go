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

func init() {
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
	// There seems to be a bug where DejaVuSansMono with hinting won't render
	// the character '2', so for now just use the built in basicfont, even though
	// it's not as pretty and doesn't have as many runes.
	//MonoFontFace = basicfont.Face7x13
}

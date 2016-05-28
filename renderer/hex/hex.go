package hex

import (
	"fmt"
	"github.com/driusan/de/demodel"
	"github.com/driusan/de/renderer"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
	"image"
	"image/color"
)

type HexRenderer struct{}

func (r *HexRenderer) InvalidateCache() {
}
func (r HexRenderer) CanRender(buf *demodel.CharBuffer) bool {
	return true
}
func (r HexRenderer) Bounds(buf *demodel.CharBuffer) image.Rectangle {
	// We can calculate the bounds for the hex editor mode easily. We know that there's
	// 1. 8 characters to represent the line addres in hex, followed by a space.
	// 2. 16 byte, each represented by two characters and every 2 followed by a space.
	//    for a total of 40 characters. (or 49 total)
	// 3. 1 character per byte of the 16 bytes (65 total)
	//    So 65*glyphAdvance = Max.X
	// We know that the number of rows is len(buf.Buffer) / 16 + 1, since there's 16
	// bytes represented per line. So numRows * glyphHeight gives us the Max.Y.
	metrics := renderer.MonoFontFace.Metrics()
	numRows := (len(buf.Buffer) / 16) + 1
	glyphSize, _ := renderer.MonoFontFace.GlyphAdvance('M') // arbitrary character
	glyphHeight := metrics.Height.Ceil()
	return image.Rectangle{
		Min: image.ZP,
		Max: image.Point{glyphSize.Ceil() * 65, numRows * glyphHeight},
	}
}

func (r HexRenderer) GetImageMap(buf *demodel.CharBuffer, viewport image.Rectangle) demodel.ImageMap {
	return nil
}
func (r HexRenderer) Render(buf *demodel.CharBuffer, viewport image.Rectangle) (image.Image, error) {

	dst := image.NewRGBA(viewport)
	metrics := renderer.MonoFontFace.Metrics()

	hexwriter := font.Drawer{
		Dst:  dst,
		Src:  &image.Uniform{color.Black},
		Face: renderer.MonoFontFace,
		Dot:  fixed.P(0, metrics.Ascent.Floor()),
	}

	charStart := 500
	charwriter := font.Drawer{
		Dst:  dst,
		Src:  &image.Uniform{color.Black},
		Face: renderer.MonoFontFace,
		Dot:  fixed.P(charStart, metrics.Ascent.Floor()),
	}

	for i, b := range buf.Buffer {
		if i%16 == 0 {
			hexwriter.Dot.X = 0
			charwriter.Dot.X = fixed.I(charStart)

			if i > 0 {
				hexwriter.Dot.Y += metrics.Height
				charwriter.Dot.Y += metrics.Height
			}

			hexwriter.DrawString(fmt.Sprintf("%0.8x ", i))
		}
		hexwriter.DrawString(fmt.Sprintf("%0.2x", b))
		if i%2 == 1 && i > 0 {
			hexwriter.DrawString(" ")
		}
		charwriter.DrawString(string(b))
	}
	return dst, nil
}

package renderer

import (
	"bytes"
	"github.com/driusan/de/demodel"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
	"image"
	"image/color"
	"image/draw"
)

// The default renderer. Performs no syntax highlighting.
type NoSyntaxRenderer struct {
	DefaultSizeCalcer
	DefaultImageMapper
}

func (r *NoSyntaxRenderer) InvalidateCache() {
	r.DefaultSizeCalcer.InvalidateCache()
	r.DefaultImageMapper.InvalidateCache()
}
func (r NoSyntaxRenderer) CanRender(*demodel.CharBuffer) bool {
	return true
}

func (r NoSyntaxRenderer) Render(buf *demodel.CharBuffer, viewport image.Rectangle) (image.Image, error) {
	//dstsize := r.Bounds(buf)
	dst := image.NewRGBA(viewport)
	metrics := MonoFontFace.Metrics()
	writer := font.Drawer{
		Dst:  dst,
		Src:  &image.Uniform{color.Black},
		Face: MonoFontFace,
		Dot:  fixed.P(0, metrics.Ascent.Floor()),
	}
	runes := bytes.Runes(buf.Buffer)

	// it's a monospace font, so only do this once outside of the for loop..
	// use an M so that space characters are based on an em-quad if we change
	// to a non-monospace font.
	_, MglyphWidth, _ := MonoFontFace.GlyphBounds('M')

	for i, r := range runes {
		_, glyphWidth, _ := MonoFontFace.GlyphBounds(r)
		runeRectangle := image.Rectangle{}
		runeRectangle.Min.X = writer.Dot.X.Ceil()
		runeRectangle.Min.Y = writer.Dot.Y.Ceil() - metrics.Ascent.Floor() + 1
		switch r {
		case '\t':
			runeRectangle.Max.X = runeRectangle.Min.X + 8*MglyphWidth.Ceil()
		case '\n':
			runeRectangle.Max.X = viewport.Max.X
		default:
			runeRectangle.Max.X = runeRectangle.Min.X + glyphWidth.Ceil()
		}
		runeRectangle.Max.Y = runeRectangle.Min.Y + metrics.Height.Ceil() + 1

		if runeRectangle.Min.Y > viewport.Max.Y {
			return dst, nil
		}
		if runeRectangle.Intersect(viewport) != image.ZR {
			if uint(i) >= buf.Dot.Start && uint(i) <= buf.Dot.End {
				draw.Draw(
					dst,
					runeRectangle,
					&image.Uniform{TextHighlight},
					image.ZP,
					draw.Over,
				)
			}
		}
		switch r {
		case '\t':
			writer.Dot.X += glyphWidth * 8
			continue
		case '\n':
			writer.Dot.Y += metrics.Height
			writer.Dot.X = 0
			continue
		}
		writer.DrawString(string(r))
	}

	return dst, nil
}

package renderer

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"

	"github.com/driusan/de/demodel"
	"github.com/driusan/de/renderer"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

// The default renderer. Performs no syntax highlighting.
type NoSyntaxRenderer struct {
	renderer.DefaultSizeCalcer
	renderer.DefaultImageMapper
}

func (r *NoSyntaxRenderer) InvalidateCache() {
	r.DefaultSizeCalcer.InvalidateCache()
	r.DefaultImageMapper.InvalidateCache()
}
func (r NoSyntaxRenderer) CanRender(*demodel.CharBuffer) bool {
	return true
}

func (r NoSyntaxRenderer) RenderInto(dst draw.Image, buf *demodel.CharBuffer, viewport image.Rectangle) error {
	bounds := dst.Bounds()
	writer := font.Drawer{
		Dst:  dst,
		Src:  &image.Uniform{color.Black},
		Face: renderer.MonoFontFace,
		Dot:  fixed.P(bounds.Min.X, bounds.Min.Y+renderer.MonoFontAscent.Floor()),
	}
	runes := bytes.Runes(buf.Buffer)

	for i, r := range runes {
		runeRectangle := image.Rectangle{}
		runeRectangle.Min.X = writer.Dot.X.Ceil()
		runeRectangle.Min.Y = writer.Dot.Y.Ceil() - renderer.MonoFontAscent.Floor() + 1
		switch r {
		case '\t':
			runeRectangle.Max.X = runeRectangle.Min.X + 8*renderer.MonoFontGlyphWidth.Ceil()
		case '\n':
			runeRectangle.Max.X = viewport.Max.X
		default:
			runeRectangle.Max.X = runeRectangle.Min.X + renderer.MonoFontGlyphWidth.Ceil()
		}
		runeRectangle.Max.Y = runeRectangle.Min.Y + renderer.MonoFontHeight.Ceil() + 1

		if runeRectangle.Min.Y > viewport.Max.Y {
			return nil
		}
		if runeRectangle.Intersect(viewport) != image.ZR {
			if uint(i) >= buf.Dot.Start && uint(i) <= buf.Dot.End {
				draw.Draw(
					dst,
					image.Rectangle{
						runeRectangle.Min.Sub(viewport.Min),
						runeRectangle.Max.Sub(viewport.Min),
					},
					&image.Uniform{renderer.TextHighlight},
					image.ZP,
					draw.Over,
				)
			}
		}
		switch r {
		case '\t':
			writer.Dot.X += renderer.MonoFontGlyphWidth * 8
			continue
		case '\n':
			writer.Dot.Y += renderer.MonoFontHeight
			writer.Dot.X = 0
			continue
		}

		// hack to draw into the dst using the viewport coordinate system
		writer.Dot.X -= fixed.I(viewport.Min.X)
		writer.Dot.Y -= fixed.I(viewport.Min.Y)
		writer.DrawString(string(r))
		writer.Dot.X += fixed.I(viewport.Min.X)
		writer.Dot.Y += fixed.I(viewport.Min.Y)
	}

	return nil
}

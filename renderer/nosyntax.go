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
	lastBuf     *demodel.CharBuffer
	lastBufSize int
	rSizeCache  image.Rectangle
}

func (r *NoSyntaxRenderer) calcImageSize(buf *demodel.CharBuffer) image.Rectangle {
	if r.rSizeCache != image.ZR && r.lastBuf == buf && r.lastBufSize == len(buf.Buffer) {
		return r.rSizeCache
	}
	r.lastBufSize = len(buf.Buffer)
	r.lastBuf = buf

	metrics := MonoFontFace.Metrics()
	runes := bytes.Runes(buf.Buffer)
	_, glyphWidth, _ := MonoFontFace.GlyphBounds('a')
	rt := image.ZR
	var lineSize fixed.Int26_6
	for _, r := range runes {
		switch r {
		case '\t':
			lineSize += glyphWidth * 8
		case '\n':
			rt.Max.Y += metrics.Height.Ceil()
			lineSize = 0
		default:
			lineSize += glyphWidth
		}
		if lineSize.Ceil() > rt.Max.X {
			rt.Max.X = lineSize.Ceil()
		}
	}
	rt.Max.Y += metrics.Height.Ceil() + 1
	r.rSizeCache = rt
	return rt
}

func (r NoSyntaxRenderer) CanRender(*demodel.CharBuffer) bool {
	return true
}

func (r NoSyntaxRenderer) Render(buf *demodel.CharBuffer, viewport image.Rectangle) (image.Image, image.Rectangle, ImageMap, error) {
	dstsize := r.calcImageSize(buf)
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
	_, glyphWidth, _ := MonoFontFace.GlyphBounds('M')
	im := ImageMap{make([]ImageLoc, 0), buf}
	for i, r := range runes {
		runeRectangle := image.Rectangle{}
		runeRectangle.Min.X = writer.Dot.X.Ceil()
		runeRectangle.Min.Y = writer.Dot.Y.Ceil() - metrics.Ascent.Floor()
		switch r {
		case '\t':
			runeRectangle.Max.X = runeRectangle.Min.X + 8*glyphWidth.Ceil()
		case '\n':
			runeRectangle.Max.X = viewport.Max.X
		default:
			runeRectangle.Max.X = runeRectangle.Min.X + glyphWidth.Ceil()
		}
		runeRectangle.Max.Y = runeRectangle.Min.Y + metrics.Height.Ceil() + 1

		if runeRectangle.Min.Y > viewport.Max.Y {
			return dst, dstsize.Bounds(), im, nil
		}
		if runeRectangle.Intersect(viewport) != image.ZR {
			im.IMap = append(im.IMap, ImageLoc{runeRectangle, uint(i)})
			if uint(i) >= buf.Dot.Start && uint(i) <= buf.Dot.End {
				draw.Draw(
					dst,
					runeRectangle,
					&image.Uniform{TextHighlight},
					image.ZP,
					draw.Src,
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

	return dst, dstsize.Bounds(), im, nil
}

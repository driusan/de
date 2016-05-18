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

// TaglineRenderer is like the default renderer which doesn't
// provide syntax highlighting, but uses a different background
// colour which doesn't change regardless of mode.
type TaglineRenderer struct {
	// The width of the window. The tag will be rendered at least this wide.
	Width int
}

func (r TaglineRenderer) calcImageSize(buf *demodel.CharBuffer) image.Rectangle {
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
	rt.Max.Y += metrics.Height.Ceil() + 5
	if rt.Max.X < r.Width {
		rt.Max.X = r.Width
	}
	return rt
}

func (r TaglineRenderer) CanRender(demodel.CharBuffer) bool {
	return true
}

func (r TaglineRenderer) Render(buf *demodel.CharBuffer) (image.Image, ImageMap, error) {
	dstSize := r.calcImageSize(buf)
	dst := image.NewRGBA(dstSize)
	draw.Draw(dst, dst.Bounds(), &image.Uniform{TaglineBackground}, image.ZP, draw.Src)
	draw.Draw(dst,
		image.Rectangle{
			Min: image.Point{0, dstSize.Max.Y - 1},
			Max: image.Point{dstSize.Max.X, dstSize.Max.Y},
		},
		&image.Uniform{color.Black},
		image.ZP,
		draw.Src,
	)

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
			runeRectangle.Max.X = dstSize.Max.X
		default:
			runeRectangle.Max.X = runeRectangle.Min.X + glyphWidth.Ceil()
		}
		runeRectangle.Max.Y = runeRectangle.Min.Y + metrics.Height.Ceil() + 1

		im.IMap = append(im.IMap, ImageLoc{runeRectangle, uint(i)})
		if buf.Dot.Start != buf.Dot.End && uint(i) >= buf.Dot.Start && uint(i) <= buf.Dot.End {
			writer.Src = &image.Uniform{color.Black}
			draw.Draw(
				dst,
				runeRectangle,
				&image.Uniform{TextHighlight},
				image.ZP,
				draw.Src,
			)
		} else {

			if buf.Dot.Start == buf.Dot.End && buf.Dot.Start == uint(i) {
				// draw a cursor at the current location.
				draw.Draw(dst,
					image.Rectangle{
						image.Point{runeRectangle.Min.X, runeRectangle.Min.Y},
						image.Point{runeRectangle.Min.X + 1, runeRectangle.Max.Y},
					},
					&image.Uniform{color.Black},
					image.ZP,
					draw.Over,
				)
			}

			// it's not within dot, so use a black font on no background.
			writer.Src = &image.Uniform{color.Black}
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

	return dst, im, nil
}

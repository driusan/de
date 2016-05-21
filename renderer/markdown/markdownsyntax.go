package renderer

import (
	"bytes"
	"github.com/driusan/de/demodel"
	//"fmt"
	"github.com/driusan/de/renderer"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
	"image"
	"unicode"
	//	"image/color"
	"image/draw"
	"strings"
)

type MarkdownSyntax struct {
	// cache a copy of what the last buffer size was, so that it can be compared
	// for the calculation in calcImageSize. If the buffer hasn't changed, and
	// the buffer size hasn't changed, we can just return the last rectangle.
	lastBuf     *demodel.CharBuffer
	lastBufSize int
	rSizeCache  image.Rectangle
}

func (rd *MarkdownSyntax) CanRender(buf *demodel.CharBuffer) bool {
	return strings.HasSuffix(buf.Filename, ".md") || strings.HasSuffix(buf.Filename, "COMMIT_EDITMSG")
}
func (rd *MarkdownSyntax) calcImageSize(buf *demodel.CharBuffer) image.Rectangle {
	if rd.rSizeCache != image.ZR && rd.lastBuf == buf && rd.lastBufSize == len(buf.Buffer) {
		// the file isn't being manipulated, just viewed, so it shouldn't
		// have changed size. Just use the cache.
		return rd.rSizeCache
	} else {
		rd.lastBuf = buf
		rd.lastBufSize = len(buf.Buffer)
	}
	metrics := renderer.MonoFontFace.Metrics()
	runes := bytes.Runes(buf.Buffer)
	_, MglyphWidth, _ := renderer.MonoFontFace.GlyphBounds('M')
	rt := image.ZR
	var lineSize fixed.Int26_6
	for _, r := range runes {
		_, glyphWidth, _ := renderer.MonoFontFace.GlyphBounds(r)
		switch r {
		case '\t':
			lineSize += MglyphWidth * 8
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
	rd.rSizeCache = rt
	return rt
}

func (rd *MarkdownSyntax) Render(buf *demodel.CharBuffer, viewport image.Rectangle) (image.Image, image.Rectangle, renderer.ImageMap, error) {
	dstSize := rd.calcImageSize(buf)
	dst := image.NewRGBA(viewport)
	metrics := renderer.MonoFontFace.Metrics()
	writer := font.Drawer{
		Dst:  dst,
		Src:  &image.Uniform{renderer.TextColour},
		Dot:  fixed.P(0, metrics.Ascent.Floor()),
		Face: renderer.MonoFontFace,
	}
	runes := bytes.Runes(buf.Buffer)

	im := renderer.ImageMap{make([]renderer.ImageLoc, 0), buf}

	// Used for calculating the size of a tab.
	_, MglyphWidth, _ := renderer.MonoFontFace.GlyphBounds('M')

	var nextColor image.Image
	// the beginning of a file is the start of the first line..
	lineStart := true

	for i, r := range runes {
		// Do this inside the loop anyways, in case someone changes it to a
		// variable width font..
		_, glyphWidth, _ := renderer.MonoFontFace.GlyphBounds(r)
		switch r {
		case '\n':
			lineStart = true
			writer.Src = &image.Uniform{renderer.TextColour}

		default:
			if lineStart {

				switch r {
				case '#':
					// heading
					writer.Src = &image.Uniform{renderer.CommentColour}
				case '*', '-', '+':
					// lists
					if i < len(runes)-1 && unicode.IsSpace(runes[i+1]) {

						writer.Src = &image.Uniform{renderer.KeywordColour}
						nextColor = &image.Uniform{renderer.TextColour}
					}
				default:
					// the \n already reset it, no need to do this.
					//writer.Src = &image.Uniform{renderer.TextColour}
				}

				lineStart = false
			}
		}

		runeRectangle := image.Rectangle{}
		runeRectangle.Min.X = writer.Dot.X.Ceil()
		runeRectangle.Min.Y = writer.Dot.Y.Ceil() - metrics.Ascent.Floor()

		if runeRectangle.Min.Y > viewport.Max.Y {
			// no point in rendering past the end of the viewport
			return dst, dstSize.Bounds(), im, nil
		}
		switch r {
		case '\t':
			runeRectangle.Max.X = runeRectangle.Min.X + 8*MglyphWidth.Ceil()
		case '\n':
			runeRectangle.Max.X = viewport.Max.X
		default:
			runeRectangle.Max.X = runeRectangle.Min.X + glyphWidth.Ceil()
		}
		runeRectangle.Max.Y = runeRectangle.Min.Y + metrics.Height.Ceil() + 1

		if runeRectangle.Intersect(viewport) != image.ZR {
			im.IMap = append(im.IMap, renderer.ImageLoc{runeRectangle, uint(i)})
			if uint(i) >= buf.Dot.Start && uint(i) <= buf.Dot.End {
				// it's in dot, so highlight the background
				draw.Draw(
					dst,
					runeRectangle,
					&image.Uniform{renderer.TextHighlight},
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

		if nextColor != nil {
			writer.Src = nextColor
			nextColor = nil
		}
	}

	return dst, dstSize.Bounds(), im, nil
}

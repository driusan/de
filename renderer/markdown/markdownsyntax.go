package renderer

import (
	"bytes"
	"image"
	"image/draw"
	"strings"
	"unicode"

	"github.com/driusan/de/demodel"
	"github.com/driusan/de/renderer"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

type MarkdownSyntax struct {
	renderer.DefaultSizeCalcer
	renderer.DefaultImageMapper
}

func (rd *MarkdownSyntax) InvalidateCache() {
	rd.DefaultImageMapper.InvalidateCache()
	rd.DefaultSizeCalcer.InvalidateCache()
}

func (rd *MarkdownSyntax) CanRender(buf *demodel.CharBuffer) bool {
	return strings.HasSuffix(buf.Filename, ".md") || strings.HasSuffix(buf.Filename, "COMMIT_EDITMSG")
}

func (rd *MarkdownSyntax) RenderInto(dst draw.Image, buf *demodel.CharBuffer, viewport image.Rectangle) error {
	bounds := dst.Bounds()
	writer := font.Drawer{
		Dst:  dst,
		Src:  &image.Uniform{renderer.TextColour},
		Dot:  fixed.P(bounds.Min.X, bounds.Min.Y+renderer.MonoFontAscent.Floor()),
		Face: renderer.MonoFontFace,
	}
	runes := bytes.Runes(buf.Buffer)

	var nextColor image.Image
	// the beginning of a file is the start of the first line..
	lineStart := true

	for i, r := range runes {
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
		runeRectangle.Min.Y = writer.Dot.Y.Ceil() - renderer.MonoFontAscent.Floor() + 1

		if runeRectangle.Min.Y > viewport.Max.Y {
			// no point in rendering past the end of the viewport
			return nil
		}
		switch r {
		case '\t':
			runeRectangle.Max.X = runeRectangle.Min.X + 8*renderer.MonoFontGlyphWidth.Ceil()
		case '\n':
			runeRectangle.Max.X = viewport.Max.X
		default:
			runeRectangle.Max.X = runeRectangle.Min.X + renderer.MonoFontGlyphWidth.Ceil()
		}
		runeRectangle.Max.Y = runeRectangle.Min.Y + renderer.MonoFontHeight.Ceil() + 1

		if runeRectangle.Intersect(viewport) != image.ZR {
			if uint(i) >= buf.Dot.Start && uint(i) <= buf.Dot.End {
				// it's in dot, so highlight the background
				draw.Draw(
					dst,
					image.Rectangle{
						runeRectangle.Min.Sub(viewport.Min),
						runeRectangle.Max.Sub(viewport.Min),
					},

					&image.Uniform{renderer.TextHighlight},
					image.ZP,
					draw.Src,
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
		// translate from viewport coordinate system to dst coordinate system
		// for drawing.
		writer.Dot.X -= fixed.I(viewport.Min.X)
		writer.Dot.Y -= fixed.I(viewport.Min.Y)
		writer.DrawString(string(r))
		writer.Dot.X += fixed.I(viewport.Min.X)
		writer.Dot.Y += fixed.I(viewport.Min.Y)

		if nextColor != nil {
			writer.Src = nextColor
			nextColor = nil
		}
	}

	return nil
}

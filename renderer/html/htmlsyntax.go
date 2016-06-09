package htmlrenderer

import (
	"bytes"
	"github.com/driusan/de/demodel"
	"unicode"
	//"fmt"
	"github.com/driusan/de/renderer"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
	"image"
	//	"image/color"
	"image/draw"
	"strings"
)

type HtmlSyntax struct {
	renderer.DefaultSizeCalcer
	renderer.DefaultImageMapper
}

func (rd *HtmlSyntax) InvalidateCache() {
	rd.DefaultSizeCalcer.InvalidateCache()
	rd.DefaultImageMapper.InvalidateCache()
}
func (rd *HtmlSyntax) CanRender(buf *demodel.CharBuffer) bool {
	return strings.HasSuffix(buf.Filename, ".css") || strings.HasSuffix(buf.Filename, ".html")
}

func (rd *HtmlSyntax) Render(buf *demodel.CharBuffer, viewport image.Rectangle) (image.Image, error) {
	dst := image.NewRGBA(viewport)
	metrics := renderer.MonoFontFace.Metrics()
	writer := font.Drawer{
		Dst:  dst,
		Src:  &image.Uniform{renderer.TextColour},
		Dot:  fixed.P(0, metrics.Ascent.Floor()),
		Face: renderer.MonoFontFace,
	}
	runes := bytes.Runes(buf.Buffer)

	// Used for calculating the size of a tab.
	_, MglyphWidth, _ := renderer.MonoFontFace.GlyphBounds('M')

	var inTag, inStringDQ, inStringSQ bool
	// Some characters (like a terminating quote) only change the active colour
	//after being rendered.
	var nextColor image.Image
	for i, r := range runes {
		inString := inStringDQ || inStringSQ
		// Do this inside the loop anyways, in case someone changes it to a
		// variable width font..
		_, glyphWidth, _ := renderer.MonoFontFace.GlyphBounds(r)
		switch r {
		case '"':
			if inStringSQ {
				// do nothing if a " is embedded in a '
			} else if inStringDQ {
				// if the " is inside a tag, it ends a string and goes back
				// to the attribute colour. If it's not in a tag, just keep
				// the default text colour
				if inTag {
					nextColor = &image.Uniform{renderer.AttributeColour}
				} else {
					nextColor = &image.Uniform{renderer.TextColour}
				}
				inStringDQ = false
			} else {
				// the " starts a string, but only if it's inside a tag.
				if inTag == true {
					writer.Src = &image.Uniform{renderer.StringColour}
					inStringDQ = true
				}
			}
		case '\'':
			if inStringDQ {
				// do nothing if a ' is embedded in a "
			} else if inStringSQ {
				if inTag {
					nextColor = &image.Uniform{renderer.AttributeColour}
				} else {
					nextColor = &image.Uniform{renderer.TextColour}
				}
				inStringSQ = false
			} else {
				if inTag == true {
					writer.Src = &image.Uniform{renderer.StringColour}
					inStringSQ = true
				}
			}
		case '=':
			if inTag && !inString {
				writer.Src = &image.Uniform{renderer.OperatorColour}
			}
		case '<':
			if inTag == false && !inString {
				writer.Src = &image.Uniform{renderer.TagDelimitorColour}
				inTag = true
				nextColor = &image.Uniform{renderer.TagColour}
			}
		case '>':
			if inTag == true && !inString {
				writer.Src = &image.Uniform{renderer.TagDelimitorColour}
				nextColor = &image.Uniform{renderer.TextColour}
				inTag = false
			}
		}

		if unicode.IsSpace(r) {
			if inTag && !inString {
				writer.Src = &image.Uniform{renderer.AttributeColour}
			}
		}
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
			// exit the loop early, we've already gotten past the part that we care about.
			return dst, nil
		}

		// Don't draw or calculate the image map if we're outside of the viewport. We can't
		// break out, because things not being drawn might still affect the rendering (ie.
		// the start of the screen might be in the middle of a comment that needs to be syntax
		// highlighted)
		//	im.IMap = append(im.IMap, renderer.ImageLoc{runeRectangle, uint(i)})
		if runeRectangle.Intersect(viewport) != image.ZR {

			if uint(i) >= buf.Dot.Start && uint(i) <= buf.Dot.End {
				// it's in dot, so highlight the background (unless it's outside of the viewport
				// clipping rectangle)
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

	return dst, nil
}

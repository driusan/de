package shell

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"os"

	"github.com/driusan/de/demodel"
	"github.com/driusan/de/renderer"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

// colours to use for ANSI escape codes
var (
	ansiBlack   image.Image = &image.Uniform{renderer.TextColour}
	ansiRed     image.Image = &image.Uniform{color.RGBA{0x80, 0, 0, 0xff}}
	ansiGreen   image.Image = &image.Uniform{color.RGBA{0, 0x80, 0, 0xff}}
	ansiYellow  image.Image = &image.Uniform{color.RGBA{0x80, 0x80, 0, 0xff}}
	ansiBlue    image.Image = &image.Uniform{color.RGBA{0, 0, 0x80, 0xff}}
	ansiMagenta image.Image = &image.Uniform{color.RGBA{0x80, 0, 0x80, 0xff}}
	ansiCyan    image.Image = &image.Uniform{color.RGBA{0, 0x80, 0x80, 0xff}}
	ansiGray    image.Image = &image.Uniform{color.RGBA{0x80, 0x80, 0x80, 0xff}}
)

func init() {
	renderer.RegisterRenderer("xterm", &TerminalRenderer{})
}

type TerminalRenderer struct {
	renderer.DefaultSizeCalcer
	renderer.DefaultImageMapper
}

func (rd *TerminalRenderer) InvalidateCache() {
	rd.DefaultSizeCalcer.InvalidateCache()
	rd.DefaultImageMapper.InvalidateCache()
}

func (rd *TerminalRenderer) CanRender(buf *demodel.CharBuffer) bool {
	return true
}

func (rd *TerminalRenderer) RenderInto(dst draw.Image, buf *demodel.CharBuffer, viewport image.Rectangle) error {
	metrics := renderer.MonoFontFace.Metrics()
	bounds := dst.Bounds()
	var foreground, background color.Color = renderer.TextColour, renderer.NormalBackground
	writer := font.Drawer{
		Dst:  dst,
		Src:  &image.Uniform{foreground},
		Dot:  fixed.P(bounds.Min.X, bounds.Min.Y+metrics.Ascent.Floor()),
		Face: renderer.MonoFontFace,
	}
	runes := bytes.Runes(buf.Buffer)

	// Used for calculating the size of a tab.
	_, MglyphWidth, _ := renderer.MonoFontFace.GlyphBounds('M')

	escapeStart := -1
	for i, r := range runes {
		// Do this inside the loop anyways, in case someone changes it to a
		// variable width font..
		_, glyphWidth, _ := renderer.MonoFontFace.GlyphBounds(r)
		switch r {
		case 0x1B:
			escapeStart = i
			continue
		case 0x8:
			// Backspace. Just move the cursor backspace, because
			// backspace and writing over it seems to be how OpenBSD
			// indicates bold in its man pages.
			writer.Dot.X -= glyphWidth
			continue
		}
		if escapeStart >= 0 {
			switch runes[escapeStart+1] {
			case ']':
				switch r {
				case '\b', '\007', '\233':
					escapeStart = -1
				}
				continue
			case '[':
				if r >= 64 && r <= 126 && i != escapeStart+1 {
					switch r {
					case 'm':
						// escapeStart is the \033 character,
						// +1 is the [, after that is what
						// we care about.
						args := string(runes[escapeStart+2 : i])

						switch args {
						case "0", "":
							// Reset everything
							writer.Face = renderer.MonoFontFace
							background = renderer.NormalBackground
							writer.Src = ansiBlack
						case "1":
							writer.Face = renderer.MonoFontFaceBold
						case "22":
							writer.Face = renderer.MonoFontFace
						case "30", "39":
							writer.Src = ansiBlack
						case "31":
							writer.Src = ansiRed
						case "32":
							writer.Src = ansiGreen
						case "33":
							writer.Src = ansiYellow
						case "34":
							writer.Src = ansiBlue
						case "35":
							writer.Src = ansiMagenta
						case "36":
							writer.Src = ansiCyan
						case "37":
							writer.Src = ansiGray
						case "38":
							fmt.Fprintf(os.Stderr, "ANSI extended colours not implemented\n")
						}
					default:
						fmt.Fprintf(os.Stderr, "ANSI command sequence %c not yet implemented.\n", r)
					}
					escapeStart = -1
				}
				continue

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
		case 27: // ESC
			continue
		default:
			runeRectangle.Max.X = runeRectangle.Min.X + glyphWidth.Ceil()
		}
		runeRectangle.Max.Y = runeRectangle.Min.Y + metrics.Height.Ceil() + 1

		if runeRectangle.Min.Y > viewport.Max.Y {
			// exit the loop early, we've already gotten past the part that we care about.
			return nil
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
					image.Rectangle{
						runeRectangle.Min.Sub(viewport.Min),
						runeRectangle.Max.Sub(viewport.Min)},
					&image.Uniform{renderer.TextHighlight},
					image.ZP,
					draw.Src,
				)
			} else {
				draw.Draw(
					dst,
					image.Rectangle{
						runeRectangle.Min.Sub(viewport.Min),
						runeRectangle.Max.Sub(viewport.Min)},
					&image.Uniform{background},
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
			writer.Dot.X = fixed.I(bounds.Min.X)
			continue
		}

		if runeRectangle.Max.Y > viewport.Min.Y {
			// as a hack to align viewport.Min with dst.Min, we subtract it from
			// Dot before drawing, then add it back.
			writer.Dot.X -= fixed.I(viewport.Min.X)
			writer.Dot.Y -= fixed.I(viewport.Min.Y)
			writer.DrawString(string(r))
			writer.Dot.X += fixed.I(viewport.Min.X)
			writer.Dot.Y += fixed.I(viewport.Min.Y)
		}
	}

	return nil
}

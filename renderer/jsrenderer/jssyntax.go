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

type JSSyntax struct {
	renderer.DefaultSizeCalcer
	renderer.DefaultImageMapper
}

func (rd *JSSyntax) InvalidateCache() {
	rd.DefaultSizeCalcer.InvalidateCache()
	rd.DefaultImageMapper.InvalidateCache()
}
func (rd *JSSyntax) CanRender(buf *demodel.CharBuffer) bool {
	return strings.HasSuffix(buf.Filename, ".js")
}

func (rd *JSSyntax) RenderInto(dst draw.Image, buf *demodel.CharBuffer, viewport image.Rectangle) error {
	bounds := dst.Bounds()
	writer := font.Drawer{
		Dst:  dst,
		Src:  &image.Uniform{renderer.TextColour},
		Dot:  fixed.P(bounds.Min.X, bounds.Min.Y+renderer.MonoFontAscent.Floor()),
		Face: renderer.MonoFontFace,
	}
	runes := bytes.Runes(buf.Buffer)

	var inLineComment, inMultilineComment, inString, inCharString, inStringLiteral bool

	// Some characters (like a terminating quote) only change the active colour
	//after being rendered.
	var nextColor image.Image
	for i, r := range runes {
		if inStringLiteral {
			if r == '`' {
				nextColor = &image.Uniform{renderer.TextColour}
				inStringLiteral = false
			}
		} else {
			switch r {
			case '\n':
				if inLineComment && !inMultilineComment && !inString {
					inLineComment = false
					writer.Src = &image.Uniform{renderer.TextColour}
				}
			case '\'':
				if !IsEscaped(i, runes) {
					if inCharString {
						// end of a string, colourize the quote too.
						nextColor = &image.Uniform{renderer.TextColour}
						inCharString = false
					} else if !inLineComment && !inMultilineComment && !inString && !inStringLiteral {
						inCharString = true
						writer.Src = &image.Uniform{renderer.StringColour}
					}
				}
			case '"':
				if !IsEscaped(i, runes) {
					if inString {
						inString = false
						nextColor = &image.Uniform{renderer.TextColour}
					} else if !inLineComment && !inMultilineComment && !inCharString && !inStringLiteral {
						inString = true
						writer.Src = &image.Uniform{renderer.StringColour}
					}
				}
			case '`':
				// \ doesn't mean anything special inside of a string literal in Go. Don't check if it's
				// escaped.
				//if !IsEscaped(i, runes) {
				if inStringLiteral {
					inStringLiteral = false
					nextColor = &image.Uniform{renderer.TextColour}
				} else if !inLineComment && !inMultilineComment && !inCharString && !inString {
					inStringLiteral = true
					writer.Src = &image.Uniform{renderer.StringColour}
				}
				//}
			case '/':
				if i+2 < len(runes) && string(runes[i:i+2]) == "//" {
					if !inCharString && !inMultilineComment && !inString {
						inLineComment = true
						writer.Src = &image.Uniform{renderer.CommentColour}
					}
				} else if i+2 < len(runes) && string(runes[i:i+2]) == "/*" {
					if !inCharString && !inString {
						inMultilineComment = true
						writer.Src = &image.Uniform{renderer.CommentColour}
					}
				}
				if i > 1 && inMultilineComment && i+1 < len(runes) && string(runes[i-1:i+1]) == "*/" {
					nextColor = &image.Uniform{renderer.TextColour}
					inMultilineComment = false
				}
			case ' ', '\t':
				if !inCharString && !inMultilineComment && !inString && !inLineComment {
					writer.Src = &image.Uniform{renderer.TextColour}
				}
			default:
				if !inCharString && !inMultilineComment && !inString && !inLineComment && !inStringLiteral {
					if IsLanguageKeyword(i, runes) {
						writer.Src = &image.Uniform{renderer.KeywordColour}
					} else if IsLanguageType(i, runes) {
						writer.Src = &image.Uniform{renderer.BuiltinTypeColour}
					} else if StartsLanguageDeliminator(r) {
						writer.Src = &image.Uniform{renderer.TextColour}
					}
				}
			}
		}

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
			}
		}

		switch r {
		case '\t':
			writer.Dot.X += renderer.MonoFontGlyphWidth * 8
			continue
		case '\n':
			writer.Dot.Y += renderer.MonoFontHeight
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

		if nextColor != nil {
			writer.Src = nextColor
			nextColor = nil
		}
	}

	return nil
}

func StartsLanguageDeliminator(r rune) bool {
	switch r {
	case '+', '-', '*', '/', '%',
		'&', '|', '^',
		'<', '>', '=', '!',
		':', '.',
		'(', ')', '[', ']', '{', '}',
		',', ';':
		return true
	}
	return unicode.IsSpace(r)
}
func IsLanguageKeyword(pos int, runes []rune) bool {
	if pos > 0 {
		prev := runes[pos-1]
		if !unicode.IsSpace(prev) && !StartsLanguageDeliminator(prev) {
			return false
		}
	}
	keywords := []string{
		"abstract",
		"arguments",
		"boolean",
		"break",
		"byte",
		"case",
		"catch",
		"char",
		"class",
		"const",
		"continue",
		"debugger",
		"default",
		"delete",
		"do",
		"double",
		"else",
		"enum",
		"eval",
		"export",
		"extends",
		"false",
		"final",
		"finally",
		"float",
		"for",
		"function",
		"goto",
		"if",
		"implements",
		"import",
		"in",
		"instanceof",
		"int",
		"interface",
		"let",
		"long",
		"native",
		"new",
		"null",
		"package",
		"private",
		"protected",
		"public",
		"return",
		"short",
		"static",
		"super",
		"switch",
		"synchronized",
		"this",
		"throw",
		"throws",
		"transient",
		"true",
		"try",
		"typeof",
		"var",
		"void",
		"volatile",
		"while",
		"with",
		"yield",
	}

	for _, keyword := range keywords {
		keyLen := len(keyword)
		// Guard against an out of bounds exception at the end of the file
		if len(runes) > pos+keyLen+1 {

			if unicode.IsSpace(runes[pos+keyLen]) ||
				StartsLanguageDeliminator(runes[pos+keyLen]) {
				if string(runes[pos:pos+keyLen]) == keyword {
					return true
				}
			}

		}
	}
	return false
}
func IsLanguageType(pos int, runes []rune) bool {
	if pos < 3 {
		return false

	}
	if !StartsLanguageDeliminator(runes[pos-1]) {
		return false
	}
	if len(runes) > pos+4 {
		if StartsLanguageDeliminator(runes[pos+3]) {
			switch string(runes[pos : pos+3]) {
			case "int":
				return true
			}
		}
	}
	if len(runes) > pos+5 {
		if StartsLanguageDeliminator(runes[pos+4]) {
			switch string(runes[pos : pos+4]) {
			case "int8", "bool", "byte", "rune", "uint":
				return true
			}
		}

	}
	if len(runes) > pos+6 {
		if StartsLanguageDeliminator(runes[pos+5]) {
			switch string(runes[pos : pos+5]) {
			case "uint8", "int16", "int32", "int64", "error":
				return true
			}
		}
	}
	if len(runes) > pos+7 {
		if StartsLanguageDeliminator(runes[pos+6]) {
			switch string(runes[pos : pos+6]) {
			case "uint16", "uint32", "uint64", "string":
				return true
			}
		}
	}
	return false
}
func IsEscaped(pos int, runes []rune) bool {
	if pos == 0 {
		return false
	}

	isEscaped := false
	for i := pos - 1; i >= 0 && runes[i] == '\\'; i-- {
		isEscaped = !isEscaped
	}
	return isEscaped
}

package renderer

import (
	"bytes"
	"github.com/driusan/de/demodel"
	"unicode"
	//"fmt"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
	"image"
	//	"image/color"
	"image/draw"
)

type GoSyntax struct{}

func (rd *GoSyntax) calcImageSize(buf demodel.CharBuffer) image.Rectangle {
	metrics := MonoFontFace.Metrics()
	runes := bytes.Runes(buf.Buffer)
	_, MglyphWidth, _ := MonoFontFace.GlyphBounds('M')
	rt := image.ZR
	var lineSize fixed.Int26_6
	for _, r := range runes {
		_, glyphWidth, _ := MonoFontFace.GlyphBounds(r)
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
	return rt
}

func (rd *GoSyntax) Render(buf demodel.CharBuffer) (image.Image, ImageMap, error) {
	dstSize := rd.calcImageSize(buf)
	dst := image.NewRGBA(dstSize)
	metrics := MonoFontFace.Metrics()
	writer := font.Drawer{
		Dst:  dst,
		Src:  &image.Uniform{TextColour},
		Dot:  fixed.P(0, metrics.Ascent.Floor()),
		Face: MonoFontFace,
	}
	runes := bytes.Runes(buf.Buffer)

	// it's a monospace font, so only do this once outside of the for loop..
	// use an M so that space characters are based on an em-quad if we change
	// to a non-monospace font.
	//writer.Src = &image.Uniform{TextColour}
	im := make(ImageMap, 0)

	var inLineComment, inMultilineComment, inString, inCharString bool

	_, MglyphWidth, _ := MonoFontFace.GlyphBounds('M')
	var nextColor image.Image
	for i, r := range runes {
		_, glyphWidth, _ := MonoFontFace.GlyphBounds(r)
		switch r {
		case '\n':
			if inLineComment && !inMultilineComment && !inString {
				inLineComment = false
				writer.Src = &image.Uniform{TextColour}
			}
		case '\'':
			if !IsEscaped(i, runes) {
				if inCharString {
					// end of a string, colourize the quote too.
					nextColor = &image.Uniform{TextColour}
					inCharString = false
				} else if !inLineComment && !inMultilineComment && !inString {
					inCharString = true
					writer.Src = &image.Uniform{StringColour}
				}
			}
		case '"':
			if !IsEscaped(i, runes) {
				if inString {
					inString = false
					nextColor = &image.Uniform{TextColour}
				} else if !inLineComment && !inMultilineComment && !inCharString {
					inString = true
					writer.Src = &image.Uniform{StringColour}
				}
			}
		case '/':
			if string(runes[i:i+2]) == "//" {
				if !inCharString && !inMultilineComment && !inString {
					inLineComment = true
					writer.Src = &image.Uniform{CommentColour}
				}
			}
		case ' ', '\t':
			if !inCharString && !inMultilineComment && !inString && !inLineComment {
				writer.Src = &image.Uniform{TextColour}
			}
		default:
			if !inCharString && !inMultilineComment && !inString && !inLineComment {
				if IsLanguageKeyword(i, runes) {
					writer.Src = &image.Uniform{KeywordColour}
				} else if IsLanguageType(i, runes) {
					writer.Src = &image.Uniform{BuiltinTypeColour}
				} else if StartsLanguageDeliminator(r) {
					writer.Src = &image.Uniform{TextColour}
				}
			}
		}

		runeRectangle := image.Rectangle{}
		runeRectangle.Min.X = writer.Dot.X.Ceil()
		runeRectangle.Min.Y = writer.Dot.Y.Ceil() - metrics.Ascent.Floor()
		switch r {
		case '\t':
			runeRectangle.Max.X = runeRectangle.Min.X + 8*MglyphWidth.Ceil()
		case '\n':
			runeRectangle.Max.X = dstSize.Max.X
		default:
			runeRectangle.Max.X = runeRectangle.Min.X + glyphWidth.Ceil()
		}
		runeRectangle.Max.Y = runeRectangle.Min.Y + metrics.Height.Ceil() + 1

		im = append(im, ImageLoc{runeRectangle, uint(i)})
		if uint(i) >= buf.Dot.Start && uint(i) <= buf.Dot.End {
			// it's in dot, so highlight the background
			draw.Draw(
				dst,
				runeRectangle,
				&image.Uniform{TextHighlight},
				image.ZP,
				draw.Src,
			)
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

	return dst, im, nil
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
	if unicode.IsSpace(r) {
		return true
	}
	return false
}
func IsLanguageKeyword(pos int, runes []rune) bool {
	if pos > 0 {
		prev := runes[pos-1]
		if !unicode.IsSpace(prev) && !StartsLanguageDeliminator(prev) {
			return false
		}
	}
	if len(runes) > pos+12 {
		if unicode.IsSpace(runes[pos+11]) || StartsLanguageDeliminator(runes[pos+11]) {
			switch string(runes[pos : pos+11]) {
			case "fallthrough":
				return true
			}
		}
	}
	if len(runes) > pos+9 {
		if unicode.IsSpace(runes[pos+8]) || StartsLanguageDeliminator(runes[pos+8]) {
			switch string(runes[pos : pos+8]) {
			case "continue":
				return true
			}
		}
	}
	if len(runes) > pos+8 {
		if unicode.IsSpace(runes[pos+7]) || StartsLanguageDeliminator(runes[pos+7]) {
			switch string(runes[pos : pos+7]) {
			case "default", "package":
				return true
			}

		}
	}
	if len(runes) > pos+7 {
		if unicode.IsSpace(runes[pos+6]) || StartsLanguageDeliminator(runes[pos+6]) {
			switch string(runes[pos : pos+6]) {
			case "import", "return", "select", "struct", "switch":
				return true
			}
		}
	}
	if len(runes) > pos+6 {
		if unicode.IsSpace(runes[pos+5]) || StartsLanguageDeliminator(runes[pos+5]) {
			switch string(runes[pos : pos+5]) {
			case "break", "const", "defer", "range", "false":
				return true
			}
		}
	}
	if len(runes) > pos+5 {
		if unicode.IsSpace(runes[pos+4]) || StartsLanguageDeliminator(runes[pos+4]) {
			switch string(runes[pos : pos+4]) {
			case "case", "chan", "else", "func", "goto", "type", "true", "iota":
				return true
			}
		}
	}
	if len(runes) > pos+4 {
		if unicode.IsSpace(runes[pos+3]) || StartsLanguageDeliminator(runes[pos+3]) {
			switch string(runes[pos : pos+3]) {
			case "for", "map", "var":
				return true
			}
		}
	}
	if len(runes) > pos+3 {
		if unicode.IsSpace(runes[pos+2]) || StartsLanguageDeliminator(runes[pos+2]) {
			switch string(runes[pos : pos+2]) {
			case "if", "go":
				return true
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
		if unicode.IsSpace(runes[pos+5]) {
			switch string(runes[pos : pos+5]) {
			case "uint8", "int16", "int32", "int64":
				return true
			}
		}
	}
	if len(runes) > pos+7 {
		if unicode.IsSpace(runes[pos+6]) {
			switch string(runes[pos : pos+6]) {
			case "uint16", "uint32", "uint64":
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

package position

import (
	"errors"
	//"fmt"
	"github.com/driusan/de/demodel"
	"unicode"
)

var Invalid error = errors.New("Invalid Position")

// DotStart can be used as a demodel.Position argument
// to actions
func DotStart(buff demodel.CharBuffer) (uint, error) {
	return buff.Dot.Start, nil
}

// DotEnd can be used as a demodel.Position argument
// to actions
func DotEnd(buff demodel.CharBuffer) (uint, error) {
	return buff.Dot.End, nil
}

// DotStart can be used as a demodel.Position argument
// to actions
func AltDotStart(buff demodel.CharBuffer) (uint, error) {
	return buff.AltDot.Start, nil
}

// DotEnd can be used as a demodel.Position argument
// to actions
func AltDotEnd(buff demodel.CharBuffer) (uint, error) {
	return buff.AltDot.End, nil
}

func TagDotStart(buff demodel.CharBuffer) (uint, error) {
	if buff.Tagline == nil || len(buff.Tagline.Buffer) == 0 {
		return 0, Invalid
	}
	return buff.Tagline.Dot.Start, nil
}
func TagDotEnd(buff demodel.CharBuffer) (uint, error) {
	if buff.Tagline == nil || len(buff.Tagline.Buffer) == 0 {
		return 0, Invalid
	}
	return buff.Tagline.Dot.End, nil
}

func TagAltDotStart(buff demodel.CharBuffer) (uint, error) {
	if buff.Tagline == nil || len(buff.Tagline.Buffer) == 0 {
		return 0, Invalid
	}
	return buff.Tagline.AltDot.Start, nil
}
func TagAltDotEnd(buff demodel.CharBuffer) (uint, error) {
	if buff.Tagline == nil || len(buff.Tagline.Buffer) == 0 {
		return 0, Invalid
	}
	return buff.Tagline.AltDot.End, nil
}

func BuffStart(buff demodel.CharBuffer) (uint, error) {
	if len(buff.Buffer) > 0 {
		return 0, nil
	}
	return 0, Invalid
}
func BuffEnd(buff demodel.CharBuffer) (uint, error) {
	if len(buff.Buffer) > 0 {
		return uint(len(buff.Buffer) - 1), nil
	}
	return 0, Invalid
}

func PrevChar(buff demodel.CharBuffer) (uint, error) {
	if buff.Dot.Start == 0 {
		return 0, Invalid
	}

	// BUG: This doesn't deal with multibyte UTF-8 runes.
	return buff.Dot.Start - 1, nil
}
func NextChar(buff demodel.CharBuffer) (uint, error) {
	if buff.Dot.End >= uint(len(buff.Buffer)) {
		return uint(len(buff.Buffer)) - 1, Invalid
	}
	return buff.Dot.End + 1, nil
}

func PrevLine(buff demodel.CharBuffer) (uint, error) {
	// how far into the current line is the current character?
	var lineIdx int = -1

	// where does the previous line start, so that we can index
	// lineIdx into it at the end?
	var prevLineStart, curLineStart int = -1, -1
	for i := buff.Dot.Start; i > 0; i-- {
		if uint(len(buff.Buffer)) <= i {
			return buff.Dot.End, Invalid
		}
		if buff.Buffer[i] == '\n' {
			if lineIdx == -1 {
				lineIdx = int(buff.Dot.Start) - int(i)
				curLineStart = int(i)
			} else {
				prevLineStart = int(i)
				// the current line was shorter than lineIdx, so move to the
				// end of the line instead of lineIdx into it.
				if prevLineStart+lineIdx > curLineStart {
					return uint(curLineStart), nil
				}
				break
			}
		}
	}

	// it was the first line, so return the first character
	if prevLineStart < 0 || lineIdx < 0 {
		return 0, nil
	}

	return uint(prevLineStart + lineIdx), nil
}

func NextLine(buff demodel.CharBuffer) (uint, error) {
	// how far into the current line is the current character?
	var lineIdx int = -1

	// calculate how far we are into the current line.
	var nextLineStart, subsequentLine int = -1, -1
	for i := buff.Dot.End; i > 0; i-- {
		if uint(len(buff.Buffer)) <= i {
			return buff.Dot.End, Invalid
		}
		if buff.Buffer[i] == '\n' {
			if lineIdx == -1 {
				lineIdx = int(buff.Dot.End - i)
				break
			}
		}
	}
	if lineIdx < 0 {
		// must be the on the first line, which means Dot
		// is actually the index.
		lineIdx = int(buff.Dot.End)
	}

	// now find the next line start, so we can add to it, and the line
	// after that, so we can check if nextLine+idx goes too far.
	for i := buff.Dot.End + 1; int(i) < len(buff.Buffer); i++ {
		if buff.Buffer[i] == '\n' {
			if nextLineStart == -1 {
				nextLineStart = int(i)
			} else {
				subsequentLine = int(i)
				break
			}
		}
	}

	// must be on the last line, so go to the end of the buffer
	if nextLineStart < 0 {
		return uint(len(buff.Buffer)) - 1, nil
	}

	// calculate the position
	pos := uint(nextLineStart + lineIdx)
	if subsequentLine >= 0 && pos > uint(subsequentLine) {
		// went too far, so return the end of the line
		return uint(subsequentLine) - 1, nil
	}

	if pos >= uint(len(buff.Buffer)) {
		// overflowed the buffer somehow, so return the end of the buffer
		return uint(len(buff.Buffer)) - 1, Invalid
	}

	return pos, nil
}

func EndOfLine(buff demodel.CharBuffer) (uint, error) {
	if len(buff.Buffer) == 0 {
		return 0, Invalid
	}

	if buff.Buffer[buff.Dot.End] == '\n' {
		return buff.Dot.End, nil
	}
	for i := buff.Dot.End; i < uint(len(buff.Buffer)); i++ {
		if buff.Buffer[i] == '\n' {
			return i - 1, nil
		}
	}
	return uint(len(buff.Buffer) - 1), nil
}
func StartOfLine(buff demodel.CharBuffer) (uint, error) {
	if len(buff.Buffer) == 0 {
		return 0, Invalid
	}
	if buff.Dot.Start == 0 {
		return 0, nil
	}

	for i := buff.Dot.Start - 1; i > 0; i-- {
		if buff.Buffer[i] == '\n' {
			return i + 1, nil
		}
	}
	return 0, nil
}

func CurWordStart(buff demodel.CharBuffer) (uint, error) {
	if len(buff.Buffer) == 0 {
		return 0, Invalid
	}
	if buff.Dot.Start == 0 {
		return 0, nil
	}
	for i := buff.Dot.Start - 1; i > 0; i-- {
		if unicode.IsSpace(rune(buff.Buffer[i])) {
			return i + 1, nil
		}
		// some non-space word borders. Note that '.'
		// isn't considered a word border so that opening
		// files by right clicking or hitting enter works.
		switch buff.Buffer[i] {
		case '(', ')', '"', '\'', ',', '/', '[', ']', ';':
			return i + 1, nil
		}

	}
	// no word boundaries found, so the first word is the start..
	return 0, nil
}

func CurWordEnd(buff demodel.CharBuffer) (uint, error) {
	switch len(buff.Buffer) {
	case 0:
		return 0, Invalid
	case 1:
		return 0, nil
	}

	for i := buff.Dot.End; i < uint(len(buff.Buffer)); i++ {
		if unicode.IsSpace(rune(buff.Buffer[i])) {
			return i - 1, nil
		}
		switch buff.Buffer[i] {
		case '(', ')', '"', '\'', ',', '/', '[', ']', ';':
			return i - 1, nil
		}
	}

	// no words found, so the end of the buffer is the end of the word.
	return uint(len(buff.Buffer)) - 1, nil

}
func CurExecutionWordStart(buff demodel.CharBuffer) (uint, error) {
	if len(buff.Buffer) == 0 {
		return 0, Invalid
	}
	if buff.Dot.Start == 0 {
		return 0, nil
	}
	for i := buff.Dot.Start - 1; i > 0; i-- {
		if unicode.IsSpace(rune(buff.Buffer[i])) {
			return i + 1, nil
		}
		// some non-space word borders. Note that '.'
		// isn't considered a word border so that opening
		// files by right clicking or hitting enter works.
		/* switch buff.Buffer[i] {
		case '(', ')', '"', '\'', ',', '/', '[', ']', ';':
			return i + 1, nil
		}*/

	}
	// no word boundaries found, so the first word is the start..
	return 0, nil
}

func CurExecutionWordEnd(buff demodel.CharBuffer) (uint, error) {
	switch len(buff.Buffer) {
	case 0:
		return 0, Invalid
	case 1:
		return 0, nil
	}

	for i := buff.Dot.End; i < uint(len(buff.Buffer)); i++ {
		if unicode.IsSpace(rune(buff.Buffer[i])) {
			return i - 1, nil
		}
		/* switch buff.Buffer[i] {
		case '(', ')', '"', '\'', ',', '/', '[', ']', ';':
			return i - 1, nil
		} */
	}

	// no words found, so the end of the buffer is the end of the word.
	return uint(len(buff.Buffer)) - 1, nil

}
func CurTagExecutionWordStart(buff demodel.CharBuffer) (uint, error) {
	if buff.Tagline == nil || len(buff.Tagline.Buffer) == 0 {
		return 0, Invalid
	}
	return CurExecutionWordStart(*buff.Tagline)
}

func CurTagExecutionWordEnd(buff demodel.CharBuffer) (uint, error) {
	if buff.Tagline == nil || len(buff.Tagline.Buffer) == 0 {
		return 0, Invalid
	}
	return CurExecutionWordEnd(*buff.Tagline)

}
func CurTagWordStart(buff demodel.CharBuffer) (uint, error) {
	if buff.Tagline == nil || len(buff.Tagline.Buffer) == 0 {
		return 0, Invalid
	}
	return CurWordStart(*buff.Tagline)
}

func CurTagWordEnd(buff demodel.CharBuffer) (uint, error) {
	if buff.Tagline == nil || len(buff.Tagline.Buffer) == 0 {
		return 0, Invalid
	}
	return CurWordEnd(*buff.Tagline)

}
func NextWordStart(buff demodel.CharBuffer) (uint, error) {
	foundSpace := false
	for i := buff.Dot.End; i < uint(len(buff.Buffer)); i++ {
		if unicode.IsSpace(rune(buff.Buffer[i])) {
			foundSpace = true
			continue
		} else {
			if foundSpace == true {
				return i, nil
			}
		}
	}
	return buff.Dot.End, Invalid
}

func PrevWordStart(buff demodel.CharBuffer) (uint, error) {
	if len(buff.Buffer) == 0 || buff.Dot.Start >= uint(len(buff.Buffer)) {
		return 0, Invalid
	}
	if buff.Dot.Start == 0 {
		return 0, nil
	}
	foundSpace := false
	foundNonSpaceBeforeSpace := false
	for i := buff.Dot.Start; i > 0; i-- {
		if unicode.IsSpace(rune(buff.Buffer[i])) {
			if foundNonSpaceBeforeSpace == true {
				return i + 1, nil
			}
			foundSpace = true
		} else if foundSpace == true {
			foundNonSpaceBeforeSpace = true
		}

	}
	// Got to the start of the buffer, so this must be the start.
	return 0, nil
}

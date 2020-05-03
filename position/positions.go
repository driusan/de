package position

import (
	"errors"
	"unicode"

	"github.com/driusan/de/demodel"
)

var ErrInvalid = errors.New("Invalid Position")

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
		return 0, ErrInvalid
	}
	return buff.Tagline.Dot.Start, nil
}

func TagDotEnd(buff demodel.CharBuffer) (uint, error) {
	if buff.Tagline == nil || len(buff.Tagline.Buffer) == 0 {
		return 0, ErrInvalid
	}
	return buff.Tagline.Dot.End, nil
}

func TagAltDotStart(buff demodel.CharBuffer) (uint, error) {
	if buff.Tagline == nil || len(buff.Tagline.Buffer) == 0 {
		return 0, ErrInvalid
	}
	return buff.Tagline.AltDot.Start, nil
}
func TagAltDotEnd(buff demodel.CharBuffer) (uint, error) {
	if buff.Tagline == nil || len(buff.Tagline.Buffer) == 0 {
		return 0, ErrInvalid
	}
	return buff.Tagline.AltDot.End, nil
}

func BuffStart(buff demodel.CharBuffer) (uint, error) {
	if len(buff.Buffer) > 0 {
		return 0, nil
	}
	return 0, ErrInvalid
}
func BuffEnd(buff demodel.CharBuffer) (uint, error) {
	if len(buff.Buffer) > 0 {
		return uint(len(buff.Buffer) - 1), nil
	}
	return 0, ErrInvalid
}

func PrevChar(buff demodel.CharBuffer) (uint, error) {
	if buff.Dot.Start == 0 {
		return 0, ErrInvalid
	}

	// BUG: This doesn't deal with multibyte UTF-8 runes.
	return buff.Dot.Start - 1, nil
}
func NextChar(buff demodel.CharBuffer) (uint, error) {
	if buff.Dot.End >= uint(len(buff.Buffer)) {
		return uint(len(buff.Buffer)) - 1, ErrInvalid
	}
	return buff.Dot.End + 1, nil
}

func PrevLine(buff demodel.CharBuffer) (uint, error) {
	// how far into the current line is the current character?
	var lineIdx = -1

	// where does the previous line start, so that we can index
	// lineIdx into it at the end?
	var prevLineStart, curLineStart int = -1, -1
	for i := buff.Dot.Start; i > 0; i-- {
		if uint(len(buff.Buffer)) <= i {
			return buff.Dot.End, ErrInvalid
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
	// special case. We're at a newline, so the next line is just the next character..
	if len(buff.Buffer) == 0 {
		return 0, ErrInvalid
	}
	if buff.Dot.End >= uint(len(buff.Buffer)) {
		return uint(len(buff.Buffer) - 1), ErrInvalid
	}
	if buff.Buffer[buff.Dot.End] == '\n' {
		return buff.Dot.End + 1, nil
	}
	// how far into the current line is the current character?
	var lineIdx = -1

	// calculate how far we are into the current line.
	var nextLineStart, subsequentLine int = -1, -1
	for i := buff.Dot.End; i > 0; i-- {
		if uint(len(buff.Buffer)) <= i {
			return buff.Dot.End, ErrInvalid
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
		// is actually the index into the line.
		lineIdx = int(buff.Dot.End)
	}
	//fmt.Printf("Line Index: %d\n", lineIdx)
	// now find the next line start, so we can add to it, and the line
	// after that, so we can check if nextLine+idx goes too far.
	for i := buff.Dot.End; i < uint(len(buff.Buffer)); i++ {
		if buff.Buffer[i] == '\n' {
			if nextLineStart == -1 {
				nextLineStart = int(i)
			} else {
				// the line starts after the newline character,
				// not on it, so add one to account for the \n.
				subsequentLine = int(i + 1)
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
	if subsequentLine > 0 && pos >= uint(subsequentLine) {
		// went too far, so return the end of the line
		return uint(subsequentLine) - 1, nil
	}

	if pos >= uint(len(buff.Buffer)) {
		// overflowed the buffer somehow, so return the end of the buffer
		return uint(len(buff.Buffer)) - 1, ErrInvalid
	}

	// the line starts at the character after the \n, not on the \n itself.
	return pos, nil
}

func EndOfLine(buff demodel.CharBuffer) (uint, error) {
	if len(buff.Buffer) == 0 {
		return 0, ErrInvalid
	}
	if buff.Dot.End >= uint(len(buff.Buffer)) {
		return uint(len(buff.Buffer)) - 1, ErrInvalid
	}
	if buff.Buffer[buff.Dot.End] == '\n' {
		return buff.Dot.End, nil
	}
	for i := buff.Dot.End; i < uint(len(buff.Buffer)); i++ {
		if buff.Buffer[i] == '\n' {
			return i, nil
		}
	}
	return uint(len(buff.Buffer) - 1), nil
}

func EndOfLineWithNewline(buff demodel.CharBuffer) (uint, error) {
	if len(buff.Buffer) == 0 {
		return 0, ErrInvalid
	}
	if buff.Dot.End >= uint(len(buff.Buffer)) {
		return uint(len(buff.Buffer)), nil
	}
	if buff.Buffer[buff.Dot.End] == '\n' {
		return buff.Dot.End + 1, nil
	}
	for i := buff.Dot.End; i < uint(len(buff.Buffer)); i++ {
		if buff.Buffer[i] == '\n' {
			return i + 1, nil
		}
	}
	return uint(len(buff.Buffer) - 1), nil
}
func StartOfLine(buff demodel.CharBuffer) (uint, error) {
	if len(buff.Buffer) == 0 {
		return 0, ErrInvalid
	}
	if buff.Dot.Start == 0 {
		return 0, nil
	}

	for i := buff.Dot.Start - 1; i > 0 && i < uint(len(buff.Buffer)); i-- {
		if buff.Buffer[i] == '\n' {
			return i + 1, nil
		}
	}
	return 0, nil
}

// StartOfLineAfterWhitespace returns the buffer index of the first character
// in the current line that is not whitespace, or the end of the line if the
// line is all whitespace.
func StartOfLineAfterWhitespace(buff demodel.CharBuffer) (uint, error) {
	if len(buff.Buffer) == 0 {
		return 0, ErrInvalid
	}

	start, err := StartOfLine(buff)
	if err != nil {
		return 0, err
	}

	for i := start; i < uint(len(buff.Buffer)); i++ {
		if buff.Buffer[i] == '\n' {
			// The line is all whitespace
			return i, nil
		}
		if buff.Buffer[i] == '\n' || !unicode.IsSpace(rune(buff.Buffer[i])) {
			return i, nil
		}
	}

	return 0, nil

}

func CurWordStart(buff demodel.CharBuffer) (uint, error) {
	if len(buff.Buffer) == 0 {
		return 0, ErrInvalid
	}
	if buff.Dot.Start == 0 {
		return 0, nil
	}
	for i := buff.Dot.Start - 1; i > 0 && i < uint(len(buff.Buffer)); i-- {
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
		return 0, ErrInvalid
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
		return 0, ErrInvalid
	}
	if buff.Dot.Start == 0 {
		return 0, nil
	}
	for i := buff.Dot.Start - 1; i > 0 && i < uint(len(buff.Buffer)); i-- {
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
		return 0, ErrInvalid
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
		return 0, ErrInvalid
	}
	return CurExecutionWordStart(*buff.Tagline)
}

func CurTagExecutionWordEnd(buff demodel.CharBuffer) (uint, error) {
	if buff.Tagline == nil || len(buff.Tagline.Buffer) == 0 {
		return 0, ErrInvalid
	}
	return CurExecutionWordEnd(*buff.Tagline)

}
func CurTagWordStart(buff demodel.CharBuffer) (uint, error) {
	if buff.Tagline == nil || len(buff.Tagline.Buffer) == 0 {
		return 0, ErrInvalid
	}
	return CurWordStart(*buff.Tagline)
}

func CurTagWordEnd(buff demodel.CharBuffer) (uint, error) {
	if buff.Tagline == nil || len(buff.Tagline.Buffer) == 0 {
		return 0, ErrInvalid
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
			if foundSpace {
				return i, nil
			}
		}
	}
	return buff.Dot.End, ErrInvalid
}

func PrevWordStart(buff demodel.CharBuffer) (uint, error) {
	if len(buff.Buffer) == 0 || buff.Dot.Start >= uint(len(buff.Buffer)) {
		return 0, ErrInvalid
	}
	if buff.Dot.Start == 0 {
		return 0, nil
	}
	foundSpace := false
	foundNonSpaceBeforeSpace := false
	for i := buff.Dot.Start; i > 0; i-- {
		if unicode.IsSpace(rune(buff.Buffer[i])) {
			if foundNonSpaceBeforeSpace {
				return i + 1, nil
			}
			foundSpace = true
		} else if foundSpace {
			foundNonSpaceBeforeSpace = true
		}

	}
	// Got to the start of the buffer, so this must be the start.
	return 0, nil
}

func MatchingBracket(buff demodel.CharBuffer) (uint, error) {
	if len(buff.Buffer) == 0 || buff.Dot.Start >= uint(len(buff.Buffer)) {
		return 0, ErrInvalid
	}
	cur := buff.Buffer[buff.Dot.Start]

	var forwardto, backwardto byte
	switch cur {
	case '{':
		forwardto = '}'
	case '}':
		backwardto = '{'
	case '(':
		forwardto = ')'
	case ')':
		backwardto = '('
	case '[':
		forwardto = ']'
	case ']':
		backwardto = '['
	case '<':
		forwardto = '>'
	case '>':
		backwardto = '<'
	}

	if forwardto != 0 {
		extralevels := 0
		for i := buff.Dot.Start + 1; i < uint(len(buff.Buffer)); i++ {
			if buff.Buffer[i] == cur {
				extralevels++
			}
			if buff.Buffer[i] == forwardto {
				if extralevels > 0 {
					extralevels--
				} else {
					return i, nil
				}
			}
		}
	} else if backwardto != 0 {
		extralevels := 0
		for i := buff.Dot.Start - 1; i >= 0; i-- {
			if buff.Buffer[i] == cur {
				extralevels++
			}

			if buff.Buffer[i] == backwardto {
				if extralevels > 0 {
					extralevels--
				} else {
					return i, nil
				}
			}
		}
	} else {
		// Go backwards to the start of the current block (for some
		// definition of block.
		extrasquigly := 0
		extraround := 0
		extrasquare := 0
		extraangle := 0
		for i := int(buff.Dot.Start - 1); i >= 0; i-- {
			// If there was a block inside this block, skip it
			if buff.Buffer[i] == '}' {
				extrasquigly++
			} else if buff.Buffer[i] == ')' {
				extraround++
			} else if buff.Buffer[i] == ']' {
				extrasquare++
			} else if buff.Buffer[i] == '>' {
				extraangle++
			} else if buff.Buffer[i] == '{' {
				if extrasquigly > 0 {
					extrasquigly--
				} else {
					return uint(i), nil
				}
			} else if buff.Buffer[i] == '(' {
				if extraround > 0 {
					extraround--
				} else {
					return uint(i), nil
				}
			} else if buff.Buffer[i] == '[' {
				if extrasquare > 0 {
					extrasquare--
				} else {
					return uint(i), nil
				}
			} else if buff.Buffer[i] == '<' {
				if extraangle > 0 {
					extraangle--
				} else {
					return uint(i), nil
				}
			}
		}
	}
	return buff.Dot.Start, ErrInvalid
}

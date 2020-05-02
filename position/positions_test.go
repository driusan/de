package position

import (
	"testing"

	"github.com/driusan/de/demodel"
)

func TestPositionsDontOverflow(t *testing.T) {
	// Start with an empty charbuffer and make sure they don't overflow
	buff := demodel.CharBuffer{}

	// We don't have to do anything with the results, if they panic the
	// test will fail.
	DotStart(buff)
	DotEnd(buff)
	AltDotStart(buff)
	AltDotEnd(buff)
	TagDotStart(buff)
	TagDotEnd(buff)
	TagAltDotStart(buff)
	TagAltDotEnd(buff)
	BuffStart(buff)
	BuffEnd(buff)
	PrevChar(buff)
	NextChar(buff)
	PrevLine(buff)
	NextLine(buff)
	EndOfLine(buff)
	EndOfLineWithNewline(buff)
	StartOfLine(buff)
	CurWordStart(buff)
	CurWordEnd(buff)
	CurExecutionWordStart(buff)
	CurExecutionWordEnd(buff)
	CurTagExecutionWordStart(buff)
	CurTagExecutionWordEnd(buff)
	CurTagWordStart(buff)
	CurTagWordEnd(buff)
	NextWordStart(buff)
	PrevWordStart(buff)
	MatchingBracket(buff)

	// Now try with a buffer and dot.End > len(buffer)
	buff.Buffer = []byte{'a'}
	buff.Dot.Start = 3
	buff.Dot.End = 3
	DotStart(buff)
	DotEnd(buff)
	AltDotStart(buff)
	AltDotEnd(buff)
	TagDotStart(buff)
	TagDotEnd(buff)
	TagAltDotStart(buff)
	TagAltDotEnd(buff)
	BuffStart(buff)
	BuffEnd(buff)
	PrevChar(buff)
	NextChar(buff)
	PrevLine(buff)
	NextLine(buff)
	EndOfLine(buff)
	EndOfLineWithNewline(buff)
	StartOfLine(buff)
	CurWordStart(buff)
	CurWordEnd(buff)
	CurExecutionWordStart(buff)
	CurExecutionWordEnd(buff)
	CurTagExecutionWordStart(buff)
	CurTagExecutionWordEnd(buff)
	CurTagWordStart(buff)
	CurTagWordEnd(buff)
	NextWordStart(buff)
	PrevWordStart(buff)
	MatchingBracket(buff)

}

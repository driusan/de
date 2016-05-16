package demodel

// Dot generally represents a selection or position in a buffer.
// It holds the start and the end of the selection.
// If Start==End, it's a position.
type Dot struct {
	Start, End uint
}

// A CharBuffer is a set of bytes being manipulated. Generally,
// a text file.
type CharBuffer struct {
	Buffer   []byte
	Dot      Dot
	Filename string

	// The most recently deleted text, which can be pasted
	SnarfBuffer []byte
}

// An action performs some sort of action to a character buffer,
// which would generally involve modifying it in some way.
type Action func(From, To Position, buf *CharBuffer) error

// A position calculates the index into a CharBuffer for
// something to use as a reference, generally to perform
// an action on it. For instance, the position of the
// start of the previous word, or the next paragraph,
// or the containing block of the cursor.
type Position func(buf CharBuffer) (uint, error)

// Renders the character buffer to a string which can
// be displayed on a non-graphical terminal.
type TermRender interface {
	Render(CharBuffer) (string, error)
}

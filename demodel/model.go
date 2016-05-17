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
	// The CharBuffer.Buffer is a slice of the bytes which are currently
	// being manipulated. It's read from the filesystem when a file is opened,
	// and written to the filesystem when it's saved.
	Buffer []byte

	// Dot represents the current cursor position or selected text inside of
	// the character buffer.
	Dot Dot

	// Filename represents the name that the CharBuffer.Buffer will be written
	// to on save.
	Filename string

	// The most recently deleted text, which can be pasted with the 'p' command
	// Each CharBuffer only has one SnarfBuffer associated with it, and it's not
	// a tree.
	SnarfBuffer []byte
}

// An action performs some sort of action to a character buffer,
// which would generally involve modifying it in some way.
type Action func(From, To Position, buf *CharBuffer) error

// A position calculates the index into a CharBuffer for
// something to use as a reference, generally to perform
// an action on it. For instance, the position of the
// start of the previous word, or the next paragraph,
// or the containing block of the cursor. Built in positions
// are in the positions package.
type Position func(buf CharBuffer) (uint, error)

// Renders the character buffer to a string which can
// be displayed on a non-graphical terminal. This is not
// yet implemented.
type TermRender interface {
	Render(CharBuffer) (string, error)
}

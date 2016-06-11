package demodel

import (
	"golang.org/x/mobile/event/key"
)

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

	// AltDot represents the selection used for middle clicking or pressing "Enter"
	// to execute a word. Since the execution might need Dot as a parameter, we
	// need a second selection to determine what we're executing, otherwise commands
	// like "Cut" would always cut the word "Cut", since it had to be selected
	// to be executed.
	AltDot Dot

	// Filename represents the name that the CharBuffer.Buffer will be written
	// to on save.
	Filename string

	// The most recently deleted text, which can be pasted with the 'p' command
	// Each CharBuffer only has one SnarfBuffer associated with it, and it's not
	// a tree.
	SnarfBuffer []byte

	// The tagline to display with this buffer. May be nil (for instance, taglines
	// are CharBuffers, but taglines don't have their own tagline..)
	Tagline *CharBuffer

	// Undo represents the previous CharBuffer that got the buffer into this state.
	// It's effectively a singly-linked list of Undos.
	Undo *CharBuffer
	// Dirty represents if the file has been modified since being open, or since the last
	// save.
	Dirty bool
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

// A Map maps a keystroke to a command. It performs a command, and then
// returns a new map which represents the keyboard mapping to be used
// for the next keystroke.
type Map interface {
	HandleKey(key.Event, *CharBuffer, Viewport) (Map, ScrollDirection, error)
}

type ScrollDirection uint8

const (
	DirectionNone = ScrollDirection(iota)
	DirectionUp
	DirectionDown
)

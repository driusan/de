package kbmap

import (
	"unicode"
	"unicode/utf8"

	"github.com/driusan/de/actions"
	"github.com/driusan/de/demodel"
	"github.com/driusan/de/position"
	"golang.org/x/mobile/event/key"
)

// Since we don't want to update the undo buffer for every single keystroke in
// insert mode, we keep a pointer to the buffer when insert mode started, and
// only update the Undo pointer upon leaving insert mode.
var insertStartCharBuffer *demodel.CharBuffer

func insertMap(e key.Event, buff *demodel.CharBuffer, v demodel.Viewport) (demodel.Map, demodel.ScrollDirection, error) {
	// special cases for Insert Mode
	switch e.Code {
	case key.CodeEscape:
		if e.Direction == key.DirPress {
			buff.Undo = insertStartCharBuffer
			insertStartCharBuffer = nil
			return NormalMode, demodel.DirectionNone, nil
		}
	case key.CodeDeleteBackspace:
		if e.Direction != key.DirRelease {
			actions.DeleteCursor(position.DotStart, position.DotEnd, buff)
			return InsertMode, demodel.DirectionUp, nil
		}
	case key.CodeLeftArrow:
		if e.Direction == key.DirPress {
			if e.Modifiers&key.ModControl != 0 {
				// ctrl is pressed, so just move the start and expand dot
				actions.MoveCursor(position.PrevChar, position.DotEnd, buff)
			} else {
				// ctrl is not pressed, so move the cursor without selecting
				actions.MoveCursor(position.PrevChar, position.PrevChar, buff)
			}
		}
		return InsertMode, demodel.DirectionUp, nil
	case key.CodeRightArrow:
		if e.Direction == key.DirPress {
			if e.Modifiers&key.ModControl != 0 {
				// ctrl is pressed, so just move the end and expand dot
				actions.MoveCursor(position.DotStart, position.NextChar, buff)
			} else {
				// ctrl is not pressed, so move the cursor without selecting
				actions.MoveCursor(position.NextChar, position.NextChar, buff)
			}
			return InsertMode, demodel.DirectionDown, nil
		}
	case key.CodeDownArrow:
		if e.Direction == key.DirPress {
			if e.Modifiers&key.ModControl != 0 {
				// ctrl is pressed, so just move the end and expand dot
				actions.MoveCursor(position.DotStart, position.NextLine, buff)
			} else {
				// ctrl is not pressed, so move the cursor without selecting
				actions.MoveCursor(position.NextLine, position.NextLine, buff)
			}
		}
		return InsertMode, demodel.DirectionDown, nil

	case key.CodeUpArrow:
		if e.Direction == key.DirPress {
			if e.Modifiers&key.ModControl != 0 {
				// ctrl is pressed, so just move the start and expand dot
				actions.MoveCursor(position.PrevLine, position.DotEnd, buff)
			} else {
				// ctrl is not pressed, so move the cursor without selecting
				actions.MoveCursor(position.PrevLine, position.PrevLine, buff)
			}
		}
		return InsertMode, demodel.DirectionUp, nil
	case key.CodeV:
		if e.Direction == key.DirPress && isCopyModifier(e) {
			insertStartCharBuffer = &demodel.CharBuffer{
				Buffer: buff.Buffer,
				Dot:    buff.Dot,
				Undo:   buff.Undo,
			}
			actions.PasteClipboard(position.DotStart, position.DotEnd, buff)
			return InsertMode, demodel.DirectionNone, nil
		}
	case key.CodeC:
		if e.Direction == key.DirPress && isCopyModifier(e) {
			actions.CopyClipboard(position.DotStart, position.DotEnd, buff)
			return InsertMode, demodel.DirectionNone, nil
		}
	case key.CodeX:
		if e.Direction == key.DirPress && isCopyModifier(e) {
			insertStartCharBuffer = &demodel.CharBuffer{
				Buffer: buff.Buffer,
				Dot:    buff.Dot,
				Undo:   buff.Undo,
			}
			actions.CutClipboard(position.DotStart, position.DotEnd, buff)
			return InsertMode, demodel.DirectionNone, nil
		}

	}

	// These events don't seem to have the rune set properly, so add it as a hack.
	if e.Code == key.CodeReturnEnter {
		e.Rune = '\n'
	}
	if e.Code == key.CodeTab {
		e.Rune = '\t'
	}

	if e.Direction == key.DirRelease {
		// add the character if it's a key release or a repeat, but not
		// if it's being released. For some reason, release seems more reliable
		// than press when typing fast.
		return InsertMode, demodel.DirectionNone, nil
	}

	// unicode.IsPrint is selective about what whitespace it considers printable.
	if !unicode.IsPrint(e.Rune) && e.Rune != '\n' && e.Rune != '\t' {
		// if it's not a printable character, don't insert it. We also
		// receive key events on things like shift being pressed..
		return InsertMode, demodel.DirectionNone, nil
	}
	// in insert the rune at the current position, overwriting Dot if applicable

	runeBytes := make([]byte, 4)
	i := utf8.EncodeRune(runeBytes, e.Rune)

	// Before we do anything, cache the start of the insert if applicable so that
	// we can set the appropriate *Undo pointer when leaving insert mode.
	if insertStartCharBuffer == nil {
		insertStartCharBuffer = &demodel.CharBuffer{
			Buffer: buff.Buffer,
			Dot:    buff.Dot,
			Undo:   buff.Undo,
		}

	}
	// inserting at the start of the file.
	if buff.Dot.End == 0 {
		buff.Buffer = append(runeBytes[:i], buff.Buffer...)
		buff.Dot.Start = uint(i)
		buff.Dot.End = buff.Dot.Start
	} else {
		newBuffer := make([]byte, len(buff.Buffer)+i)
		copy(newBuffer, buff.Buffer)
		copy(newBuffer[buff.Dot.Start:], runeBytes[:i])
		copy(newBuffer[buff.Dot.Start+uint(i):], buff.Buffer[buff.Dot.End:])
		//copy(newBuffer[:buff.Dot.Start+uint(i)+1], buff.Buffer[buff.Dot.End:])

		buff.Buffer = newBuffer
		buff.Dot.Start += uint(i)
		buff.Dot.End = buff.Dot.Start
	}
	buff.Dirty = true
	return InsertMode, demodel.DirectionDown, nil
}

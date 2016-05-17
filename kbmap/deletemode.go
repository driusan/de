package kbmap

import (
	"github.com/driusan/de/actions"
	"github.com/driusan/de/demodel"
	"github.com/driusan/de/position"
	"golang.org/x/mobile/event/key"
)

func deleteMap(e key.Event, buff *demodel.CharBuffer) (Map, error) {
	// things only happen on key press in normal mode, if it's a release
	// or a repeat, ignore it. It's not an error
	if e.Direction != key.DirPress {
		return DeleteMode, nil
	}
	if buff == nil {
		return NormalMode, Invalid
	}
	switch e.Code {
	case key.CodeEscape:
		return NormalMode, nil
	case key.CodeDeleteBackspace:
		if e.Direction == key.DirPress {
			actions.DeleteCursor(position.DotStart, position.DotEnd, buff)
		}
		return NormalMode, nil
	case key.CodeK:
		if Repeat == 0 {
			Repeat = 1
		}
		for ; Repeat > 0; Repeat-- {
			actions.MoveCursor(position.PrevLine, position.DotEnd, buff)
		}

		actions.DeleteCursor(position.DotStart, position.DotEnd, buff)
		return NormalMode, nil
	case key.CodeH:
		if Repeat == 0 {
			Repeat = 1
		}
		for ; Repeat > 0; Repeat-- {
			actions.MoveCursor(position.PrevChar, position.DotEnd, buff)
		}
		actions.DeleteCursor(position.DotStart, position.DotEnd, buff)
		return NormalMode, nil
	case key.CodeL:
		if Repeat == 0 {
			Repeat = 1
		}
		for ; Repeat > 0; Repeat-- {
			actions.MoveCursor(position.DotStart, position.NextChar, buff)
		}
		actions.DeleteCursor(position.DotStart, position.DotEnd, buff)
		return NormalMode, nil
	case key.CodeJ:
		if Repeat == 0 {
			Repeat = 1
		}
		for ; Repeat > 0; Repeat-- {
			actions.MoveCursor(position.DotStart, position.NextLine, buff)
		}
		actions.DeleteCursor(position.DotStart, position.DotEnd, buff)
		return NormalMode, nil
	case key.CodeX:
		// x just deletes the selected text, similar to vi. Repeat only does
		// anything in normal mode.
		Repeat = 0
		actions.DeleteCursor(position.DotStart, position.DotEnd, buff)
		return NormalMode, nil
	case key.CodeW:
		if Repeat == 0 {
			Repeat = 1
		}
		for ; Repeat > 0; Repeat-- {
			actions.MoveCursor(position.DotStart, position.NextWordStart, buff)
		}
		actions.DeleteCursor(position.DotStart, position.DotEnd, buff)
		return NormalMode, nil
	case key.CodeB:
		if Repeat == 0 {
			Repeat = 1
		}
		for ; Repeat > 0; Repeat-- {
			actions.MoveCursor(position.PrevWordStart, position.DotEnd, buff)
		}
		actions.DeleteCursor(position.DotStart, position.DotEnd, buff)
		return NormalMode, nil
	case key.CodeRightArrow:
		return DeleteMode, ScrollRight
	case key.CodeLeftArrow:
		return DeleteMode, ScrollLeft
	case key.CodeDownArrow:
		return DeleteMode, ScrollDown
	case key.CodeUpArrow:
		return DeleteMode, ScrollUp
	case key.Code4:
		// $ is pressed, in most key maps..

		if e.Modifiers&key.ModShift != 0 {
			Repeat = 0
			actions.DeleteCursor(position.DotStart, position.EndOfLine, buff)
		}
		return NormalMode, nil
	case key.Code6:
		// ^ is pressed, on most keyboards..
		if e.Modifiers&key.ModShift != 0 {
			Repeat = 0
			actions.DeleteCursor(position.StartOfLine, position.DotEnd, buff)
		}
		return NormalMode, nil
	case key.CodeG:
		// capital G
		if e.Modifiers&key.ModShift != 0 {
			Repeat = 0
			actions.MoveCursor(position.DotStart, position.BuffEnd, buff)
		}
		return NormalMode, nil
	case key.CodeD:
		// don't need to handle Repeat = 0 case, because the first movement will
		// take care of it.
		actions.MoveCursor(position.StartOfLine, position.EndOfLine, buff)
		for ; Repeat > 0; Repeat-- {
			actions.MoveCursor(position.DotStart, position.NextLine, buff)
		}
		actions.DeleteCursor(position.DotStart, position.DotEnd, buff)
		return NormalMode, nil
	}
	return DeleteMode, Invalid
}

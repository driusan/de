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
		actions.DeleteCursor(position.PrevLine, position.DotEnd, buff)
		return NormalMode, nil
	case key.CodeH:
		actions.DeleteCursor(position.PrevChar, position.DotEnd, buff)
		return NormalMode, nil
	case key.CodeL:
		actions.DeleteCursor(position.DotStart, position.NextChar, buff)
		return NormalMode, nil
	case key.CodeJ:
		actions.DeleteCursor(position.DotStart, position.NextLine, buff)
		return NormalMode, nil
	case key.CodeX:
		actions.DeleteCursor(position.DotStart, position.NextChar, buff)
		return NormalMode, nil
	case key.CodeW:
		actions.DeleteCursor(position.DotStart, position.NextWordStart, buff)
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
			actions.DeleteCursor(position.DotStart, position.EndOfLine, buff)
		}
		return NormalMode, nil
	case key.Code6:
		// ^ is pressed, on most keyboards..
		if e.Modifiers&key.ModShift != 0 {
			actions.DeleteCursor(position.StartOfLine, position.DotEnd, buff)
		}
		return NormalMode, nil
	case key.CodeG:
		// capital G
		if e.Modifiers&key.ModShift != 0 {
			actions.MoveCursor(position.DotStart, position.BuffEnd, buff)
		}
		return NormalMode, nil
	case key.CodeD:
		actions.DeleteCursor(position.StartOfLine, position.EndOfLine, buff)
		return NormalMode, nil
	}
	return DeleteMode, Invalid
}

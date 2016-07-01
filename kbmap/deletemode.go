package kbmap

import (
	"github.com/driusan/de/actions"
	"github.com/driusan/de/demodel"
	"github.com/driusan/de/position"
	"golang.org/x/mobile/event/key"
)

func deleteMap(e key.Event, buff *demodel.CharBuffer, v demodel.Viewport) (demodel.Map, demodel.ScrollDirection, error) {
	// things only happen on key press in normal mode, if it's a release
	// or a repeat, ignore it. It's not an error
	if e.Direction == key.DirRelease {
		return DeleteMode, demodel.DirectionNone, nil
	}
	if buff == nil {
		return NormalMode, demodel.DirectionNone, ErrInvalid
	}

	// since some keys we're modify dot before calling DeleteCursor, we need to back it up
	// and set it properly for the undo buffer after deleting. We just back it up once here,
	// for simplicity, even for the delete modes that don't need it.
	undoDot := buff.Dot

	switch e.Code {
	case key.CodeEscape:
		return NormalMode, demodel.DirectionNone, nil
	case key.CodeDeleteBackspace:
		if e.Direction != key.DirRelease {
			actions.DeleteCursor(position.DotStart, position.DotEnd, buff)
		}
		return DeleteMode, demodel.DirectionUp, nil
	case key.CodeK:
		if Repeat == 0 {
			Repeat = 1
		}

		for ; Repeat > 0; Repeat-- {
			actions.MoveCursor(position.PrevLine, position.DotEnd, buff)
		}

		actions.DeleteCursor(position.DotStart, position.DotEnd, buff)
		buff.Undo.Dot = undoDot
		return NormalMode, demodel.DirectionUp, nil
	case key.CodeH:
		if Repeat == 0 {
			Repeat = 1
		}
		for ; Repeat > 0; Repeat-- {
			actions.MoveCursor(position.PrevChar, position.DotEnd, buff)
		}
		actions.DeleteCursor(position.DotStart, position.DotEnd, buff)
		buff.Undo.Dot = undoDot
		return NormalMode, demodel.DirectionUp, nil
	case key.CodeL:
		if Repeat == 0 {
			Repeat = 1
		}
		for ; Repeat > 0; Repeat-- {
			actions.MoveCursor(position.DotStart, position.NextChar, buff)
		}
		actions.DeleteCursor(position.DotStart, position.DotEnd, buff)
		buff.Undo.Dot = undoDot
		return NormalMode, demodel.DirectionUp, nil
	case key.CodeJ:
		if Repeat == 0 {
			Repeat = 1
		}
		for ; Repeat > 0; Repeat-- {
			actions.MoveCursor(position.DotStart, position.NextLine, buff)
		}
		actions.DeleteCursor(position.DotStart, position.DotEnd, buff)
		buff.Undo.Dot = undoDot
		return NormalMode, demodel.DirectionUp, nil
	case key.CodeX:
		if e.Direction == key.DirPress && isCopyModifier(e) {
			actions.CutClipboard(position.DotStart, position.DotEnd, buff)
			return DeleteMode, demodel.DirectionNone, nil
		}
		// x just deletes the selected text, similar to vi. Repeat only does
		// anything in normal mode.
		Repeat = 0
		actions.DeleteCursor(position.DotStart, position.DotEnd, buff)
		return NormalMode, demodel.DirectionUp, nil
	case key.CodeW:
		if Repeat == 0 {
			Repeat = 1
		}

		for ; Repeat > 0; Repeat-- {
			actions.MoveCursor(position.DotStart, position.NextWordStart, buff)
		}
		actions.DeleteCursor(position.DotStart, position.DotEnd, buff)
		buff.Undo.Dot = undoDot
		return NormalMode, demodel.DirectionUp, nil
	case key.CodeB:
		if Repeat == 0 {
			Repeat = 1
		}
		for ; Repeat > 0; Repeat-- {
			actions.MoveCursor(position.PrevWordStart, position.DotEnd, buff)
		}
		actions.DeleteCursor(position.DotStart, position.DotEnd, buff)
		buff.Undo.Dot = undoDot
		return NormalMode, demodel.DirectionUp, nil
	case key.CodeRightArrow:
		return DeleteMode, demodel.DirectionUp, ErrScrollRight
	case key.CodeLeftArrow:
		return DeleteMode, demodel.DirectionUp, ErrScrollLeft
	case key.CodeDownArrow:
		return DeleteMode, demodel.DirectionUp, ErrScrollDown
	case key.CodeUpArrow:
		return DeleteMode, demodel.DirectionUp, ErrScrollUp
	case key.Code4:
		// $ is pressed, in most key maps..

		if e.Modifiers&key.ModShift != 0 {
			Repeat = 0
			actions.DeleteCursor(position.DotStart, position.EndOfLine, buff)
		}
		return NormalMode, demodel.DirectionUp, nil
	case key.Code6:
		// ^ is pressed, on most keyboards..
		if e.Modifiers&key.ModShift != 0 {
			Repeat = 0
			actions.DeleteCursor(position.StartOfLine, position.DotEnd, buff)
		}
		return NormalMode, demodel.DirectionUp, nil
	case key.CodeG:
		// capital G
		if e.Modifiers&key.ModShift != 0 {
			Repeat = 0
			actions.MoveCursor(position.DotStart, position.BuffEnd, buff)
		}
		actions.DeleteCursor(position.DotStart, position.DotEnd, buff)
		buff.Undo.Dot = undoDot
		return NormalMode, demodel.DirectionUp, nil
	case key.CodeD:
		// don't need to handle Repeat = 0 case, because the first movement will
		// take care of it.
		actions.MoveCursor(position.StartOfLine, position.DotEnd, buff)

		for ; Repeat > 1; Repeat-- {
			actions.MoveCursor(position.DotStart, position.NextLine, buff)
		}
		actions.DeleteCursor(position.DotStart, position.EndOfLineWithNewline, buff)
		buff.Undo.Dot = undoDot
		Repeat = 0
		return NormalMode, demodel.DirectionUp, nil
	case key.CodeV:
		if e.Direction == key.DirPress && isCopyModifier(e) {
			actions.PasteClipboard(position.DotStart, position.DotEnd, buff)
			return DeleteMode, demodel.DirectionNone, nil
		}
	case key.CodeC:
		if e.Direction == key.DirPress && isCopyModifier(e) {
			actions.CopyClipboard(position.DotStart, position.DotEnd, buff)
			return DeleteMode, demodel.DirectionNone, nil
		}
	}
	return DeleteMode, demodel.DirectionNone, ErrInvalid
}

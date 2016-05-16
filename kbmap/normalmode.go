package kbmap

import (
	//	"fmt"
	"github.com/driusan/de/actions"
	"github.com/driusan/de/demodel"
	"github.com/driusan/de/position"
	"golang.org/x/mobile/event/key"
)

var Repeat uint

func normalMap(e key.Event, buff *demodel.CharBuffer) (Map, error) {
	// things only happen on key press in normal mode, if it's a release
	// or a repeat, ignore it. It's not an error
	if e.Direction != key.DirPress {
		return NormalMode, nil
	}
	if buff == nil {
		return NormalMode, Invalid
	}
	switch e.Code {
	case key.CodeEscape:
		actions.SaveFile(position.BuffStart, position.BuffEnd, buff)
		return NormalMode, ExitProgram
	case key.CodeDeleteBackspace:
		if e.Direction == key.DirPress {
			actions.DeleteCursor(position.DotStart, position.DotEnd, buff)
		}
		return NormalMode, nil
	case key.CodeI:
		return InsertMode, nil
	case key.CodeA:
		actions.MoveCursor(position.NextChar, position.NextChar, buff)
		return InsertMode, nil
	case key.CodeK:
		if Repeat == 0 {
			Repeat = 1
		}
		for ; Repeat > 0; Repeat-- {
			if e.Modifiers&key.ModControl != 0 {
				// ctrl is pressed, so just move the start and expand dot
				actions.MoveCursor(position.PrevLine, position.DotEnd, buff)
			} else {
				// ctrl is not pressed, so move the cursor without selecting
				actions.MoveCursor(position.PrevLine, position.PrevLine, buff)

			}
		}

		return NormalMode, nil
	case key.CodeH:
		if Repeat == 0 {
			Repeat = 1
		}
		for ; Repeat > 0; Repeat-- {
			if e.Modifiers&key.ModControl != 0 {
				// ctrl is pressed, so just move the end and expand dot
				actions.MoveCursor(position.PrevChar, position.DotEnd, buff)
			} else {
				// ctrl is not pressed, so move the cursor without selecting
				actions.MoveCursor(position.PrevChar, position.PrevChar, buff)

			}
		}

		return NormalMode, nil

	case key.CodeL:
		if Repeat == 0 {
			Repeat = 1
		}
		for ; Repeat > 0; Repeat-- {
			if e.Modifiers&key.ModControl != 0 {
				// ctrl is pressed, so just move the end and expand dot
				actions.MoveCursor(position.DotStart, position.NextChar, buff)
			} else {
				// ctrl is not pressed, so move the cursor without selecting
				actions.MoveCursor(position.NextChar, position.NextChar, buff)
			}
		}

		return NormalMode, nil
	case key.CodeJ:
		if Repeat == 0 {
			Repeat = 1

		}
		for ; Repeat > 0; Repeat-- {
			if e.Modifiers&key.ModControl != 0 {
				actions.MoveCursor(position.DotStart, position.NextLine, buff)
			} else {
				actions.MoveCursor(position.NextLine, position.NextLine, buff)
			}
		}

		return NormalMode, nil
	case key.CodeX:
		if Repeat == 0 {
			Repeat = 1
		}
		for ; Repeat > 0; Repeat-- {
			actions.MoveCursor(position.DotStart, position.NextChar, buff)
		}
		actions.DeleteCursor(position.DotStart, position.DotEnd, buff)
		return NormalMode, nil
	case key.CodeW:
		if Repeat == 0 {
			Repeat = 1
		}
		for ; Repeat > 0; Repeat-- {
			if e.Modifiers&key.ModControl != 0 {
				actions.MoveCursor(position.DotStart, position.NextWordStart, buff)
			} else {
				actions.MoveCursor(position.NextWordStart, position.NextWordStart, buff)
			}
		}
	case key.CodeB:
		if Repeat == 0 {
			Repeat = 1
		}
		for ; Repeat > 0; Repeat-- {
			if e.Modifiers&key.ModControl != 0 {
				actions.MoveCursor(position.PrevWordStart, position.DotEnd, buff)
			} else {
				actions.MoveCursor(position.PrevWordStart, position.PrevWordStart, buff)
			}
		}
	case key.CodeP:
		actions.InsertSnarfBuffer(position.DotStart, position.DotEnd, buff)
	case key.CodeRightArrow:
		return NormalMode, ScrollRight
	case key.CodeLeftArrow:
		return NormalMode, ScrollLeft
	case key.CodeDownArrow:
		return NormalMode, ScrollDown
	case key.CodeUpArrow:
		return NormalMode, ScrollUp
	case key.Code0:
		if e.Modifiers == 0 {
			Repeat *= 10
		}
		return NormalMode, nil
	case key.Code1:
		if e.Modifiers == 0 {
			Repeat = Repeat*10 + 1
		}
		return NormalMode, nil
	case key.Code2:
		if e.Modifiers == 0 {
			Repeat = Repeat*10 + 2

		}
		return NormalMode, nil
	case key.Code3:
		if e.Modifiers == 0 {
			Repeat = Repeat*10 + 3

		}
		return NormalMode, nil
	case key.Code4:
		if e.Modifiers == 0 {
			Repeat = Repeat*10 + 4

		} else if e.Modifiers&key.ModShift != 0 {
			// $ is pressed, in most key maps..
			// it doesn't mean anything to repeat "End of line", so just
			// reset the repeat counter.
			Repeat = 0

			if e.Modifiers&key.ModControl != 0 {
				// ctrl is pressed, so just move the end and expand dot
				// BUG: This doesn't seem to work. Probably a bug in the shiny driver.
				actions.MoveCursor(position.DotStart, position.EndOfLine, buff)
			} else {
				actions.MoveCursor(position.EndOfLine, position.EndOfLine, buff)
			}

		}
		return NormalMode, nil
	case key.Code5:
		if e.Modifiers == 0 {
			Repeat = Repeat*10 + 5
		}
		return NormalMode, nil
	case key.Code6:
		if e.Modifiers == 0 {
			Repeat = Repeat*10 + 6
		} else if e.Modifiers&key.ModShift != 0 {
			// ^ is pressed, on most keyboards..
			// It doesn't mean anything to repeat "Start of line" either
			Repeat = 0

			if e.Modifiers&key.ModControl != 0 {
				// BUG: This doesn't seem to work. Probably a bug in the shiny driver.
				actions.MoveCursor(position.StartOfLine, position.DotEnd, buff)
			} else {
				actions.MoveCursor(position.StartOfLine, position.StartOfLine, buff)
			}

		}
		return NormalMode, nil
	case key.Code7:
		if e.Modifiers == 0 {
			Repeat = Repeat*10 + 7
		}
		return NormalMode, nil
	case key.Code8:
		if e.Modifiers == 0 {
			Repeat = Repeat*10 + 8
		}
		return NormalMode, nil
	case key.Code9:
		if e.Modifiers == 0 {
			Repeat = Repeat*10 + 9
		}
		return NormalMode, nil
	case key.CodeG:
		if e.Modifiers&key.ModShift != 0 {
			// G treats the repeat counter differently. If it's set, it means "Go to
			// lineno", and if it's not it means "Go to end of buffer"
			if Repeat > 0 {
				actions.MoveCursor(position.BuffStart, position.BuffStart, buff)
				for ; Repeat > 0; Repeat-- {
					actions.MoveCursor(position.NextLine, position.NextLine, buff)
				}
			} else {
				actions.MoveCursor(position.BuffEnd, position.BuffEnd, buff)
			}
		}
	case key.CodeD:
		return DeleteMode, nil
	case key.CodeSlash:
		if buff.Dot.Start == buff.Dot.End {
			actions.FindNext(position.CurWordStart, position.CurWordEnd, buff)

		} else {
			actions.FindNext(position.DotStart, position.DotEnd, buff)
		}
		return NormalMode, nil
	case key.CodeReturnEnter:
		if buff.Dot.Start == buff.Dot.End {
			actions.OpenOrPerformAction(position.CurWordStart, position.CurWordEnd, buff)
		} else {
			actions.OpenOrPerformAction(position.DotStart, position.DotEnd, buff)
		}
		return NormalMode, nil
	}
	return NormalMode, Invalid
}

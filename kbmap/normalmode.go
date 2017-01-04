package kbmap

import (
	//"fmt"

	"github.com/driusan/de/actions"
	"github.com/driusan/de/demodel"
	"github.com/driusan/de/position"
	"golang.org/x/mobile/event/key"
)

var Repeat uint

func normalMap(e key.Event, buff *demodel.CharBuffer, v demodel.Viewport) (demodel.Map, demodel.ScrollDirection, error) {
	// things only happen on key press in normal mode, if it's a release
	// or a repeat, ignore it. It's not an error
	if e.Direction == key.DirRelease {
		return NormalMode, demodel.DirectionNone, nil
	}
	if buff == nil {
		return NormalMode, demodel.DirectionNone, ErrInvalid
	}
	switch e.Code {
	case key.CodeEscape:
		actions.Do("SaveOrExit", buff, v)
		return NormalMode, demodel.DirectionUp, nil
	case key.CodeDeleteBackspace:
		if e.Direction == key.DirPress {
			actions.DeleteCursor(position.DotStart, position.DotEnd, buff)
		}
		// There's a potential that we're pressing backspace at the start of the
		// screen and may need to scroll up
		return NormalMode, demodel.DirectionUp, nil
	case key.CodeI:
		if e.Modifiers&key.ModShift == key.ModShift {
			actions.MoveCursor(position.StartOfLineAfterWhitespace, position.StartOfLineAfterWhitespace, buff)
		}

		return InsertMode, demodel.DirectionNone, nil
	case key.CodeA:
		if buff.Dot.End >= uint(len(buff.Buffer)) {
			return InsertMode, demodel.DirectionDown, nil
		}

		if e.Modifiers&key.ModShift == key.ModShift {
			actions.MoveCursor(position.EndOfLine, position.EndOfLine, buff)
		} else {
			if buff.Buffer[buff.Dot.End] == '\n' {
				// vi doesn't actually let you navigate your cursor *on top*
				// of the newline, while de does.
				// It feels weird to push 'a' there and end up on the next
				// line, so if we append to the end of a line, just leave
				// the cursor end where it is,
				actions.MoveCursor(position.DotEnd, position.DotEnd, buff)
			} else {
				actions.MoveCursor(position.NextChar, position.NextChar, buff)

			}
		}

		// There's a potentially we pressed 'a' at the very end of the screen and
		// need to scroll down
		return InsertMode, demodel.DirectionDown, nil
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

		return NormalMode, demodel.DirectionUp, nil
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

		return NormalMode, demodel.DirectionUp, nil

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

		return NormalMode, demodel.DirectionDown, nil
	case key.CodeJ:
		if e.Modifiers&key.ModShift != 0 {
			// 'J' for Join

			undoDot := -1
			if buff.Dot.Start == buff.Dot.End {
				if Repeat <= 1 {
					Repeat = 2
				}
				undoDot = int(buff.Dot.Start)
				// Join n lines
				actions.MoveCursor(position.StartOfLine, position.DotEnd, buff)
				for ; Repeat > 1; Repeat-- {
					actions.MoveCursor(position.DotStart, position.NextLine, buff)
				}
				actions.MoveCursor(position.DotStart, position.EndOfLine, buff)
			}

			buff.JoinLines(buff.Dot.Start, buff.Dot.End)
			if undoDot >= 0 {
				buff.Undo.Dot.Start = uint(undoDot)
				buff.Undo.Dot.End = uint(undoDot)
				buff.Dot = buff.Undo.Dot
			}
			v.InvalidateCache()
			Repeat = 0
			return NormalMode, demodel.DirectionNone, nil
		}

		// 'j' for movement
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

		return NormalMode, demodel.DirectionDown, nil
	case key.CodeX:
		if e.Direction == key.DirPress && isCopyModifier(e) {
			actions.CutClipboard(position.DotStart, position.DotEnd, buff)
			return NormalMode, demodel.DirectionNone, nil
		}

		if Repeat == 0 {
			Repeat = 1
		}
		for ; Repeat > 0; Repeat-- {
			actions.MoveCursor(position.DotStart, position.NextChar, buff)
		}
		actions.DeleteCursor(position.DotStart, position.DotEnd, buff)
		return NormalMode, demodel.DirectionNone, nil
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
		return NormalMode, demodel.DirectionDown, nil
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
		return NormalMode, demodel.DirectionUp, nil
	case key.CodeP:
		actions.InsertSnarfBuffer(position.DotStart, position.DotEnd, buff)
		return NormalMode, demodel.DirectionDown, nil
	case key.CodeRightArrow:
		// Arrow keys indicate their scroll direction via the error return value,
		// they return demodel.DirectionNone to make sure both code paths don't accidentally
		// get triggered
		return NormalMode, demodel.DirectionNone, ErrScrollRight
	case key.CodeLeftArrow:
		return NormalMode, demodel.DirectionNone, ErrScrollLeft
	case key.CodeDownArrow:
		return NormalMode, demodel.DirectionNone, ErrScrollDown
	case key.CodeUpArrow:
		return NormalMode, demodel.DirectionNone, ErrScrollUp
	case key.Code0:
		if e.Modifiers == 0 {
			Repeat *= 10
		}
		return NormalMode, demodel.DirectionNone, nil
	case key.Code1:
		if e.Modifiers == 0 {
			Repeat = Repeat*10 + 1
		}
		return NormalMode, demodel.DirectionNone, nil
	case key.Code2:
		if e.Modifiers == 0 {
			Repeat = Repeat*10 + 2

		}
		return NormalMode, demodel.DirectionNone, nil
	case key.Code3:
		if e.Modifiers == 0 {
			Repeat = Repeat*10 + 3

		}
		return NormalMode, demodel.DirectionNone, nil
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
		return NormalMode, demodel.DirectionNone, nil
	case key.Code5:
		if e.Modifiers == 0 {
			Repeat = Repeat*10 + 5
		}
		return NormalMode, demodel.DirectionNone, nil
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
		return NormalMode, demodel.DirectionNone, nil
	case key.Code7:
		if e.Modifiers == 0 {
			Repeat = Repeat*10 + 7
		}
		return NormalMode, demodel.DirectionNone, nil
	case key.Code8:
		if e.Modifiers == 0 {
			Repeat = Repeat*10 + 8
		}
		return NormalMode, demodel.DirectionNone, nil
	case key.Code9:
		if e.Modifiers == 0 {
			Repeat = Repeat*10 + 9
		}
		return NormalMode, demodel.DirectionNone, nil
	case key.CodeG:
		if e.Modifiers&key.ModShift != 0 {
			// G treats the repeat counter differently. If it's set, it means "Go to
			// lineno", and if it's not it means "Go to end of buffer"
			if Repeat > 0 {
				/* This seems to be buggy, so do it manually, at least until there's better
				   tests.
				actions.MoveCursor(position.BuffStart, position.BuffStart, buff)
				for ; Repeat > 0; Repeat-- {
					actions.MoveCursor(position.NextLine, position.NextLine, buff)
				}*/
				lineNo := uint(1)
				for i, c := range buff.Buffer {
					if c == '\n' {
						lineNo++
					}
					// should be equal, but if we've went too far at least it's
					// better to stop now.
					if lineNo >= Repeat {
						// we're on the \n itself, so add 1 to get to the actual line.
						if Repeat == 1 {
							buff.Dot.Start = 0
						} else {
							buff.Dot.Start = uint(i + 1)
						}
						buff.Dot.End = buff.Dot.Start
						Repeat = 0
						// select the line.
						actions.MoveCursor(position.DotStart, position.EndOfLine, buff)
						break
					}
				}
				// We don't know if the line being moved to is above or below, but send an "up"
				// hint so that it centers it at the top of the screen based on dot.
				return NormalMode, demodel.DirectionUp, nil
			}
			// moving to the end of the file because Repeat isn't set.
			actions.MoveCursor(position.BuffEnd, position.BuffEnd, buff)
			// send a down hint so it centers it at the bottom.
			return NormalMode, demodel.DirectionDown, nil
		}

	case key.CodeD:
		return DeleteMode, demodel.DirectionNone, nil
	case key.CodeU:
		actions.Do("Undo", buff, v)
		return NormalMode, demodel.DirectionNone, nil
	case key.CodeO:
		actions.MoveCursor(position.EndOfLineWithNewline, position.EndOfLineWithNewline, buff)

		newBuffer := make([]byte, len(buff.Buffer)+1)
		copy(newBuffer, buff.Buffer[:buff.Dot.Start])
		newBuffer[buff.Dot.Start] = '\n'
		copy(newBuffer[buff.Dot.Start+1:], buff.Buffer[buff.Dot.End:])

		buff.Buffer = newBuffer
		return InsertMode, demodel.DirectionDown, nil
	case key.CodeSlash:
		if buff.Dot.Start == buff.Dot.End {
			actions.FindNext(position.CurWordStart, position.CurWordEnd, buff)

		} else {
			actions.FindNext(position.DotStart, position.DotEnd, buff)
		}
		return NormalMode, demodel.DirectionDown, nil
	case key.CodeReturnEnter:
		if buff.Dot.Start == buff.Dot.End {
			actions.OpenOrPerformAction(position.CurExecutionWordStart, position.CurExecutionWordEnd, buff, v)
		} else {
			actions.OpenOrPerformAction(position.DotStart, position.DotEnd, buff, v)
		}
		// There's a possibility OpenOrPerformAction opened a new file, in which case
		// we should scroll to the top, or inserted text, in which case we should scroll
		// to the top of the inserted text. Either way, it's an "Up" hint so that the
		// scrolling is based on Dot.Start
		return v.GetKeyboardMode(), demodel.DirectionUp, nil
	case key.CodeSemicolon:
		if e.Modifiers&key.ModShift != 0 {
			if buff.Tagline != nil {
				buff.Tagline.Buffer = append(buff.Tagline.Buffer, ' ')

				buff.Tagline.Dot.Start = uint(len(buff.Tagline.Buffer))
				buff.Tagline.Dot.End = buff.Tagline.Dot.Start
			}
		}
		return TagMode, demodel.DirectionNone, nil
	case key.CodeV:
		if e.Direction == key.DirPress && isCopyModifier(e) {
			actions.PasteClipboard(position.DotStart, position.DotEnd, buff)
			return NormalMode, demodel.DirectionNone, nil
		}
	case key.CodeC:
		if e.Direction == key.DirPress && isCopyModifier(e) {
			actions.CopyClipboard(position.DotStart, position.DotEnd, buff)
			return NormalMode, demodel.DirectionNone, nil
		}
	}
	return NormalMode, demodel.DirectionNone, ErrInvalid
}

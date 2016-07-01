package kbmap

import (
	"unicode"
	"unicode/utf8"

	"github.com/driusan/de/actions"
	"github.com/driusan/de/demodel"
	"github.com/driusan/de/position"
	"golang.org/x/mobile/event/key"
)

func tagMap(e key.Event, buff *demodel.CharBuffer, v demodel.Viewport) (demodel.Map, demodel.ScrollDirection, error) {
	// special cases for Insert Mode
	switch e.Code {
	case key.CodeEscape:
		actions.Do("SaveOrExit", buff, v)
		return NormalMode, demodel.DirectionUp, nil
	case key.CodeDeleteBackspace:
		if e.Direction == key.DirPress {
			if e.Direction == key.DirPress {
				actions.DeleteCursor(position.DotStart, position.DotEnd, buff.Tagline)
			}
		}
		return TagMode, demodel.DirectionNone, nil
	case key.CodeLeftArrow:
		if e.Direction == key.DirPress {
			if e.Modifiers&key.ModControl != 0 {
				// ctrl is pressed, so just move the start and expand dot
				actions.MoveCursor(position.PrevChar, position.DotEnd, buff.Tagline)
			} else {
				// ctrl is not pressed, so move the cursor without selecting
				actions.MoveCursor(position.PrevChar, position.PrevChar, buff.Tagline)
			}
		}
		return TagMode, demodel.DirectionNone, nil
	case key.CodeRightArrow:
		if e.Direction == key.DirPress {
			if e.Modifiers&key.ModControl != 0 {
				// ctrl is pressed, so just move the end and expand dot
				actions.MoveCursor(position.DotStart, position.NextChar, buff.Tagline)
			} else {
				// ctrl is not pressed, so move the cursor without selecting
				actions.MoveCursor(position.NextChar, position.NextChar, buff.Tagline)
			}
		}
		return TagMode, demodel.DirectionNone, nil

	case key.CodeDownArrow:
		if e.Direction == key.DirPress {
			if e.Modifiers&key.ModControl != 0 {
				// ctrl is pressed, so just move the end and expand dot
				actions.MoveCursor(position.DotStart, position.NextLine, buff.Tagline)
			} else {
				// ctrl is not pressed, so move the cursor without selecting
				actions.MoveCursor(position.NextLine, position.NextLine, buff.Tagline)
			}
		}
		return TagMode, demodel.DirectionNone, nil

	case key.CodeUpArrow:
		if e.Direction == key.DirPress {
			if e.Modifiers&key.ModControl != 0 {
				// ctrl is pressed, so just move the start and expand dot
				actions.MoveCursor(position.PrevLine, position.DotEnd, buff.Tagline)
			} else {
				// ctrl is not pressed, so move the cursor without selecting
				actions.MoveCursor(position.PrevLine, position.PrevLine, buff.Tagline)
			}
		}
		return TagMode, demodel.DirectionNone, nil
	case key.CodeReturnEnter:
		if buff.Tagline.Dot.Start == buff.Tagline.Dot.End {
			//fmt.Printf("Executing tag from %s\n", *buff.Tagline)
			actions.PerformTagAction(position.CurTagExecutionWordStart, position.CurTagExecutionWordEnd, buff, v)
		} else {
			actions.PerformTagAction(position.TagDotStart, position.TagDotEnd, buff, v)
		}
		// if an action was performed, potentially scroll to the start of dot if it's off
		// the screen.
		return NormalMode, demodel.DirectionUp, nil

	}

	// These events don't seem to have the rune set properly, so add it as a hack
	if e.Code == key.CodeTab {
		e.Rune = '\t'
	}

	if e.Direction != key.DirPress {
		// add the character if it's a key release or a repeat, but not
		// if it's being released. For some reason, release seems more reliable
		// than press when typing fast.
		return TagMode, demodel.DirectionNone, nil
	}

	// unicode.IsPrint is selective about what whitespace it considers printable.
	if !unicode.IsPrint(e.Rune) && e.Rune != '\n' && e.Rune != '\t' {
		// if it's not a printable character, don't insert it. We also
		// receive key events on things like shift being pressed..
		return TagMode, demodel.DirectionNone, nil
	}
	// in insert the rune at the current position, overwriting Dot if applicable

	runeBytes := make([]byte, 4)
	i := utf8.EncodeRune(runeBytes, e.Rune)

	// inserting at the start of the file.
	if buff.Tagline.Dot.End == 0 {
		buff.Tagline.Buffer = append(runeBytes[:i], buff.Tagline.Buffer...)
		buff.Tagline.Dot.Start = uint(i)
		buff.Tagline.Dot.End = buff.Tagline.Dot.Start
	} else {
		newBuffer := make([]byte, len(buff.Tagline.Buffer)+i)
		copy(newBuffer, buff.Tagline.Buffer)
		copy(newBuffer[buff.Tagline.Dot.Start:], runeBytes[:i])
		copy(newBuffer[buff.Tagline.Dot.Start+uint(i):], buff.Tagline.Buffer[buff.Tagline.Dot.End:])
		//copy(newBuffer[:buff.Dot.Start+uint(i)+1], buff.Buffer[buff.Dot.End:])

		buff.Tagline.Buffer = newBuffer
		buff.Tagline.Dot.Start += uint(i)
		buff.Tagline.Dot.End = buff.Tagline.Dot.Start
	}
	return TagMode, demodel.DirectionNone, nil
}

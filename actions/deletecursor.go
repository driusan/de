package actions

import (
	"github.com/driusan/de/demodel"
)

func DeleteCursor(From, To demodel.Position, buff *demodel.CharBuffer) {
	if buff == nil {
		return
	}

	dot := demodel.Dot{}
	i, err := From(*buff)
	if err != nil {
		return
	}
	dot.Start = i

	i, err = To(*buff)
	if err != nil {
		return
	}
	dot.End = i

	// Update the undo buffer before doing anything.
	buff.Undo = &demodel.CharBuffer{
		Dot:    buff.Dot,
		Undo:   buff.Undo,
		Buffer: make([]byte, len(buff.Buffer)),
	}
	copy(buff.Undo.Buffer, buff.Buffer)
	if dot.Start == dot.End {
		// nothing selected, so delete the previous character

		if dot.Start == 0 {
			// can't delete past the beginning
			return
		}

		buff.Buffer = append(
			buff.Buffer[:dot.Start-1], buff.Buffer[dot.Start:]...,
		)

		// now adjust dot if it was inside the deleted range..
		if buff.Dot.Start == dot.Start {
			buff.Dot.Start--
			buff.Dot.End = buff.Dot.Start
		}
	} else {
		// delete the selected text.
		buff.SnarfBuffer = make([]byte, dot.End-dot.Start)
		copy(buff.SnarfBuffer, buff.Buffer[dot.Start:dot.End])

		buff.Buffer = append(buff.Buffer[:dot.Start], buff.Buffer[dot.End:]...)

		// now adjust dot if it was inside the deleted range..
		if buff.Dot.Start >= dot.Start && buff.Dot.End <= dot.End {
			buff.Dot.Start = dot.Start
			buff.Dot.End = buff.Dot.Start
		}
	}
	buff.Dirty = true
}

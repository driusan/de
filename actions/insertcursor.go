package actions

import (
	"github.com/driusan/de/demodel"
)

func InsertSnarfBuffer(From, To demodel.Position, buff *demodel.CharBuffer) {
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
	// inserting at the start of the file.
	if dot.End == 0 {
		newBuffer := make([]byte, len(buff.Buffer)+len(buff.SnarfBuffer))
		copy(newBuffer, buff.SnarfBuffer)
		copy(newBuffer[len(buff.SnarfBuffer):], buff.Buffer)

		buff.Buffer = newBuffer
		buff.Dot.Start = 0
		buff.Dot.End = buff.Dot.Start
	} else {
		newBuffer := make([]byte, len(buff.Buffer)+len(buff.SnarfBuffer)-int(buff.Dot.End-buff.Dot.Start))
		copy(newBuffer, buff.Buffer)
		copy(newBuffer[buff.Dot.Start:], buff.SnarfBuffer)
		copy(newBuffer[buff.Dot.Start+uint(len(buff.SnarfBuffer)):], buff.Buffer[buff.Dot.End:])

		buff.Buffer = newBuffer
		buff.Dot.End = buff.Dot.Start + uint(len(buff.SnarfBuffer))
		buff.Dot.Start = buff.Dot.End
	}
	buff.Dirty = true
}

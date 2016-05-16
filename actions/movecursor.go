package actions

import (
	"github.com/driusan/de/demodel"
)

func MoveCursor(From, To demodel.Position, buff *demodel.CharBuffer) {
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
	buff.Dot = dot

}

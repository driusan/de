package actions

import (
	"github.com/atotto/clipboard"
	"github.com/driusan/de/demodel"
)

func PasteClipboard(From, To demodel.Position, buff *demodel.CharBuffer) {
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

	clipboardData, err := clipboard.ReadAll()
	if err != nil {
		buff.AppendTag("\nError reading clipboard: " + err.Error())
		return
	}
	// Update the undo buffer before doing anything.
	buff.Undo = &demodel.CharBuffer{
		Buffer: buff.Buffer,
		Dot:    buff.Dot,
		Undo:   buff.Undo,
	}

	clipLen := uint(len([]byte(clipboardData)))
	newBuffer := make([]byte, uint(len(buff.Buffer))-(dot.End-dot.Start)+clipLen)

	copy(newBuffer, buff.Buffer[0:dot.Start])
	copy(newBuffer[dot.Start:dot.Start+clipLen], []byte(clipboardData))
	copy(newBuffer[dot.Start+clipLen:], buff.Buffer[dot.End:])

	buff.Buffer = newBuffer
	buff.Dot.End = dot.Start + clipLen
	buff.Dirty = true

}

func CopyClipboard(From, To demodel.Position, buff *demodel.CharBuffer) {
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
	if err := clipboard.WriteAll(string(buff.Buffer[dot.Start:dot.End])); err != nil {
		buff.AppendTag("\nError writing clipboard: " + err.Error())
	}

}
func CutClipboard(From, To demodel.Position, buff *demodel.CharBuffer) {
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

	if err := clipboard.WriteAll(string(buff.Buffer[dot.Start:dot.End])); err != nil {
		buff.AppendTag("\nError cutting to clipboard: " + err.Error())
		return
	}

	// Update the undo buffer before doing anything.
	buff.Undo = &demodel.CharBuffer{
		Buffer: buff.Buffer,
		Dot:    buff.Dot,
		Undo:   buff.Undo,
	}

	newBuffer := make([]byte, uint(len(buff.Buffer))-(dot.End-dot.Start))

	copy(newBuffer, buff.Buffer[0:dot.Start])
	copy(newBuffer[dot.Start:], buff.Buffer[dot.End:])

	buff.Buffer = newBuffer
	buff.Dot.End = dot.Start
	buff.Dirty = true
}

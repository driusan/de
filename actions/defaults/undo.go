package defaults

import (
	"github.com/driusan/de/actions"
	"github.com/driusan/de/demodel"
)

func init() {
	// Basic Save and Discard
	actions.RegisterAction("Undo", Undo)

}

func Undo(args string, buff *demodel.CharBuffer, v demodel.Viewport) {
	if buff.Undo != nil {
		buff.Buffer = buff.Undo.Buffer
		buff.Dot = buff.Undo.Dot
		buff.Undo = buff.Undo.Undo
		v.InvalidateCache()
		v.Rerender()
	}
}

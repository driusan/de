package defaults

import (
	"github.com/driusan/de/actions"
	"github.com/driusan/de/demodel"
	"github.com/driusan/de/position"
	"strconv"
)

func init() {
	actions.RegisterAction("Join", Join)
}

// Join joins multiple lines together as if you had pressed 'J' in vi,
// but the number of lines to join is an optional argument instead of
// a repeat command. If no argument is specified, the selected lines are
// joined together.
func Join(args string, buff *demodel.CharBuffer, v demodel.Viewport) {
	undoDot := -1
	if buff.Dot.Start == buff.Dot.End {
		// If there's an argument, only use it if nothing is selected.
		// Otherwise, join the selected lines.
		var LinesArg int
		var err error
		if args == "" {
			LinesArg = 2
		} else if LinesArg, err = strconv.Atoi(args); err != nil {
			buff.AppendTag("\nInvalid argument to Join. Must be number: " + err.Error())
			return

		}

		if LinesArg <= 1 {
			LinesArg = 2
		}
		undoDot = int(buff.Dot.Start)
		// Join n lines
		actions.MoveCursor(position.StartOfLine, position.DotEnd, buff)
		for ; LinesArg > 1; LinesArg-- {
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
	v.GetRenderer().InvalidateCache()
}

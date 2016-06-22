package defaults

import (
	"github.com/driusan/de/actions"
	"github.com/driusan/de/demodel"
)

func init() {
	actions.RegisterAction("Autoindent", Autoindent)
}

var AutoIndentEnabled bool

func Autoindent(args string, buff *demodel.CharBuffer, v demodel.Viewport) {
	switch args {
	default:
		fallthrough
	case "on", "On":
		AutoIndentEnabled = true
		buff.AppendTag("\nAutoindent:on is set")

	case "off", "Off":
		AutoIndentEnabled = false
		buff.AppendTag("\nAutoindent:off is set")
	}
}

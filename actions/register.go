package actions

import (
	"github.com/driusan/de/demodel"
	//"github.com/driusan/de/viewer"
)

func init() {
	actions = make(map[string]func(string, *demodel.CharBuffer, demodel.Viewport))
}

var actions map[string]func(string, *demodel.CharBuffer, demodel.Viewport)

func RegisterAction(cmd string, f func(string, *demodel.CharBuffer, demodel.Viewport)) {
	actions[cmd] = f

}

package actions

import (
	"github.com/driusan/de/demodel"
)

func init() {
	actions = make(map[string]func(string, *demodel.CharBuffer))
}

var actions map[string]func(string, *demodel.CharBuffer)

func RegisterAction(cmd string, f func(string, *demodel.CharBuffer)) {
	actions[cmd] = f

}

package actions

import (
	"github.com/driusan/de/demodel"
	//"github.com/driusan/de/viewer"
)

func init() {
	actions = make(map[string]func(string, *demodel.CharBuffer, demodel.Viewport))

	// This needs to go here instead of in an init function where the Alias
	// command is defined in order to make sure that the above map is created
	// first, since it's in the same package.
	RegisterAction("Alias", Alias)
}

var actions map[string]func(string, *demodel.CharBuffer, demodel.Viewport)

func RegisterAction(cmd string, f func(string, *demodel.CharBuffer, demodel.Viewport)) {
	actions[cmd] = f

}

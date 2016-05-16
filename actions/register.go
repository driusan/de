package actions

import (
	"fmt"
	"github.com/driusan/de/demodel"
	"io/ioutil"
	"os"
)

func init() {
	actions = make(map[string]func(string, *demodel.CharBuffer))

	RegisterAction("Put", Put)
	RegisterAction("Save", Put)

	RegisterAction("Get", Get)
	RegisterAction("Discard", Get)

	RegisterAction("Exit", Quit)
	RegisterAction("Quit", Quit)
}

var actions map[string]func(string, *demodel.CharBuffer)

func RegisterAction(cmd string, f func(string, *demodel.CharBuffer)) {
	actions[cmd] = f

}

// Some default actions.
func Put(args string, buff *demodel.CharBuffer) {
	SaveFile(nil, nil, buff)
}

func Get(args string, buff *demodel.CharBuffer) {
	if buff == nil {
		return
	}
	b, err := ioutil.ReadFile(buff.Filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}
	buff.Buffer = b
}

func Quit(args string, buff *demodel.CharBuffer) {
	os.Exit(0)
}

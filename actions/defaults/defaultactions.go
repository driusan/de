package defaults

import (
	"fmt"
	"github.com/driusan/de/actions"
	"github.com/driusan/de/demodel"
	"io/ioutil"
	"os"
)

func init() {
	actions.RegisterAction("Put", Put)
	actions.RegisterAction("Save", Put)

	actions.RegisterAction("Get", Get)
	actions.RegisterAction("Discard", Get)

	actions.RegisterAction("Exit", Quit)
	actions.RegisterAction("Quit", Quit)
}

func Put(args string, buff *demodel.CharBuffer) {
	// Just use the savefile movement command that's used saving with
	// escape.
	actions.SaveFile(nil, nil, buff)
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

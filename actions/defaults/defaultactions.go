package defaults

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/driusan/de/actions"
	"github.com/driusan/de/demodel"
	"github.com/driusan/de/position"
	"github.com/driusan/de/renderer"
)

func init() {
	// Basic Save and Discard
	actions.RegisterAction("Put", Put)
	actions.RegisterAction("Save", Put)
	actions.RegisterAction("w", Put)

	actions.RegisterAction("Get", Get)
	actions.RegisterAction("Discard", Get)

	// Quit if clean, provide a warning otherwise
	actions.RegisterAction("Quit", Exit)
	actions.RegisterAction("Exit", Exit)
	actions.RegisterAction("q", Exit)

	// Quit regardless of cleanliness
	actions.RegisterAction("ForceExit", ForceQuit)
	actions.RegisterAction("ForceQuit", ForceQuit)
	actions.RegisterAction("q!", ForceQuit)

	// Save AND Exit
	actions.RegisterAction("SaveExit", SaveAndExit)
	actions.RegisterAction("SaveQuit", SaveAndExit)
	actions.RegisterAction("wq", SaveAndExit)
	actions.RegisterAction("wq!", SaveAndExit)
	actions.RegisterAction("x", SaveAndExit)

	// Save OR Exit, depending on dirty buffer status,
	// default for Escape key.
	actions.RegisterAction("SaveOrExit", SaveOrExit)
	actions.RegisterAction("SaveOrQuit", SaveOrExit)
	actions.RegisterAction("Paste", Paste)
	actions.RegisterAction("ResetTagline", ResetTagline)
	actions.RegisterAction("Tagline", Tagline)

	// Change the renderer by name
	actions.RegisterAction("Renderer", SwitchRenderer)
}

func Put(args string, buff *demodel.CharBuffer, v demodel.Viewport) {
	// Just use the savefile movement command that's used saving with
	// escape.
	actions.SaveFile(nil, nil, buff)
}

func Get(args string, buff *demodel.CharBuffer, v demodel.Viewport) {
	if buff == nil {
		return
	}
	b, err := ioutil.ReadFile(buff.Filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}
	buff.Buffer = b
	buff.Dirty = false
}

func Exit(args string, buff *demodel.CharBuffer, v demodel.Viewport) {
	if buff.Dirty {
		buff.AppendTag("\nFile has been modified. Save or Discard changes or ForceQuit")
		return
	}
	buff.SaveSnarfBuffer()
	os.Exit(0)
}
func ForceQuit(args string, buff *demodel.CharBuffer, v demodel.Viewport) {
	buff.SaveSnarfBuffer()
	if buff.Dirty {
		os.Exit(1)
	}
	os.Exit(0)
}

func SaveAndExit(args string, buff *demodel.CharBuffer, v demodel.Viewport) {
	buff.SaveSnarfBuffer()
	actions.SaveFile(nil, nil, buff)
	os.Exit(0)
}

func SaveOrExit(args string, buff *demodel.CharBuffer, v demodel.Viewport) {
	if buff.Dirty {
		actions.SaveFile(nil, nil, buff)
		return
	}
	buff.SaveSnarfBuffer()
	os.Exit(0)
}
func Paste(args string, buff *demodel.CharBuffer, v demodel.Viewport) {
	actions.InsertSnarfBuffer(position.DotStart, position.DotEnd, buff)
}

func ResetTagline(args string, buff *demodel.CharBuffer, v demodel.Viewport) {
	buff.ResetTagline()
}

func Tagline(args string, buff *demodel.CharBuffer, v demodel.Viewport) {
	buff.SetTagline(args)
}

func SwitchRenderer(args string, buff *demodel.CharBuffer, v demodel.Viewport) {
	r := renderer.GetNamedRenderer(args)
	if r == nil || v == nil {
		return
	}
	v.SetRenderer(r)
	v.Rerender()
}

package viewer

import (
	"errors"
	//	"fmt"
	//"github.com/driusan/de/actions"
	"github.com/driusan/de/demodel"
	//	"github.com/driusan/de/kbmap"
	"golang.org/x/exp/shiny/screen"
	"image"
)

type Viewport struct {
	demodel.Map
	demodel.Renderer
	Location image.Point

	kbLocked bool
	Window   screen.Window

	//	KeyboardMode demodel.Map
}

type RequestRerender struct{}

func (v *Viewport) Rerender() {
	v.Window.Send(RequestRerender{})

}
func (v *Viewport) GetKeyboardMode() demodel.Map {
	return v.Map
}

var KBLockedError error = errors.New("Keyboard mode is locked")

// Sets the keyboard mode in a way that future requests to SetKeyboardMode
// will fail.
func (v *Viewport) LockKeyboardMode(m demodel.Map) error {
	if v.kbLocked {
		return KBLockedError
	}
	v.kbLocked = true
	v.Map = m
	return nil
}

// Unlocks a locked keyboard mode. The locked mode must be passed as a
// parameter to ensure that it's the same caller that locked it.
// It's not secure, but it prevents plugins from accidentally unlocking
// someone else's locked keyboard.
// Returns KBLockedError if d doesn't equal the mode that it's locked to.
func (v *Viewport) UnlockKeyboardMode(d demodel.Map) error {
	if v.Map == d {
		v.kbLocked = false
		return nil
	}
	return KBLockedError
}
func (v *Viewport) SetKeyboardMode(m demodel.Map) error {
	if v.kbLocked {
		return KBLockedError
	}
	v.Map = m
	//fmt.Print("Set keyboard mode!")
	return nil
}

func (v *Viewport) GetRenderer() demodel.Renderer {
	return v.Renderer
}

func (v *Viewport) SetRenderer(r demodel.Renderer) error {
	v.Renderer = r
	return nil
}

package kbmap

import (
	"errors"
	"github.com/driusan/de/demodel"
	"golang.org/x/mobile/event/key"
)

// A ScrollDirection is a hint to the viewport for the direction that a keystroke
// might potentially scroll the buffer, so that it knows if it should scroll from
// the start, or the end Dot when it goes out of view.

var ErrInvalid = errors.New("Invalid keyboard map.")
var ErrExitProgram = errors.New("Keystroke wants to exit the program.")
var ErrScrollDown = errors.New("Keystroke wants to scroll the window down.")
var ErrScrollUp = errors.New("Keystroke wants to scroll the window up.")
var ErrScrollLeft = errors.New("Keystroke wants to scroll the window left.")
var ErrScrollRight = errors.New("Keystroke wants to scroll the window right.")

type defaultMaps uint

const (
	NormalMode = defaultMaps(iota)
	InsertMode
	DeleteMode
	TagMode
)

func (m defaultMaps) HandleKey(e key.Event, buff *demodel.CharBuffer, v demodel.Viewport) (demodel.Map, demodel.ScrollDirection, error) {
	switch m {
	case NormalMode:
		return normalMap(e, buff, v)
	case InsertMode:
		return insertMap(e, buff, v)
	case DeleteMode:
		return deleteMap(e, buff, v)
	case TagMode:
		return tagMap(e, buff, v)
	}
	return nil, demodel.DirectionNone, ErrInvalid

}

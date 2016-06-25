// +build !darwin

package kbmap

import (
	"golang.org/x/mobile/event/key"
)

func isCopyModifier(e key.Event) bool {
	return e.Modifiers&key.ModControl != 0
}

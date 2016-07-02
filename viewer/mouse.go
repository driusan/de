package viewer

import (
	"github.com/driusan/de/demodel"
	"github.com/driusan/de/renderer"
	"golang.org/x/mobile/event/mouse"
	"image"
)

// Handles a mousewheel event. Returns true if the event did something that
// should result in a repainting of the screen.
func (v *Viewport) HandleMouseWheel(e mouse.Event, buff *demodel.CharBuffer, wSize image.Point) bool {
	var scrollY int
	imgSize := v.Renderer.Bounds(buff)
	lineSize := renderer.MonoFontFace.Metrics().Height.Ceil()
	switch e.Button {
	case mouse.ButtonWheelUp:
		if v.Location.Y == 0 {
			return false
		}
		scrollY = -lineSize
	case mouse.ButtonWheelDown:
		if v.Location.Y == wSize.Y-50 {
			return false
		}
		scrollY = lineSize
	}

	if scrollY == 0 {
		// This should be unreachable.
		return false
	}

	v.Location.Y += scrollY
	if v.Location.Y < 0 || imgSize.Size().Y < wSize.Y {
		v.Location.Y = 0
	} else if v.Location.Y+wSize.Y > imgSize.Max.Y+50 {
		v.Location.Y = imgSize.Max.Y - wSize.Y + 50
	}
	return true
}

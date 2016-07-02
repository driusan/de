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
	var scrollAmt int
	var scrollAxis *int
	lineSize := renderer.MonoFontFace.Metrics().Height.Ceil()
	switch e.Button {
	case mouse.ButtonWheelUp:
		if v.Location.Y == 0 {
			return false
		}
		scrollAxis = &v.Location.Y
		scrollAmt = -lineSize
	case mouse.ButtonWheelDown:
		if v.Location.Y == wSize.Y-50 {
			return false
		}
		scrollAxis = &v.Location.Y
		scrollAmt = lineSize
	case mouse.ButtonWheelLeft:
		scrollAxis = &v.Location.X
		scrollAmt = -lineSize
	case mouse.ButtonWheelRight:
		scrollAxis = &v.Location.X
		scrollAmt = lineSize
	default:
		scrollAmt = 0
		scrollAxis = nil
	}

	if scrollAmt == 0 || scrollAxis == nil {
		// This should be unreachable.
		return false
	}

	*scrollAxis += scrollAmt
	imgSize := v.Renderer.Bounds(buff).Size()
	if v.Location.Y < 0 || imgSize.Y < wSize.Y {
		v.Location.Y = 0
	} else if v.Location.Y+wSize.Y > imgSize.Y+50 {
		v.Location.Y = imgSize.Y - wSize.Y + 50
	}

	if v.Location.X < 0 || imgSize.X < wSize.X {
		v.Location.X = 0
	} else if v.Location.X+wSize.X > imgSize.X+50 {
		v.Location.X = imgSize.X - wSize.X + 50
	}

	return true
}

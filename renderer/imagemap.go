package renderer

import (
	"errors"
	"image"

	"github.com/driusan/de/demodel"
	"golang.org/x/image/math/fixed"
)

var ErrNoCharacter = errors.New("No character under the mouse cursor.")

type ImageLoc struct {
	Loc fixed.Rectangle26_6

	Idx uint
}
type ImageMap struct {
	IMap []ImageLoc
	Buf  *demodel.CharBuffer
}

func inRectangle(x, y int, r fixed.Rectangle26_6) bool {
	return x >= r.Min.X.Ceil() && x <= r.Max.X.Floor() && y >= r.Min.Y.Ceil() && y <= r.Max.Y.Floor()
}
func (im ImageMap) At(x, y int) (uint, error) {
	for _, m := range im.IMap {
		if inRectangle(x, y, m.Loc) {
			return m.Idx, nil
		}
	}
	return 0, ErrNoCharacter
}

// Returns the bounding rectangle for the character at index idx
// in the character buffer.
func (im ImageMap) Get(idx uint) (image.Rectangle, error) {
	for _, chr := range im.IMap {
		if chr.Idx == idx {
			return image.Rectangle{
				Min: image.Point{chr.Loc.Min.X.Ceil(), chr.Loc.Min.Y.Ceil()},
				Max: image.Point{chr.Loc.Max.X.Ceil(), chr.Loc.Max.Y.Ceil()},
			}, nil
		}
	}
	return image.ZR, ErrNoCharacter
}

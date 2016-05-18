package renderer

import (
	"errors"
	"github.com/driusan/de/demodel"
	"image"
	"image/color"
)

var NoCharacter error = errors.New("No character under the mouse cursor.")

type ImageLoc struct {
	Loc image.Rectangle
	Idx uint
}
type ImageMap struct {
	IMap []ImageLoc
	Buf  *demodel.CharBuffer
}

func (im ImageMap) At(x, y int) (uint, error) {
	for _, m := range im.IMap {
		if m.Loc.At(x, y) == color.Opaque {
			return m.Idx, nil
		}
	}
	return 0, NoCharacter
}

// Returns the bounding rectangle for the character at index idx
// in the character buffer.
func (im ImageMap) Get(idx uint) (image.Rectangle, error) {
	for _, chr := range im.IMap {
		if chr.Idx == idx {
			return chr.Loc, nil
		}
	}
	return image.ZR, NoCharacter
}

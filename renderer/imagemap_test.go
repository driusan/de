package renderer

import (
	"fmt"
	"image"
	"testing"

	"github.com/driusan/de/demodel"
)

func TestSimpleDefaultImageMap(t *testing.T) {
	buffer := &demodel.CharBuffer{
		Buffer: []byte("Hello\nThere\nFriend"),
	}

	type expected struct {
		Idx uint
		Val image.Rectangle
	}

	expectedValues := []expected{
		expected{
			Idx: 0,
			Val: image.Rectangle{
				Min: image.ZP,
				Max: image.Point{MonoFontAdvance.Ceil(), MonoFontHeight.Ceil()},
			},
		},
		expected{
			Idx: 1,
			Val: image.Rectangle{
				Min: image.Point{MonoFontAdvance.Ceil(), 0},
				Max: image.Point{2 * MonoFontAdvance.Ceil(), MonoFontHeight.Ceil()},
			},
		},
		expected{
			Idx: 6,
			Val: image.Rectangle{
				Min: image.Point{0, MonoFontHeight.Ceil()},
				Max: image.Point{MonoFontAdvance.Ceil(), 2 * MonoFontHeight.Ceil()},
			},
		},
	}

	var defIMap DefaultImageMapper
	imageMap := defIMap.GetImageMap(buffer, image.ZR)

	for _, e := range expectedValues {
		r, err := imageMap.Get(e.Idx)

		if r != e.Val {
			fmt.Printf("Incorrect value for index %d. Got %v expected %v\n", e.Idx, r, e.Val)
			t.Fail()
		}
		if err != nil {
			fmt.Printf("Unexpected error getting character %d: %v", e.Idx, err)
			t.Fail()
		}

	}

}

package viewer

import (
	"fmt"
	"image"

	"github.com/driusan/de/demodel"
	"github.com/driusan/de/renderer"
)

type offsetImageMap struct {
	buf             *demodel.CharBuffer
	visibleViewport image.Rectangle
	viewport        *Viewport
}

func (v *Viewport) GetImageMap(buf *demodel.CharBuffer, viewport image.Rectangle) demodel.ImageMap {
	return &offsetImageMap{buf, viewport, v}
}

func (i *offsetImageMap) At(x, y int) (uint, error) {
	if i == nil || i.viewport == nil || i.viewport.Renderer == nil {
		return 0, fmt.Errorf("Viewport renderer is nil")
	}
	switch i.viewport.lineNumberMode {
	case NoLineNumbers:
		im := i.viewport.Renderer.GetImageMap(i.buf, i.visibleViewport)
		if im == nil {
			return 0, fmt.Errorf("Renderer does not provide an image map.")
		}
		return im.At(x, y)
	default:
		lineNumberOffset := renderer.MonoFontAdvance * 6
		if x < lineNumberOffset.Ceil() {
			imap := i.viewport.Renderer.GetImageMap(i.buf, i.visibleViewport)
			if imap == nil {
				return 0, fmt.Errorf("Renderer imagemap is nil")
			}
			return imap.At(lineNumberOffset.Ceil(), y)
		}
		return i.viewport.Renderer.GetImageMap(i.buf, i.visibleViewport).At(x-lineNumberOffset.Floor(), y)
	}

}

func (i *offsetImageMap) Get(idx uint) (image.Rectangle, error) {
	switch i.viewport.lineNumberMode {
	case NoLineNumbers:
		return i.viewport.Renderer.GetImageMap(i.buf, i.visibleViewport).Get(idx)
	default:
		lineNumberOffset := renderer.MonoFontAdvance * 6
		r, err := i.viewport.Renderer.GetImageMap(i.buf, i.visibleViewport).Get(idx)
		r.Min.X += lineNumberOffset.Floor()
		r.Max.X += lineNumberOffset.Floor()
		return r, err
	}
}

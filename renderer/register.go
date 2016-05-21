package renderer

import (
	"github.com/driusan/de/demodel"
	"image"
)

type Renderer interface {
	// Determines whether this renderer knows how to render this character buffer.
	// The most recently registered plugin that can handle it wins (because otherwise
	// everything would use the default renderer, which claims it can render anything.)
	CanRender(*demodel.CharBuffer) bool

	// Given a character buffer, and a clipping region (viewport) to be displayed, renders an image that
	// should be shown on the screen, a rectangle determining what the size of the *entire* buffer rendered
	// would be (even if it didn't render it), an ImageMap that, at least for the portion rendered, can
	// be used to determine what any pixel represents, and an error (hopefully nil.)
	Render(buffer *demodel.CharBuffer, viewport image.Rectangle) (image.Image, image.Rectangle, ImageMap, error)
}

var renderers []Renderer

func init() {
	// Make sure renderers is initialized with at least 1 renderer
	// that can render anything.
	renderers = []Renderer{NoSyntaxRenderer{}}

}
func RegisterRenderer(r Renderer) {
	renderers = append(renderers, r)
}

func GetRenderer(buff *demodel.CharBuffer) Renderer {
	for i := len(renderers) - 1; i >= 0; i-- {
		if renderers[i].CanRender(buff) {
			return renderers[i]
		}
	}
	return renderers[0]
}

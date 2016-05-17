package renderer

import (
	"github.com/driusan/de/demodel"
	"image"
)

type Renderer interface {
	CanRender(demodel.CharBuffer) bool
	Render(demodel.CharBuffer) (image.Image, ImageMap, error)
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

func GetRenderer(buff demodel.CharBuffer) Renderer {
	for i := len(renderers) - 1; i >= 0; i-- {
		if renderers[i].CanRender(buff) {
			return renderers[i]
		}
	}
	return renderers[0]
}

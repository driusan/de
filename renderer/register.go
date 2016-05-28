package renderer

import (
	"github.com/driusan/de/demodel"
)

// maps don't range in any particular order, so we store the renderers
// in an array and just range through them checking the name. That way,
// we can either get them in order of registration or by name and don't
// need to maintain two structures.
type namedRenderer struct {
	name     string
	renderer demodel.Renderer
}

var renderers []namedRenderer

func RegisterRenderer(name string, r demodel.Renderer) {
	renderers = append(renderers, namedRenderer{name, r})
}

func GetRenderer(buff *demodel.CharBuffer) demodel.Renderer {
	for i := len(renderers) - 1; i >= 0; i-- {
		if renderers[i].renderer.CanRender(buff) {
			return renderers[i].renderer
		}
	}
	return renderers[0].renderer
}
func GetNamedRenderer(name string) demodel.Renderer {
	for i := len(renderers) - 1; i >= 0; i-- {
		if renderers[i].name == name {
			return renderers[i].renderer
		}
	}
	return renderers[0].renderer
}

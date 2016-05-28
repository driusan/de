package imagerenderer

import (
	"github.com/driusan/de/renderer"
)

func init() {
	renderer.RegisterRenderer("image", &ImageRenderer{})
}

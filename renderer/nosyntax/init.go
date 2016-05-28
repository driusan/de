package renderer

import (
	"github.com/driusan/de/renderer"
)

func init() {
	nsrender := &NoSyntaxRenderer{}
	renderer.RegisterRenderer("nosyntax", nsrender)
	renderer.RegisterRenderer("default", nsrender)
}

package htmlrenderer

import (
	"github.com/driusan/de/renderer"
)

func init() {
	renderer.RegisterRenderer("html", &HTMLSyntax{})
	renderer.RegisterRenderer("css", &HTMLSyntax{})
}

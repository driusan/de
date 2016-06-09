package htmlrenderer

import (
	"github.com/driusan/de/renderer"
)

func init() {
	renderer.RegisterRenderer("html", &HtmlSyntax{})
	renderer.RegisterRenderer("css", &HtmlSyntax{})
}

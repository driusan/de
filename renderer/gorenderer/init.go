package renderer

import (
	"github.com/driusan/de/renderer"
)

func init() {
	renderer.RegisterRenderer(&GoSyntax{})
}

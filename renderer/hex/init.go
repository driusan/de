package hex

import (
	"github.com/driusan/de/renderer"
)

func init() {
	renderer.RegisterRenderer("hex", &HexRenderer{})
}

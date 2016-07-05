// +build !darwin

package viewer

import (
	"github.com/driusan/de/renderer"
)

func getScrollAmt() int {
	// by default, a scrollwheel event should scroll by 1 line. This varies
	// on other operating systems that don't have linear scrolling events
	// (ie. Mac)
	return renderer.MonoFontFace.Metrics().Height.Ceil()
}

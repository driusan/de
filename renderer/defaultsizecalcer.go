package renderer

import (
	"bytes"
	"github.com/driusan/de/demodel"
	"golang.org/x/image/math/fixed"
	"image"
)

// A DefaultSizeCalcer is a type that can be mixed in to
// renders to provide a default implementation of Bounds,
// for calculating what the bounds would be with the NoSyntaxRenderer.
// This is useful for renderers which only provide syntax
// highlighting but do nothing that would change the size of
// the image rendered.
type DefaultSizeCalcer struct {
	LastBuf     *demodel.CharBuffer
	LastBufSize int
	SizeCache   image.Rectangle
}

func (r *DefaultSizeCalcer) InvalidateCache() {
	r.SizeCache = image.ZR
}
func (r *DefaultSizeCalcer) Bounds(buf *demodel.CharBuffer) image.Rectangle {
	if r.SizeCache != image.ZR && r.LastBuf == buf && r.LastBufSize == len(buf.Buffer) {
		return r.SizeCache
	}
	r.LastBufSize = len(buf.Buffer)
	r.LastBuf = buf

	metrics := MonoFontFace.Metrics()
	runes := bytes.Runes(buf.Buffer)
	_, glyphWidth, _ := MonoFontFace.GlyphBounds('a')
	rt := image.ZR
	var lineSize fixed.Int26_6
	for _, r := range runes {
		switch r {
		case '\t':
			lineSize += glyphWidth * 8
		case '\n':
			rt.Max.Y += metrics.Height.Ceil()
			lineSize = 0
		default:
			lineSize += glyphWidth
		}
		if lineSize.Ceil() > rt.Max.X {
			rt.Max.X = lineSize.Ceil()
		}
	}
	rt.Max.Y += metrics.Height.Ceil() + 1
	r.SizeCache = rt
	return rt
}

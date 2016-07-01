package imagerenderer

import (
	"bytes"
	//"fmt"
	"github.com/driusan/de/demodel"
	"github.com/driusan/de/renderer"
	"image"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
)

type ImageRenderer struct {
	cache image.Image
}

func (r ImageRenderer) InvalidateCache() {
	// there is no cache, this is just here to meet the interface requirements.
}
func (r ImageRenderer) Bounds(buf *demodel.CharBuffer) image.Rectangle {
	if r.cache == nil {
		return image.ZR
	}
	return r.cache.Bounds()
}
func (r ImageRenderer) CanRender(buf *demodel.CharBuffer) bool {
	// Check for a PNG signature
	if len(buf.Buffer) > 8 && bytes.Equal(buf.Buffer[:8], []byte{
		// PNG magic bytes/signature according to wikipedia.
		0x89,
		0x50, 0x4E, 0x47, // PNG
		0x0D, 0x0A, // \r\n
		0x1A,
		0x0A,
	}) {
		return true
	}

	// Check for a JPEG signature.
	if len(buf.Buffer) > 3 && bytes.Equal(buf.Buffer[:3], []byte{
		0xFF, 0xD8, 0xFF,
	}) {
		// technically there's more to the signature than this, but different variations
		// of JPEG have different signatures according to wikipedia and this is the part
		// that's in common. It should be enough, since the high bit it set it's unlikely to
		// be handled by another renderer anyways.
		return true
	}
	return false
}

func (r *ImageRenderer) RenderInto(dst draw.Image, buf *demodel.CharBuffer, viewport image.Rectangle) error {
	bReader := bytes.NewReader(buf.Buffer)
	img, _, err := image.Decode(bReader)
	r.cache = img
	draw.Draw(dst, dst.Bounds(), img, viewport.Min, draw.Src)
	return err
}

func (r *ImageRenderer) GetImageMap(buf *demodel.CharBuffer, viewport image.Rectangle) demodel.ImageMap {
	return &renderer.ImageMap{}
}

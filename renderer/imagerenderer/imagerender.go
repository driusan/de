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
	return r.cache.Bounds()
}
func (r ImageRenderer) CanRender(buf *demodel.CharBuffer) bool {
	// Check for a PNG signature
	if len(buf.Buffer) > 8 && bytes.Compare(buf.Buffer[:8], []byte{
		// PNG magic bytes/signature according to wikipedia.
		0x89,
		0x50, 0x4E, 0x47, // PNG
		0x0D, 0x0A, // \r\n
		0x1A,
		0x0A,
	}) == 0 {
		return true
	}

	// Check for a JPEG signature.
	if len(buf.Buffer) > 3 && bytes.Compare(buf.Buffer[:3], []byte{
		0xFF, 0xD8, 0xFF,
	}) == 0 {
		// technically there's more to the signature than this, but different variations
		// of JPEG have different signatures according to wikipedia and this is the part
		// that's in common. It should be enough, since the high bit it set it's unlikely to
		// be handled by another renderer anyways.
		return true
	}
	return false
}

func (r *ImageRenderer) Render(buf *demodel.CharBuffer, viewport image.Rectangle) (image.Image, error) {
	bReader := bytes.NewReader(buf.Buffer)
	img, _, err := image.Decode(bReader)
	r.cache = img
	// image.Image doesn't have SubImage(r). Most types do, but not jpeg/png.
	// so we need to allocate a new image, draw to it, and return that.
	dst := image.NewRGBA(viewport)
	draw.Draw(dst, viewport, img, viewport.Min, draw.Src)
	return dst, err
}

func (r *ImageRenderer) GetImageMap(buf *demodel.CharBuffer, viewport image.Rectangle) demodel.ImageMap {
	return &renderer.ImageMap{}
}

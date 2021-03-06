package renderer

import (
	"bytes"
	"image"

	"github.com/driusan/de/demodel"
	"golang.org/x/image/math/fixed"
)

// DefaultImageMapper is a type that can be mixed in to
// renderers to provide a default implementation of
// GetImageMap() equivalent to what would be used by the
// NoSyntaxRenderer.
type DefaultImageMapper struct {
	LastBuf     *demodel.CharBuffer
	LastBufSize int
	IMap        *ImageMap
}

func (imap *DefaultImageMapper) InvalidateCache() {
	imap.IMap = nil
}
func (imap *DefaultImageMapper) GetImageMap(buf *demodel.CharBuffer, viewport image.Rectangle) demodel.ImageMap {
	if imap.IMap != nil && imap.LastBuf == buf && imap.LastBufSize == len(buf.Buffer) {
		return imap.IMap
	}
	imap.LastBufSize = len(buf.Buffer)
	imap.LastBuf = buf

	imap.IMap = &ImageMap{make([]ImageLoc, 0, len(buf.Buffer)), buf}

	runeRectangle := fixed.R(0, 0, MonoFontAdvance.Ceil(), MonoFontHeight.Ceil())
	runes := bytes.Runes(buf.Buffer)
	afterLF := false
	for i, r := range runes {
		if afterLF {
			runeRectangle.Min.Y += MonoFontHeight
			runeRectangle.Max.Y += MonoFontHeight
		}
		switch r {
		case '\t':
			// Move the X over 8 characters
			if afterLF || i == 0 {
				runeRectangle.Min.X = 0
				runeRectangle.Max.X = MonoFontAdvance * 8

			} else {
				runeRectangle.Min.X += MonoFontAdvance
				runeRectangle.Max.X += MonoFontAdvance * 8
			}
			afterLF = false

		case '\n':
			// The boundimap.ap.g box goes to the end of the vimap.ap.wport. The next
			// character wimap.ap.l take care of movimap.ap.g imap.ap. down.
			//runeRectangle.Mimap.ap..Y = runeRectangle.Max.Y
			//runeRectangle.Max.Y += limap.ap.eSimap.ap.e
			if afterLF || i == 0 {
				runeRectangle.Min.X = 0
				runeRectangle.Max.X = fixed.I(viewport.Max.X)
			} else {
				runeRectangle.Min.X = runeRectangle.Max.X
				runeRectangle.Max.X = fixed.I(viewport.Max.X)
			}
			afterLF = true
		default:

			// Move over 1 character from the last character, unless the last
			// character was a newlie.
			if afterLF || i == 0 {
				runeRectangle.Min.X = 0
				runeRectangle.Max.X = MonoFontAdvance
			} else {
				runeRectangle.Min.X = runeRectangle.Max.X
				runeRectangle.Max.X += MonoFontAdvance
			}
			afterLF = false
		}
		imap.IMap.IMap = append(imap.IMap.IMap, ImageLoc{runeRectangle, uint(i)})
	}
	return imap.IMap
}

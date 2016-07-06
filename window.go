package main

import (
	"image"
	"image/draw"
	"sync"

	"github.com/driusan/de/demodel"
	"github.com/driusan/de/kbmap"
	"github.com/driusan/de/renderer"
	"github.com/driusan/de/viewer"
	"golang.org/x/exp/shiny/screen"
	"golang.org/x/mobile/event/size"
)

// dewindow encapsulates the shiny window of de.
type dewindow struct {
	sync.Mutex
	screen.Window

	painting bool
	buffer   screen.Buffer
	sz       size.Event
}

func (w *dewindow) paint(buf *demodel.CharBuffer, viewport *viewer.Viewport) {
	if w.painting {
		return
	}
	w.realpaint(buf, viewport)
}

// paints buf into the viewport attached to this window
func (w *dewindow) realpaint(buf *demodel.CharBuffer, viewport *viewer.Viewport) {
	w.Lock()
	defer w.Unlock()

	defer func() {
		w.painting = false
	}()
	w.painting = true

	if w.buffer == nil {
		return
	}
	dst := w.buffer.RGBA()

	// Fill the buffer with the window background colour before
	// drawing the web page on top of it.
	switch viewport.GetKeyboardMode() {
	case kbmap.InsertMode:
		draw.Draw(dst, dst.Bounds(), &image.Uniform{renderer.InsertBackground}, image.ZP, draw.Src)
	case kbmap.DeleteMode:
		draw.Draw(dst, dst.Bounds(), &image.Uniform{renderer.DeleteBackground}, image.ZP, draw.Src)
	default:
		draw.Draw(dst, dst.Bounds(), &image.Uniform{renderer.NormalBackground}, image.ZP, draw.Src)
	}

	s := w.sz.Size()

	contentBounds := dst.Bounds()
	tagBounds := tagSize
	// ensure that the tag takes no more than half the window, so that the content doesn't get
	// drowned out by commands that output more to stderr than they should.
	if wHeight := s.Y; tagBounds.Max.Y > wHeight/2 {
		tagBounds.Max.Y = wHeight / 2
	}
	contentBounds.Min.Y = tagBounds.Max.Y

	tagline.RenderInto(dst.SubImage(image.Rectangle{image.ZP, image.Point{s.X, tagBounds.Max.Y}}).(*image.RGBA), buf.Tagline, clipRectangle(w.sz, viewport))
	viewport.RenderInto(dst.SubImage(image.Rectangle{image.Point{0, tagBounds.Max.Y}, s}).(*image.RGBA), buf, clipRectangle(w.sz, viewport))

	w.Upload(image.Point{0, 0}, w.buffer, dst.Bounds())
	w.Publish()
	return
}

func (w *dewindow) setSize(s size.Event, sc screen.Screen) error {
	w.Lock()
	defer w.Unlock()

	w.sz = s
	if w.buffer != nil {
		// Release the old buffer.
		w.buffer.Release()
	}

	sbuffer, err := sc.NewBuffer(s.Size())
	w.buffer = sbuffer
	return err

}

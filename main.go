package main

import (
	"fmt"
	"github.com/driusan/de/actions"
	"github.com/driusan/de/demodel"
	"github.com/driusan/de/kbmap"
	"github.com/driusan/de/position"
	"github.com/driusan/de/renderer"
	"github.com/driusan/de/viewer"
	"golang.org/x/exp/shiny/driver"
	"golang.org/x/exp/shiny/screen"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/mouse"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
	"image"
	"image/draw"
	"os"
	"runtime/pprof"
)

const (
	ButtonLeft = iota
	ButtonMiddle
	ButtonRight
	MouseWheelUp
	MouseWheelDown
)

func clipRectangle(sz size.Event, viewport *viewer.Viewport) image.Rectangle {
	wSize := sz.Size()
	return image.Rectangle{
		Min: viewport.Location,
		Max: image.Point{viewport.Location.X + wSize.X, viewport.Location.Y + wSize.Y},
	}
}
func paintWindow(s screen.Screen, w screen.Window, sz size.Event, buf image.Image, tagimage image.Image, viewport *viewer.Viewport) {
	b, err := s.NewBuffer(sz.Size())
	defer b.Release()
	if err != nil {
		panic(err)
	}
	dst := b.RGBA()

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

	tagBounds := tagimage.Bounds()
	contentBounds := dst.Bounds()
	contentBounds.Min.Y = tagBounds.Max.Y

	draw.Draw(dst, contentBounds, buf, viewport.Location, draw.Over)
	draw.Draw(dst, tagBounds, tagimage, image.ZP, draw.Over)

	w.Upload(image.Point{0, 0}, b, dst.Bounds())
	w.Publish()
	return
}

func main() {
	f, _ := os.Create("test.profile")
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()
	var sz size.Event
	var filename string
	if len(os.Args) <= 1 {
		// no file given on the command line, so open the curent directory and give a
		// file listing that can be clicked on.
		filename = "."
	} else {
		filename = os.Args[1]
	}
	buff := demodel.CharBuffer{Filename: filename, Tagline: &demodel.CharBuffer{Buffer: make([]byte, 0)}}
	if err := actions.OpenFile(filename, &buff); err != nil {

		// An unhandled error occured
		if !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			return
		}
		// the error was just that the file doesn't exist, it'll be created on
		// save
		buff.Buffer = make([]byte, 0)
	}
	buff.ResetTagline()
	var imap renderer.ImageMap
	var MouseButtonMask [6]bool

	// hack so that things don't get confused on DirRelease when a button transitions keyboard modes
	//var lastKeyboardMode demodel.Map = kbmap.NormalMode

	viewport := &viewer.Viewport{Map: kbmap.NormalMode}

	lastKeyboardMode := viewport.GetKeyboardMode()

	render := renderer.GetRenderer(&buff)

	img, imgSize, imap, err := render.Render(&buff, clipRectangle(sz, viewport))

	if err != nil {
		panic(err)
	}
	tagline := renderer.TaglineRenderer{}

	tagimg, tagmap, _ := tagline.Render(buff.Tagline)

	var lastCharIdx uint
	var dpi float64
	var mLoc image.Point // used for determining if we should honour getting out of tag mode
	// when hitting enter from the tagline. If the mouse is still over
	// the tagline, we should stay in tagmode.
	driver.Main(func(s screen.Screen) {

		w, err := s.NewWindow(nil)
		if err != nil {
			return
		}
		defer w.Release()
		viewport.Window = w

		for {

			img, imgSize, imap, _ = render.Render(&buff, clipRectangle(sz, viewport))
			if buff.Tagline != nil {
				tagimg, tagmap, _ = tagline.Render(buff.Tagline)
			}
			paintWindow(s, w, sz, img, tagimg, viewport)

			switch e := w.NextEvent().(type) {
			case lifecycle.Event:
				if e.To == lifecycle.StageDead {
					return
				}
			case key.Event:
				if lastKeyboardMode != viewport.GetKeyboardMode() && e.Direction == key.DirRelease {
					lastKeyboardMode = viewport.GetKeyboardMode()
					// don't repeat the same key in a different direction if a keystroke changed
					// modes.
					continue
				}

				oldFilename := buff.Filename
				newKbmap, scrolldir, err := viewport.HandleKey(e, &buff, viewport)

				wSize := sz.Size()
				//	imgSize := img.Bounds().Size()
				// special error codes that a keystroke can send to control the
				// program
				switch err {
				case kbmap.ExitProgram:
					if buff.Dirty == false {
						return
					}
				case kbmap.ScrollUp:
					scrollSize := sz.Size().Y / 2
					viewport.Location.Y -= scrollSize
					if viewport.Location.Y < 0 {
						viewport.Location.Y = 0
					}
				case kbmap.ScrollDown:
					scrollSize := sz.Size().Y / 2
					viewport.Location.Y += scrollSize

					if viewport.Location.Y+wSize.Y > imgSize.Max.Y+50 {
						// we can scroll a *little* past the end, so that it's easier to read
						// the last
						viewport.Location.Y = imgSize.Max.Y - wSize.Y + 50
					}
				case kbmap.ScrollLeft:
					scrollSize := sz.Size().X / 2
					viewport.Location.X -= scrollSize
					if viewport.Location.X < 0 {
						viewport.Location.X = 0
					}
				case kbmap.ScrollRight:
					scrollSize := sz.Size().X / 2
					viewport.Location.X += scrollSize
					// we've scrolled too far down

					if viewport.Location.X+wSize.X > imgSize.Max.X+50 {
						// we can scroll a *little* past the end, so that it's easier to read
						// the longest line
						viewport.Location.X = imgSize.Max.X - wSize.X + 50
					}
				}

				// There's nothing special about the tagline size since the viewport location already
				// takes it into account, it just happens to be a good size (slightly more than 1 line)
				// to use as a buffer at the top and bottom so that we don't trigger the scroll at the
				// very last pixel
				tagEnd := tagimg.Bounds().Max.Y
				switch scrolldir {
				case demodel.DirectionUp:
					// check if has moved so that it's before the top left corner.
					if idx, err := imap.At(viewport.Location.X, viewport.Location.Y+tagEnd); err == nil && buff.Dot.Start < idx {
						if newViewport, gerr := imap.Get(buff.Dot.Start); gerr == nil {
							viewport.Location.Y = newViewport.Min.Y - 50
						}
						//fmt.Printf("Should be scrolling up (Char at viewport: %d, buff start: %d!\n", idx, buff.Dot.Start)
					} else if idx, err := imap.At(0, viewport.Location.Y+wSize.Y-tagEnd); err == nil && buff.Dot.Start > idx {
						// if't not before the top-left, so check if it's after the bottom-right
						// this might have happened if we manually scrolled the window, or used a command like <lineno>G
						if newViewport, gerr := imap.Get(buff.Dot.Start); gerr == nil {
							viewport.Location.Y = newViewport.Min.Y - 50
						}
					}
				case demodel.DirectionDown:
					// check if dot moved so that it's end is end is after the bottom right corner.
					wSize := sz.Size()
					//imgbounds := img.Bounds()
					if idx, err := imap.At(0, viewport.Location.Y+wSize.Y-tagEnd); err == nil && buff.Dot.End > idx {
						if newViewport, gerr := imap.Get(buff.Dot.End); gerr == nil {
							// scroll to about the middle of the screen
							viewport.Location.Y = newViewport.Min.Y - (wSize.Y / 2)
						}
					} else if idx, err := imap.At(0, viewport.Location.Y-tagEnd); err == nil && buff.Dot.End < idx {
						if newViewport, gerr := imap.Get(buff.Dot.End); gerr == nil {
							viewport.Location.Y = newViewport.Min.Y - (wSize.Y / 2)
						}
					}
				}

				if wSize.X >= imgSize.Max.X {
					viewport.Location.X = 0
				}
				if wSize.Y >= imgSize.Max.Y || viewport.Location.Y < 0 {
					viewport.Location.Y = 0
				}

				if mLoc.Y > tagimg.Bounds().Max.Y {
					// now apply the new map and repaint the window to incorporate
					// whatever changes the keystroke may have changed.
					// but if the mouse is over the tagline, stay in tagline mode.

					lastKeyboardMode = viewport.GetKeyboardMode()
					viewport.SetKeyboardMode(newKbmap)
				}

				if oldFilename != buff.Filename {
					render = renderer.GetRenderer(&buff)
				}
				img, imgSize, imap, _ = render.Render(&buff, clipRectangle(sz, viewport))
				if buff.Tagline != nil {
					tagimg, tagmap, _ = tagline.Render(buff.Tagline)
				}
				paintWindow(s, w, sz, img, tagimg, viewport)
			case mouse.Event:
				mLoc = image.Point{int(e.X), int(e.Y)}
				tagEnd := tagimg.Bounds().Max.Y

				// the buffer that the mouse is over. Generally either the tagline,
				// or the standard buffer
				var evtBuff *demodel.CharBuffer = &buff
				// the index into that buffer that's being pointed at
				var charIdx uint
				var err error
				if int(e.Y) < tagEnd {
					if viewport.GetKeyboardMode() != kbmap.TagMode {
						lastKeyboardMode = viewport.GetKeyboardMode()
						viewport.SetKeyboardMode(kbmap.TagMode)
					}
					evtBuff = evtBuff.Tagline
					charIdx, err = tagmap.At(int(e.X), int(e.Y))
				} else {
					// focus follows pointer
					if viewport.GetKeyboardMode() == kbmap.TagMode {
						viewport.SetKeyboardMode(kbmap.NormalMode)
					}
					// translate the mouse event by an appropriate amount, taking
					// the size of the tagline, and scrolling of the viewport into
					// consideration
					charIdx, err = imap.At(int(e.X)+viewport.Location.X, int(e.Y)+viewport.Location.Y-tagEnd)
				}

				if charIdx == lastCharIdx && e.Direction == mouse.DirNone {
					continue
				} else {
					lastCharIdx = charIdx
				}

				if err != nil {
					continue
				}

				// eDot determines which dot is being used for this event, based on the
				// mouse button. Left or right clicking uses the normal dot, middle uses
				// the alternate so that things can be executed with parameters other than
				// themselves.
				var eDot *demodel.Dot = &evtBuff.Dot
				/*
						This doesn't seem to be working well enough to use yet.
					if e.Button == mouse.ButtonMiddle {
						eDot = &evtBuff.AltDot
					}*/
				var pressed bool

				switch e.Direction {
				case mouse.DirPress:
					pressed = true
				case mouse.DirRelease:
					pressed = false
				}

				if e.Direction != mouse.DirNone {
					switch e.Button {
					case mouse.ButtonLeft:
						// this is the start of a mouse click. Reset Dot
						// to whatever was clicked on.
						if pressed && MouseButtonMask[ButtonLeft] == false {
							eDot.Start = charIdx
							eDot.End = charIdx
						}
						MouseButtonMask[ButtonLeft] = pressed
					case mouse.ButtonMiddle:
						if pressed && MouseButtonMask[ButtonMiddle] == false {
							eDot.Start = charIdx
							eDot.End = charIdx
						}
						MouseButtonMask[ButtonMiddle] = pressed
					case mouse.ButtonRight:
						if pressed && MouseButtonMask[ButtonRight] == false {
							eDot.Start = charIdx
							eDot.End = charIdx
						}
						MouseButtonMask[ButtonRight] = pressed
					case mouse.ButtonWheelUp:
						viewport.Location.Y -= 50
						if viewport.Location.Y < 0 {
							viewport.Location.Y = 0
						}
						// scrolling can't affect the content, so just rerender the window.
						img, imgSize, imap, _ = render.Render(&buff, clipRectangle(sz, viewport))
						if buff.Tagline != nil {
							tagimg, tagmap, _ = tagline.Render(buff.Tagline)
						}
						paintWindow(s, w, sz, img, tagimg, viewport)

					case mouse.ButtonWheelDown:
						viewport.Location.Y += 50
						wSize := sz.Size()
						//imgSize := img.Bounds().Size()
						if viewport.Location.Y+wSize.Y > imgSize.Max.Y+50 {
							// we can scroll a *little* past the end, so that it's easier to read
							// the last
							viewport.Location.Y = imgSize.Max.Y - wSize.Y + 50
						}
						img, imgSize, imap, _ = render.Render(&buff, clipRectangle(sz, viewport))
						if buff.Tagline != nil {
							tagimg, tagmap, _ = tagline.Render(buff.Tagline)
						}

						paintWindow(s, w, sz, img, tagimg, viewport)
					}
				}

				// nothing is pressed, so don't rerender. There's no possibility of something having changed in the view.
				if e.Direction == mouse.DirNone &&
					MouseButtonMask[ButtonLeft] == false &&
					MouseButtonMask[ButtonRight] == false &&
					MouseButtonMask[ButtonMiddle] == false {
					continue
				}
				if MouseButtonMask[ButtonLeft] == true || MouseButtonMask[ButtonRight] == true || MouseButtonMask[ButtonMiddle] == true {
					// if it's outside the current selection, expand the selection.
					if charIdx < eDot.Start {
						eDot.Start = charIdx
					}
					if charIdx > eDot.End {
						eDot.End = charIdx
					}

					// if it's inside the current selection, shrink the selection.
					if charIdx < eDot.End && charIdx > eDot.Start {
						if eDot.End-charIdx < eDot.Start-charIdx {
							// it's slower to the end
							eDot.End = charIdx
						} else {
							// it's slower to the start
							eDot.Start = charIdx
						}
					}

					// the highlighted portion of the image may have changed, so
					// rerender everything.
					img, imgSize, imap, _ = render.Render(&buff, clipRectangle(sz, viewport))
					if buff.Tagline != nil {
						tagimg, tagmap, _ = tagline.Render(buff.Tagline)
					}
					paintWindow(s, w, sz, img, tagimg, viewport)
				}
				if e.Direction == mouse.DirRelease && e.Button == mouse.ButtonRight {
					oldFilename := buff.Filename
					if evtBuff == buff.Tagline {
						if eDot.Start == eDot.End {
							actions.FindNextOrOpenTag(position.CurTagWordStart, position.CurTagWordEnd, &buff)
						} else {
							actions.FindNextOrOpenTag(position.TagDotStart, position.TagDotEnd, &buff)
						}
					} else {
						if eDot.Start == eDot.End {
							actions.FindNextOrOpen(position.CurWordStart, position.CurWordEnd, evtBuff)

						} else {
							actions.FindNextOrOpen(position.DotStart, position.DotEnd, evtBuff)
						}
					}
					if oldFilename != buff.Filename {
						// make sure the syntax highlighting gets updated if it needs to be.
						render = renderer.GetRenderer(&buff)
					}

					img, imgSize, imap, _ = render.Render(&buff, clipRectangle(sz, viewport))
					if buff.Tagline != nil {
						tagimg, tagmap, _ = tagline.Render(buff.Tagline)
					}
					paintWindow(s, w, sz, img, tagimg, viewport)
				}
				if e.Direction == mouse.DirRelease && e.Button == mouse.ButtonMiddle {
					if evtBuff == buff.Tagline {
						// executing from the tagline is a little special, because it uses the word
						// from a tagline buffer to perform an action in a non-tagline buffer.
						if eDot.Start == eDot.End {
							actions.PerformTagAction(position.CurTagExecutionWordStart, position.CurTagExecutionWordEnd, &buff, viewport)
						} else {
							actions.PerformTagAction(position.TagDotStart, position.TagDotEnd, &buff, viewport)
						}
					} else {
						if eDot.Start == eDot.End {
							// otherwise, just perform the action normally.
							actions.PerformAction(position.CurExecutionWordStart, position.CurExecutionWordEnd, evtBuff, viewport)

						} else {
							actions.PerformAction(position.DotStart, position.DotEnd, evtBuff, viewport)
						}
					}
					img, imgSize, imap, _ = render.Render(&buff, clipRectangle(sz, viewport))
					if buff.Tagline != nil {
						tagimg, tagmap, _ = tagline.Render(buff.Tagline)
					}
					paintWindow(s, w, sz, img, tagimg, viewport)
				}
				//paintWindow(s, w, sz, buff)
			case paint.Event:
				paintWindow(s, w, sz, img, tagimg, viewport)
			case size.Event:
				sz = e
				wSize := e.Size()
				tagline.Width = wSize.X
				if dpi == 0 {
					dpi = float64(sz.PixelsPerPt) * 72
					renderer.RecalculateFontFace(dpi)
					render.InvalidateCache()
				}
				img, imgSize, imap, _ = render.Render(&buff, clipRectangle(sz, viewport))
				imgSize := img.Bounds().Size()
				if wSize.X >= imgSize.X {
					viewport.Location.X = 0
				}
				if wSize.Y >= imgSize.Y {
					viewport.Location.Y = 0
				}
				if buff.Tagline != nil {
					tagimg, tagmap, _ = tagline.Render(buff.Tagline)
				}
				paintWindow(s, w, sz, img, tagimg, viewport)

			}
		}
	})

}

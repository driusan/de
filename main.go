package main

import (
	"fmt"
	"github.com/driusan/de/actions"
	"github.com/driusan/de/demodel"
	"github.com/driusan/de/kbmap"
	"github.com/driusan/de/position"
	"github.com/driusan/de/renderer"
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
)

var viewport struct {
	Location     image.Point
	KeyboardMode kbmap.Map
}

const (
	ButtonLeft = iota
	ButtonMiddle
	ButtonRight
	MouseWheelUp
	MouseWheelDown
)

func paintWindow(s screen.Screen, w screen.Window, sz size.Event, buf image.Image, tagimage image.Image) {
	b, err := s.NewBuffer(sz.Size())
	defer b.Release()
	if err != nil {
		panic(err)
	}
	dst := b.RGBA()

	// Fill the buffer with the window background colour before
	// drawing the web page on top of it.
	switch viewport.KeyboardMode {
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
	var sz size.Event
	var filename string
	if len(os.Args) <= 1 {
		// no file given on the command line, so open the curent directory and give a
		// file listing that can be clicked on.
		filename = "."
	} else {
		filename = os.Args[1]
	}
	buff := demodel.CharBuffer{Filename: filename, Tagline: &demodel.CharBuffer{}}
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

	var imap renderer.ImageMap
	var MouseButtonMask [6]bool

	// hack so that things don't get confused on DirRelease when a button transitions keyboard modes
	var lastKeyboardMode kbmap.Map = kbmap.NormalMode
	viewport.KeyboardMode = kbmap.NormalMode

	render := renderer.GetRenderer(buff)
	img, imap, err := render.Render(&buff)

	if err != nil {
		panic(err)
	}
	tagline := renderer.TaglineRenderer{}

	tagimg, tagmap, _ := tagline.Render(buff.Tagline)

	var lastCharIdx uint
	driver.Main(func(s screen.Screen) {
		w, err := s.NewWindow(nil)
		if err != nil {
			return
		}
		defer w.Release()

		for {
			switch e := w.NextEvent().(type) {
			case lifecycle.Event:
				if e.To == lifecycle.StageDead {
					return
				}
			case key.Event:
				if lastKeyboardMode != viewport.KeyboardMode && e.Direction == key.DirRelease {
					lastKeyboardMode = viewport.KeyboardMode
					// don't repeat the same key in a different direction if a keystroke changed
					// modes.
					continue
				}

				oldFilename := buff.Filename
				newKbmap, scrolldir, err := viewport.KeyboardMode.HandleKey(e, &buff)

				wSize := sz.Size()
				imgSize := img.Bounds().Size()
				// special error codes that a keystroke can send to control the
				// program
				switch err {
				case kbmap.ExitProgram:
					return
				case kbmap.ScrollUp:
					scrollSize := sz.Size().Y / 2
					viewport.Location.Y -= scrollSize
					if viewport.Location.Y < 0 {
						viewport.Location.Y = 0
					}
				case kbmap.ScrollDown:
					scrollSize := sz.Size().Y / 2
					viewport.Location.Y += scrollSize
					if viewport.Location.Y+wSize.Y > imgSize.Y+50 {
						// we can scroll a *little* past the end, so that it's easier to read
						// the last
						viewport.Location.Y = imgSize.Y - wSize.Y + 50
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
					if viewport.Location.X+wSize.X > imgSize.X+50 {
						// we can scroll a *little* past the end, so that it's easier to read
						// the longest line
						viewport.Location.X = imgSize.X - wSize.X + 50
					}
				}

				// There's nothing special about the tagline size since the viewport location already
				// takes it into account, it just happens to be a good size (slightly more than 1 line)
				// to use as a buffer at the top and bottom so that we don't trigger the scroll at the
				// very last pixel
				tagEnd := tagimg.Bounds().Max.Y
				switch scrolldir {
				case kbmap.DirectionUp:
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
				case kbmap.DirectionDown:
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

				case kbmap.DirectionNone:
				}

				if wSize.X >= imgSize.X {
					viewport.Location.X = 0
				}
				if wSize.Y >= imgSize.Y || viewport.Location.Y < 0 {
					viewport.Location.Y = 0
				}

				// now apply the new map and repaint the window to incorporate
				// whatever changes the keystroke may have changed.
				lastKeyboardMode = viewport.KeyboardMode
				viewport.KeyboardMode = newKbmap

				if oldFilename != buff.Filename {
					render = renderer.GetRenderer(buff)
				}
				img, imap, _ = render.Render(&buff)
				if buff.Tagline != nil {
					tagimg, tagmap, _ = tagline.Render(buff.Tagline)
				}
				paintWindow(s, w, sz, img, tagimg)
			case mouse.Event:
				tagEnd := tagimg.Bounds().Max.Y

				// the buffer that the mouse is over. Generally either the tagline,
				// or the standard buffer
				var evtBuff *demodel.CharBuffer = &buff
				// the index into that buffer that's being pointed at
				var charIdx uint
				var err error
				if int(e.Y) < tagEnd {
					if viewport.KeyboardMode != kbmap.TagMode {
						lastKeyboardMode = viewport.KeyboardMode
						viewport.KeyboardMode = kbmap.TagMode
					}
					evtBuff = evtBuff.Tagline
					charIdx, err = tagmap.At(int(e.X), int(e.Y))
				} else {
					// focus follows pointer
					if viewport.KeyboardMode == kbmap.TagMode {
						viewport.KeyboardMode = kbmap.NormalMode
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
							evtBuff.Dot.Start = charIdx
							evtBuff.Dot.End = charIdx
						}
						MouseButtonMask[ButtonLeft] = pressed
					case mouse.ButtonMiddle:
						if pressed && MouseButtonMask[ButtonMiddle] == false {
							evtBuff.Dot.Start = charIdx
							evtBuff.Dot.End = charIdx
						}
						MouseButtonMask[ButtonMiddle] = pressed
					case mouse.ButtonRight:
						if pressed && MouseButtonMask[ButtonRight] == false {
							evtBuff.Dot.Start = charIdx
							evtBuff.Dot.End = charIdx
						}
						MouseButtonMask[ButtonRight] = pressed
					case mouse.ButtonWheelUp:
						viewport.Location.Y -= 50
						if viewport.Location.Y < 0 {
							viewport.Location.Y = 0
						}
						//img, imap, _ = render.Render(buff)
						// scrolling can't affect the content, so just rerender the window.
						paintWindow(s, w, sz, img, tagimg)

					case mouse.ButtonWheelDown:
						viewport.Location.Y += 50
						wSize := sz.Size()
						imgSize := img.Bounds().Size()
						if viewport.Location.Y+wSize.Y > imgSize.Y+50 {
							// we can scroll a *little* past the end, so that it's easier to read
							// the last
							viewport.Location.Y = imgSize.Y - wSize.Y + 50
						}
						/*
							img, imap, _ = render.Render(buff)
							if buff.Tagline != nil {
								tagimg, tagmap, _ = tagline.Render(*buff.Tagline)
							}
						*/
						paintWindow(s, w, sz, img, tagimg)
					}
				}

				if MouseButtonMask[ButtonLeft] == true || MouseButtonMask[ButtonRight] == true || MouseButtonMask[ButtonMiddle] == true {
					// if it's outside the current selection, expand the selection.
					if charIdx < evtBuff.Dot.Start {
						evtBuff.Dot.Start = charIdx
					}
					if charIdx > evtBuff.Dot.End {
						evtBuff.Dot.End = charIdx
					}

					// if it's inside the current selection, shrink the selection.
					if charIdx < evtBuff.Dot.End && charIdx > evtBuff.Dot.Start {
						if evtBuff.Dot.End-charIdx < evtBuff.Dot.Start-charIdx {
							// it's slower to the end
							evtBuff.Dot.End = charIdx
						} else {
							// it's slower to the start
							evtBuff.Dot.Start = charIdx
						}
					}

					// the highlighted portion of the image may have changed, so
					// rerender everything.
					img, imap, _ = render.Render(&buff)
					if buff.Tagline != nil {
						tagimg, tagmap, _ = tagline.Render(buff.Tagline)
					}
					paintWindow(s, w, sz, img, tagimg)
				}
				if e.Direction == mouse.DirRelease && e.Button == mouse.ButtonRight {
					oldFilename := buff.Filename
					if evtBuff == buff.Tagline {
						if evtBuff.Dot.Start == evtBuff.Dot.End {
							actions.FindNextOrOpenTag(position.CurTagWordStart, position.CurTagWordEnd, &buff)
						} else {
							actions.FindNextOrOpenTag(position.TagDotStart, position.TagDotEnd, &buff)
						}
					} else {

						if evtBuff.Dot.Start == evtBuff.Dot.End {
							actions.FindNextOrOpen(position.CurWordStart, position.CurWordEnd, evtBuff)

						} else {
							actions.FindNextOrOpen(position.DotStart, position.DotEnd, evtBuff)
						}
					}
					if oldFilename != buff.Filename {
						// make sure the syntax highlighting gets updated if it needs to be.
						render = renderer.GetRenderer(buff)
					}

					img, imap, _ = render.Render(&buff)
					if buff.Tagline != nil {
						tagimg, tagmap, _ = tagline.Render(buff.Tagline)
					}
					paintWindow(s, w, sz, img, tagimg)
				}
				if e.Direction == mouse.DirRelease && e.Button == mouse.ButtonMiddle {
					if evtBuff == buff.Tagline {
						// executing from the tagline is a little special, because it uses the word
						// from a tagline buffer to perform an action in a non-tagline buffer.
						if evtBuff.Dot.Start == evtBuff.Dot.End {
							actions.PerformTagAction(position.CurTagWordStart, position.CurTagWordEnd, &buff)
						} else {
							actions.PerformTagAction(position.TagDotStart, position.TagDotEnd, &buff)
						}
					} else {
						if evtBuff.Dot.Start == evtBuff.Dot.End {

							// otherwise, just perform the action normally.
							actions.PerformAction(position.CurWordStart, position.CurWordEnd, evtBuff)

						} else {
							actions.PerformAction(position.DotStart, position.DotEnd, evtBuff)
						}
					}
					img, imap, _ = render.Render(&buff)
					if buff.Tagline != nil {
						tagimg, tagmap, _ = tagline.Render(buff.Tagline)
					}
					paintWindow(s, w, sz, img, tagimg)
				}
				//paintWindow(s, w, sz, buff)
			case paint.Event:
				paintWindow(s, w, sz, img, tagimg)
			case size.Event:
				sz = e
				wSize := e.Size()
				tagline.Width = wSize.X
				img, imap, _ = render.Render(&buff)
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
				paintWindow(s, w, sz, img, tagimg)
			}
		}
	})
}

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
	"io/ioutil"
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

func paintWindow(s screen.Screen, w screen.Window, sz size.Event, buf image.Image) {
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

	draw.Draw(dst, dst.Bounds(), buf, viewport.Location, draw.Over)

	w.Upload(image.Point{0, 0}, b, dst.Bounds())
	w.Publish()
	return
}

func main() {
	var sz size.Event
	if len(os.Args) <= 1 {
		fmt.Fprintf(os.Stderr, "Need filename to open.\n")
		return
	}
	filename := os.Args[1]
	b, err := ioutil.ReadFile(filename)
	buff := demodel.CharBuffer{}
	if err != nil {
		// An unhandled error occured
		if !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			return
		}
		// the error was just that the file doesn't exist, it'll be created on
		// save
		buff.Buffer = make([]byte, 0)
	} else {
		buff.Buffer = b
	}

	buff.Filename = filename

	var imap renderer.ImageMap
	var MouseButtonMask [6]bool

	// hack so that things don't get confused on DirRelease when a button transitions keyboard modes
	var lastKeyboardMode kbmap.Map = kbmap.NormalMode
	viewport.KeyboardMode = kbmap.NormalMode

	render := renderer.GetRenderer(buff)
	img, imap, err := render.Render(buff)

	if err != nil {
		panic(err)
	}
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

				newKbmap, err := viewport.KeyboardMode.HandleKey(e, &buff)

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

				if wSize.X >= imgSize.X {
					viewport.Location.X = 0
				}
				if wSize.Y >= imgSize.Y {
					viewport.Location.Y = 0
				}

				// TODO: Autoscroll if the cursor has moved past the end of the window.

				// now apply the new map and repaint the window to incorporate
				// whatever changes the keystroke may have changed.
				lastKeyboardMode = viewport.KeyboardMode
				viewport.KeyboardMode = newKbmap

				img, imap, _ = render.Render(buff)
				paintWindow(s, w, sz, img)
			case mouse.Event:
				charIdx, err := imap.At(int(e.X)+viewport.Location.X, int(e.Y)+viewport.Location.Y)
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
							buff.Dot.Start = charIdx
							buff.Dot.End = charIdx
						}
						MouseButtonMask[ButtonLeft] = pressed
					case mouse.ButtonMiddle:
						if pressed && MouseButtonMask[ButtonMiddle] == false {
							buff.Dot.Start = charIdx
							buff.Dot.End = charIdx
						}
						MouseButtonMask[ButtonMiddle] = pressed
					case mouse.ButtonRight:
						if pressed && MouseButtonMask[ButtonRight] == false {
							buff.Dot.Start = charIdx
							buff.Dot.End = charIdx
						}
						MouseButtonMask[ButtonRight] = pressed
					case mouse.ButtonWheelUp:
						viewport.Location.Y -= 50
						if viewport.Location.Y < 0 {
							viewport.Location.Y = 0
						}
						img, imap, _ = render.Render(buff)
						paintWindow(s, w, sz, img)
					case mouse.ButtonWheelDown:
						viewport.Location.Y += 50
						wSize := sz.Size()
						imgSize := img.Bounds().Size()
						if viewport.Location.Y+wSize.Y > imgSize.Y+50 {
							// we can scroll a *little* past the end, so that it's easier to read
							// the last
							viewport.Location.Y = imgSize.Y - wSize.Y + 50
						}
						img, imap, _ = render.Render(buff)
						paintWindow(s, w, sz, img)
					}
				}

				if MouseButtonMask[ButtonLeft] == true || MouseButtonMask[ButtonRight] == true || MouseButtonMask[ButtonMiddle] == true {
					// if it's outside the current selection, expand the selection.
					if charIdx < buff.Dot.Start {
						buff.Dot.Start = charIdx
					}
					if charIdx > buff.Dot.End {
						buff.Dot.End = charIdx
					}

					// if it's inside the current selection, shrink the selection.
					if charIdx < buff.Dot.End && charIdx > buff.Dot.Start {
						if buff.Dot.End-charIdx < buff.Dot.Start-charIdx {
							// it's slower to the end
							buff.Dot.End = charIdx
						} else {
							// it's slower to the start
							buff.Dot.Start = charIdx
						}
					}
					img, imap, _ = render.Render(buff)
					paintWindow(s, w, sz, img)
				}
				if e.Direction == mouse.DirRelease && e.Button == mouse.ButtonRight {
					if buff.Dot.Start == buff.Dot.End {
						actions.FindNextOrOpen(position.CurWordStart, position.CurWordEnd, &buff)

					} else {
						actions.FindNextOrOpen(position.DotStart, position.DotEnd, &buff)
					}
					img, imap, _ = render.Render(buff)
					paintWindow(s, w, sz, img)
				}
				if e.Direction == mouse.DirRelease && e.Button == mouse.ButtonMiddle {
					if buff.Dot.Start == buff.Dot.End {
						actions.PerformAction(position.CurWordStart, position.CurWordEnd, &buff)
					} else {
						actions.PerformAction(position.DotStart, position.DotEnd, &buff)
					}
				}
				//paintWindow(s, w, sz, buff)
			case paint.Event:
				paintWindow(s, w, sz, img)
			case size.Event:
				sz = e
				wSize := e.Size()
				img, imap, _ = render.Render(buff)
				imgSize := img.Bounds().Size()
				if wSize.X >= imgSize.X {
					viewport.Location.X = 0
				}
				if wSize.Y >= imgSize.Y {
					viewport.Location.Y = 0
				}
				paintWindow(s, w, sz, img)
			}
		}
	})
}

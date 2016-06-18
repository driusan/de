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
	"io/ioutil"
	"os"
	"os/user"
	"strings"
	//"runtime/pprof"
)

const (
	ButtonLeft = iota
	ButtonMiddle
	ButtonRight
)

func clipRectangle(sz size.Event, viewport *viewer.Viewport) image.Rectangle {
	wSize := sz.Size()
	return image.Rectangle{
		Min: viewport.Location,
		Max: image.Point{viewport.Location.X + wSize.X, viewport.Location.Y + wSize.Y},
	}
}

var tagline demodel.Renderer
var tagSize image.Rectangle

func paintWindow(b screen.Buffer, w screen.Window, sz size.Event, buf *demodel.CharBuffer, viewport *viewer.Viewport) {
	if b == nil {
		return
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

	s := sz.Size()

	contentBounds := dst.Bounds()
	tagBounds := tagSize
	// ensure that the tag takes no more than half the window, so that the content doesn't get
	// drowned out by commands that output more to stderr than they should.
	if wHeight := sz.Size().Y; tagBounds.Max.Y > wHeight/2 {
		tagBounds.Max.Y = wHeight / 2
	}
	contentBounds.Min.Y = tagBounds.Max.Y

	//	draw.Draw(dst, contentBounds, buf, viewport.Location, draw.Over)
	//	draw.Draw(dst, tagBounds, tagimage, image.ZP, draw.Over)
	tagline.RenderInto(dst.SubImage(image.Rectangle{image.ZP, image.Point{s.X, tagBounds.Max.Y}}).(*image.RGBA), buf.Tagline, clipRectangle(sz, viewport))
	viewport.RenderInto(dst.SubImage(image.Rectangle{image.Point{0, tagBounds.Max.Y}, s}).(*image.RGBA), buf, clipRectangle(sz, viewport))

	w.Upload(image.Point{0, 0}, b, dst.Bounds())
	w.Publish()
	return
}
func getKbScrollSizeY(e key.Event, wSize image.Point) int {
	switch e.Modifiers {
	case (key.ModControl | key.ModAlt):
		return 1
	case key.ModControl, key.ModAlt:
		return renderer.MonoFontFace.Metrics().Height.Ceil()
	default:
		return wSize.Y / 2
	}

}
func getKbScrollSizeX(e key.Event, wSize image.Point) int {
	switch e.Modifiers {
	case (key.ModControl | key.ModAlt):
		return 1
	case key.ModControl, key.ModAlt:
		return renderer.MonoFontFace.Metrics().Height.Ceil()
	default:
		return wSize.X / 2
	}

}

func runStartupCommands(b *demodel.CharBuffer, v demodel.Viewport) {
	u, err := user.Current()
	if err != nil {
		return
	}
	file := u.HomeDir + "/.de/startup"

	data, err := ioutil.ReadFile(file)
	if err != nil {
		return
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		actions.RunOrExec(strings.TrimSpace(line), b, v)
	}
}

func main() {
	/*
		f, _ := os.Create("test.profile")
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	*/
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
	if err := actions.OpenFile(filename, &buff, nil); err != nil {

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

	buff.LoadSnarfBuffer()
	defer buff.SaveSnarfBuffer()
	var MouseButtonMask [6]bool

	viewport := &viewer.Viewport{
		Map:      kbmap.NormalMode,
		Renderer: renderer.GetRenderer(&buff),
	}
	viewport.SetOption("TermWidth", 80)

	// the renderer wasn't set yet when OpenFile was called, so do it now
	actions.FocusViewport(buff.Dot.Start, &buff, viewport)

	// hack so that things don't get confused on DirRelease when a button transitions keyboard modes
	lastKeyboardMode := viewport.GetKeyboardMode()

	imap := viewport.GetImageMap(&buff, clipRectangle(sz, viewport))
	imgSize := viewport.Bounds(&buff)

	// Don't use type inference because we want to make sure that we keep implementing
	// the renderer interface for taglines for consistency, but also don't want to register
	// it so that no one accidentally uses it.
	tagline = &renderer.TaglineRenderer{}
	tagSize = tagline.Bounds(buff.Tagline)

	//tagimg, _ := tagline.Render(buff.Tagline, clipRectangle(sz, viewport))
	tagmap := tagline.GetImageMap(buff.Tagline, clipRectangle(sz, viewport))

	var lastCharIdx uint
	var dpi float64
	// when hitting enter from the tagline. If the mouse is still over
	// the tagline, we should stay in tagmode.
	var screenBuffer screen.Buffer

	runStartupCommands(&buff, viewport)
	driver.Main(func(s screen.Screen) {

		w, err := s.NewWindow(nil)
		if err != nil {
			return
		}
		defer w.Release()
		viewport.Window = w

		for {
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
					if !buff.Dirty {
						return
					}
				case kbmap.ScrollUp:
					scrollSize := getKbScrollSizeY(e, wSize)

					viewport.Location.Y -= scrollSize
					if viewport.Location.Y < 0 {
						viewport.Location.Y = 0
					}
				case kbmap.ScrollDown:
					scrollSize := getKbScrollSizeY(e, wSize)

					viewport.Location.Y += scrollSize
					if viewport.Location.Y+wSize.Y > imgSize.Max.Y+50 {
						// we can scroll a *little* past the end, so that it's easier to read
						// the last
						viewport.Location.Y = imgSize.Max.Y - wSize.Y + 50
					}
				case kbmap.ScrollLeft:
					scrollSize := getKbScrollSizeX(e, wSize)
					viewport.Location.X -= scrollSize
					if viewport.Location.X < 0 {
						viewport.Location.X = 0
					}
				case kbmap.ScrollRight:
					scrollSize := getKbScrollSizeX(e, wSize)
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
				tagEnd := tagSize.Max.Y
				switch scrolldir {
				case demodel.DirectionUp:
					// check if has moved so that it's before the top left corner.
					if imap == nil {
						continue
					}
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
					if imap == nil {
						continue
					}
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

				lastKeyboardMode = viewport.GetKeyboardMode()
				viewport.SetKeyboardMode(newKbmap)

				if oldFilename != buff.Filename {
					viewport.SetRenderer(
						renderer.GetRenderer(&buff),
					)
				}
				//img, _ = viewport.Render(&buff, clipRectangle(sz, viewport))
				imap = viewport.GetImageMap(&buff, clipRectangle(sz, viewport))
				imgSize = viewport.Bounds(&buff)
				if buff.Tagline != nil {
					tagmap = tagline.GetImageMap(buff.Tagline, clipRectangle(sz, viewport))
					tagSize = tagline.Bounds(buff.Tagline)
				}
				paintWindow(screenBuffer, w, sz, &buff, viewport)
			case mouse.Event:
				// Some drivers (notable Mac and Windows)
				// seem to only send scroll events as a DirNone event,
				// which trips up by the code that tries to ensure we don't
				// do anything if nothing changed. As a result, this hack
				// fakes a DirPress event by just modifying the event struct
				// before we do anything.
				if e.Direction == mouse.DirNone &&
					(e.Button == mouse.ButtonWheelUp ||
						e.Button == mouse.ButtonWheelDown) {
					e.Direction = mouse.DirPress
				}

				// No button changed state, and nothing is pressed.
				// Skip doing anything in this mouse event, because
				// there's nothing that could have changed which needs
				// a rerender.

				if e.Direction == mouse.DirNone &&
					!MouseButtonMask[ButtonLeft] &&
					!MouseButtonMask[ButtonMiddle] &&
					!MouseButtonMask[ButtonRight] {
					continue
				}
				tagEnd := tagSize.Max.Y

				// the buffer that the mouse is over. Generally either the tagline,
				// or the standard buffer
				evtBuff := &buff
				// the index into that buffer that's being pointed at
				var charIdx uint
				var err error
				if int(e.Y) < tagEnd {
					evtBuff = evtBuff.Tagline
					charIdx, err = tagmap.At(int(e.X), int(e.Y))
				} else {
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
					if int(e.Y) < tagEnd {
						// clicked in the tag bar
						if viewport.GetKeyboardMode() != kbmap.TagMode {
							lastKeyboardMode = viewport.GetKeyboardMode()
							viewport.SetKeyboardMode(kbmap.TagMode)
						}
					} else {
						// clicked outside of the tagbar
						if viewport.GetKeyboardMode() == kbmap.TagMode {
							lastKeyboardMode = kbmap.TagMode
							viewport.SetKeyboardMode(kbmap.NormalMode)
						}
					}
				case mouse.DirRelease:
					pressed = false
				}

				if e.Direction != mouse.DirNone {
					switch e.Button {
					case mouse.ButtonLeft:
						// this is the start of a mouse click. Reset Dot
						// to whatever was clicked on.
						if pressed && !MouseButtonMask[ButtonLeft] {
							eDot.Start = charIdx
							eDot.End = charIdx
						}
						MouseButtonMask[ButtonLeft] = pressed
					case mouse.ButtonMiddle:
						if pressed && !MouseButtonMask[ButtonMiddle] {
							eDot.Start = charIdx
							eDot.End = charIdx
						}
						MouseButtonMask[ButtonMiddle] = pressed
					case mouse.ButtonRight:
						if pressed && !MouseButtonMask[ButtonRight] {
							eDot.Start = charIdx
							eDot.End = charIdx
						}
						MouseButtonMask[ButtonRight] = pressed
					case mouse.ButtonWheelUp:
						viewport.Location.Y -= 50
						if viewport.Location.Y < 0 {
							viewport.Location.Y = 0
						}
						paintWindow(screenBuffer, w, sz, &buff, viewport)

					case mouse.ButtonWheelDown:
						viewport.Location.Y += 50
						wSize := sz.Size()

						if viewport.Location.Y+wSize.Y > imgSize.Max.Y+50 {
							// we can scroll a *little* past the end, so that it's easier to read
							// the last
							viewport.Location.Y = imgSize.Max.Y - wSize.Y + 50
						}
						//img, _ = viewport.Render(&buff, clipRectangle(sz, viewport))
						imgSize = viewport.Bounds(&buff)
						paintWindow(screenBuffer, w, sz, &buff, viewport)
					}
				}

				// nothing is pressed, so don't rerender. There's no possibility of something having changed in the view.
				if e.Direction == mouse.DirNone &&
					!MouseButtonMask[ButtonLeft] &&
					!MouseButtonMask[ButtonRight] &&
					!MouseButtonMask[ButtonMiddle] {
					continue
				}
				if MouseButtonMask[ButtonLeft] || MouseButtonMask[ButtonRight] || MouseButtonMask[ButtonMiddle] {
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
					// reviewport everything.
					//img, _ = viewport.Render(&buff, clipRectangle(sz, viewport))
					imgSize = viewport.Bounds(&buff)
					paintWindow(screenBuffer, w, sz, &buff, viewport)
				}
				if e.Direction == mouse.DirRelease && e.Button == mouse.ButtonRight {
					oldFilename := buff.Filename
					if evtBuff == buff.Tagline {
						if eDot.Start == eDot.End {
							actions.FindNextOrOpenTag(position.CurTagWordStart, position.CurTagWordEnd, &buff, viewport)
						} else {
							actions.FindNextOrOpenTag(position.TagDotStart, position.TagDotEnd, &buff, viewport)
						}
					} else {
						if eDot.Start == eDot.End {
							actions.FindNextOrOpen(position.CurWordStart, position.CurWordEnd, evtBuff, viewport)

						} else {
							actions.FindNextOrOpen(position.DotStart, position.DotEnd, evtBuff, viewport)
						}
					}
					if oldFilename != buff.Filename {
						// make sure the syntax highlighting gets updated if it needs to be.
						viewport.SetRenderer(
							renderer.GetRenderer(&buff),
						)
					}

					//img, _ = viewport.Render(&buff, clipRectangle(sz, viewport))
					imap = viewport.GetImageMap(&buff, clipRectangle(sz, viewport))
					imgSize = viewport.Bounds(&buff)
					if buff.Tagline != nil {
						tagSize = tagline.Bounds(buff.Tagline)
						tagmap = tagline.GetImageMap(buff.Tagline, clipRectangle(sz, viewport))
					}
					paintWindow(screenBuffer, w, sz, &buff, viewport)
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
					//img, _ = viewport.Render(&buff, clipRectangle(sz, viewport))
					imap = viewport.GetImageMap(&buff, clipRectangle(sz, viewport))
					imgSize = viewport.Bounds(&buff)
					if buff.Tagline != nil {
						//tagimg, _ = tagline.Render(buff.Tagline, clipRectangle(sz, viewport))
						tagSize = tagline.Bounds(buff.Tagline)
						tagmap = tagline.GetImageMap(buff.Tagline, clipRectangle(sz, viewport))
					}
					paintWindow(screenBuffer, w, sz, &buff, viewport)
				}
				//paintWindow(s, w, sz, buff)
			case paint.Event:
				paintWindow(screenBuffer, w, sz, &buff, viewport)
			case size.Event:
				sz = e
				wSize := e.Size()
				//tagline.Width = wSize.X
				if dpi == 0 {
					dpi = float64(sz.PixelsPerPt) * 72
					renderer.RecalculateFontFace(dpi)
					viewport.InvalidateCache()
					tagline.InvalidateCache()

				}

				if wSize.X >= imgSize.Max.X {
					viewport.Location.X = 0
				}
				if wSize.Y >= imgSize.Max.Y {
					viewport.Location.Y = 0
				}
				imgSize = viewport.Bounds(&buff)
				tagSize = tagline.Bounds(buff.Tagline)
				/*
					img, _ = viewport.Render(&buff, clipRectangle(sz, viewport))
					imap = viewport.GetImageMap(&buff, clipRectangle(sz, viewport))
					imgSize = viewport.Bounds(&buff)
					if buff.Tagline != nil {
						tagimg, _ = tagline.Render(buff.Tagline, clipRectangle(sz, viewport))
						tagmap = tagline.GetImageMap(buff.Tagline, clipRectangle(sz, viewport))
					}
				*/
				if screenBuffer != nil {
					// Release the old buffer.
					screenBuffer.Release()
				}
				screenBuffer, err = s.NewBuffer(wSize)
				if err != nil {
					fmt.Fprintf(os.Stderr, "%v\n", err)
					continue
				}
				paintWindow(screenBuffer, w, sz, &buff, viewport)
			case viewer.RequestRerender:
				//fmt.Printf("Requesting rerender\n")
				//img, _ = viewport.Render(&buff, clipRectangle(sz, viewport))
				imap = viewport.GetImageMap(&buff, clipRectangle(sz, viewport))
				imgSize = viewport.Bounds(&buff)
				if buff.Tagline != nil {
					tagSize = tagline.Bounds(buff.Tagline)
					//tagimg, _ = tagline.Render(buff.Tagline, clipRectangle(sz, viewport))
					tagmap = tagline.GetImageMap(buff.Tagline, clipRectangle(sz, viewport))
				}
				wSize := sz.Size()

				if viewport.Location.Y+wSize.Y > imgSize.Max.Y+50 {
					// we can scroll a *little* past the end, so that it's easier to read
					// the last
					viewport.Location.Y = imgSize.Max.Y - wSize.Y + 50
				}
				if wSize.X >= imgSize.Max.X {
					viewport.Location.X = 0
				}
				if wSize.Y >= imgSize.Max.Y {
					viewport.Location.Y = 0
				}

				paintWindow(screenBuffer, w, sz, &buff, viewport)

			}
		}
	})
}

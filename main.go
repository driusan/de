package main

import (
	"flag"
	"fmt"
	"image"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

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

func getKbScrollSizeY(e key.Event, wSize image.Point) int {
	switch e.Modifiers {
	case (key.ModControl | key.ModAlt):
		return 1
	case key.ModControl, key.ModAlt:
		return renderer.MonoFontHeight.Ceil()
	default:
		return wSize.Y / 2
	}

}
func getKbScrollSizeX(e key.Event, wSize image.Point) int {
	switch e.Modifiers {
	case (key.ModControl | key.ModAlt):
		return 1
	case key.ModControl, key.ModAlt:
		return renderer.MonoFontHeight.Ceil()
	default:
		return wSize.X / 2
	}

}

func runStartupCommands(b *demodel.CharBuffer, v demodel.Viewport) {
	data, err := ioutil.ReadFile(demodel.ConfigHome() + "/startup")
	if err != nil {
		return
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		actions.RunOrExec(strings.TrimSpace(line), b, v)
	}
}

func main() {
	delete := flag.Bool("delete", false, "Delete the file passed as a parameter after opening it")
	fname := flag.String("filename", "", "Use filename as the buffer's filename after opening, instead of basing it on the file's real filename")

	dirtyChan := make(chan bool)

	plumber := &plumbService{}
	plumber.Connect(dirtyChan)

	var sz size.Event
	var filename string

	_, cmd := filepath.Split(os.Args[0])

	var pager bool
	switch cmd {
	case "dep", "less", "more":
		pager = true
	}

	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		// no file given on the command line, so open the curent directory and give a
		// file listing that can be clicked on.
		//
		// If in pager mode, read from stdin instead. Pagers are usually
		// passed a filename in /tmp, but as long as we're here we might
		// as well give people an easy way to invoke de so that it reads
		// from stdin
		if pager {
			filename = "-"
		} else {
			filename = "."
		}
	} else {
		filename = args[0]
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

	// -delete was passed, so unless it's stdin or no
	// parameters, delete it.
	if *delete && filename != "-" && len(args) >= 1 {
		os.Remove(filename)
	}
	// We were told to pretend the buffer has a different
	// filename, so do so.
	if *fname != "" {
		buff.Filename = *fname
	}

	buff.ResetTagline()

	buff.LoadSnarfBuffer()
	defer buff.SaveSnarfBuffer()

	var MouseButtonMask [6]bool

	viewport := &viewer.Viewport{
		Map: kbmap.NormalMode,
	}

	// If invoked as a pager, assume the xterm emulating renderer which
	// understands (at least some) ANSI escape codes.
	if pager {
		viewport.Renderer = renderer.GetNamedRenderer("xterm")
	} else {
		viewport.Renderer = renderer.GetRenderer(&buff)
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

	runStartupCommands(&buff, viewport)

	// cache the last dirty status, so that we only send a message when
	// it changes, instead of every single event
	lastDirty := buff.Dirty

	var w screen.Window

	// Monitor the plumber service, and send it the buffer status
	// as soon as it's available.
	go func() {
		// If it's not ready yet, wait for either an error or it to
		// become available.
		for !plumber.Available() {
			// Select on the different channels that the plumbService
			// may have sent on, and convert them to.
			select {
			case err := <-plumber.ErrorChan:
				// If there was an error connecting before
				// it became available, print it and abort
				// the goroutine.
				buff.AppendTag("\n" + err.Error())
				return
			default:
				// Wait 200 milliseconds before trying again,
				// there's no reason to be too greedy.
				time.Sleep(200 * time.Millisecond)
			}
		}

		// Buffer must be available if we got here, so make sure the buffer
		// dirty status is up to date.
		dirtyChan <- buff.Dirty
		lastDirty = buff.Dirty

		actions.PlumbingReady = true

		// All that's left to do is continually wait for messages
		// on either OpenChan or ErrorChan and act on them.
		for {
			select {
			case err := <-plumber.ErrorChan:
				// Print any errors that come in to the tagline.
				buff.AppendTag("\n" + err.Error())
			case filename := <-plumber.OpenChan:
				// Open the requested file if we got something on openChan.
				if err := actions.OpenFile(filename, &buff, viewport); err != nil {
					buff.AppendTag("\n" + err.Error())
					continue
				}

				// Reset the viewport to 0,0, get a new
				// renderer for the correct content type,
				// and request a new render of the
				// current window.
				viewport.Location.X = 0
				viewport.Location.Y = 0

				viewport.SetRenderer(
					renderer.GetRenderer(&buff),
				)
				if w != nil {
					w.Send(viewer.RequestRerender{})
				}

			}
		}
	}()

	// Main shiny event loop
	driver.Main(func(s screen.Screen) {

		win, err := s.NewWindow(&screen.NewWindowOptions{Title: "de"})
		if err != nil {
			return
		}
		// We need to store the window as w, so that it's available
		// for the plumbService to use.
		w = win
		defer w.Release()
		window := dewindow{
			Window: w,
		}
		viewport.Window = w

		for {
			if plumber.Available() {
				// Send the dirty status if it changes.
				if buff.Dirty != lastDirty {
					dirtyChan <- buff.Dirty
					lastDirty = buff.Dirty
				}
			}

			// Handle the next actual event
			switch e := w.NextEvent().(type) {
			case lifecycle.Event:
				switch e.To {
				case lifecycle.StageFocused:
					if plumber.Available() {
						dirtyChan <- buff.Dirty
						lastDirty = buff.Dirty
					}
				case lifecycle.StageDead:
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
				case kbmap.ErrExitProgram:
					if !buff.Dirty {
						return
					}
				case kbmap.ErrScrollUp:
					scrollSize := getKbScrollSizeY(e, wSize)

					viewport.Location.Y -= scrollSize
					if viewport.Location.Y < 0 {
						viewport.Location.Y = 0
					}
				case kbmap.ErrScrollDown:
					scrollSize := getKbScrollSizeY(e, wSize)

					viewport.Location.Y += scrollSize
					if viewport.Location.Y+wSize.Y > imgSize.Max.Y+50 {
						// we can scroll a *little* past the end, so that it's easier to read
						// the last
						viewport.Location.Y = imgSize.Max.Y - wSize.Y + 50
					}
				case kbmap.ErrScrollLeft:
					scrollSize := getKbScrollSizeX(e, wSize)
					viewport.Location.X -= scrollSize
					if viewport.Location.X < 0 {
						viewport.Location.X = 0
					}
				case kbmap.ErrScrollRight:
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
				go window.paint(&buff, viewport)
			case mouse.Event:
				switch e.Direction {
				case mouse.DirStep:
					// Scroll wheel events don't really share
					// much in common with other mouse wheel events,
					// so just handle them and repaint
					if viewport.HandleMouseWheel(e, &buff, clipRectangle(sz, viewport).Size()) {
						go window.paint(&buff, viewport)
					}
					viewport.NotifyMouseListeners(e)
					continue
				case mouse.DirNone:
					// If no button changed state, and nothing is pressed,
					// then skip doing anything in this mouse event, because
					// there's nothing that could have changed which needs
					// a rerender.
					if !MouseButtonMask[ButtonLeft] &&
						!MouseButtonMask[ButtonMiddle] &&
						!MouseButtonMask[ButtonRight] {
						viewport.NotifyMouseListeners(e)
						continue
					}

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
					viewport.NotifyMouseListeners(e)
					continue
				} else {
					lastCharIdx = charIdx
				}

				if err != nil {
					viewport.NotifyMouseListeners(e)
					continue
				}

				// eDot determines which dot is being used for this event, based on the
				// mouse button. Left or right clicking uses the normal dot, middle uses
				// the alternate so that things can be executed with parameters other than
				// themselves.
				var eDot = &evtBuff.Dot
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
					}

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
					go window.paint(&buff, viewport)
				}
				if e.Direction == mouse.DirRelease && e.Button == mouse.ButtonRight {
					oldFilename := buff.Filename
					if evtBuff == buff.Tagline {
						if plumber.Available() {
							if eDot.Start == eDot.End {
								actions.TagPlumbOrFindNext(position.CurTagWordStart, position.CurTagWordEnd, &buff, viewport)
							} else {
								actions.TagPlumbOrFindNext(position.TagDotStart, position.TagDotEnd, &buff, viewport)
							}
						} else {
							if eDot.Start == eDot.End {
								actions.FindNextOrOpenTag(position.CurTagWordStart, position.CurTagWordEnd, &buff, viewport)
							} else {
								actions.FindNextOrOpenTag(position.TagDotStart, position.TagDotEnd, &buff, viewport)
							}
						}
					} else {
						if plumber.Available() {
							actions.PlumbOrFindNext(position.DotStart, position.DotEnd, evtBuff, viewport)
						} else {
							if eDot.Start == eDot.End {
								actions.FindNextOrOpen(position.CurWordStart, position.CurWordEnd, evtBuff, viewport)

							} else {
								actions.FindNextOrOpen(position.DotStart, position.DotEnd, evtBuff, viewport)
							}
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
					go window.paint(&buff, viewport)
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
					go window.paint(&buff, viewport)
				}
				viewport.NotifyMouseListeners(e)
			case paint.Event:
				go window.paint(&buff, viewport)
			case size.Event:
				sz = e
				wSize := e.Size()
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
				imap = viewport.GetImageMap(&buff, clipRectangle(sz, viewport))
				tagmap = tagline.GetImageMap(buff.Tagline, clipRectangle(sz, viewport))

				if err := window.setSize(sz, s); err != nil {
					fmt.Fprintf(os.Stderr, "Error: allocating buffer: %v\n", err)
					continue
				}
				go window.paint(&buff, viewport)
			case viewer.RequestRerender:
				imap = viewport.GetImageMap(&buff, clipRectangle(sz, viewport))
				imgSize = viewport.Bounds(&buff)
				if buff.Tagline != nil {
					tagSize = tagline.Bounds(buff.Tagline)
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
				go window.paint(&buff, viewport)

			}
		}

	})
}

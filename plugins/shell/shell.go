package shell

import (
	//"bytes"
	"log"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	//"sync"
	"syscall"

	"github.com/pkg/term/termios"

	"github.com/driusan/de/actions"
	"github.com/driusan/de/demodel"
	"github.com/driusan/de/kbmap"
	"golang.org/x/mobile/event/key"
)

type shellKbmap struct {
	stdin io.Writer
}

// Keymap that sends everything to the Shell command, except Escape (quit the shell and return to
// normal mode), arrow keys (scroll the viewport)
func (s shellKbmap) HandleKey(e key.Event, buff *demodel.CharBuffer, v demodel.Viewport) (demodel.Map, demodel.ScrollDirection, error) {
	switch e.Code {
	/*
		case key.CodeEscape:
			// TODO: Quit the shell as the documentation claims happens.
			if e.Direction != key.DirPress {
				//return kbmap.NormalMode, demodel.DirectionNone, nil
			}
			return s, demodel.DirectionDown, nil
	*/
	// Still honour the viewport manipulation keys
	case key.CodeRightArrow:
		// Arrow keys indicate their scroll direction via the error return value,
		// they return demodel.DirectionNone to make sure both code paths don't accidentally
		// get triggered
		if e.Direction == key.DirPress {
			return s, demodel.DirectionNone, kbmap.ErrScrollRight
		}
		return s, demodel.DirectionNone, nil
	case key.CodeLeftArrow:
		if e.Direction == key.DirPress {
			return s, demodel.DirectionNone, kbmap.ErrScrollLeft
		}
		return s, demodel.DirectionNone, nil
	case key.CodeDownArrow:
		if e.Direction == key.DirPress {
			return s, demodel.DirectionNone, kbmap.ErrScrollDown
		}
		return s, demodel.DirectionNone, nil
	case key.CodeUpArrow:
		if e.Direction == key.DirPress {
			return s, demodel.DirectionNone, kbmap.ErrScrollUp
		}
		return s, demodel.DirectionNone, nil
	// Special cases for control characters.

	case key.CodeTab:
		if e.Direction != key.DirRelease {
			//buff.Buffer = append(buff.Buffer, '\t')
			fmt.Fprintf(s.stdin, "%c", '\t')
			buff.Dot.End = uint(len(buff.Buffer)) - 1
			buff.Dot.Start = buff.Dot.End
		}
		return s, demodel.DirectionDown, nil
		//	fmt.Printf("Pressed key %s. Rune is %x", e, e.Rune
	case key.CodeDeleteBackspace:
		if e.Direction != key.DirRelease {
			//buff.Buffer = buff.Buffer[:len(buff.Buffer)-1] //append(buff.Buffer, "\t")
			fmt.Fprintf(s.stdin, "%c", '\b')
			buff.Dot.End = uint(len(buff.Buffer)) - 1
			buff.Dot.Start = buff.Dot.End
		}
		return s, demodel.DirectionDown, nil
	case key.CodeReturnEnter:
		if e.Direction != key.DirRelease {
			//buff.Buffer = append(buff.Buffer, '\n')
			fmt.Fprintf(s.stdin, "%c", '\n')
			buff.Dot.End = uint(len(buff.Buffer)) - 1
			buff.Dot.Start = buff.Dot.End
			return s, demodel.DirectionDown, nil
		}
		return s, demodel.DirectionDown, nil
	default:
		if e.Direction != key.DirRelease && e.Rune > 0 {
			// Hacks for some control codes where it's
			// not set properly in e.Rune.
			if e.Modifiers&key.ModControl != 0 {
				switch e.Rune {
				case 'd':
					e.Rune = 4
				case 'z':
					e.Rune = 26
				}
			}
			fmt.Fprintf(s.stdin, "%c", e.Rune)
			buff.Dot.End = uint(len(buff.Buffer)) - 1
			buff.Dot.Start = buff.Dot.End

		}
	}
	return s, demodel.DirectionDown, nil
}
func init() {
	actions.RegisterAction("Shell", Shell)
}

// Shell invokes an interactive shell terminal similarly to "win" in ACME
func Shell(args string, buff *demodel.CharBuffer, viewport demodel.Viewport) {
	//cmd.Start()
	//scanner := bufio.NewScanner(cmdReader)
	//buff.KeyboardMode = &shellKbmap{}
	shell := os.Getenv("SHELL")
	if shell == "" {
		switch runtime.GOOS {
		case "plan9":
			shell = "rc"
		default:
			shell = "sh"
		}
	}

	c := exec.Command(shell, "-i")

	buff.Filename = ""

	go func() {
		masterTTY, slaveTTY, err := termios.Pty()
		// _, slaveTTY, err := termios.Pty()
		kbMap := &shellKbmap{masterTTY}
		viewport.LockKeyboardMode(kbMap)
		if err != nil {
			log.Print(err)
			return
		}
		procs := syscall.SysProcAttr{
			Setsid: true,
		}
		c.SysProcAttr = &procs
		c.Stdout = slaveTTY
		c.Stderr = slaveTTY
		c.Stdin = slaveTTY
		c.Start()

		// buffer to read lines into. Allocate this out of the loop to go
		// easier on the GC.
		termline := make([]byte, 1024)
		viewport.SetRenderer(&TerminalRenderer{})

		var quit bool
		go func() {
			c.Wait()
			quit = true
		}()
		for !quit {
			if buff.Filename != "" {
				// The user must have clicked on a filename and opened it.
				// Stop the Shell.
				break
			}

			n, _ := masterTTY.Read(termline)

			if n > 0 {
				buff.Buffer = append(buff.Buffer, termline[:n]...)
				if l := len(buff.Buffer); l > 65536 {
					// be relatively conservative in how large the buffer
					// can get, so that the rendering doesn't slow everything
					// down.
					buff.Buffer = buff.Buffer[l-65536:]
				}
				buff.Dot.End = uint(len(buff.Buffer)) - 1
				buff.Dot.Start = buff.Dot.End
				//fmt.Printf("Requesting rerender\n")
				//viewport.Rerender()
			}
		}
		viewport.UnlockKeyboardMode(kbMap)
		viewport.SetKeyboardMode(kbmap.NormalMode)
		viewport.Rerender()

	}()
}

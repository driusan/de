package shell

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"sync"
	//"time"

	"github.com/kr/pty"

	"github.com/driusan/de/actions"
	"github.com/driusan/de/demodel"
	"github.com/driusan/de/kbmap"
	"golang.org/x/mobile/event/key"
)

// Create a thread safe wrapper around bytes.Buffer, so that our go routine can read
// from it despite the fact that the shell might be writing to it.
type TSBuffer struct {
	b bytes.Buffer
	m sync.Mutex
}

func (b *TSBuffer) Read(p []byte) (n int, err error) {
	b.m.Lock()
	defer b.m.Unlock()
	return b.b.Read(p)
}
func (b *TSBuffer) Write(p []byte) (n int, err error) {
	b.m.Lock()
	defer b.m.Unlock()
	return b.b.Write(p)
}
func (b *TSBuffer) String() string {
	b.m.Lock()
	defer b.m.Unlock()
	return b.b.String()
}

type shellKbmap struct {
	//runeChan chan rune
	stdin io.Writer
}

// Keymap that sends everything to the Shell command, except Escape (quit the shell and return to
// normal mode), arrow keys (scroll the viewport)
func (s shellKbmap) HandleKey(e key.Event, buff *demodel.CharBuffer, v demodel.Viewport) (demodel.Map, demodel.ScrollDirection, error) {
	switch e.Code {
	case key.CodeEscape:
		// TODO: Quit the shell as the documentation claims happens.
		if e.Direction != key.DirPress {
			//return kbmap.NormalMode, demodel.DirectionNone, nil
		}
		return s, demodel.DirectionDown, nil
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
		if e.Direction != key.DirPress {
			buff.Buffer = append(buff.Buffer, '\t')
			fmt.Fprintf(s.stdin, "%c", '\t')
			buff.Dot.End = uint(len(buff.Buffer)) - 1
			buff.Dot.Start = buff.Dot.End
		}
		return s, demodel.DirectionDown, nil
		//	fmt.Printf("Pressed key %s. Rune is %x", e, e.Rune
	case key.CodeDeleteBackspace:
		if e.Direction != key.DirPress {
			buff.Buffer = buff.Buffer[:len(buff.Buffer)-1] //append(buff.Buffer, "\t")
			fmt.Fprintf(s.stdin, "%c", '\b')
			buff.Dot.End = uint(len(buff.Buffer)) - 1
			buff.Dot.Start = buff.Dot.End
		}
		return s, demodel.DirectionDown, nil
	case key.CodeReturnEnter:
		if e.Direction != key.DirPress {
			buff.Buffer = append(buff.Buffer, '\n')
			fmt.Fprintf(s.stdin, "%c", '\n')
			buff.Dot.End = uint(len(buff.Buffer)) - 1
			buff.Dot.Start = buff.Dot.End
			return s, demodel.DirectionDown, nil
		}
		return s, demodel.DirectionDown, nil
	default:
		println("Pressed", e.Rune)
		if e.Direction != key.DirPress && e.Rune > 0 {
			// send the rune to the buffer and to the shell
			//rbytes := make([]byte, 4)
			//n := utf8.EncodeRune(rbytes, e.Rune)
			//	fmt.Printf("Sent to stdin: %c %d", e.Rune, e.Rune)
			// bash and zsh echo the character typed back when invoked with $SHELL -i
			// and it's not a tty.
			// dash doesn't.
			// Don't append the rune to the buffer, because odds are high it'll
			// get echoed back, though there's no way to know for sure.
			//buff.Buffer = append(buff.Buffer, rbytes[:n]...)

			fmt.Fprintf(s.stdin, "%c", e.Rune)
			buff.Dot.End = uint(len(buff.Buffer)) - 1
			buff.Dot.Start = buff.Dot.End

		} else {
			/*
					for debugging only. This otherwise triggers errors on things like
					the user pressing a control key.

				if e.Rune <= 0 {
					fmt.Printf("Invalid rune %d from %s\n", e.Rune, e)
				}*/
		}
	}
	return s, demodel.DirectionDown, nil
}
func init() {
	actions.RegisterAction("Shell", Shell)
}

// Shell invokes an interactive shell terminal similarly to "win" in ACME
func Shell(args string, buff *demodel.CharBuffer, viewport demodel.Viewport) {
	go func() {
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
		master, slave, err := pty.Open()
		if err != nil {
			// FIXME: add better error handling.
			panic(err)
		}
		kbMap := &shellKbmap{slave}
		viewport.LockKeyboardMode(kbMap)

		buff.Filename = ""

		c.Stdout = master
		c.Stderr = master
		c.Stdin = master
		c.Start()
		viewport.SetRenderer(&TerminalRenderer{})
		for {
			if buff.Filename != "" {
				// The user must have clicked on a filename and opened it.
				// Stop the Shell.
				master.Close()
				slave.Close()
				break
			}

			io.Copy(buff, slave)
		}
		c.Wait()
		fmt.Fprintf(os.Stderr, "Shell exited\n")
		viewport.UnlockKeyboardMode(kbMap)
		viewport.SetKeyboardMode(kbmap.NormalMode)
	}()
}

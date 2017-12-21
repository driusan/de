// +build !windows,!plan9

package shell

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"syscall"
	//"time"

	"github.com/kr/pty"

	"github.com/driusan/de/actions"
	"github.com/driusan/de/demodel"
	"github.com/driusan/de/kbmap"
	"golang.org/x/mobile/event/key"
)

type shellKbmap struct {
	process *os.Process
	stdin   io.WriteCloser
}

// Keymap that sends everything to the Shell command, except Escape (quit the shell and return to
// normal mode), arrow keys (scroll the viewport)
func (s shellKbmap) HandleKey(e key.Event, buff *demodel.CharBuffer, v demodel.Viewport) (demodel.Map, demodel.ScrollDirection, error) {
	switch e.Code {
	case key.CodeEscape:
		if e.Direction == key.DirPress {
			s.process.Kill()
			return s, demodel.DirectionNone, nil
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
		if e.Modifiers&key.ModControl != 0 {
			if e.Direction != key.DirRelease {
				fmt.Fprintf(s.stdin, "%c[B", '\033')

				println("send up arrow")
			}
			return s, demodel.DirectionNone, nil
		}

		if e.Direction == key.DirPress {
			return s, demodel.DirectionNone, kbmap.ErrScrollDown
		}
		return s, demodel.DirectionNone, nil
	case key.CodeUpArrow:
		if e.Modifiers&key.ModControl != 0 {
			if e.Direction != key.DirRelease {
				fmt.Fprintf(s.stdin, "%c[A", '\033')

				println("send up arrow")
			}
			return s, demodel.DirectionNone, nil
		}
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
			fmt.Fprintf(s.stdin, "%c", '\b')
			buff.Dot.End = uint(len(buff.Buffer)) - 1
			buff.Dot.Start = buff.Dot.End
		}
		return s, demodel.DirectionDown, nil
	case key.CodeReturnEnter:
		if e.Direction != key.DirRelease {
			buff.Buffer = append(buff.Buffer, '\n')
			fmt.Fprintf(s.stdin, "%c", '\n')
			buff.Dot.End = uint(len(buff.Buffer)) - 1
			buff.Dot.Start = buff.Dot.End
			return s, demodel.DirectionDown, nil
		}
		return s, demodel.DirectionDown, nil
	default:
		if e.Direction != key.DirRelease && e.Rune > 0 {
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

		c := exec.Command(shell)
		master, slave, err := pty.Open()
		if err != nil {
			// FIXME: add better error handling.
			panic(err)
		}
		c.SysProcAttr = &syscall.SysProcAttr{
			Setsid:  true,
			Setctty: true,
			Ctty:    int(master.Fd()),
		}

		buff.Filename = ""

		c.Stdout = master
		c.Stderr = master
		c.Stdin = master
		c.Start()
		kbMap := &shellKbmap{c.Process, slave}
		viewport.LockKeyboardMode(kbMap)
		defer func() {
			println("Unlocking keyboard")
			viewport.UnlockKeyboardMode(kbMap)
			viewport.SetKeyboardMode(kbmap.NormalMode)
			println("waiting")
			c.Wait()
			println("shell exited")
		}()

		viewport.SetRenderer(&TerminalRenderer{})
		buf := make([]byte, 1024)

		mouseChan := make(chan interface{})
		viewport.RegisterMouseListener(mouseChan)
		go func() {
			for {
				select {
				case <-mouseChan:
					if buff.Filename != "" {
						viewport.DeregisterMouseListener(mouseChan)
						// The user must have clicked on a filename and
						// opened it. Stop the Shell.
						c.Process.Kill()
					}
				}
			}
		}()
		for {
			if buff.Filename != "" {
				master.Close()
				slave.Close()
				break
			}

			n, err := slave.Read(buf)
			if n > 0 {
				buff.Buffer = append(buff.Buffer, buf[:n]...)

			}
			if err != nil {
				master.Close()
				slave.Close()
				break
			}

			buff.Dot.End = uint(len(buff.Buffer)) - 1
			buff.Dot.Start = buff.Dot.End
			viewport.Rerender()
			buf = make([]byte, 1024)

		}
	}()
}

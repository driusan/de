package actions

import (
	"bytes"
	"fmt"
	"github.com/driusan/de/demodel"
	"github.com/driusan/de/position"
	"io/ioutil"
	"os"
	"os/exec"
)

func PerformAction(From, To demodel.Position, buff *demodel.CharBuffer) {
	if buff == nil {
		return
	}

	dot := demodel.Dot{}
	i, err := From(*buff)
	if err != nil {
		return
	}

	dot.Start = i
	i, err = To(*buff)
	if err != nil {
		return
	}
	dot.End = i

	cmd := string(buff.Buffer[dot.Start : dot.End+1])
	runOrExec(cmd, buff)
}

func PerformTagAction(From, To demodel.Position, buff *demodel.CharBuffer) {
	if buff == nil || buff.Tagline == nil {
		return
	}
	dot := demodel.Dot{}
	// From and To should be the tag variant of Position functions, so run
	// it directly on buff.
	i, err := From(*buff)
	if err != nil {
		return
	}

	dot.Start = i
	i, err = To(*buff)
	if err != nil {
		return
	}
	dot.End = i

	if l := uint(len(buff.Tagline.Buffer)); dot.Start >= l || dot.End+1 >= l {
		// one last check for safety
		return
	}
	cmd := string(buff.Tagline.Buffer[dot.Start : dot.End+1])

	// now that the command has been extracted from the tagline, perform the command
	// on the real buffer.
	runOrExec(cmd, buff)
}

func OpenOrPerformAction(From, To demodel.Position, buff *demodel.CharBuffer) {
	if buff == nil {
		return
	}

	dot := demodel.Dot{}
	i, err := From(*buff)
	if err != nil {
		return
	}

	dot.Start = i
	i, err = To(*buff)
	if err != nil {
		return
	}
	dot.End = i

	var cmd string
	if dot.End+1 >= uint(len(buff.Buffer)) {
		cmd = string(buff.Buffer[dot.Start:])
	} else {
		cmd = string(buff.Buffer[dot.Start : dot.End+1])
	}

	if err := OpenFile(cmd, buff); err == nil {
		return
	}

	runOrExec(cmd, buff)
}

type replaceMode uint8

const (
	appendDot = replaceMode(iota)
	replaceBuffer
	replaceDot
)

func runOrExec(cmd string, buff *demodel.CharBuffer) {

	if f, ok := actions[cmd]; ok {
		// it was an internal command, so run it.
		f("", buff)
		return
	}
	if len(cmd) <= 0 || buff == nil {
		return
	}
	var mode replaceMode
	switch cmd[0] {
	case '<':
		mode = replaceBuffer
		cmd = cmd[1:]
	case '|':
		mode = replaceDot
		cmd = cmd[1:]
	default:
		mode = appendDot
	}
	// it wasn't an internal command, so run it through a shell.
	gocmd := exec.Command("sh", "-c", cmd)
	stdout, err := gocmd.StdoutPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return
	}
	stdin, err := gocmd.StdinPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return
	}
	stderr, err := gocmd.StderrPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return
	}
	if err := gocmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return
	}

	// send the selection to the processes's stdin, or the file if nothing
	// is selected.
	if buff.Dot.Start != buff.Dot.End {
		fmt.Fprintf(stdin, "%s", buff.Buffer[buff.Dot.Start:buff.Dot.End])
	} else {
		fmt.Fprintf(stdin, "%s", buff.Buffer)
	}
	// if we don't close Stdin, the shell won't exit.
	stdin.Close()
	output, err := ioutil.ReadAll(stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return
	}

	// print stderr to the tagline if anything was printed to stderr. This needs to be
	// done before waiting for it finish. See godoc for StderrPipe in go docs for os/exec
	erroutput, err := ioutil.ReadAll(stderr)
	if err == nil && len(erroutput) > 0 && buff.Tagline != nil {

		tagBuf := bytes.NewBuffer(buff.Tagline.Buffer)
		fmt.Fprintf(tagBuf, "\n%s", erroutput)
		buff.Tagline.Buffer = tagBuf.Bytes()
	}
	fmt.Fprintf(os.Stderr, "Error: %s\n", erroutput)
	exiterr := gocmd.Wait()

	if exiterr != nil {
		// Something went wrong, so log it and return without modifying the real
		// buffer.
		fmt.Fprintf(os.Stderr, "%s\n", exiterr)
		return
	}

	if len(output) <= 0 {
		// there was no output, so we don't need to do anything
		return
	}

	switch mode {
	case replaceBuffer:
		buff.Dot.Start = 0
		buff.Dot.End = 0
		buff.Buffer = output
	case replaceDot:

		newBuffer := make([]byte, len(buff.Buffer)-int(buff.Dot.End-buff.Dot.Start)+len(output))
		copy(newBuffer, buff.Buffer)
		copy(newBuffer[buff.Dot.Start:], output)
		copy(newBuffer[buff.Dot.Start+uint(len(output)):], buff.Buffer[buff.Dot.End:])
		buff.Dot.End = buff.Dot.Start + uint(len(output))

		buff.Buffer = newBuffer
	default:
		if buff.Dot.Start == buff.Dot.End {
			// if it was executed by pressing enter, move dot.End to the end of the word, so that
			// the output doesn't get inserted in the middle of the command.
			if newEnd, err := position.CurWordEnd(*buff); err == nil {
				buff.Dot.End = newEnd
			}
		}
		// otherwise, insert the output after the end of the currently selected text, and adjust the
		// selection to be the selected text.
		newBuffer := make([]byte, len(buff.Buffer)+len(output))
		copy(newBuffer, buff.Buffer)
		copy(newBuffer[buff.Dot.End+1:], output)
		copy(newBuffer[int(buff.Dot.End+1)+len(output):], buff.Buffer[buff.Dot.End:])

		// adjust the selected text. It starts at the end of the old selected text, and goes
		// for len(output)
		buff.Dot.Start = buff.Dot.End + 1
		buff.Dot.End = buff.Dot.Start + uint(len(output))

		buff.Buffer = newBuffer
	}
}

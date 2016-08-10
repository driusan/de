package actions

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/driusan/de/demodel"
	"github.com/driusan/de/position"
)

// Do does a built in action that's been registered.
func Do(cmd string, buff *demodel.CharBuffer, v demodel.Viewport) {
	if f, ok := actions[cmd]; ok {
		// it was an internal command, so run it.
		f("", buff, v)
	}
}
func PerformAction(From, To demodel.Position, buff *demodel.CharBuffer, v demodel.Viewport) {
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
	RunOrExec(cmd, buff, v)
}

func PerformTagAction(From, To demodel.Position, buff *demodel.CharBuffer, v demodel.Viewport) {
	if buff == nil || buff.Tagline == nil {
		return
	}
	dot := demodel.Dot{}
	// using := doesn't seem to be working with delve, so use the manual
	// initialization..
	var i, j uint
	var err error
	// From and To should be the tag variant of Position functions, so run
	// it directly on buff.
	i, err = From(*buff)
	if err != nil {
		return
	}
	dot.Start = i

	j, err = To(*buff)
	if err != nil {
		return
	}
	dot.End = j

	/*
		if l := uint(len(buff.Tagline.Buffer)); dot.Start >= l || dot.End+1 >= l {
			// one last check for safety
			return
		}
	*/
	cmd := string(buff.Tagline.Buffer[dot.Start : dot.End+1])

	// now that the command has been extracted from the tagline, perform the command
	// on the real buffer.
	RunOrExec(cmd, buff, v)
}

func OpenOrPerformAction(From, To demodel.Position, buff *demodel.CharBuffer, v demodel.Viewport) {
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

	if err := OpenFile(cmd, buff, v); err == nil {
		return
	}

	RunOrExec(cmd, buff, v)
}
func OpenOrPerformTagAction(From, To demodel.Position, buff *demodel.CharBuffer, v demodel.Viewport) {
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
		cmd = string(buff.Tagline.Buffer[dot.Start:])
	} else {
		cmd = string(buff.Tagline.Buffer[dot.Start : dot.End+1])
	}

	if err := OpenFile(cmd, buff, v); err == nil {
		return
	}

	RunOrExec(cmd, buff, v)
}

type replaceMode uint8

const (
	appendDot = replaceMode(iota)
	replaceBuffer
	replaceDot
)

func RunOrExec(cmd string, buff *demodel.CharBuffer, v demodel.Viewport) {
	// replace aliases before doing anything else
	for name, value := range aliases {
		if strings.HasPrefix(cmd, name) {
			cmd = strings.Replace(cmd, name, value, 1)
		}
	}

	var actionCmd, actionArgs string = cmd, ""

	// separate the argument from the value
	if i := strings.Index(cmd, ":"); i >= 0 {
		actionCmd = cmd[:i]
		actionArgs = cmd[i+1:]
	}

	if f, ok := actions[actionCmd]; ok {
		// replace : with spaces in internal command arguments unless
		// they're escaped, so that it's easier to click. This needs
		// to be done here, since we escape before getting to the code
		// otherwise.
		newArgs := make([]byte, len(actionArgs))
		for i, r := range []byte(actionArgs) {
			if r == ':' && i > 0 && actionArgs[i-1] != '\\' {
				newArgs[i] = ' '
			} else {
				newArgs[i] = r
			}
		}

		// it was an internal command, so run it.
		f(string(newArgs), buff, v)
		return
	}
	if len(cmd) <= 0 || buff == nil {
		return
	}
	var ignoreReturnCode bool
	if cmd[0] == '!' {
		ignoreReturnCode = true
		cmd = cmd[1:]
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

	// replace : with spaces unless they're escaped, so that it's easier
	// to click
	newCmd := make([]byte, len(cmd))
	for i, r := range []byte(cmd) {
		if r == ':' && i > 0 && cmd[i-1] != '\\' {
			newCmd[i] = ' '
		} else {
			newCmd[i] = r
		}
	}

	// it wasn't an internal command, so run it through a shell.
	shellCmd, shellArgs := getShellCmd()
	gocmd := exec.Command(shellCmd, shellArgs, string(newCmd))
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
	if len(erroutput) > 0 {
		fmt.Fprintf(os.Stderr, "Error: %s\n", erroutput)
	}
	exiterr := gocmd.Wait()

	if exiterr != nil && !ignoreReturnCode {
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
		if buff.Dot.Start == buff.Dot.End {
			// replacing dot means replacing the whole file, because nothing was selected
			buff.Dot.Start = 0
			buff.Dot.End = 0
			buff.Buffer = output
		} else {
			newBuffer := make([]byte, len(buff.Buffer)-int(buff.Dot.End-buff.Dot.Start)+len(output))
			copy(newBuffer, buff.Buffer)
			copy(newBuffer[buff.Dot.Start:], output)
			copy(newBuffer[buff.Dot.Start+uint(len(output)):], buff.Buffer[buff.Dot.End:])
			buff.Dot.End = buff.Dot.Start + uint(len(output))

			buff.Buffer = newBuffer
		}
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
	buff.Dirty = true
}

var aliases map[string]string

// Alias registers an alias for execution commands. It requires two parameters:
// a name, and a value. When a command is executed by either pressing enter or
// middle clicking, if the name is a prefix of the command, it's replaced by
// the value before getting executed. This allows you to make replacements such
// as Alias:s/:|sed:s/ in order to make the command s/foo/bar/ replaced by
// |sed:s/foo/bar/ and pipe the current selection through sed.
// There are no restrictions on names or values, other than that the name can
// not contain a space, since it's used as a delimiter when registering the
// alias.
func Alias(args string, buff *demodel.CharBuffer, v demodel.Viewport) {
	idx := strings.Index(args, " ")
	if idx < 1 || idx+1 >= len(args) {
		buff.AppendTag("\nAlias requires both a name and a value.")
		return
	}
	name := args[0:idx]
	val := args[idx+1:]

	if aliases == nil {
		aliases = make(map[string]string)
	}
	aliases[name] = val
}

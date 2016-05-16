package actions

import (
	"fmt"
	"github.com/driusan/de/demodel"
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

	if _, err := os.Stat(cmd); err == nil {
		// the file exists, so open it
		b, ferr := ioutil.ReadFile(cmd)
		if ferr != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			return
		}
		buff.Buffer = b
		buff.Filename = cmd
		buff.Dot.Start = 0
		buff.Dot.End = 0
		return
	}
	runOrExec(cmd, buff)
}

func runOrExec(cmd string, buff *demodel.CharBuffer) {

	if f, ok := actions[cmd]; ok {
		// it was an internal command, so run it.
		f("", buff)
		return
	}

	// it wasn't an internal command, so run the shell command
	gocmd := exec.Command(cmd)
	stdout, err := gocmd.StdoutPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return
	}
	if err := gocmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return
	}
	output, err := ioutil.ReadAll(stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return
	}
	if err := gocmd.Wait(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
	}
	// nothing selected, so insert at dot.Start
	newBuffer := make([]byte, len(buff.Buffer)-int((buff.Dot.End-buff.Dot.Start))+len(output))
	copy(newBuffer, buff.Buffer)
	copy(newBuffer[buff.Dot.Start:], output)
	copy(newBuffer[int(buff.Dot.Start)+len(output):], buff.Buffer[buff.Dot.End:])

	buff.Buffer = newBuffer

	fmt.Printf("Output: %s\n", string(output))
}

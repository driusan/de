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

func runOrExec(cmd string, buff *demodel.CharBuffer) {

	if f, ok := actions[cmd]; ok {
		// it was an internal command, so run it.
		f("", buff)
		return
	}

	// it wasn't an internal command, so run the shell command
	gocmd := exec.Command("sh", "-c", cmd)
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

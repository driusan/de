// The deplumber is a service which coordinates the running de processes
// to ensure that they don't act on the same plumber message. When it reads
// a plumbing message from the edit port, it will open it in the active de
// window if the buffer is clean, and open a new de window if it's not.
//
// Messages that don't come from de always spawn a new window.
package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"time"

	"9fans.net/go/plan9"
	"9fans.net/go/plumb"
)

const ORCLOSE = 64

// We keep track of the most recent de process that we got a message from,
// and make the assumption that that's the window that should be used for
// plumbing messages. Since plumbing messages are generated from the user
// doing something, we can assume that the user did something in that window
// in order to generate the message.
var activeProcess struct {
	Conn  *os.File
	PID   uint64
	Dirty bool
}

// Deletes the ~/.de/deplumber file just before dying. This should be called
// before log.Fatal() or from any signal handler.
func cleanupBeforeDeath() {
	if activeProcess.Conn != nil {
		activeProcess.Conn.Close()
	}
}
func plumbListener() {
	f, err := plumb.Open("edit", plan9.OREAD)
	if err != nil {
		cleanupBeforeDeath()
		log.Fatal("Could not open the edit plumbing port :(")
	}

	var m *plumb.Message = &plumb.Message{}
	for {
		if err := m.Recv(f); err != nil {
			cleanupBeforeDeath()
			log.Fatal(err)
			return
		}

		//fmt.Printf("%v", m)
		var filename, fakefile, linespec string
		for attr := m.Attr; attr != nil; attr = attr.Next {
			// If there's a filename attribute, it's a synthetic
			// file and Data is the content, not the filename.
			switch attr.Name {
			case "filename":
				filename = attr.Value
			case "addr":
				linespec = attr.Value
			default:
				//fmt.Printf("%v\n", attr)
			}
		}
		// If we've received a plumbing message from de, we determine
		// whether or not to receive a new de window by checking if the
		// active window (which is the one that sent it) has a dirty
		// buffer.
		//
		// If the message didn't come from de, we always spawn a new
		// window.
		if false && !activeProcess.Dirty && activeProcess.PID != 0 && m.Src == "de" {
			// This needs to be reworked, so for now just open everything in a new
			// window..
			fmt.Fprintf(activeProcess.Conn, "\n%d:%s\n", activeProcess.PID, m.Data)
			if err := ioutil.WriteFile(
				fmt.Sprintf("/proc/%d/note", activeProcess.PID),
				[]byte(fmt.Sprintf("open file: %v", m.Data)),
				0700,
			); err != nil {
				log.Println(err)
			}
		} else {
			var cmd *exec.Cmd
			var args []string
			if m.Dir != "" {
				args = append(args, "-cd", m.Dir)
			}
			if filename == "" {
				// There was no filename attribute, so data
				// is the filename.
				//
				// First, make it relative to dir if possible.
				fname := strings.TrimPrefix(string(m.Data), m.Dir+"/")
				if linespec != "" {
					args = append(args, "de", fname+":"+linespec)
				} else {
					args = append(args, "de", fname)
				}
				cmd = exec.Command("window", args...)
			} else {
				// There was a filename attribute, so data
				// is the content of that "file"

				// We can't just use stdin, because window starts
				// a new process, so write to a temp file and
				// then use that.
				f, err := ioutil.TempFile("", "syntheticplumb")
				if err != nil {
					fmt.Fprintf(os.Stderr, "%v", err)
					continue
				}
				if _, err := f.Write(m.Data); err != nil {
					fmt.Fprintf(os.Stderr, "%v", err)
					continue
				}

				// Tell de to delete the temp file after opening,
				// and pretend that the filename is whatever was
				// in the plumb message.
				fakefile = f.Name()
				var args []string
				if m.Dir != "" {
					args = append(args, "-cd", m.Dir)
				}
				args = append(args, "de", "-delete", "-filename", `'`+filename+`'`, fakefile)
				cmd = exec.Command("window", args...)
			}
			cmd.Dir = m.Dir
			if err := cmd.Start(); err != nil {
				if fakefile != "" {
					os.Remove(fakefile)
				}
				fmt.Fprintf(os.Stderr, "%v", err)
			}
		}
	}
}

func main() {
	// If there was nothing else running, it should be a not found error.
	if _, err := os.Stat("/tmp/deplumber"); err == nil {
		log.Fatalf("Another deplumber instance appears to already be running. If this is not correct, please delete the file /tmp/deplumber\n")
	}
	data, err := os.OpenFile("/tmp/deplumber", os.O_RDONLY|ORCLOSE|os.O_CREATE, 0600)
	if err != nil {
		log.Fatal(err)
	}
	activeProcess.Conn = data

	// Catch any note that will kill the process, so that we can
	// clean up.
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c)

		s := <-c
		// If we get past here, we caught a signal telling us to die.
		cleanupBeforeDeath()
		log.Fatalf("Killed by signal %s", s)
	}()

	// Start a goroutine to listen for plumbing events
	go plumbListener()
	msg := make([]byte, 1024)
	msgRE, err := regexp.Compile(`([\d]+):(.+)`)
	if err != nil {
		panic(err)
	}
	for {
		n, err := data.Read(msg)
		if err == io.EOF || n == 0 {
			time.Sleep(1 * time.Second)
		} else if err != nil {
			panic(err)
		} else {
			if match := msgRE.FindStringSubmatch(string(msg)); len(match) == 3 {
				if pid, err := strconv.Atoi(match[1]); err == nil {
					activeProcess.PID = uint64(pid)
				}
				switch match[2] {
				case "Clean":
					activeProcess.Dirty = false
				case "Dirty":
					activeProcess.Dirty = true
				default:
					fmt.Fprintf(os.Stderr, "Unknown state for PID %v\n", match[1])
				}
				//fmt.Printf("%v", match)
			} else {
				fmt.Printf("%v %v", n, string(msg))
			}

		}
	}
}

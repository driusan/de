// The deplumber is a service which coordinates the running de processes
// to ensure that they don't act on the same plumber message. When it reads
// a plumbing message from the edit port, it will open it in the active de
// window if the buffer is clean, and open a new de window if it's not.
//
// Messages that don't come from de always spawn a new window.
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"

	"9fans.net/go/plan9"
	"9fans.net/go/plumb"
)

// We keep track of the most recent de process that we got a message from,
// and make the assumption that that's the window that should be used for
// plumbing messages. Since plumbing messages are generated from the user
// doing something, we can assume that the user did something in that window
// in order to generate the message.
var activeProcess struct {
	Conn  net.Conn
	PID   uint64
	Dirty bool
}

// Deletes the ~/.de/deplumber file just before dying. This should be called
// before log.Fatal() or from any signal handler.
func cleanupBeforeDeath() {
	os.Remove("/tmp/deplumber")
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

		var filename, fakefile string
		for attr := m.Attr; attr != nil; attr = attr.Next {
			// If there's a filename attribute, it's a synthetic
			// file and Data is the content, not the filename.
			if attr.Name == "filename" {
				filename = attr.Value
			}
			fmt.Printf("Filename: %v", attr)
		}
		// If we've received a plumbing message from de, we determine
		// whether or not to receive a new de window by checking if the
		// active window (which is the one that sent it) has a dirty
		// buffer.
		//
		// If the message didn't come from de, we always spawn a new
		// window.
		if !activeProcess.Dirty && activeProcess.PID != 0 && m.Src == "de" {
			fmt.Fprintf(activeProcess.Conn, "%s\n", m.Data)
		} else {
			var cmd *exec.Cmd
			if filename == "" {
				// There was no filename attribute, so data
				// is the filename.
				cmd = exec.Command("window", "de", string(m.Data))
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
				cmd = exec.Command("window", "de", "-delete", "-filename", `'`+filename+`'`, fakefile)
			}
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
	if data, err := ioutil.ReadFile("/tmp/deplumber"); err == nil && string(data) != "" {
		log.Fatalf("Another deplumber instance appears to already be running. If this is not correct, please delete the file /tmp/deplumber\n")
	}
	ioutil.WriteFile("/tmp/deplumber", []byte("I'm alive!"), 0600)

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

	plumbListener()
}

package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/user"
	"strings"
	"time"
)

// a plumbService coordinates communication between a de process, and a deplumber
// process.
//
// (The deplumber process communicates with the p9p plumber and receives messages.
// It decides whether or not to spawn a new window or re-use the existing one as
// best as it can.)
type plumbService struct {

	// A channel which communicates that a file named
	// string should be opened.
	OpenChan chan string

	// A channel on which the main thread communicates the
	// clean/dirty status of its buffer to us.
	DirtyChan chan bool

	// A channel which the plumbService communicates errors over.
	ErrorChan chan error

	// The Unix Domain Socket connection to the deplumber service.
	conn net.Conn

	// Set to true at the end of initialization, so that Connect()
	// doesn't need to block and Available() will work.
	ready bool
}

// Connect connects the plumbService to a deplumber instance, and returns
// channels that it may asynchronously communicate with the main thread
// over.
//
// In particular, if there's any errors they will be sent over the errors
// channel instead of returned directly, to avoid having to block when calling
// connect.
//
// dirtyChan is a channel that de communicates the buffer dirty status to
// the plumbService over.
func (p *plumbService) Connect(dirtyChan chan bool) {
	p.ErrorChan = make(chan error, 1)
	p.OpenChan = make(chan string)
	p.DirtyChan = dirtyChan
	// Read the ~/.de/deplumber file to see where we should connect
	u, err := user.Current()
	socket, err := ioutil.ReadFile(u.HomeDir + "/.de/deplumber")
	if err != nil {
		p.ErrorChan <- fmt.Errorf("deplumber not started. Plumbing not available.")
		close(p.ErrorChan)
		return
	}

	// Connect.
	// It's a Unix Domain socket. If it takes longer than a second to connect, there's a problem
	p.conn, err = net.DialTimeout("unix", string(socket), time.Second)
	if err != nil {
		p.ErrorChan <- fmt.Errorf("Could not connect to deplumber at %v: %v", string(socket), err)
		close(p.ErrorChan)
		return
	}

	// Monitor the dirtyChan for messages from the main thread saying
	// our dirty bit has changed, and inform the deplumber as appropriate.
	go p.dirtyMonitor()

	// We're now ready to receive messages and monitor the connection for
	// new files that we should open.
	go p.monitorOpenChan()
	return
}

// Returns whether the plumbing service is available and ready
// to plumb messages
func (p *plumbService) Available() bool {
	return p.ready && p.conn != nil
}

// Goes into an infinite loop monitoring the dirtyChan
// for changes in buffer status, and forwards them to the
// deplumber connection.
//
// This should only be called from a goroutine after the
// connection is initialized. It communicates from de to
// the deplumber.
func (p *plumbService) dirtyMonitor() {
	for {
		dirty := <-p.DirtyChan

		if dirty {
			fmt.Fprintf(p.conn, "Dirty\n")
		} else {
			fmt.Fprintf(p.conn, "Clean\n")
		}
	}
}

// Goes into an infinite loop, reading messages from the socket connection
// and sending them across the OpenChan
//
// This should also only be called from a goroutine after the connection is
// initialized. It communicates from the deplumber to de.
func (p *plumbService) monitorOpenChan() {
	// We've connected to the deplumber service, so send it our PID and tell
	// it we have a clean buffer
	fmt.Fprintf(p.conn, "%d\nClean\n", os.Getpid())
	r := bufio.NewReader(p.conn)

	p.ready = true

	for {
		file, err := r.ReadString('\n')
		if err != nil {
			p.ErrorChan <- err
			p.ready = false
			return
		}
		file = strings.TrimSpace(file)
		p.OpenChan <- file
	}
}

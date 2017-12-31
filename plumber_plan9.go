package main

import (
	//"bufio"
	"fmt"
	"os"
	//	"strings"
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

	// The file used for communication with the deplumber service
	conn *os.File

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
	file, err := os.OpenFile("/tmp/deplumber", os.O_RDWR|os.O_APPEND, 0600)

	if err != nil {
		p.ErrorChan <- fmt.Errorf("deplumber not started. Plumbing not available.")
		close(p.ErrorChan)
		return
	}

	p.conn = file

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
	return p.ready
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
			fmt.Fprintf(p.conn, "%d:Dirty\n", os.Getpid())
		} else {
			fmt.Fprintf(p.conn, "%d:Clean\n", os.Getpid())
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
	fmt.Fprintf(p.conn, "%d:Clean\n", os.Getpid())

	p.ready = true
	c := make(chan os.Signal)

	for {
		select {
		case s := <-c:
			fmt.Printf("%v", s)
			p.ErrorChan <- fmt.Errorf("%v", s)
		}
	}
	//r := bufio.NewReader(p.conn)

	/*
		for {
			file, err := r.ReadString('\n')
			switch err {
				case io.EOF:
			}
			if err != nil {
				println("foo")
				p.ErrorChan <- err
				p.ready = false
				return
			}
			file = strings.TrimSpace(file)
			p.OpenChan <- file
		}
	*/
}

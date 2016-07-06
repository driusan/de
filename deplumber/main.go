// The deplumber is a service which coordinates the running de processes
// to ensure that they don't act on the same plumber message. When it reads
// a plumbing message from the edit port, it will open it in the active de
// window if the buffer is clean, and open a new de window if it's not.
//
// Messages that don't come from de always spawn a new window.
package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"strconv"
	"strings"
	"sync"
	"syscall"

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

// A mutex for any modifications to the activeProcess.
var mu sync.Mutex

var listener *net.UnixListener

// Deletes the ~/.de/deplumber file just before dying. This should be called
// before log.Fatal() or from any signal handler.
func cleanupBeforeDeath() {
	// Delete our ~/.de/deplumber file, since we're no longer using it.
	if u, err := user.Current(); err == nil {
		os.Remove(u.HomeDir + "/.de/deplumber")
	}

	if listener != nil {
		listener.Close()
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	buffReader := bufio.NewReader(conn)
	val, err := buffReader.ReadString('\n')
	if err != nil {
		return
	}
	val = strings.TrimSpace(val)
	pid, err := strconv.ParseUint(val, 10, 64)
	if err != nil {
		log.Printf("Invalid PID %s\n", val)
		return
	}
	activeProcess.PID = pid
	activeProcess.Conn = conn

	for {
		val, err := buffReader.ReadString('\n')
		if err != nil {
			return
		}
		val = strings.TrimSpace(val)

		mu.Lock()
		activeProcess.PID = pid
		activeProcess.Conn = conn
		switch val {
		case "Clean":
			activeProcess.Dirty = false
		case "Dirty":
			activeProcess.Dirty = true
		default:
			log.Printf("Unknown status ", val)

		}
		mu.Unlock()
	}
}

func plumbListener() {
	f, err := plumb.Open("edit", plan9.OREAD)
	if err != nil {
		cleanupBeforeDeath()
		log.Fatal("Could not connect to p9p plumber. Is it running?")
		return
	}
	var m *plumb.Message = &plumb.Message{}
	for {
		err := m.Recv(f)
		if err != nil {
			cleanupBeforeDeath()
			log.Fatal(err)
			return
		}
		mu.Lock()

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
			cmd := exec.Command("de", string(m.Data))
			cmd.Dir = m.Dir
			err := cmd.Start()
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v", err)
			}
		}
		mu.Unlock()
	}

}

func main() {
	if u, err := user.Current(); err == nil {
		if data, err := ioutil.ReadFile(u.HomeDir + "/.de/deplumber"); err == nil && string(data) != "" {
			log.Fatalf("Another deplumber instance appears to already be running at %v. If this is not correct, please delete the file ~/.de/deplumber\n", string(data))
		}
	}

	// Get a temporary filename to use as the unix domain socket.
	// We don't actually use it, we just delete it and then create a
	// UDS socket at the same address.
	f, err := ioutil.TempFile("", "delistener")
	if err != nil {
		log.Fatal(err)
	}
	f.Close()
	udsFile := f.Name()
	os.Remove(udsFile)

	// Listen on the temporary file name we just created.
	addr, err := net.ResolveUnixAddr("unix", udsFile)
	if err != nil {
		log.Fatal(err)
	}
	ln, err := net.ListenUnix(
		"unix",
		addr,
	)
	if err != nil {
		log.Fatal(err)
		return
	}
	listener = ln

	// Write the temporary filename to ~/.de/deplumber, so that new de
	// instances know where to send their plumbing status.
	if u, err := user.Current(); err == nil {
		ioutil.WriteFile(u.HomeDir+"/.de/deplumber", []byte(udsFile), 0600)
		// This isn't done in a defer, because there's no code path
		// where we exit normally. We call cleanupBeforeDeath() to do
		// this instead.
		//defer os.Remove(u.HomeDir + "/.de/deplumber")
	}

	// Start a goroutine to listen for plumbing events
	go plumbListener()

	// Start a thread to catch signals. We don't want to do it in the main
	// thread, because ln.Accept() may be blocking.
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c,
			syscall.SIGHUP,
			syscall.SIGINT,
			syscall.SIGTERM,
			syscall.SIGABRT,
			syscall.SIGQUIT,
		)
		s := <-c
		// If we get past here, we caught a signal telling us to die.
		cleanupBeforeDeath()
		log.Fatalf("Killed by signal %s", s)
	}()

	// Wait for dead child processes instead of leaving zombies around
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGCHLD)
		for {
			<-c
			syscall.Wait4(-1, nil, 0, nil)
		}
	}()

	for {
		// Just keep accepting connections on our socket.
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go handleConnection(conn)
	}
}

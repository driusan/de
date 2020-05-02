package actions

import (
	"fmt"
	"os"
	"strconv"

	"9fans.net/go/plan9"
	plumblib "9fans.net/go/plumb"

	"github.com/driusan/de/demodel"
)

// This boolean indicates whether plumbing is ready to be used. It should
// generally only be called by the main thread to flag that plumb messages
// are fine to use.
//
// Until the main thread does that, plumb will fail, assuming that it hasn't
// been properly initialized.
var PlumbingReady bool

func plumb(content []byte, buff *demodel.CharBuffer, v demodel.Viewport, click int) error {
	if !PlumbingReady {
		return fmt.Errorf("Plumbing unavailable")
	}
	fid, err := plumblib.Open("send", plan9.OWRITE)
	if err != nil {
		fmt.Printf("Plumbing error: %v", err)
		return err
	}

	wd, _ := os.Getwd()
	m := plumblib.Message{
		Src:  "de",
		Dst:  "",
		Dir:  wd,
		Type: "text",
		Data: content,
	}
	if click != 0 {
		m.Attr = &plumblib.Attribute{Name: "click", Value: strconv.Itoa(click)}
	}
	return m.Send(fid)
}

func PlumbExecuteOrFindNext(From, To demodel.Position, buff *demodel.CharBuffer, v demodel.Viewport) {
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
	dot.End = i + 1

	word := string(buff.Buffer[dot.Start:dot.End])

	// Don't bother trying if there's nothing listening yet.
	if PlumbingReady {
		// Create a new []byte for the plumber, because if nothing
		// was selected we want to send a message with a click
		// attribute, and if plumbing fails we don't want to have
		// touched the word for the fallbacks.
		var plumbword []byte
		var click int
		var pdot demodel.Dot

		if dot.Start == dot.End-1 {
			// Nothing was selected, to add a "click" attribute
			if dot.Start < 100 {
				click = int(dot.Start)
				pdot.Start = 0
			} else {
				click = 100
				pdot.Start = dot.Start - 100
			}
			if dot.End+100 < uint(len(buff.Buffer)) {
				pdot.End = dot.End + 100
			} else {
				pdot.End = uint(len(buff.Buffer))
			}
			plumbword = buff.Buffer[pdot.Start:pdot.End]
		} else {
			// Default to "word" (the selected text)
			// with no click attribute
			plumbword = []byte(word)
		}

		// If the message was successfully plumbed, we're done.
		if err := plumb(plumbword, buff, v, click); err == nil {
			return
		}
	}
	// Try executing the command. If it works, we're done.
	if err := RunOrExec(word, buff, v); err == nil {
		return
	}

	// We couldn't plumb it, we couldn't execute it, so give up and search for
	// the word
	lenword := dot.End - dot.Start
	for i := dot.End; i < uint(len(buff.Buffer))-lenword; i++ {
		if string(buff.Buffer[i:i+lenword]) == word {
			buff.Dot.Start = i
			buff.Dot.End = i + lenword - 1
			return
		}
	}

}
func PlumbOrFindNext(From, To demodel.Position, buff *demodel.CharBuffer, v demodel.Viewport) {
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
	dot.End = i + 1

	var word string

	// If nothing is selected, instead send 100 characters before and after
	// and include a "click" attribute in the plumbing message.
	var click int
	if dot.Start == dot.End-1 {
		if dot.Start < 100 {
			click = int(dot.Start)
			dot.Start = 0
		} else {
			click = 100
			dot.Start -= 100
		}
		if dot.End+100 < uint(len(buff.Buffer)) {
			dot.End += 100
		} else {
			dot.End = uint(len(buff.Buffer))
		}
	}
	word = string(buff.Buffer[dot.Start:dot.End])

	if PlumbingReady {
		if err := plumb([]byte(word), buff, v, click); err == nil {
			return
		}
	}

	// the file doesn't exist, so find the next instance of word.
	lenword := dot.End - dot.Start
	for i := dot.End; i < uint(len(buff.Buffer))-lenword; i++ {
		if string(buff.Buffer[i:i+lenword]) == word {
			buff.Dot.Start = i
			buff.Dot.End = i + lenword - 1
			return
		}
	}
}

func TagPlumbOrFindNext(From, To demodel.Position, buff *demodel.CharBuffer, v demodel.Viewport) {
	if buff == nil || buff.Tagline == nil {
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
	dot.End = i + 1

	// find the word between From and To in the tagline
	word := string(buff.Tagline.Buffer[dot.Start:dot.End])

	if PlumbingReady {
		if err := plumb([]byte(word), buff, v, 0); err == nil {
			return
		}
	}

	// the file doesn't exist, so find the next instance of word inside
	// the *non-tag* buffer.
	lenword := dot.End - dot.Start
	for i := buff.Dot.End; i < uint(len(buff.Buffer))-lenword; i++ {
		if string(buff.Buffer[i:i+lenword]) == word {
			buff.Dot.Start = i
			buff.Dot.End = i + lenword - 1
			return
		}
	}
}

package demodel

import (
	"errors"
	"io/ioutil"
	"os"
	"os/user"
	"unicode"
)

var ErrNoTagline = errors.New("No tagline exists for buffer")

func (c *CharBuffer) AppendTag(val string) error {
	if c == nil || c.Tagline == nil {
		return ErrNoTagline
	}

	c.Tagline.Buffer = append(c.Tagline.Buffer, []byte(val)...)
	return nil
}

func (c *CharBuffer) ResetTagline() error {
	c.Tagline = &CharBuffer{Buffer: []byte(c.Filename)}
	c.AppendTag(" | " + getDefaultTagline())
	c.Tagline.Dot.Start = uint(len(c.Tagline.Buffer))
	c.Tagline.Dot.End = c.Tagline.Dot.Start
	return nil
}

func getSnarfSaveDir() string {
	u, err := user.Current()
	if err != nil {
		return ""
	}
	return u.HomeDir + "/.de/snarf/"
}

func getDefaultTagline() string {
	u, err := user.Current()
	if err != nil {
		return ""
	}
	file := u.HomeDir + "/.de/tagline"
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return "Save Discard Cut Copy Paste Undo Exit"
	}
	return string(content)
}
func (c *CharBuffer) LoadSnarfBuffer() {
	dir := getSnarfSaveDir()
	if dir == "" {
		return
	}
	sbuf, err := ioutil.ReadFile(dir + "default")
	if err != nil {
		return
	}
	//fmt.Printf("Loading %s into snarf\n", sbuf)
	c.SnarfBuffer = sbuf
}

func (c *CharBuffer) JoinLines(from, to uint) {
	var replaced []byte
	if to >= uint(len(c.Buffer)) {
		to = uint(len(c.Buffer)) - 1
	}
	if from < 0 {
		from = 0
	}
	lineStart := false
	for i := from; i < to; i++ {
		chr := c.Buffer[i]
		switch chr {
		case '\n':
			replaced = append(replaced, ' ')
			lineStart = true
		case '\r':
			// ignore, just in case someone does something
			// stupid like use \n\r for a line ending, we
			// don't want to duplicate spaces.
		default:
			if lineStart && unicode.IsSpace(rune(chr)) {
				continue
			}
			replaced = append(replaced, chr)
			lineStart = false
		}
	}

	newSize := uint(len(replaced))
	newBuffer := make([]byte, from+newSize+uint(len(c.Buffer))-to)
	copy(newBuffer, c.Buffer[:from])
	copy(newBuffer[from:from+newSize], replaced)
	copy(newBuffer[from+newSize:], c.Buffer[to:])

	c.Undo = &CharBuffer{
		Buffer: c.Buffer,
		Dot:    c.Dot,
		Undo:   c.Undo,
	}
	c.Buffer = newBuffer
	c.Dot.End = c.Dot.Start + uint(len(replaced))
}
func (c *CharBuffer) SaveSnarfBuffer() {
	dir := getSnarfSaveDir()
	if dir == "" {
		return
	}
	//fmt.Printf("Saving %s\n", c.SnarfBuffer)
	os.MkdirAll(getSnarfSaveDir(), 0700)
	ioutil.WriteFile(dir+"default", c.SnarfBuffer, 0600)
}

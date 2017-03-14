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

// getHomeDir always returns some sort of home directory
func getHomeDir() string {
	u, err := user.Current()
	if err != nil {
		return os.Getenv("HOME")
	}
	return u.HomeDir
}

// ConfigHome returns the directory where configuration files are written
func ConfigHome() string {
	return getHomeDir() + "/.de"
}

// DataHome returns the directory where data files may be written
func DataHome() string {
	return getHomeDir() + "/.de"
}

func getSnarfSaveDir() string {
	return DataHome() + "/snarf"
}

func getSnarfFile() string {
	return getSnarfSaveDir() + "/default"
}

func getDefaultTagline() string {
	content, err := ioutil.ReadFile(ConfigHome() + "/tagline")
	if err != nil {
		return "Save Discard Cut Copy Paste Undo Exit"
	}
	return string(content)
}

func (c *CharBuffer) LoadSnarfBuffer() {
	sbuf, err := ioutil.ReadFile(getSnarfFile())
	if err != nil {
		return
	}
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
	c.Dirty = true
}

func (c *CharBuffer) SaveSnarfBuffer() {
	os.MkdirAll(getSnarfSaveDir(), 0700)
	ioutil.WriteFile(getSnarfFile(), c.SnarfBuffer, 0600)
}

func (c *CharBuffer) Write(p []byte) (int, error) {
	c.Buffer = append(c.Buffer, p...)
	return len(p), nil
}

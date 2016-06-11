package demodel

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
)

var NoTagline = errors.New("No tagline exists for buffer")

func (c *CharBuffer) AppendTag(val string) error {
	if c == nil || c.Tagline == nil {
		return NoTagline
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

func (c *CharBuffer) SaveSnarfBuffer() {
	dir := getSnarfSaveDir()
	if dir == "" {
		fmt.Printf("No directory to save buffer?")
		return
	}
	//fmt.Printf("Saving %s\n", c.SnarfBuffer)
	os.MkdirAll(getSnarfSaveDir(), 0700)
	ioutil.WriteFile(dir+"default", c.SnarfBuffer, 0600)
}

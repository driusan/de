package demodel

import (
	"errors"
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
	c.AppendTag(" | Save Discard Cut Copy Paste Exit")
	c.Tagline.Dot.Start = uint(len(c.Tagline.Buffer))
	c.Tagline.Dot.End = c.Tagline.Dot.Start
	return nil
}

package defaults

import (
	"bytes"

	"github.com/driusan/de/actions"
	"github.com/driusan/de/demodel"
)

func init() {
	actions.RegisterAction("File", File)
}

// File changes the file name of the current buffer. It will not reload or
// refresh the buffer in any way, only change the file name to used for future
// commands like Save or Discard. File with no arguments will *not* set the
// filename to empty, but rather print the current filename to the tagline.
func File(args string, buff *demodel.CharBuffer, v demodel.Viewport) {
	if args == "" {
		buff.AppendTag("\nFile:" + buff.Filename)
		return
	}

	// Update tagline if it starts with the current filename. Otherwise
	// people will get confused.
	if bytes.HasPrefix(buff.Tagline.Buffer, []byte(buff.Filename)) {
		newTag := make(
			[]byte,
			len(buff.Tagline.Buffer)-len(buff.Filename)+len(args),
		)
		copy(newTag, []byte(args))
		copy(newTag[len(args):], buff.Tagline.Buffer[len(buff.Filename):])
		buff.Tagline.Buffer = newTag
	}
	buff.Filename = args

	// assume that the new filename doesn't have the exact same contents
	// as the current buffer, and just mark it dirty.
	buff.Dirty = true
}

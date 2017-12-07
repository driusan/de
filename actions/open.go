package actions

import (
	"bytes"
	"fmt"
	"image"
	"io/ioutil"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/driusan/de/demodel"
	"github.com/driusan/de/viewer"
)

// setOpenDot sets dot when opening a file to the spec defined by
// params. Understood specs are:
//
// filename:lineno        -- selects entire line
// filename:lineno:column -- sets cursor to column at lineNo
// filename:/regex/       -- selects the first match of regex in the file
func setOpenDot(buff *demodel.CharBuffer, params string) {
	if buff == nil {
		return
	}

	if params == "" {
		buff.Dot.Start = 0
		buff.Dot.End = 0
		return
	}
	pArray := strings.Split(params, ":")
	if len(pArray) == 0 {
		// this shouldn't happen, at least we should have had 1 element
		// with the same content
		panic("Split a non-empty string and got nothing back.")
	}
	if params[0] == '/' {
		var re *regexp.Regexp
		var err error

		// make sure there's 2 slashes, otherwise it's not a valid regex as far
		// as we're concerned.
		if i := strings.LastIndex(params, "/"); i > 0 {
			re, err = regexp.Compile(params[1:i])
			if err != nil {
				fmt.Fprintf(os.Stderr, "Invalid regex in file open semantics: %s\n", params)
				return
			}
		} else {
			fmt.Fprintf(os.Stderr, "Invalid regex in file open semantics: usage filename:/re/\n")
			return

		}

		match := re.FindIndex(buff.Buffer)
		if match == nil {
			buff.Dot.Start = 0
			buff.Dot.End = 0
			return
		}
		buff.Dot.Start = uint(match[0])
		buff.Dot.End = uint(match[1] - 1)
		return
	}

	var lineNo, columnNo int
	var err error
	lineNo, err = strconv.Atoi(pArray[0])
	if err != nil {
		// wasn't a number, wasn't a regex. Don't know what's going on, so
		// abort.
		return
	}

	if lineNo >= 0 && len(pArray) >= 2 {
		columnNo, err = strconv.Atoi(pArray[1])
		if err != nil {
			// a lineNo was set and the column was invalid,
			// to just ignore and pretend it was 0
			columnNo = 0
		}
	}
	if lineNo >= 0 {
		currentLine := 1
		// if endNext is true, we've found the start of the line and should set dot.End to
		// the next line found.
		endNext := false
		for i := 0; i < len(buff.Buffer); i++ {
			if buff.Buffer[i] == '\n' {
				currentLine++
				if endNext {
					buff.Dot.End = uint(i)
					return
				}
			}

			if currentLine >= lineNo && !endNext {
				// a column was provided, to set the cursor to that point
				if columnNo > 0 {
					buff.Dot.Start = uint(i + columnNo)
					buff.Dot.End = buff.Dot.Start
					return
				}
				// no column provided, so set dot to the whole line
				buff.Dot.Start = uint(i + 1)
				endNext = true
			}
		}
		// was the last line, so set the end to the end of the buffer.
		if endNext {
			buff.Dot.End = uint(len(buff.Buffer)) - 1
		}
	}
}

func OpenFile(filename string, buff *demodel.CharBuffer, v demodel.Viewport) error {
	var pager bool
	fstat, err := os.Stat(filename)

	// when opening a file as file:lineno or file:lineno:column: or file:regex
	// dot params is the part after the colon, denoting where to move dot after
	// opening the file.
	var dotparams string
	if filename == "-" {
		pager = true
		goto consideredgood
	}
	if err != nil {
		// check if it's a go style pathspec and try that instead.
		if path := strings.Index(filename, ":"); path > 0 {
			if path+1 < len(filename) {
				dotparams = filename[path+1:]
			}

			filename = filename[:path]

			var err2 error

			if fstat, err2 = os.Stat(filename[:path]); err2 != nil {
				// return the original error, this one was mostly just
				// a stab in the dark.
				return err
			}

			// You're not the boss of me, Dijkstra!
			goto consideredgood
		}
		return err
	}

	// filename has been stat'ed and can generally be considered good to open.
consideredgood:

	if !pager && fstat.IsDir() {
		// it's a directory
		files, err := ioutil.ReadDir(filename)
		if err != nil {
			return err
		}
		os.Chdir(filename)

		var bBuff bytes.Buffer
		if runtime.GOOS != "windows" {
			fmt.Fprintf(&bBuff, "Shell\n\n./\n../\n")
		}

		for _, f := range files {

			if f.IsDir() {
				fmt.Fprintf(&bBuff, "%s/\n", f.Name())
			} else {
				fmt.Fprintf(&bBuff, "%s\n", f.Name())
			}
		}

		// save a reference to the old filename so we can strip the prefix.
		oldFilename := buff.Filename

		buff.Buffer = bBuff.Bytes()
		buff.Filename = filename
		buff.Dot.Start = 0
		buff.Dot.End = 0

		buff.Tagline.Buffer = append(
			[]byte(filename+" "),
			bytes.TrimPrefix(buff.Tagline.Buffer, []byte(oldFilename))...,
		)
	} else {
		// it's a file
		var b []byte
		var ferr error
		if pager == true {
			b, ferr = ioutil.ReadAll(os.Stdin)
		} else {
			b, ferr = ioutil.ReadFile(filename)
		}
		if ferr != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			return ferr
		}

		oldFilename := buff.Filename
		buff.Buffer = b
		buff.Filename = filename
		buff.Dot.Start = 0
		buff.Dot.End = 0
		buff.Dirty = false
		if buff.Tagline.Buffer == nil {
			buff.Tagline.Buffer = make([]byte, 0)
		}
		// if the tagline starts with the filename, update it, otherwise,
		// add it as a prefix.
		buff.Tagline.Buffer = append(
			[]byte(filename+" "),
			bytes.TrimPrefix(buff.Tagline.Buffer, []byte(oldFilename))...,
		)
	}

	// new file, nothing to undo yet..
	buff.Undo = nil

	setOpenDot(buff, dotparams)
	FocusViewport(buff.Dot.Start, buff, v)
	return nil
}

// Focus the viewport v (which is assumed to be rendering buff), such that
// the character at buff.Buffer[idx] is visible somewhere on the screen.
func FocusViewport(idx uint, buff *demodel.CharBuffer, v demodel.Viewport) error {
	if v == nil {
		return fmt.Errorf("No viewport")
	}

	im := v.GetImageMap(buff, image.ZR)
	if im == nil {
		return fmt.Errorf("Could not get image map for buffer.")
	}
	rect, err := im.Get(idx)
	if err != nil {
		return fmt.Errorf("Could not find location of character %d from image map", idx)
	}
	// this should be a method in the demodel.Viewport
	// interface, but it isn't..
	if vp, ok := v.(*viewer.Viewport); ok {
		vp.Location = rect.Min

		// don't put the character *directly* at the top unless it's the very start
		// of the file. If the screen is less than 100px tall, this probably isn't
		//the only thing that will break.
		if vp.Location.Y > 100 {
			vp.Location.Y -= 100
		}
		return nil
	}
	return fmt.Errorf("Could not set viewport location.")
}

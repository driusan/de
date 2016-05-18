package actions

import (
	"bytes"
	"fmt"
	"github.com/driusan/de/demodel"
	"io/ioutil"
	"os"
)

func OpenFile(filename string, buff *demodel.CharBuffer) error {
	fstat, err := os.Stat(filename)
	if err != nil {
		return err
	}
	// file exists, so open it.
	switch fstat.IsDir() {
	case false:
		// it's a file

		b, ferr := ioutil.ReadFile(filename)
		if ferr != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			return ferr
		}

		oldFilename := buff.Filename
		buff.Buffer = b
		buff.Filename = filename
		buff.Dot.Start = 0
		buff.Dot.End = 0

		if buff.Tagline.Buffer == nil {
			buff.Tagline.Buffer = make([]byte, 0)
		}
		// if the tagline starts with the filename, update it, otherwise,
		// add it as a prefix.
		buff.Tagline.Buffer = append(
			[]byte(filename+" "),
			bytes.TrimPrefix(buff.Tagline.Buffer, []byte(oldFilename))...,
		)
	case true:
		// it's a directory
		files, err := ioutil.ReadDir(filename)
		if err != nil {
			return err
		}
		os.Chdir(filename)

		var bBuff bytes.Buffer
		fmt.Fprintf(&bBuff, "./\n../\n")

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

	}
	return nil
}

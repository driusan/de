package actions

import (
	"fmt"
	"github.com/driusan/de/demodel"
	"io/ioutil"
	"os"
)

// Changes Dot to next instance of the character sequence between From
// and To
func FindNext(From, To demodel.Position, buff *demodel.CharBuffer) {
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
	lenword := dot.End - dot.Start
	for i := dot.End; i < uint(len(buff.Buffer))-lenword; i++ {
		if string(buff.Buffer[i:i+lenword]) == word {
			buff.Dot.Start = i
			buff.Dot.End = i + lenword - 1
			return
		}
	}

}

func FindNextOrOpen(From, To demodel.Position, buff *demodel.CharBuffer) {
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

	fmt.Printf("Word: %s\n", word)
	if _, err := os.Stat(word); err == nil {
		// the file exists, so open it
		b, ferr := ioutil.ReadFile(word)
		if ferr != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			return
		}
		buff.Buffer = b
		buff.Filename = word
		buff.Dot.Start = 0
		buff.Dot.End = 0
		return
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

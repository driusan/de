package actions

import (
	"github.com/driusan/de/demodel"
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

	if err := OpenFile(word, buff); err == nil {
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

func FindNextOrOpenTag(From, To demodel.Position, buff *demodel.CharBuffer) {
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

	if err := OpenFile(word, buff); err == nil {
		return
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

package actions

import (
	"errors"
	"fmt"
	"github.com/driusan/de/demodel"
	"io/ioutil"
	"os"
)

var NoFile error = errors.New("Can not save empty filename or nil buffer.")

func SaveFile(From, To demodel.Position, buff *demodel.CharBuffer) error {
	if buff == nil || buff.Filename == "" {
		return NoFile
	}

	// we don't care about positions, just write the file
	err := ioutil.WriteFile(buff.Filename, buff.Buffer, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}
	return err

}

package actions

import (
	"fmt"
	"github.com/driusan/de/demodel"
	"io/ioutil"
	"os"
)

func SaveFile(From, To demodel.Position, buff *demodel.CharBuffer) {
	if buff == nil || buff.Filename == "" {
		return
	}

	// we don't care about positions, just write the file
	err := ioutil.WriteFile(buff.Filename, buff.Buffer, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}

}

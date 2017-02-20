// +build plan9

// Package plumb provides routines for sending and receiving messages for the plumber.
package plumb // import "9fans.net/go/plumb"

import (
	"fmt"
	"io"
	"os"

	"9fans.net/go/plan9"
)

type ByteReadWriteCloser interface {
	io.ReadWriteCloser
	io.ByteReader
}

type byter os.File

func (b *byter) ReadByte() (byte, error) {
	f := make([]byte, 1)
	_, err := b.Read(f)
	return f[0], err
}

func (b *byter) Close() error {
	return (*os.File)(b).Close()
}

func (b *byter) Read(buf []byte) (int, error) {
	return (*os.File)(b).Read(buf)
}

func (b *byter) Write(buf []byte) (int, error) {
	return (*os.File)(b).Write(buf)
}

// Open opens the plumbing file with the given name and open mode.
func Open(name string, mode int) (ByteReadWriteCloser, error) {
	switch mode {
	case plan9.OREAD:
		f, err := os.Open("/mnt/plumb/" + name)
		return (*byter)(f), err
	}
	return nil, fmt.Errorf("Unsupported mode.")
}

package jsonarray

import (
	"bufio"
	"bytes"
	"io"
)

type reader interface {
	io.Reader
	io.ByteReader
}

// stackReader is sort like io.MultiReader, but it insists on
// io.ByteReader and supports inserting readers in front of the queue.
type stackReader struct {
	// front of the queue is highest index, for easy insertion at head
	readers []reader
}

var _ = io.Reader(&stackReader{})
var _ = io.ByteReader(&stackReader{})

func (mr *stackReader) Read(p []byte) (n int, err error) {
	for len(mr.readers) > 0 {
		n, err = mr.readers[len(mr.readers)-1].Read(p)
		// Pop readers that have become empty. Strive to do this one
		// round earlier than when we'd see io.EOF; that way the stack
		// remains at max 2 entries in practice.
		if err == io.EOF || isEmpty(mr.readers[len(mr.readers)-1]) {
			mr.readers = mr.readers[:len(mr.readers)-1]
		}
		if n > 0 || err != io.EOF {
			if err == io.EOF {
				// Don't return io.EOF yet. There may be more bytes
				// in the remaining readers.
				err = nil
			}
			return n, err
		}
	}
	return 0, io.EOF
}

func (mr *stackReader) ReadByte() (c byte, err error) {
	for len(mr.readers) > 0 {
		c, err = mr.readers[len(mr.readers)-1].ReadByte()
		if err == io.EOF {
			mr.readers = mr.readers[:len(mr.readers)-1]
			continue
		}
		return c, err
	}
	return 0, io.EOF
}

// isEmpty peeks inside a bytes.Reader to see if it has been drained.
func isEmpty(r io.Reader) bool {
	if br, ok := r.(*bytes.Reader); ok {
		if br.Len() == 0 {
			return true
		}
	}
	return false
}

// Insert reader in front of the queue. If the reader does not
// implement io.ByteReader, it will be wrapped in a bufio.Reader.
func (mr *stackReader) Insert(r io.Reader) {
	if isEmpty(r) {
		return
	}
	rr, ok := r.(reader)
	if !ok {
		rr = bufio.NewReader(r)
	}
	mr.readers = append(mr.readers, rr)
}

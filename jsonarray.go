package jsonarray

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
)

// NewDecoder returns a new decoder that reads items of a JSON array
// from r. Only one item is held in memory at a time, and items are
// decoded as soon as they are completed, without waiting for the
// whole array.
func NewDecoder(r io.Reader) *Decoder {
	b := bufio.NewReader(r)
	dec := &Decoder{
		state: start,
		r:     &stackReader{readers: []reader{b}},
	}
	return dec
}

type state int

const (
	start  state = iota
	after  state = iota
	broken state = iota
)

// A Decoder reads and decodes JSON array items from an input stream.
type Decoder struct {
	state state
	r     *stackReader
	err   error
}

// ErrNotArray is the type of an error returned when the stream did not
// contain a JSON array.
type ErrNotArray struct {
	Bad byte
}

func (n *ErrNotArray) Error() string {
	return fmt.Sprintf("not an array: starts with %q", n.Bad)
}

// ErrNotCommaSeparated is the type of an error returned when the array
// items in the stream were not comma separated.
type ErrNotCommaSeparated struct {
	Bad byte
}

func (n *ErrNotCommaSeparated) Error() string {
	return fmt.Sprintf("not comma-separated: %q", n.Bad)
}

// Decode unmarshals the next item in the array.
//
// On reaching the end of the array, Decode returns io.EOF. This does
// not mean that the underlying stream would have reached EOF.
//
// If EOF is seen before the JSON array closes, returns
// io.ErrUnexpectedEOF.
func (d *Decoder) Decode(v interface{}) error {
	switch d.state {
	case broken:
		return d.err

	case start:
		c, err := d.readNonWhitespace()
		if err != nil {
			return d.breaks(err)
		}
		switch c {
		case '[':
			// nothing
		default:
			return d.breaks(&ErrNotArray{Bad: c})
		}
		d.state = after

	case after:
		c, err := d.readNonWhitespace()
		if err == io.EOF {
			// did not see closing `]`
			return d.breaks(io.ErrUnexpectedEOF)
		}
		if err != nil {
			return d.breaks(err)
		}
		switch c {
		case ',':
			// nothing
		case ']':
			// end of array
			return d.breaks(io.EOF)
		default:
			return d.breaks(&ErrNotCommaSeparated{Bad: c})
		}
	}

	dec := json.NewDecoder(d.r)
	err := dec.Decode(v)
	if err == io.EOF {
		// did not see closing `]`
		return d.breaks(io.ErrUnexpectedEOF)
	}
	if err != nil {
		return d.breaks(err)
	}
	// patch the parts already buffered back into the reader
	d.r.Insert(dec.Buffered())
	return nil
}

// Read a non-whitespace byte.
func (d *Decoder) readNonWhitespace() (byte, error) {
	for {
		c, err := d.r.ReadByte()
		if err != nil {
			return c, err
		}
		switch c {
		// http://tools.ietf.org/html/rfc7159#section-2
		case 0x20, 0x09, 0x0A, 0x0D:
			continue
		}
		return c, nil
	}
}

func (d *Decoder) breaks(err error) error {
	d.err = err
	d.state = broken
	return err
}

package jsonarray_test

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/tv42/jsonarray"
)

func Example() {
	// simulate streaming by serving reads one byte at a time
	stream := iotest.OneByteReader(
		strings.NewReader(`[{"Greeting": "hell"},{"Greeting": "o, w"},{"Greeting": "orld"}]`),
	)

	type Message struct {
		Greeting string
	}
	dec := jsonarray.NewDecoder(stream)
	for {
		var msg Message
		if err := dec.Decode(&msg); err != nil {
			if err == io.EOF {
				break
			}
			fmt.Printf("decode error: %v\n", err)
			return
		}
		fmt.Printf("%s", msg.Greeting)
	}
	fmt.Printf("\nbye!\n")

	// Output:
	// hello, world
	// bye!
}

type T struct {
	X int
}

func decode(t *testing.T, dec *jsonarray.Decoder, want T) {
	var got T
	if err := dec.Decode(&got); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if got != want {
		t.Fatalf("decode error: %#v != %#v", got, want)
	}
}

func eof(t *testing.T, dec *jsonarray.Decoder) {
	var got T
	err := dec.Decode(&got)
	if err == nil {
		t.Fatalf("expected EOF, got: %#v", got)
	}
	if err != io.EOF {
		t.Fatalf("unexpected decode error: %v", err)
	}
}

func erring(t *testing.T, dec *jsonarray.Decoder, want error) {
	var got T
	err := dec.Decode(&got)
	if err == nil {
		t.Fatalf("expected %v, got: %#v", want, got)
	}
	if err != want {
		t.Fatalf("unexpected decode error: %v != %v", err, want)
	}
}

func TestSimple(t *testing.T) {
	const input = `[{"X":1},  {"X" :2  }  ,  { "X" : 3  } ]  `
	r := strings.NewReader(input)
	dec := jsonarray.NewDecoder(r)
	decode(t, dec, T{X: 1})
	decode(t, dec, T{X: 2})
	decode(t, dec, T{X: 3})
	eof(t, dec)
}

func TestLong(t *testing.T) {
	var buf bytes.Buffer
	buf.Write([]byte(`[`))
	for i := 0; i < 999; i++ {
		fmt.Fprintf(&buf, `{"X":%d},`, i)
	}
	buf.Write([]byte(`{"X":999}]`))
	dec := jsonarray.NewDecoder(&buf)
	for i := 0; i < 1000; i++ {
		decode(t, dec, T{X: i})
	}
	eof(t, dec)
}

func TestBadEarlyEOF(t *testing.T) {
	const input = `[{"X":1},`
	r := strings.NewReader(input)
	dec := jsonarray.NewDecoder(r)
	decode(t, dec, T{X: 1})
	erring(t, dec, io.ErrUnexpectedEOF)
}

# `jsonarray` -- Streaming decoder for JSON arrays

Go library for decoding very large or streaming JSON arrays.

Many streaming JSON APIs give you newline-separated JSON. That's easy
to parse, just keep calling
[`json.Decoder.Decode`](http://golang.org/pkg/encoding/json/#Decoder.Decode).

Sometimes, streaming APIs, and especially JSON databases, just return
a very large JSON array as their result. This is not as easy to
handle. `jsonarray` makes it easy.

(If the large array isn't the outermost JSON object, that's still
harder to get right. Ideas for API that can handle that are welcome.)

Use the Go import path

    github.com/tv42/jsonarray

Documentation at http://godoc.org/github.com/tv42/jsonarray

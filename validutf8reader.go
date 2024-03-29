// Original code:
// https://github.com/kisielk/gorge/util/util.go
// Copyright 2012 Kamil Kisiel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fileutils

import (
	"bufio"
	"io"
	"unicode"
	"unicode/utf8"
)

// ValidUTF8Reader implements a Reader which reads
// only bytes that constitute valid UTF-8.
type ValidUTF8Reader struct {
	buffer *bufio.Reader
}

// Read reads bytes into the byte array passed in.
// It returns `n`, the number of bytes read.
func (rd ValidUTF8Reader) Read(b []byte) (n int, err error) {
	for {
		var r rune
		var size int
		r, size, err = rd.buffer.ReadRune()
		if err != nil {
			return
		}
		if r == unicode.ReplacementChar && size == 1 {
			continue
		} else if n+size < len(b) {
			utf8.EncodeRune(b[n:], r)
			n += size
		} else {
			rd.buffer.UnreadRune()
			break
		}
	}
	return
}

// NewValidUTF8Reader constructs a new `ValidUTF8Reader`
// that wraps an existing `io.Reader`.
func NewValidUTF8Reader(rd io.Reader) ValidUTF8Reader {
	return ValidUTF8Reader{bufio.NewReader(rd)}
}

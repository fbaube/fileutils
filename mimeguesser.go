package fileutils

import (

	// NOTE The hosom/gomagic library is licensed BSD-3,
	// and this file (mimeguesser.go) borrows heavily from it.
	//
	// Three different libraries for determining MIME types were evaluated.
	// All three produced the same results on common files.
	// "hosom" is the easiest to use because it does its own cleanup.

	h2non "github.com/h2non/filetype"

	// hosom "github.com/hosom/gomagic"
)

// GoMagic is based on
//  https://godoc.org/github.com/hosom/gomagic#Magic.Buffer
//  func (m *Magic) Buffer(binary []byte) (string, error)
func GoMagic(s string) string {
	/*
	m, e := mmagic.Buffer([]byte(s))
	if e != nil {
		panic(e)
	}
	return m
	*/
	return "OBSOLETE"
}

// H2N returns:
//  type Type struct { MIME MIME ; Extension string }
//  type MIME struct { Type, Subtype, Value string }
func H2N(s string) string {
	m, e := h2non.Match([]byte(s))
	if e != nil {
		panic("H2N")
	}
	return m.MIME.Type + "/" + m.MIME.Subtype +
		"/" + m.MIME.Value + ";" + m.Extension
}

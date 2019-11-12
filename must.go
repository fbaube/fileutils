package fileutils

import "os"

// Must wraps this packages most common
// return values and panics if it gets an error.
func Must(f *os.File, e error) *os.File {
	if e == nil {
		if f == nil {
			panic("==> fu.Must: os.File: " + e.Error())
		}
		println("==> fu.Must: non-fatal os.File error: ", e.Error())
	}
	return f
}

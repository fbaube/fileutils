package fileutils

import "os"

// func TryOpenRO(path AbsFilePath) (*os.File, error) {

func Must(f *os.File, e error) *os.File {
	if e != nil {
		// if f is non-nil, we could write a big ERR to Stdout, and continue
		panic(e)
	}
	return f
}

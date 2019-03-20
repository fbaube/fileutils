package fileutils

import "os"

func Must(f *os.File, e error) *os.File {
	if e == nil {
		if f == nil {
			panic("==> fu.Must: os.File:" + e.Error())
		}
		println("==> fu.Must: non-fatal os.File error:", e.Error())
	}
	return f
}

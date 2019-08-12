package fileutils

import (
	"io/ioutil"
	"os"
)

// GetStringFromStdin reads `os.Stdin` completely and returns a new
// `InputFile`.
// func NewInputFileFromStdin() (*InputFile, error) {
func GetStringFromStdin() string {
	/*
		p := new(InputFile)
		p.RelFilePath = "-"
		// p.FileFullName is left at "nil"
	*/
	bb, e := ioutil.ReadAll(os.Stdin)
	if e != nil {
		return "STDIN READ FAILURE: " + e.Error() // nil, errors.Wrap(e, "Can't read standard input")
	}
	// p.FileContent = FileContent(S.TrimSpace(string(bb)))
	// p.MagicMimeType = "text/plain"
	return string(bb) // p, nil
}

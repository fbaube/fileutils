package fileutils

import (
	"io/ioutil"
	"os"
)

// GetStringFromStdin reads "os.Stdin" completely and returns a string.
func GetStringFromStdin() string {
	bb, e := ioutil.ReadAll(os.Stdin)
	if e != nil {
		return "STDIN READ FAILURE: " + e.Error()
	}
	// p.FileContent = FileContent(S.TrimSpace(string(bb)))
	// p.MagicMimeType = "text/plain"
	return string(bb)
}

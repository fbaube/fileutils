package fileutils

import (
	"errors"
	"io"
	"os"
)

// GetStringFromStdin reads "os.Stdin" completely and returns a string.
func GetStringFromStdin() (string, error) {
	bb, e := io.ReadAll(os.Stdin)
	if e != nil {
		return "(ERROR!)", errors.New("Cannot read from Stdin: " + e.Error())
	}
	// p.FileContent = FileContent(S.TrimSpace(string(bb)))
	// p.MagicMimeType = "text/plain"
	return string(bb), nil
}

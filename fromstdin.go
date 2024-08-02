package fileutils

import (
	"fmt"
	"io"
	"os"
)

// GetStringFromStdin reads "os.Stdin" completely
// (i.e. until "\n^D") and returns a string.
func GetStringFromStdin() (string, error) {
	bb, e := io.ReadAll(os.Stdin)
	if e != nil {
		return "(ERROR!)", fmt.Errorf("Cannot read from Stdin: %w", e)
	}
	// p.FileContent = FileContent(S.TrimSpace(string(bb)))
	// p.MagicMimeType = "text/plain"
	return string(bb), nil
}

package fileutils

import (
	"fmt"
)

// Echo implements Markupper.
func (p PathInfo) Echo() string {
	return p.AbsFP() // FilePathParts.Echo()
}

// String implements Markupper.
func (p CheckedContent) String() string {
	if p.IsOkayDir() {
		return fmt.Sprintf("PathInfo: DIR[%d]: %s | %s",
			p.Size, p.RelFilePath, p.AbsFP()) // FilePathParts.Echo())
	}
	var isXML string
	if p.IsXml != 0 {
		isXML = "[XML] "
	}
	s := fmt.Sprintf("ChP: %sLen<%d>MType<%s>",
		isXML, p.Size, p.MType) // Mstring())
	s += fmt.Sprintf("\n\t %s | %s",
		p.RelFilePath, Tilded(p.AbsFP())) // ilePathParts.Echo()))
	s += fmt.Sprintf("\n\t (snift) %s ", p.MimeType)
	return s
}

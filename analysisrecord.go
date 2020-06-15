package fileutils

import (
	S  "strings"
	XM "github.com/fbaube/xmlmodels"
)

type AnalysisRecord struct {
	IsXml       int
	MimeType    string
	MType       string
	RootTag     string // e.g. "html", enclosing both <head> and <body>
	RootAtts    string // e.g. <html lang="en"> yields << lang="en" >> 
  MarkdownFlavor string
	XM.XmlInfo
	XM.DitaInfo
}

// FileType returns "XML", "MKDN", "HTML", or future stuff TBD.
func (p AnalysisRecord) FileType() string {
	// Exceptional case
	if S.HasPrefix(p.MType,    "xml/html/") { return "HTML" }
	if S.HasPrefix(p.MimeType, "text/html") { return "HTML" }
	// Normal case
	return S.ToUpper(MTypeSub(p.MType, 0))
}

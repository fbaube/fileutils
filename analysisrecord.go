package fileutils

import (
	S "strings"

	XM "github.com/fbaube/xmlmodels"
)

// AnalysisRecord is the results of content analysis. It is named "Record"
// because it is meant to be persisted to the database.
type AnalysisRecord struct {
	FileExt        string
	MimeType       string
	MType          string
	RootTag        string // e.g. "html", enclosing both <head> and <body>
	RootAtts       string // e.g. <html lang="en"> yields << lang="en" >>
	MarkdownFlavor string
	XM.XmlInfo
	XM.DitaInfo
}

// IsXML is true for all XML, including all HTML.
func (p AnalysisRecord) IsXML() bool {
	s := p.FileType()
	return s == "XML" || s == "HTML"
}

// FileType returns "XML", "MKDN", "HTML", or future stuff TBD.
func (p AnalysisRecord) FileType() string {
	// Exceptional case
	if S.HasPrefix(p.MType, "xml/html/") {
		return "HTML"
	}
	if S.HasPrefix(p.MimeType, "text/html") {
		return "HTML"
	}
	// Normal case
	return S.ToUpper(MTypeSub(p.MType, 0))
}

package fileutils

import (
	"net/http"
	S "strings"
)

/*
https://www.xml.com/pub/a/2007/02/28/what-does-xml-smell-like.html

If the document has a DOCTYPE with a public identifier containing "XHTML,"
such as -//W3C//DTD XHTML 1.0 Transitional//EN, then it is definitely XML.

On the other hand, a DOCTYPE with a public identifier containing "HTML,"
such as -//W3C//DTD HTML 4.01 Transitional//EN, means it is HTML, not XML.

If the DOCTYPE has a system identifier but no public identifier, then it
must be XML, cos XML removed the need for a public identifier in DOCTYPEs.

If the document has an empty DOCTYPE of <!DOCTYPE html>, then it is HTML5.

If we reach the first start tag in the document and none of the heuristic
rules have matched yet, then examine the attributes on the root element.
Any xmlns, xmlns:*, or xml:* attributes, such as xml:lang or xml:base,
mean that the document is XML.
*/

// SetContype comprises four steps:
//
// * use stdlib and third-party libraries to make initial guesses
// * dump those guesses for the purpose of evaluating those libraries
// * call custom code to evaluate more deeply XML and/or as mixed content
// * dump those results for the purpose of refining the code
//
// The fields of interest in `struct fileutiles.InputFile`:
//
// - Set using various heuristics of our own devising: IsXML bool
// - Set using Golang stdlib: SniftMimeType string
// - Set using 3rd-party lib: MagicMimeType string
// - Set by our own code, based on the results set
// in the preceding string fields: Mtype []string
//
func (pIF *InputFile) SetContype() {
	// We're gonna need the file extension and the content itself.
	var filext string
	var content string
	content = string(pIF.FileContent)
	filext = pIF.FileExt
	if !S.HasPrefix(filext, ".") {
		filext = "." + filext
	}
	pIF.MagicMimeType = GoMagic(content)
	// Trim long JPEG descriptions
	if s := pIF.MagicMimeType; S.HasPrefix(s, "JPEG") {
		if i := S.Index(s, "xres"); i > 0 {
			pIF.MagicMimeType = "JPEG, " + s[i:]
		}
	}
	// This next call assigns "text/html" to DITA maps :-/
	contyp := http.DetectContentType([]byte(content))
	pIF.SniftMimeType = S.TrimSuffix(contyp, "; charset=utf-8")
}

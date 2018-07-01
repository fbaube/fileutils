package fileutils

import (
	S "strings"

	SU "github.com/fbaube/stringutils"
)

func (p InputFile) MMCstring() string {
	if p.MMCtype == nil {
		return "-/-/-"
	}
	return p.MMCtype[0] + "/" + p.MMCtype[1] + "/" + p.MMCtype[2]
}

// SetMMCtype works as follows:
// - inputs:
// - - file extension
// - - file mimetype (as already analyzed by a simple third party library)
// - - file content
// - - NOT an input: the DOCTYPE (it is looked at later)
// - outputs:
// - - MMCtype (a [3]string that works like a Mime-Type)
// - - IsXML // , DeclaresDoctype, GuessedDoctype (three booleans, for XML only)
// - - // DeclaresDoctype and GuessedDoctype are mutually exclusive
// Reference material re. MDITA:
// https://github.com/jelovirt/dita-ot-markdown/wiki/Syntax-reference
// The format of local link targets is detected based on file extension.
// The following extensions are treated as DITA files:
// .dita =>	dita ; .xml => dita ; .md => markdown ; .markdown => markdown
func (p *InputFile) SetMMCtype() {

	var theMimeType = p.MimeType
	var theFileExt = p.FileFullName.FileExt
	var theContent = S.TrimSpace(string(p.FileContent))

	if p.MMCtype == nil {
		p.MMCtype = make([]string, 3)
	}
	// Quick exit: IMAGES
	if S.HasPrefix(theMimeType, "image/") {
		p.MMCtype[0] = "image"
		p.IsXML = S.Contains(theMimeType, "xml") ||
			S.Contains(theMimeType, "svg")
		isText :=
			S.Contains(theMimeType, "xml") ||
				S.Contains(theMimeType, "svg") ||
				S.Contains(theMimeType, "eps") ||
				S.Contains(theMimeType, "text")
		if isText {
			p.MMCtype[1] = "text"
			// TODO
			println("Q: What is MMCtype(2) for image/text:", theMimeType)
		} else {
			p.MMCtype[1] = "bin"
			p.MMCtype[2] = S.TrimPrefix(theMimeType, "image/")
		}
		return
	}

	if S.HasPrefix(theMimeType, "app") {
		println("(DD) 3P MIME type is APP(!):", theMimeType)
	} else if !S.HasPrefix(theMimeType, "text/") {
		println("(DD) 3P MIME type is ODD:", theMimeType)
	}
	// Quick exit: DTDs ( .dtd .mod .ent )
	if S.HasPrefix(theContent, "<!") &&
		SU.IsInSliceIgnoreCase(theFileExt, DTDtypeFileExtensions) {
		p.MimeType = "application/xml-dtd"
		p.MMCtype[0] = "schema"
		p.MMCtype[1] = "dtd"
		p.MMCtype[2] = theFileExt[1:4]
		p.IsXML = true
		return
	}
	// Markdown is a tough case, because it's basically a text file.
	// There is no string that definitively declares "This is Markdown",
	// and there might not be YAML metadata at the start of the file,
	// and at this point here we don't want to scan ALL the file content,
	// at least not more than the first few characters.
	// So, the best we can do is check for a known file extension.
	if S.HasPrefix(theMimeType, "text/") &&
		SU.IsInSliceIgnoreCase(theFileExt, MarkdownFileExtensions) {
		p.MimeType = "text/markdown"
		p.MMCtype[0] = "lwdita"
		p.MMCtype[1] = "mdita"
		p.MMCtype[2] = "[TBS]"
		return
	}
	if theMimeType == "text/html" {
		// Can this be HDITA ?
		p.MMCtype[0] = "html"
		p.MMCtype[1] = "[TBS]"
		p.MMCtype[2] = "(nil)"
		p.IsXML = true
		return
	}
	if S.HasPrefix(theMimeType, "text/") &&
		theFileExt == ".dita" { // S.HasPrefix(theFileExt, ".dita") {
		p.MimeType = "application/dita+xml"
		p.MMCtype[0] = "dita"
		p.MMCtype[1] = "[TBS]"
		p.MMCtype[2] = "topic"
		p.IsXML = true
		return
	}
	if S.HasPrefix(theMimeType, "text/") &&
		theFileExt == ".ditamap" {
		p.MimeType = "application/dita+xml"
		p.MMCtype[0] = "lwdita"
		p.MMCtype[1] = "(or_dita)"
		p.MMCtype[2] = "map"
		p.IsXML = true
		return
	}

	// Now we really have to make some guesses.

	var detectedXML bool
	// XML preamble ? Note that whitespace has been trim'd from theContent.
	if theMimeType == "text/xml" || S.HasSuffix(theMimeType, "xml") {
		detectedXML = true
	}
	if S.HasPrefix(theContent, "<?xml ") {
		detectedXML = true
	}
	if S.HasPrefix(theContent, "<!DOCTYPE") {
		detectedXML = true
	}
	// If XML not detected yet, maybe try to detect a tag in the first line
	if !detectedXML {
		// if S.HasPrefix(string(pIF.Contents), "<") && S.
	}
	if detectedXML {
		p.IsXML = true
		if "" != p.MMCtype[0] {
			panic("SetMMCtype")
		}
		p.MMCtype[0] = "xml"
		p.MMCtype[1] = "xml"
	}
	// fmt.Printf("(DD) (%s:%s) MMCtype(%s) \n",
	// 	theFileExt, theMimeType, p.MMCstring())
}

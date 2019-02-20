package fileutils

import (
	S "strings"

	SU "github.com/fbaube/stringutils"
)

// MMCstring extracts (as user-readable text) the MMC type set for the file.
// Note that the MMC type can be set by analyzing the file extension and
// contents, and then later revised if there is an XML `DOCTYPE` declaration.
func (p InputFile) MMCstring() string {
	if p.MMCtype == nil {
		return "-/-/-"
	}
	var ss = []string{p.MMCtype[0], p.MMCtype[1], p.MMCtype[2]}
	for i, s := range p.MMCtype {
		if s == "" {
			ss[i] = "-"
		}
	}
	return ss[0] + "/" + ss[1] + "/" + ss[2]
}

// SetMMCtype works as follows:
//
// Inputs:
// - file extension
// - file mimetype (as already analyzed by a simple third party library)
// - file content
// - NOT an input: the `DOCTYPE` (it is looked at later)
//
// Outputs:
// - `MMCtype` (a `[3]string` slice that works like a MIME type)
// - `IsXML` // , DeclaresDoctype, GuessedDoctype (three booleans, for XML only)
// - // DeclaresDoctype and GuessedDoctype are mutually exclusive
//
// Reference material re. MDITA: <br/>
// https://github.com/jelovirt/dita-ot-markdown/wiki/Syntax-reference
// <br/> The format of local link targets is detected based on file
// extension. <br/>
// The following extensions are treated as DITA files: <br/>
// `.dita` =>	dita ; `.xml` => dita ; `.md` => markdown ; `.markdown` => markdown
//
func (p *InputFile) SetMMCtype() *InputFile {

	// MimeType can be set by `InputFile.OpenAndLoadContent()`
	var theMimeType = p.SniftMimeType
	// theFileExt includes a leading "."
	var theFileExt = p.FileFullName.FileExt

	if p.MMCtype == nil {
		p.MMCtype = []string{"-", "-", "-"}
		// p.MMCtype = make([]string, 0, 3)
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
		return p
	}

	var theContent = S.TrimSpace(string(p.FileContent))

	// Quick exit: DTDs ( .dtd .mod .ent )
	if S.HasPrefix(theContent, "<!") &&
		SU.IsInSliceIgnoreCase(theFileExt, DTDtypeFileExtensions) {
		p.MagicMimeType = "application/xml-dtd"
		p.MMCtype[0] = "schema"
		p.MMCtype[1] = "dtd"
		p.MMCtype[2] = theFileExt[1:4]
		p.IsXML = true
		return p
	}
	// Markdown is a tough case, because it's basically a text file.
	// There is no string that definitively declares "This is Markdown",
	// and there might not be YAML metadata at the start of the file,
	// and at this point here we don't want to scan ALL the file content,
	// at least not more than the first few characters.
	// So, the best we can do is check for a known file extension.
	if S.HasPrefix(theMimeType, "text/") &&
		SU.IsInSliceIgnoreCase(theFileExt, MarkdownFileExtensions) {
		p.MagicMimeType = "text/markdown"
		p.MMCtype[0] = "lwdita"
		p.MMCtype[1] = "mdita"
		p.MMCtype[2] = "[TBS]"
		return p
	}
	if theMimeType == "text/html" {
		// Can this be HDITA ?
		p.MMCtype[0] = "html"
		p.MMCtype[1] = "[TBS]"
		p.MMCtype[2] = "(nil)"
		p.IsXML = true
		return p
	}
	if S.HasPrefix(theMimeType, "text/") &&
		theFileExt == ".dita" { // S.HasPrefix(theFileExt, ".dita") {
		p.MagicMimeType = "application/dita+xml"
		p.MMCtype[0] = "dita"
		p.MMCtype[1] = "[TBS]"
		p.MMCtype[2] = "topic"
		p.IsXML = true
		return p
	}
	if S.HasPrefix(theMimeType, "text/") &&
		theFileExt == ".ditamap" {
		p.MagicMimeType = "application/dita+xml"
		p.MMCtype[0] = "lwdita"
		p.MMCtype[1] = "(or_dita)"
		p.MMCtype[2] = "map"
		p.IsXML = true
		return p
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
			println("SetMMCtype: misjudged XML ?")
		}
		p.MMCtype[0] = "xml"
		p.MMCtype[1] = "xml"
	}
	// fmt.Printf("(DD) (%s:%s) MMCtype(%s) \n",
	// 	theFileExt, theMimeType, p.MMCstring())
	return p
}

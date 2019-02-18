package fileutils

import (
	S "strings"

	SU "github.com/fbaube/stringutils"
)

// Mstring extracts (as user-readable text) the M-type set for the file.
// Note that the M-type can be set by analyzing the file extension and
// contents, and then later revised if there is an XML `DOCTYPE` declaration.
func (p InputFile) Mstring() string {
	if p.Mtype == nil {
		return "-/-/-"
	}
	var ss = []string{p.Mtype[0], p.Mtype[1], p.Mtype[2]}
	for i, s := range p.Mtype {
		if s == "" {
			ss[i] = "-"
		}
	}
	return ss[0] + "/" + ss[1] + "/" + ss[2]
}

// SetMtype works as follows:
//
// Inputs:
// - file extension
// - file mimetype (as already analyzed by a simple third party library)
// - file content
// - NOT an input: the `DOCTYPE` (it is looked at later)
//
// Outputs:
// - `Mtype` (a `[3]string` slice that works like a MIME type)
// -and NOTE that it actually uses lower case
// [0] XML, BIN, TXT, MD
// [1] IMG, CNT, TOC, SCH(ema); maybe others TBD
// [2] fmt/filext: XML Pub-ID, BIN, MD flavor, SCH (DTD/MOD/XSL)
// [3] IFF XML, TBS: Full Public ID string
//
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
func (p *InputFile) SetMtype() *InputFile {

	// theFileExt includes a leading "."
	var theFileExt = p.FileFullName.FileExt

	if p.Mtype == nil {
		p.Mtype = []string{"-", "-", "-"}
		// p.Mtype = make([]string, 0, 3)
	}
	// Quick exit: IMAGES
	if S.HasPrefix(p.MagicMimeType, "image/") {
		p.Mtype[1] = "img"
		p.IsXML = S.Contains(p.MagicMimeType, "xml") ||
			S.Contains(p.MagicMimeType, "svg")
		isXML :=
			S.Contains(p.MagicMimeType, "xml") ||
				S.Contains(p.MagicMimeType, "svg")
		isTXT :=
			S.Contains(p.MagicMimeType, "eps") ||
				S.Contains(p.MagicMimeType, "text")
		if isXML {
			p.Mtype[0] = "xml"
			println("Q: What is Mtype(2) for xml/img:", p.MagicMimeType)
		} else if isTXT {
			p.Mtype[0] = "txt"
			// TODO
			println("Q: What is Mtype(2) for txt/img:", p.MagicMimeType)
		} else {
			p.Mtype[0] = "bin"
			p.Mtype[2] = S.TrimPrefix(p.MagicMimeType, "image/")
			println("Mtype(2)", p.Mtype[2], "for txt/img:", p.MagicMimeType)
		}
		return p
	}

	var theContent = S.TrimSpace(string(p.FileContent))

	// Quick exit: DTDs ( .dtd .mod .ent )
	if S.HasPrefix(theContent, "<!") &&
		SU.IsInSliceIgnoreCase(theFileExt, DTDtypeFileExtensions) {
		p.MagicMimeType = "application/xml-dtd"
		p.Mtype[1] = "sch"
		p.Mtype[0] = "xml"
		p.Mtype[2] = theFileExt[1:4]
		p.IsXML = true
		return p
	}
	// Markdown is a tough case, because it's basically a text file.
	// There is no string that definitively declares "This is Markdown",
	// and there might not be YAML metadata at the start of the file,
	// and at this point here we don't want to scan ALL the file content,
	// at least not more than the first few characters.
	// So, the best we can do is check for a known file extension.
	if S.HasPrefix(p.MagicMimeType, "text/") &&
		SU.IsInSliceIgnoreCase(theFileExt, MarkdownFileExtensions) {
		p.MagicMimeType = "text/markdown"
		p.Mtype[0] = "mdita"
		p.Mtype[1] = "md"
		p.Mtype[2] = "[TBS]"
		return p
	}
	if p.MagicMimeType == "text/html" {
		// Can this be HDITA ?
		p.Mtype[0] = "hdita"
		p.Mtype[1] = "xml"
		p.Mtype[2] = "(nil)"
		p.IsXML = true
		return p
	}
	if S.HasPrefix(p.MagicMimeType, "text/") &&
		theFileExt == ".dita" { // S.HasPrefix(theFileExt, ".dita") {
		p.MagicMimeType = "application/dita+xml"
		p.Mtype[0] = "dita"
		p.Mtype[1] = "xml"
		p.Mtype[2] = "topic"
		p.IsXML = true
		return p
	}
	if S.HasPrefix(p.MagicMimeType, "text/") &&
		theFileExt == ".ditamap" {
		p.MagicMimeType = "application/dita+xml"
		p.Mtype[0] = "toc"
		p.Mtype[1] = "xml"
		p.Mtype[2] = "map"
		p.IsXML = true
		return p
	}

	// Now we really have to make some guesses.

	var detectedXML bool
	// XML preamble ? Note that whitespace has been trim'd from theContent.
	// if theMimeType == "text/xml" || S.HasSuffix(theMimeType, "xml") {
	if S.Contains(p.MagicMimeType, "xml") {
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
		if "" != p.Mtype[0] {
			println("SetMtype: IsXML but empty mime[0]")
		}
		p.Mtype[0] = "xml"
		p.Mtype[1] = "TBD"
		p.Mtype[2] = "TBD"
	}
	// fmt.Printf("(DD) (%s:%s) Mtype(%s) \n",
	// 	theFileExt, theMimeType, p.MMCstring())
	return p
}

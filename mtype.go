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
// - file extension (not really helpful OR reliable)
// - file mimetype (as already analyzed by a simple third party library)
// - file content
// - NOT an input: the `DOCTYPE` (it is looked at later)
//
// Outputs:
// - `Mtype` (a `[3]string` slice that works like a MIME type)
// [0] xml, md, txt, bin
// [1] img, cnt, map, sch(ema); maybe others TBD
// [2] fmt/filext: XML Pub-ID/filext, MD flavor(?), SCH DTD/MOD/XSL, BIN filext
// [3] IFF XML, TBS: Full Public ID string
//
// - `IsXML` // , DeclaredDoctype, GuessedDoctype (three bool's, for XML only)
// - // DeclaredDoctype and GuessedDoctype are mutually exclusive
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
	}
	// For easier checks, incl. "contains"
	sniftMT := S.ToLower(p.SniftMimeType) // redundant
	magicMT := S.ToLower(p.MagicMimeType)
	// For later use
	hasXML := S.Contains(sniftMT, "xml") || S.Contains(magicMT, "xml")
	hasSVG := S.Contains(sniftMT, "svg") || S.Contains(magicMT, "svg")

	// Quick exit: IMAGES (including SVG!)
	if S.HasPrefix(sniftMT, "image/") {
		p.Mtype[1] = "img"
		hasEPS := S.Contains(sniftMT, "eps") || S.Contains(magicMT, "eps")
		hasTXT := S.Contains(sniftMT, "eps") || S.Contains(sniftMT, "text")
		p.IsXML = hasXML || hasSVG
		if hasXML {
			p.Mtype[0] = "xml"
			println("Q: What is Mtype(2) for image/:xml", sniftMT, magicMT)
		} else if hasTXT || hasEPS {
			p.Mtype[0] = "txt"
			// TODO
			println("Q: What is Mtype(2) for image/:text", sniftMT, magicMT)
		} else {
			p.Mtype[0] = "bin"
			p.Mtype[2] = S.TrimPrefix(magicMT, "image/")
			println("Mtype(2)", p.Mtype[2], "for img/bin:", magicMT)
		}
		return p
	}

	var theContent = S.TrimSpace(string(p.FileContent))

	// Quick exit: DTDs ( .dtd .mod .ent )
	if S.HasPrefix(theContent, "<!") &&
		SU.IsInSliceIgnoreCase(theFileExt, DTDtypeFileExtensions) {
		p.SniftMimeType = "application/xml-dtd"
		p.Mtype[1] = "sch"
		p.Mtype[0] = "xml"
		p.Mtype[2] = theFileExt[1:]
		p.IsXML = true
		return p
	}
	// Markdown is a tough case, because it's basically a text file.
	// There is no string that definitively declares "This is Markdown",
	// and there might not be YAML metadata at the start of the file,
	// and at this point here we don't want to scan ALL the file content,
	// at least not more than the first few characters.
	// So, the best we can do is check for a known file extension.
	if S.HasPrefix(sniftMT, "text/") &&
		SU.IsInSliceIgnoreCase(theFileExt, MarkdownFileExtensions) {
		p.SniftMimeType = "text/markdown"
		p.Mtype[0] = "md"
		p.Mtype[1] = "cnt" // or might be "map" ?
		p.Mtype[2] = "[flavr:TBS]"
		return p
	}
	if sniftMT == "text/html" {
		// Can this be HDITA ?
		p.Mtype[0] = "xml"
		p.Mtype[1] = "html" // "hdita"
		p.Mtype[2] = "(nil?!)"
		p.IsXML = true
		return p
	}
	if S.HasPrefix(sniftMT, "text/") &&
		theFileExt == ".dita" { // S.HasPrefix(theFileExt, ".dita") {
		p.SniftMimeType = "application/dita+xml"
		p.Mtype[0] = "xml"
		p.Mtype[1] = "dita"
		p.Mtype[2] = "topic"
		p.IsXML = true
		return p
	}
	if S.HasPrefix(magicMT, "text/") &&
		theFileExt == ".ditamap" {
		p.MagicMimeType = "application/dita+xml"
		p.Mtype[0] = "xml"
		p.Mtype[1] = "dita"
		p.Mtype[2] = "map"
		p.IsXML = true
		return p
	}

	// Now we really have to make some guesses.

	var detectedXML bool
	// XML preamble ? Note that whitespace has been trim'd from theContent.
	// if theMimeType == "text/xml" || S.HasSuffix(theMimeType, "xml") {
	if S.Contains(sniftMT, "ml") { // ""xml") {
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
		println("TODO: Look for an XML tag (fu.mtype.set.L155)")
		// if S.HasPrefix(string(pIF.Contents), "<") && S.
	}
	if detectedXML {
		p.IsXML = true
		if "-" != p.Mtype[0] {
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

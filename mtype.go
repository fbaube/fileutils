package fileutils

import (
	S "strings"
	SU "github.com/fbaube/stringutils"
)

// Mstring extracts (as user-readable text) the file's MType. Note that the
// MType is initially set by analyzing the file extension and contents, but
// can then later revised if there is an XML `DOCTYPE` declaration.
func (p CheckedContent) Mstring() string {
	if p.MType == nil {
		return "-/-/-"
	}
	var ss = []string{p.MType[0], p.MType[1], p.MType[2]}
	for i, s := range p.MType {
		if s == "" {
			ss[i] = "-"
		}
	}
	return ss[0] + "/" + ss[1] + "/" + ss[2]
}

// SetFileMtype works as follows:
//
// Inputs:
// - file extension (not necessarily helpful OR reliable)
// - file mimetype (as already analyzed by a simple third party library)
// - file content
// - NOT an input: the `DOCTYPE` (it is looked at later)
//
// Outputs:
// - `Mtype` (a `[3]string` slice that works like a MIME type)
// [0] xml, md, txt, bin
// [1] img, c(onte)nt, map, sch(ema); maybe others TBD
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
func (p *CheckedContent) SetFileMtype() *CheckedContent {
	if p.error != nil || !p.IsOkayFile() { //p.PathType() != "FILE" {
		return p
	}
	// theFileExt includes a leading "."
	var theFileExt = p.AbsFilePathParts.FileExt

	if p.MType == nil {
		p.MType = []string{"-", "-", "-"}
	}
	// For easier checks, incl. "contains"
	sniftMT := S.ToLower(p.SniftMimeType) // redundant
	magicMT := S.ToLower(p.MagicMimeType)
	// For later use
	hasXML := S.Contains(sniftMT, "xml") || S.Contains(magicMT, "xml")
	hasSVG := S.Contains(sniftMT, "svg") || S.Contains(magicMT, "svg")

	// Quick exit: IMAGES (including SVG!)
	if S.HasPrefix(sniftMT, "image/") {
		p.MType[1] = "img"
		hasEPS := S.Contains(sniftMT, "eps") || S.Contains(magicMT, "eps")
		hasTXT := S.Contains(sniftMT, "eps") || S.Contains(sniftMT, "text")
		p.IsXML = hasXML || hasSVG
		if hasXML {
			p.MType[0] = "xml"
			println("Q: What is Mtype(2) for image/:xml", sniftMT, magicMT)
		} else if hasTXT || hasEPS {
			p.MType[0] = "txt"
			// TODO
			println("Q: What is Mtype(2) for image/:text", sniftMT, magicMT)
		} else {
			p.MType[0] = "bin"
			p.MType[2] = S.TrimPrefix(magicMT, "image/")
			println("Mtype(2)", p.MType[2], "for img/bin:", magicMT)
		}
		return p
	}

	var theContent = S.TrimSpace(p.Raw)
	if len(theContent) == 0 {
		println("zero content:", p.Raw)
	}

	// Quick exit: DTDs ( .dtd .mod .ent )
	if // S.HasPrefix(theContent, "<!") &&
		SU.IsInSliceIgnoreCase(theFileExt, DTDtypeFileExtensions) {
		p.SniftMimeType = "application/xml-dtd"
		p.MType[1] = "sch"
		p.MType[0] = "xml"
		p.MType[2] = theFileExt[1:]
		p.IsXML = true
		return p
	}
	// Markdown is a tough case, because it's basically a text file.
	// There is no string that definitively declares "This is Markdown",
	// and there might not be YAML metadata at the start of the file,
	// and at this point here we don't want to scan ALL the file content,
	// at least not more than the first few characters.
	// So, about the best we can do is check for a known file extensions.
	if S.HasPrefix(sniftMT, "text/") &&
		SU.IsInSliceIgnoreCase(theFileExt, MarkdownFileExtensions) {
		p.SniftMimeType = "text/markdown"
		p.MType[0] = "mkdn"
		p.MType[1] = "tpcOrMap" // or might be "map" ?
		p.MType[2] = "[flavr:TBS]"
		return p
	}
	if sniftMT == "text/html" {
		if !S.Contains(p.Raw, "<!DOCTYPE HTML ") {
			println("--> text/html has no DOCTYPE HTML")
		}
		// Can this be HDITA ?
		p.MType[0] = "xml"
		p.MType[1] = "html" // "hdita"
		p.MType[2] = "(nil?!)"
		p.IsXML = true
		return p
	}
	if S.HasPrefix(sniftMT, "text/") &&
		theFileExt == ".dita" { // S.HasPrefix(theFileExt, ".dita") {
		p.SniftMimeType = "application/dita+xml"
		p.MType[0] = "xml"
		p.MType[1] = "dita"
		p.MType[2] = "topic"
		p.IsXML = true
		return p
	}
	if S.HasPrefix(magicMT, "text/") &&
		theFileExt == ".ditamap" {
		p.MagicMimeType = "application/dita+xml"
		p.MType[0] = "xml"
		p.MType[1] = "dita"
		p.MType[2] = "map"
		p.IsXML = true
		return p
	}

	// Now we REALLY have to do some guesswork.

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
		println("TODO: Look for an XML tag (fu.mtype.set.L164)")
		// if S.HasPrefix(string(pIF.Contents), "<") && S.
	}
	if detectedXML {
		p.IsXML = true
		if "-" != p.MType[0] {
			println("SetMtype: IsXML but empty mime[0]")
		}
		p.MType[0] = "xml"
		p.MType[1] = "TBD"
		p.MType[2] = "TBD"
		// Check for "<!DOCTYPE HTML "
		if S.Contains(p.Raw, "<!DOCTYPE HTML ") {
			p.MType[1] = "html" // "hdita"
			p.MType[2] = "(nil?!)"
		}
	}
	// fmt.Printf("(DD) (%s:%s) Mtype(%s) \n",
	// 	theFileExt, theMimeType, p.MMCstring())
	return p
}

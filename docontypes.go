package fileutils

import (
	S "strings"

	SU "github.com/fbaube/stringutils"
	XM "github.com/fbaube/xmlmodels"
)

// DoContentTypes sets MType, MimeType, and IsXml, as follows:
//
// Inputs:
//  * file extension (not necessarily helpful OR reliable)
//  * file mimetype (as already analyzed by a simple third party library)
//  * file content
//  * NOT an input: the `DOCTYPE` (it is looked at later)
//
// Outputs: (NOTE: Obsolete comments ?)
//  `Mtype` (a `[3]string` slice that works like a MIME type)
//  [0] xml, md, txt, bin
//  [1] img, c(onte)nt, map, sch(ema); maybe others TBD
//  [2] fmt/filext: XML Pub-ID/filext, MD flavor(?), SCH DTD/MOD/XSL, BIN filext
//  [3] IFF XML, TBS: Full Public ID string
//
// - `IsXML` // , DeclaredDoctype, GuessedDoctype (three bool's, for XML only)
// - // DeclaredDoctype and GuessedDoctype are mutually exclusive
//
// Reference material re. MDITA:
//  https://github.com/jelovirt/dita-ot-markdown/wiki/Syntax-reference
//  The format of local link targets is detected based on file
//  extension.
//  The following extensions are treated as DITA files:
//  `.dita` =>	dita ; `.xml` => dita ; `.md` => markdown ; `.markdown` => markdown
//
func DoContentTypes(sniftMimeType, sCont, theFileExt string) (retMimeType, retMType string) {
	// theFileExt includes a leading "."
	retMType = "-/-/-"
	sniftMT := S.ToLower(sniftMimeType)
	// For later use
	hasXML := S.Contains(sniftMT, "xml")
	// !! hasSVG := S.Contains(sniftMT, "svg")

	// Quick exit: IMAGES (including SVG!)
	if S.HasPrefix(sniftMT, "image/") {
		hasEPS := S.Contains(sniftMT, "eps")
		hasTXT := S.Contains(sniftMT, "eps") || S.Contains(sniftMT, "text")
		// !! !! p.IsXml = boolToInt(hasXML || hasSVG)
		if hasXML {
			println("Q: What is Mtype(2) for image/:xml", sniftMT)
			return sniftMT, "xml/img/???"
		} else if hasTXT || hasEPS {
			// TODO
			println("Q: What is Mtype(2) for image/:text", sniftMT)
			return sniftMT, "txt/img/???"
		} else {
			// p.MType[2] = S.TrimPrefix(magicMT, "image/")
			println("Q: What is Mtype(2) for bin/*", sniftMT)
			return sniftMT, "bin/img/???"
		}
		return
	}

	var theContent = S.TrimSpace(sCont)
	if len(theContent) == 0 {
		println("zero content:", sCont)
	}

	// Quick exit: DTDs ( .dtd .mod .ent )
	if // S.HasPrefix(theContent, "<!") &&
	SU.IsInSliceIgnoreCase(theFileExt, XM.DTDtypeFileExtensions) {
		return "application/xml-dtd", "xml/sch/" + theFileExt[1:]
	}
	// Markdown is a tough case, because it's basically a text file.
	// There is no string that definitively declares "This is Markdown",
	// and there might not be YAML metadata at the start of the file,
	// and at this point here we don't want to scan ALL the file content,
	// at least not more than the first few characters.
	// So, about the best we can do is check for a known file extensions.
	if S.HasPrefix(sniftMT, "text/") &&
		SU.IsInSliceIgnoreCase(theFileExt, XM.MarkdownFileExtensions) {
		return "text/markdown", "mkdn/tpcOrMap/?fmt"
	}
	if sniftMT == "text/html" {
		// FIXME: Make case-insensitive
		if !S.Contains(sCont, "<!DOCTYPE HTML") {
			println("--> text/html has no DOCTYPE HTML")
		}
		// Can this be HDITA ?
		return sniftMT, "xml/html/(nil?!)" // or hdita not html
	}
	if S.HasPrefix(sniftMT, "text/") &&
		theFileExt == ".dita" { // S.HasPrefix(theFileExt, ".dita") {
		return "application/dita+xml", "xml/dita/topic"
	}
	if S.HasPrefix(sniftMT, "text/") &&
		(theFileExt == ".ditamap" || theFileExt == ".map") {
		return "application/dita+xml", "xml/dita/map"
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
		retMimeType = "xml/TBD/TBD"
		// Check for "<!DOCTYPE HTML "
		if S.Contains(sCont, "<!DOCTYPE HTML") {
			retMimeType = "xml/html/(nil?!)" // or hdita for html
		}
	}
	// fmt.Printf("(DD) (%s:%s) Mtype(%s) \n",
	// 	theFileExt, theMimeType, p.MMCstring())
	return retMimeType, retMType
}

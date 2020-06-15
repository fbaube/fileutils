package fileutils

import (
	S "strings"
	SU "github.com/fbaube/stringutils"
)

// DoContentTypes_non_xml is TBS.
//
func DoContentTypes_non_xml(sniftMT, sCont, filext string) (retMT, retMType string) {
	// theFileExt includes a leading "."
	retMType = "-/-/-"

	// Quick exit: IMAGES (including SVG!?)
	if S.HasPrefix(sniftMT, "image/") {
		hasEPS := S.Contains(sniftMT, "eps")
		hasTXT := S.Contains(sniftMT, "eps") || S.Contains(sniftMT, "text")
		// !! !! p.IsXml = boolToInt(hasXML || hasSVG)
		/* if hasXML {
			println("Q: What is Mtype(2) for image/:xml", sniftMT)
			return sniftMT, "xml/img/???"
		} else */ if hasTXT || hasEPS {
			// TODO
			println("Q: What is Mtype(2) for image/:text", sniftMT)
			return sniftMT, "txt/img/???"
		} else {
			// p.MType[2] = S.TrimPrefix(magicMT, "image/")
			println("Q: What is Mtype(2) for bin/*", sniftMT)
			return sniftMT, "bin/img/???"
		}
	}
	// Markdown is a tough case, because it's basically a text file.
	// There is no string that definitively declares "This is Markdown",
	// and there might not be YAML metadata at the start of the file,
	// and at this point here we don't want to scan ALL the file content,
	// at least not more than the first few characters.
	// So, about the best we can do is check for a known file extensions.
	if S.HasPrefix(sniftMT, "text/") &&
		SU.IsInSliceIgnoreCase(filext, MarkdownFileExtensions) {
		return "text/markdown", "mkdn/tpcOrMap/[flavr:TBS]"
	}
	// fmt.Printf("(DD) (%s:%s) Mtype(%s) \n",
	// 	theFileExt, theMimeType, p.MMCstring())
	return retMT, retMType
}

package fileutils

import (
	"fmt"
	"net/http"
	FP "path/filepath"
	S "strings"

	SU "github.com/fbaube/stringutils"
	XM "github.com/fbaube/xmlmodels"
)

// AnalyseFile has drastically different handling for XML content versus
// non-XML content, so for the sake of clarity and simplicity, it tries
// to first process non-XML content. For XML files we have much more
// info available, so processing is both simpler and more complicated.
//
// The second argument "filext" can be any filepath; the Go stdlib is used
// to split off the file extension. It can also be "", if (for example) the
// content is entered interactively, without a file name or assigned MIME type.
//
// If the first argument "sCont" (the content) is zero-length, return (nil, nil).
//
func AnalyseFile(sCont string, filext string) (*AnalysisRecord, error) {

	var pCntpg *XM.Contyping
	var pAnlRec *AnalysisRecord
	var hasKeyElm bool

	if sCont == "" {
		println("==>", "Cannot analyze zero-length content")
		return nil, nil
	}
	// A trailing dot in the filename provides no filetype info.
	filext = FP.Ext(filext)
	if filext == "." {
		filext = ""
	}
	fmt.Printf("==> Starting file content analysis: len<%d> filext<%s> \n", len(sCont), filext)

	// =======================================
	//  stdlib HTTP content detection library
	// =======================================
	// Note that it assigns "text/html" to DITA maps :-/
	var httpContype string
	httpContype = http.DetectContentType([]byte(sCont))
	httpContype = S.TrimSuffix(httpContype, "; charset=utf-8")
	println("-->", "Contype per HTTP stdlib:", httpContype)

	// Preliminary
	pCntpg = new(XM.Contyping)
	pAnlRec = new(AnalysisRecord)
	// pAnlRec.MType = ""
	pCntpg.FileExt = filext
	pCntpg.MimeType = httpContype

	// =======================================
	//  Quick check for top-level XML stuff
	// =======================================
	var keyElms []string
	keyElms = []string{
		"html",
		"topic",
		"map",
		"reference",
		"task",
		"bookmap",
		"glossentry",
		"glossgroup",
	}
	Peek := XM.PeekAtStructure_xml(sCont, keyElms)
	if Peek.HasError() {
		return nil, fmt.Errorf("fu.peekXml failed: %w", Peek.GetError())
	}
	hasKeyElm = (Peek.KeyElmTag != "")
	if hasKeyElm {
		var pos int
		pos = Peek.KeyElmPos.Pos
		fmt.Printf("D=> fu.AF: keyElm: %s / %+v \n", Peek.KeyElmTag, pos)
		// fmt.Printf("D=> Key Elm <%s> position: %s (%d) \n",
		// 	Peek.KeyElmTag, Peek.KeyElmPos, pos)
		fmt.Printf("D=> Key Elm <%s> at |%s| \n",
			Peek.KeyElmTag, sCont[pos:pos+10])
	}
	var gotXml, gotPreamble, gotDoctype, gotRootTag, gotDTDstuff bool
	gotPreamble = (Peek.Preamble != "")
	gotDoctype = (Peek.Doctype != "")
	gotRootTag = (Peek.RootTag != "")
	gotDTDstuff = Peek.HasDTDstuff
	gotXml = gotRootTag || gotDTDstuff || gotDoctype || gotPreamble
	if !gotXml {
		println("--> Does not seem to be XML")
	} else {
		fmt.Printf("--> xm.peek: preamble<%s> doctype<%s> DTDstuff<%s> RootTag<%s> \n",
			SU.Yn(gotPreamble), SU.Yn(gotDoctype), SU.Yn(gotDTDstuff), SU.Yn(gotRootTag))
	}
	bb, ss := SeemsToBeXml(httpContype, filext)
	if !gotXml && bb {
		gotXml = true
		println("--> Seems to be XML after all, based on file ext. and HTTP stdlib:", ss)
	}
	if gotXml && !(gotRootTag || gotDTDstuff) {
		println("--> XML file has no root tag (and is not DTD)")
	}
	if gotDTDstuff && SU.IsInSliceIgnoreCase(filext, XM.DTDtypeFileExtensions) {
		fmt.Printf("--> DTD content detected (filext<%s>) \n", filext)
		pAnlRec.MimeType = "application/xml-dtd"
		pAnlRec.MType = "xml/sch/" + filext[1:]
		// Could allocate and fill field XmlInfo
		return pAnlRec, nil
	}
	/*
		TAKE A DUMP HERE !!
		A reminder of what we should be setting:
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
		Also at this point we should have the sep.pt btwn head/meta and body/text,
		and the struct Sections filled in.
	*/
	// =============
	//   NOT XML ?
	// =============
	if !gotXml {
		var nonXmlCntpg *XM.Contyping
		nonXmlCntpg = DoContentTypes_non_xml(httpContype, sCont, filext)
		fmt.Printf("--> NON-XML: %s \n", nonXmlCntpg)
		pAnlRec.Contyping = *nonXmlCntpg
		return pAnlRec, nil
	}

	// ===============
	//   YES XML !!!
	// ===============
	// var isLwDita bool
	var pPRF *XM.XmlPreambleFields

	var e error
	if gotPreamble {
		// println("preamble:", preamble)
		pPRF, e = XM.NewXmlPreambleFields(Peek.Preamble)
		if e != nil {
			println("xm.peek: preamble failure")
			return nil, fmt.Errorf("xm.peek: preamble failure: %w", e)
		}
		print("--> Parsed XML preamble: " + pPRF.String())
	}

	// At this point, we have pCntpg.
	// Is this repetitive ?
	pCntpg.FileExt = filext
	pCntpg.MimeType = httpContype

	// Got DOCTYPE ? If so, it is gospel.
	if Peek.Doctype != "" {
		// We are here if we got a DOCTYPE; we also have a file extension,
		// and we should have a root tag (or else the DOCTYPE makes no sense !)
		var pDF *XM.DoctypeFields
		pDF = pCntpg.AnalyzeDoctype(Peek.Doctype)
		if pDF.HasError() {
			panic("FIXME:" + pDF.Error())
		}
		println("--> AnlzDT: Contpg:" + pCntpg.String())
		println("--> AnlzDT: DTflds:" + pDF.String())

		println("==> TODO: cp from DT fields to AnlRec")

		return pAnlRec, nil
	}
	// We don't have a DOCTYPE, so it's gonna be a PITA !
	// // So let's at least try to set the MType.
	if !gotRootTag {
		fmt.Printf("==> Got no root tag; filext: %s \n", filext)
		return pAnlRec, nil
	}
	// We are here if we have only a root tag and a file extension.
	var rutag string
	if rutag == "" {
		panic("Got nil root tag")
	}
	rutag = S.ToLower(Peek.RootTag)
	fmt.Printf("fu.AF: rutag<%s> filext<%s> ?mtype<%s> \n",
		rutag, filext, pAnlRec.MType)
	pCntpg.MType = pAnlRec.MType
	// Do some easy cases
	if rutag == "html" && (filext == ".html" || filext == ".htm") {
		pCntpg.MType = "html/cnt/html5"
	} else if rutag == "html" && S.HasPrefix(filext, ".xht") {
		pCntpg.MType = "html/cnt/xhtml"
	} else if SU.IsInSliceIgnoreCase(rutag, XM.DITArootElms) &&
		SU.IsInSliceIgnoreCase(filext, XM.DITAtypeFileExtensions) {
		pCntpg.MType = "xml/cnt/" + rutag
		if rutag == "bookmap" && S.HasSuffix(filext, "map") {
			pCntpg.MType = "xml/map/" + rutag
		}
	}
	println("--> MType guessing, XML, no Doctype:", rutag, filext)
	pAnlRec.Contyping = *pCntpg
	if pAnlRec.MType == "-/-/-" {
		pAnlRec.MType = "xml/???/" + rutag
	}

	// At this point, mt should be valid !
	println("--> fu.AF: Contyping (derived both ways):", pAnlRec.Contyping.String())

	// Now we should fill in all the detail fields.
	/*
	  type XmlInfo struct {
	    XmlContype
	    XmlPreambleFields
	    XmlDoctype
	   *XmlDoctypeFields
	  }
	  type DitaInfo struct {
	    DitaMarkupLg
	    DitaContype
	  } */
	pAnlRec.XmlContype = "RootTagData"
	pAnlRec.XmlDoctype = XM.XmlDoctype("DOCTYPE " + Peek.Doctype)
	// pAnlRec.DoctypeFields = pDF
	if pPRF != nil {
		pAnlRec.XmlPreambleFields = pPRF
	} else {
		// SKIP
		// pBA.XmlPreambleFields = XM.STD_PreambleFields
	}
	// pBA.DoctypeIsDeclared  =  true
	pAnlRec.DitaMarkupLg = "TBS"
	pAnlRec.DitaContype = "TBS"

	fmt.Printf("--> fu.analyzeFile: \n--> 1) MType: %s \n--> 2) XmlInfo: %s \n--> 3) DitaInfo: %s \n",
		pAnlRec.MType, pAnlRec.XmlInfo, pAnlRec.DitaInfo)

	return pAnlRec, nil
}

func CollectKeysOfNonNilMapValues(M map[string]*XM.FilePosition) []string {
	var ss []string
	for K, V := range M {
		if V != nil {
			ss = append(ss, K)
		}
	}
	return ss
}

func SeemsToBeXml(httpContype string, filext string) (isXml bool, msg string) {

	if S.Contains(httpContype, "xml") {
		return true, "HTTP contype-detection got XML (in MIME type)"
	}
	if httpContype == "text/html" {
		return true, "HTTP contype-detection got XML (text/html, HDITA?)"
	}
	if S.HasPrefix(httpContype, "text/") &&
		(filext == ".dita" || filext == ".ditamap" || filext == ".map") {
		return true, "HTTP contype-detection got XML (text/dita-filext)"
	}
	if S.Contains(httpContype, "ml") {
		return true, "HTTP contype-detection got <ml>"
	}
	if S.Contains(httpContype, "svg") {
		return true, "HTTP contype-detection got <svg>"
	}
	return false, ""
}

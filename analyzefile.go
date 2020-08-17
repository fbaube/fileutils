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

	var pBA *AnalysisRecord
	if sCont == "" {
		println("==>", "Cannot analyze zero-length content")
		return nil, nil
	}
	filext = FP.Ext(filext)
	if filext == "." {
		filext = ""
	}
	fmt.Printf("==> Content analysis: len<%d> filext<%s> \n", len(sCont), filext)

	// =======================================
	//  stdlib HTTP content detection library
	// =======================================
	// Note that it assigns "text/html" to DITA maps :-/
	httpContype := http.DetectContentType([]byte(sCont))
	httpContype = S.TrimSuffix(httpContype, "; charset=utf-8")
	println("-->", "HTTP stdlib:", httpContype)

	pBA = new(AnalysisRecord)
	pBA.MType = "-/-/-"
	pBA.FileExt = filext
	pBA.MimeType = httpContype

	// =======================================
	//  Quick check for top-level XML stuff
	// =======================================
	var elmMap map[string]*XM.FilePosition
	elmMap = map[string]*XM.FilePosition{
		"html":       nil,
		"topic":      nil,
		"concept":    nil,
		"reference":  nil,
		"task":       nil,
		"bookmap":    nil,
		"glossentry": nil,
		"glossgroup": nil,
	}
	Peek := XM.PeekAtStructure_xml(sCont, elmMap)
	if Peek.HasError() {
		return pBA, fmt.Errorf("fu.peekXml failed: %w", Peek.GetError())
	}
	var isXml, gotPreamble, gotDoctype, gotRootTag, gotDTDstuff bool
	gotPreamble = (Peek.Preamble != "")
	gotDoctype = (Peek.Doctype != "")
	gotRootTag = (Peek.RootTag.Name.Local != "")
	gotDTDstuff = Peek.HasDTDstuff
	isXml = gotRootTag || gotDTDstuff || gotDoctype || gotPreamble
	if !isXml {
		println("--> Does not seem to be XML")
	} else {
		fmt.Printf("--> xm.peek: preamble<%s> doctype<%s> DTDstuff<%s> RootTag<%s> \n",
			SU.Yn(gotPreamble), SU.Yn(gotDoctype), SU.Yn(gotDTDstuff), SU.Yn(gotRootTag))
	}
	bb, ss := SeemsToBeXml(httpContype, filext)
	if !isXml && bb {
		isXml = true
		println("--> Seems to be XML after all:", ss)
	}
	if isXml && !(gotRootTag || gotDTDstuff) {
		println("-->", "XML file has no root tag (and is not DTD)")
	}
	if gotDTDstuff && SU.IsInSliceIgnoreCase(filext, XM.DTDtypeFileExtensions) {
		fmt.Printf("--> DTD type detected (filext<%s>) \n", filext)
		pBA.MimeType = "application/xml-dtd"
		pBA.MType = "xml/sch/" + filext[1:]
		return pBA, nil
	}

	// =============
	//   NOT XML ?
	// =============
	if !isXml {
		var mimetype, mtype string
		mimetype, mtype = DoContentTypes_non_xml(httpContype, sCont, filext)
		fmt.Printf("--> NON-XML: filext<%s> mtype<%s> mimetype<%s> \n",
			filext, mtype, mimetype)
		pBA.FileExt = filext
		pBA.MimeType = mimetype
		pBA.MType = mtype
		return pBA, nil
	}

	// ===============
	//   YES XML !!!
	// ===============
	// var isLwDita bool
	var pPRF *XM.XmlPreambleFields
	var pDTF *XM.XmlDoctypeFields

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

	mt := "none"
	dtmt := "none"
	// isLwdita := false

	// If we don't have a DOCTYPE, it's gonna be a PITA !
	// So le's at least try to se the MType.
	if Peek.Doctype == "" {
		if gotRootTag {
			var rutag string
			rutag = S.ToLower(Peek.RootTag.Name.Local)
			if rutag == "" {
				panic("Nil root tag")
			}
			if rutag == "html" && S.HasPrefix(filext, ".ht") {
				pBA.MType = "html/cnt/html5"
			}
			if rutag == "html" && S.HasPrefix(filext, ".xht") {
				pBA.MType = "html/cnt/xhtml"
			}
			if SU.IsInSliceIgnoreCase(rutag, XM.DITArootElms) &&
				SU.IsInSliceIgnoreCase(filext, XM.DITAtypeFileExtensions) {
				pBA.MType = "xml/cnt/" + rutag
				if rutag == "bookmap" && S.HasSuffix(filext, "map") {
					pBA.MType = "xml/map/" + rutag
				}
			}
			println("--> MType guessing, XML, no Doctype:", rutag, filext)
			pBA.MType = "xml/???/" + rutag
		}
	} else {
		println("-->", Peek.Doctype)
		mt, _ = XM.GetMTypeByDoctype(Peek.Doctype)
		// println("-->", "Doctype/MType search results:", mt)

		// If we got an MType, we don't really need to make this call,
		// but for now let's do it anyways.
		pDTF, e = XM.NewXmlDoctypeFieldsInclMType(Peek.Doctype)
		if e != nil {
			println("xm.peek: doctype failure")
		}
		println("-->", "Parsed doctype:", pDTF.String())
		dtmt = pDTF.DoctypeMType

		if pDTF.TopTag != "" && Peek.RootTag.Name.Local != "" &&
			S.ToLower(pDTF.TopTag) != S.ToLower(Peek.RootTag.Name.Local) {
			fmt.Printf("--> RootTag MISMATCH: doctype<%s> bodytext<%s> \n",
				pDTF.TopTag, Peek.RootTag.Name.Local)
			panic("ROOT TAG MISMATCH")
		}
		if mt != dtmt {
			fmt.Printf("--> M-Type MISMATCH: contype<%s> doctype<%s> \n", mt, dtmt)
			panic("M-TYPE MISMATCH")
		}
	}
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
	pBA.XmlInfo.XmlContype = "RootTagData"
	pBA.XmlInfo.XmlDoctype = XM.XmlDoctype("DOCTYPE " + Peek.Doctype)
	pBA.XmlDoctypeFields = pDTF
	if pPRF != nil {
		pBA.XmlPreambleFields = *pPRF
	} else {
		pBA.XmlPreambleFields = XM.STD_PreambleFields
	}
	// pBA.DoctypeIsDeclared  =  true
	pBA.DitaInfo.DitaMarkupLg = "TBS"
	pBA.DitaInfo.DitaContype = "TBS"

	fmt.Printf("--> MType<%s>\n--> XmlInfo<%s> \n--> DitaInfo<%s> \n",
		pBA.MType, pBA.XmlInfo, pBA.DitaInfo)

	return pBA, nil
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

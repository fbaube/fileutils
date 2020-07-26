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
func AnalyseFile(sCont string, filext string) (*BasicAnalysis, error) {

	var pBA *BasicAnalysis
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

	pBA = NewBasicAnalysis()
	pBA.FileExt = filext
	pBA.MimeType = httpContype

	// =======================================
	//  Quick check for top-level XML stuff
	// =======================================
	preamble, doctype, rootTag, dtdStuff, e := XM.Peek_xml(sCont)
	var gotRootTag bool
	gotRootTag = (rootTag.Name.Local != "")

	fmt.Printf("--> xm.peek: preamble<%s> doctype<%s> DTDstuff<%s> RootTag<%s> \n",
		SU.Yn(preamble != ""), SU.Yn(doctype != ""),
		SU.Yn(dtdStuff), SU.Yn(rootTag.Name.Local != ""))
	if rootTag.Name.Local == "" && !dtdStuff && (preamble != "" || doctype != "") {
		println("-->", "XML file has no root tag")
	}
	if dtdStuff && SU.IsInSliceIgnoreCase(filext, XM.DTDtypeFileExtensions) {
		println("--> DTD type detected (filext<%s>)", filext)
		pBA.MimeType = "application/xml-dtd"
		pBA.MType = "xml/sch/" + filext[1:]
		return pBA, nil
	}

	var isXml bool
	// Note that this check means that if an error was encountered,
	// content that might be XML will instead be processed as non-XML.
	isXml = (e == nil && (preamble != "" || doctype != "" || rootTag.Name.Local != ""))
	// We can do a bit more checking for XML
	if !isXml {
		if S.Contains(httpContype, "xml") {
			println("--> Peek_xml oops: HTTP contype-detection got XML (in MIME type)")
			isXml = true
		}
		if httpContype == "text/html" {
			println("--> Peek_xml oops: HTTP contype-detection got XML (text/html) (HDITA?)")
			isXml = true
		}
		if S.HasPrefix(httpContype, "text/") &&
			(filext == ".dita" || filext == ".ditamap" || filext == ".map") {
			println("--> Peek_xml oops: HTTP contype-detection got XML (text/dita-filext)")
			isXml = true
		}
		if S.Contains(httpContype, "ml") {
			println("--> Peek_xml oops: HTTP contype-detection got <ml>")
			isXml = true
		}
		if S.Contains(httpContype, "svg") {
			println("--> Peek_xml oops: HTTP contype-detection got <svg>")
			isXml = true
		}
	}

	// ===============
	//   NOT XML !!!
	// ===============
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

	if preamble != "" {
		// println("preamble:", preamble)
		pPRF, e = XM.NewXmlPreambleFields(preamble)
		if e != nil {
			println("xm.peek: preamble failure")
			return nil, fmt.Errorf("xm.peek: preamble failure: %w", e)
		}
		print("--> Parsed XML preamble: " + pPRF.String())
	}

	mt := "none"
	dtmt := "none"
	// isLwdita := false

	// If we don't have a DOCTYPE, we have trouble!
	if doctype == "" {
		if gotRootTag {
			var rutag string
			rutag = S.ToLower(rootTag.Name.Local)
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
			println("MType guessing, XML, no Doctype:", rutag, filext)
			pBA.MType = "xml/???/" + rutag
		}
	} else {
		println("-->", doctype)
		mt, _ = XM.GetMTypeByDoctype(doctype)
		// println("-->", "Doctype/MType search results:", mt)

		// If we got an MType, we don't really need to make this call,
		// but for now let's do it anyways.
		pDTF, e = XM.NewXmlDoctypeFieldsInclMType(doctype)
		if e != nil {
			println("xm.peek: doctype failure")
		}
		println("-->", "Parsed doctype:", pDTF.String())
		dtmt = pDTF.DoctypeMType

		if pDTF.TopTag != "" && rootTag.Name.Local != "" &&
			pDTF.TopTag != rootTag.Name.Local {
			fmt.Printf("--> RootTag MISMATCH: doctype<%s> bodytext<%s> \n",
				pDTF.TopTag, rootTag.Name.Local)
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
	pBA.XmlInfo.XmlDoctype = XM.XmlDoctype("DOCTYPE " + doctype)
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

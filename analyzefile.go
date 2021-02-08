package fileutils

import (
	"fmt"
	"net/http"
	FP "path/filepath"
	S "strings"

	SU "github.com/fbaube/stringutils"
	XM "github.com/fbaube/xmlmodels"
)

// <!ELEMENT map     (topicmeta?, (topicref | keydef)*)  >
// <!ELEMENT topicmeta (navtitle?, linktext?, data*) >

// AnalyseFile has drastically different handling for XML content versus
// non-XML content. Most of the function is mkaing several checks for the
// presence of XML. For XML files we have much more info available, so
// processing is both simpler and more complicated.
//
// The second argument "filext" can be any filepath; the Go stdlib is used
// to split off the file extension. It can also be "", if (for example) the
// content is entered interactively, without a file name or assigned MIME type.
//
// If the first argument "sCont" (the content) is zero-length, return (nil, nil).
//
func AnalyseFile(sCont string, filext string) (*XM.AnalysisRecord, error) {

	// pCntpg is ContypingInfo is FileExt MimeType MType Doctype IsLwDita IsProcbl
	var pCntpg *XM.ContypingInfo
	// pAnlRec is AnalysisRecord is basicly all analysis results, incl. ContypingInfo
	var pAnlRec *XM.AnalysisRecord

	if sCont == "" {
		println("==>", "Cannot analyze zero-length content")
		return nil, nil
	}
	// A trailing dot in the filename provides no filetype info.
	filext = FP.Ext(filext)
	if filext == "." {
		filext = ""
	}
	fmt.Printf("==> Analysing file: len<%d> filext<%s> \n", len(sCont), filext)

	// =======================================
	//  ANALYSIS #1: Use the stdlib
	//  HTTP content detection library
	// =======================================
	// Note that it assigns "text/html" to DITA maps :-/
	var httpContype string
	httpContype = http.DetectContentType([]byte(sCont))
	httpContype = S.TrimSuffix(httpContype, "; charset=utf-8")
	println("-->", "Contype acrdg to HTTP stdlib:", httpContype)
	htCntpIsXml, htCntpMsg := HttpContypeIsXml(httpContype, filext)

	// Preliminaries for deeper analysis
	pCntpg = new(XM.ContypingInfo)
	pAnlRec = new(XM.AnalysisRecord)
	pCntpg.FileExt = filext
	pCntpg.MimeType = httpContype
	// pAnlRec.MType = ""

	// =======================================
	//  ANALYSIS #2: Check for an XML preamble
	// =======================================
	hasXmlPre := S.HasPrefix(sCont, "<?xml ")

	// =======================================
	//  ANALYSIS #3: Quick check for root
	//  tag and other top-level XML stuff
	// =======================================
	var Peek *XM.XmlStructurePeek
	// Peek also sets KeyElms (Root,Meta,Text)
	Peek = XM.PeekAtStructure_xml(sCont)
	if Peek.HasError() {
		return nil, fmt.Errorf("fu.peekXml failed: %w", Peek.GetError())
	}
	// =======================================
	//  Now ANALYSE the analyses.
	// =======================================
	foundRootElm := Peek.KeyElms.CheckXmlSections()
	gotDoctype := (Peek.Doctype != "")
	gotXml := foundRootElm || Peek.HasDTDstuff || gotDoctype || hasXmlPre
	if !gotXml {
		println("--> Does not seem to be XML")
		if htCntpIsXml {
			gotXml = true
			println("--> BUT seems to be XML after all, based on file ext'n and HTTP stdlib:", htCntpMsg)
		}
	}
	if gotXml {
		fmt.Printf("--> is XML: preamble<%s> doctype<%s> RootTag<%s> DTDstuff<%s> \n",
			SU.Yn(hasXmlPre), SU.Yn(gotDoctype), SU.Yn(foundRootElm), SU.Yn(Peek.HasDTDstuff))
		if !(foundRootElm || Peek.HasDTDstuff) {
			println("--> Warning! XML file has no root tag (and is not DTD)")
		}
	}
	// =======================================
	//  If it's DTD stuff, we're done.
	// =======================================
	if Peek.HasDTDstuff && SU.IsInSliceIgnoreCase(filext, XM.DTDtypeFileExtensions) {
		fmt.Printf("--> DTD content detected (and filext<%s>) \n", filext)
		pAnlRec.MimeType = "application/xml-dtd"
		pAnlRec.MType = "xml/sch/" + filext[1:]
		// Could allocate and fill field XmlInfo
		return pAnlRec, nil
	}
	// ==============================
	//  If it's not XML, we're done.
	// ==============================
	if !gotXml {
		var nonXmlCntpg *XM.ContypingInfo
		nonXmlCntpg = DoContentTypes_non_xml(httpContype, sCont, filext)
		fmt.Printf("==> NON-XML: %s \n", nonXmlCntpg)
		pAnlRec.ContypingInfo = *nonXmlCntpg
		return pAnlRec, nil
	}
	// ==============
	//  YES IT'S XML
	// ==============
	// var isLwDita bool
	var pPRF *XM.XmlPreambleFields
	var e error
	if hasXmlPre {
		// println("preamble:", preamble)
		pPRF, e = XM.NewXmlPreambleFields(Peek.Preamble)
		if e != nil {
			println("xm.peek: preamble failure")
			return nil, fmt.Errorf("xm.peek: preamble failure: %w", e)
		}
		// print("--> Parsed XML preamble: " + pPRF.String())
	}
	// ================================
	//  Time to do some heavy lifting.
	// ================================
	println("==> Now split the file")
	pAnlRec.KeyElms = Peek.KeyElms
	pAnlRec.MakeContentitySections(sCont)
	fmt.Printf("--> meta pos<%d>len<%d> text pos<%d>len<%d> \n",
		pAnlRec.MetaElm.BegPos.Pos, len(pAnlRec.Meta_raw),
		pAnlRec.TextElm.BegPos.Pos, len(pAnlRec.Text_raw))
	if !Peek.IsSplittable() {
		println("--> Can't split into meta and text")
	}
	// =================================
	//  If we have DOCTYPE,
	//  it is gospel (and we are done).
	// =================================
	if Peek.Doctype != "" {
		// We are here if we got a DOCTYPE; we also have a file extension,
		// and we should have a root tag (or else the DOCTYPE makes no sense !)
		var pXDTF *XM.XmlDoctypeFields
		pXDTF = pCntpg.AnalyzeXmlDoctype(Peek.Doctype)
		if pXDTF.HasError() {
			panic("FIXME:" + pXDTF.Error())
		}
		println("--> Contyping: " + pCntpg.String())
		println("--> DTDfields: " + pXDTF.String())

		// What does AnalysisRecord need from Contyping and DoctypeFields ?
		pAnlRec.ContypingInfo = pXDTF.ContypingInfo
		pAnlRec.XmlDoctypeFields = pXDTF

		return pAnlRec, nil
	}
	// =====================
	//  No DOCTYPE. Bummer.
	// =====================
	if !foundRootElm {
		fmt.Printf("==> Got no root tag; filext: %s \n", filext)
		return pAnlRec, nil
	}
	// ==========================================
	//  Let's at least try to set the MType.
	//  We have a root tag and a file extension.
	// ==========================================
	rutag := S.ToLower(Peek.RootElm.Name)
	fmt.Printf("Guessing for: roottag<%s> filext<%s> ?mtype<%s> \n",
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
	pAnlRec.ContypingInfo = *pCntpg
	if pAnlRec.MType == "-/-/-" {
		pAnlRec.MType = "xml/???/" + rutag
	}
	// At this point, mt should be valid !
	println("--> Contyping (derived both ways):",
		pAnlRec.ContypingInfo.String())

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
	// Redundant!
	// pAnlRec.XmlDoctype = XM.XmlDoctype("DOCTYPE " + Peek.Doctype)
	// ?? pAnlRec.DoctypeFields = pDF
	if pPRF != nil {
		pAnlRec.XmlPreambleFields = pPRF
	} else {
		// SKIP
		// pBA.XmlPreambleFields = XM.STD_PreambleFields
	}
	// pBA.DoctypeIsDeclared  =  true
	pAnlRec.DitaMarkupLg = "TBS"
	pAnlRec.DitaContype = "TBS"

	println("D=> fu.af: Would have set up XmlInfo...")
	// pAnlRec.XmlInfo = *new(XM.XmlInfo)
	/* Fields to set:
			type XmlInfo struct {
				XmlContype // "Unknown", "DTD", "DTDmod", "DTDent", "RootTagData",
	        "RootTagMixedContent", "MultipleRootTags", "INVALID"
				*XmlPreambleFields
				XmlDoctype // type string // this is probly unnecessary
				// XmlDoctypeFields is a ptr - nil if there is no DOCTYPE declaration.
				*DoctypeFields
	*/
	fmt.Printf("--> fu.af: \n--> 1) MType: %s \n--> 2) XmlInfo: cntp:%s prmbl:%s DT:%s \n--> 3) DitaInfo: ML:%s cntp:%s \n",
		pAnlRec.MType, pAnlRec.XmlContype, pAnlRec.XmlPreambleFields,
		pAnlRec.XmlDoctypeFields, pAnlRec.DitaMarkupLg, pAnlRec.DitaContype)
	println("--> fu.af: Meta_raw:", pAnlRec.Meta_raw)
	println("--> fu.af: Text_raw:", pAnlRec.Text_raw)

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

func HttpContypeIsXml(httpContype string, filext string) (isXml bool, msg string) {

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

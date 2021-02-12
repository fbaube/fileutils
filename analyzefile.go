package fileutils

import (
	"fmt"
	"net/http"
	FP "path/filepath"
	S "strings"

	SU "github.com/fbaube/stringutils"
	XM "github.com/fbaube/xmlmodels"
)

// <!ELEMENT  map     (topicmeta?, (topicref | keydef)*)  >
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
// The return value is an XM.AnalysisRecord, which has a BUNCH of fields.
//
func AnalyseFile(sCont string, filext string) (*XM.AnalysisRecord, error) {

	// pCntpg is ContypingInfo is FileExt MimeType MType Doctype IsLwDita IsProcbl
	var pCntpg *XM.ContypingInfo
	// pAnlRec is AnalysisRecord is basicly all analysis results, incl. ContypingInfo
	var pAnlRec *XM.AnalysisRecord
	pAnlRec = new(XM.AnalysisRecord)

	if sCont == "" {
		println("==>", "Cannot analyze zero-length content")
		return nil, nil
	}
	// A trailing dot in the filename provides no filetype info.
	filext = FP.Ext(filext)
	if filext == "." {
		filext = ""
	}
	fmt.Printf("--> Analysing file: len<%d> filext<%s> \n", len(sCont), filext)

	// ========================================
	//  MAIN PRELIMINARY ANALYSIS: Check for
	//  root tag and other top-level XML stuff
	// ========================================
	var peek *XM.XmlStructurePeek
	// Peek also sets KeyElms (Root,Meta,Text)
	peek = XM.PeekAtStructure_xml(sCont)
	// NOTE! An error from oeeking might be from, for example, applying XML
	// parsing to a binary file. So, an error should not be fatal.
	var xmlParsingFailed bool
	if peek.HasError() {
		// return nil, fmt.Errorf("fu.peekXml failed: %w", Peek.GetError())
		println("--> XML parsing got error:", peek.GetError())
		xmlParsingFailed = true
	}
	// =======================================
	//  If it's DTD stuff, we're done.
	// =======================================
	if peek.HasDTDstuff && SU.IsInSliceIgnoreCase(filext, XM.DTDtypeFileExtensions) {
		fmt.Printf("--> DTD content detected (& filext<%s>) \n", filext)
		pAnlRec.MimeType = "application/xml-dtd"
		pAnlRec.MType = "xml/sch/" + filext[1:]
		// Could allocate and fill field XmlInfo
		return pAnlRec, nil
	}
	// ===============================
	//  Set variables, including
	//  supporting analysis by stdlib
	// ===============================
	gotRootElm := (peek.KeyElms.CheckXmlSections())
	gotDoctype := (peek.Doctype != "")
	gotPreambl := (peek.Preamble != "")
	gotSomeXml := (gotRootElm || gotDoctype || gotPreambl)
	// Note that stdlib assigns "text/html" to DITA maps :-/
	var httpContype string
	httpContype = http.DetectContentType([]byte(sCont))
	httpContype = S.TrimSuffix(httpContype, "; charset=utf-8")
	println("-->", "Contype acrdg to HTTP stdlib:", httpContype)
	htCntpIsXml, htCntpMsg := HttpContypeIsXml(httpContype, filext)

	// ==============================
	//  If it's not XML, we're done.
	// ==============================
	if xmlParsingFailed || !gotSomeXml {
		pAnlRec.ContypingInfo = *DoContentTypes_non_xml(httpContype, sCont, filext)
		fmt.Printf("==> NON-XML: %s \n", pAnlRec.ContypingInfo)
		println("!!> Fix the content extents")
		return pAnlRec, nil
	}

	// ======================================
	//  Handle a possible pathological case.
	// ======================================
	if xmlParsingFailed {
		println("--> Does not seem to be XML")
		if htCntpIsXml {
			println("--> Although HTTP stdlib seems to think it is:", htCntpMsg)
		}
	}

	// =========================================
	//  So from this point onward, WE HAVE XML.
	// =========================================
	var sP, sD, sR, sDtd string
	if gotPreambl {
		sP = "<?xml..> "
	}
	if gotDoctype {
		sD = "<!DOCTYPE..> "
	}
	if gotRootElm {
		sR = "<rootTag> "
	}
	if peek.HasDTDstuff {
		sDtd = "<!DTD stuff> "
	}
	fmt.Printf("--> IS-XML: %s%s%s%s \n", sP, sD, sR, sDtd)
	if !(gotRootElm || peek.HasDTDstuff) {
		println("--> WARNING! XML file has no root tag (and is not DTD)")
	}

	// Preliminaries for deeper analysis
	pCntpg = new(XM.ContypingInfo)
	pCntpg.FileExt = filext
	pCntpg.MimeType = httpContype
	var e error
	// pAnlRec.MType = ""
	// var isLwDita bool

	var pPRF *XM.XmlPreambleFields
	if gotPreambl {
		// println("preamble:", preamble)
		pPRF, e = XM.NewXmlPreambleFields(peek.Preamble)
		if e != nil {
			println("xm.peek: preamble failure in:", peek.Preamble)
			return nil, fmt.Errorf("xm<>>e<> preamble failure: %w", e)
		}
		// print("--> Parsed XML preamble: " + pPRF.String())
	}
	// ================================
	//  Time to do some heavy lifting.
	// ================================
	println("==> Now split the file")
	pAnlRec.KeyElms = peek.KeyElms
	pAnlRec.MakeXmlContentitySections(sCont)
	fmt.Printf("--> meta pos<%d>len<%d> text pos<%d>len<%d> \n",
		pAnlRec.MetaElm.BegPos.Pos, len(pAnlRec.Meta_raw),
		pAnlRec.TextElm.BegPos.Pos, len(pAnlRec.Text_raw))
	if !peek.IsSplittable() {
		println("--> Can't split into meta and text")
	}
	// =================================
	//  If we have DOCTYPE,
	//  it is gospel (and we are done).
	// =================================
	if peek.Doctype != "" {
		// We are here if we got a DOCTYPE; we also have a file extension,
		// and we should have a root tag (or else the DOCTYPE makes no sense !)
		var pXDTF *XM.XmlDoctypeFields
		pXDTF = pCntpg.AnalyzeXmlDoctype(peek.Doctype)
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
	if !gotRootElm {
		return pAnlRec, fmt.Errorf("Got no XML root tag in file with ext <%s>", filext)
	}
	// ==========================================
	//  Let's at least try to set the MType.
	//  We have a root tag and a file extension.
	// ==========================================
	rutag := S.ToLower(peek.RootElm.Name)
	fmt.Printf("Guessing XML typing for: roottag<%s> filext<%s> ?mtype<%s> \n",
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

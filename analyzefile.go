package fileutils

import (
	"errors"
	"fmt"
	"net/http"
	FP "path/filepath"
	S "strings"

	"github.com/gabriel-vasile/mimetype"
	"golang.org/x/tools/godoc/util"

	L "github.com/fbaube/mlog"
	SU "github.com/fbaube/stringutils"
	XM "github.com/fbaube/xmlmodels"
)

// <!ELEMENT  map     (topicmeta?, (topicref | keydef)*)  >
// <!ELEMENT topicmeta (navtitle?, linktext?, data*) >

// AnalyseFile has drastically different handling for XML content versus
// non-XML content. Most of the function is making several checks for the
// presence of XML. When a file is identified as XML, we have much more
// info available, so processing becomes both simpler and more complicated.
//
// Binary content is tagged as such and no further examination is done.
// So, the basic top-level classificaton of content is:
// (1) Binary
// (2) XML (when DOCTYPE is detected)
// (3) Everything else (incl. plain text, Markdown, and XML/HTML that lacks DOCTYPE)
//
// The second argument "filext" can be any filepath; the Go stdlib is used
// to split off the file extension. It can also be "", if (for example) the
// content is entered interactively, without a file name or assigned MIME type.
//
// If the first argument "sCont" (the content) is less than six bytes, return (nil, nil).
//
// The return value is an XM.AnalysisRecord, which has a BUNCH of fields.
//
func AnalyseFile(sCont string, filext string) (*XM.AnalysisRecord, error) {

	// ===========================
	//  Handle pathological cases
	// ===========================
	if sCont == "" {
		L.L.Warning("Cannot analyze zero-length content")
		return nil, errors.New("cannot analyze zero-length content")
	}
	if len(sCont) < 6 {
		L.L.Warning("Content is too short (%d bytes) to analyse", len(sCont))
		return nil, errors.New(fmt.Sprintf("content is too short (%d bytes) to analyse", len(sCont)))
	}
	// ===================
	//  Prepare variables
	// ===================
	// pAnlRec is AnalysisRecord is basicly all of our analysis results, incl. ContypingInfo
	var pAnlRec *XM.AnalysisRecord
	// pCntpg is ContypingInfo is all of: FileExt MimeType MType Doctype IsLwDita IsProcbl
	var pCntpg *XM.ContypingInfo
	pAnlRec = new(XM.AnalysisRecord)
	// A trailing dot in the filename provides no filetype info.
	filext = FP.Ext(filext)
	if filext == "." {
		filext = ""
	}
	L.L.Dbg("Analysing file: len<%d> filext<%s>", len(sCont), filext)
	// ========================
	//  Try a coupla shortcuts
	// ========================
	cheatYaml := S.HasPrefix(sCont, "---\n")
	cheatXml := S.HasPrefix(sCont, "<?xml")
	// ========================
	//  Content type detection
	// ========================
	var httpContype string
	var mimeLibTree *mimetype.MIME
	var mimeLibTreeS string // *mimetype.MIME
	var mimeLibIsBinary, stdUtilIsBinary, isBinary bool
	// MIME type ?
	httpContype = http.DetectContentType([]byte(sCont))
	mimeLibTree = mimetype.Detect([]byte(sCont))
	mimeLibTreeS = mimeLibTree.String()
	// Binary ?
	stdUtilIsBinary = !util.IsText([]byte(sCont))
	mimeLibIsBinary = true
	for mime := mimeLibTree; mime != nil; mime = mime.Parent() {
		if mime.Is("text/plain") {
			mimeLibIsBinary = false
		}
	}
	httpContype = S.TrimSuffix(httpContype, "; charset=utf-8")
	mimeLibTreeS = S.TrimSuffix(mimeLibTreeS, "; charset=utf-8")
	var sMime string
	if httpContype == mimeLibTreeS {
		sMime = httpContype
	} else {
		sMime = httpContype + "/" + mimeLibTreeS
	}
	L.L.Info("filext<%s> snift-MIME-type: %s", filext, sMime)

	// ===========================
	//  Check for & handle BINARY
	// ===========================
	isBinary = mimeLibIsBinary
	if stdUtilIsBinary != mimeLibIsBinary {
		L.L.Warning("MIME disagreement re is-binary: http-stdlib<%t> mime-lib<%t>",
			stdUtilIsBinary, mimeLibIsBinary)
	}
	if isBinary {
		if cheatYaml || cheatXml {
			L.L.Panic("analyzefile: binary + yaml/xml")
		}
		// For BINARY we won't ourselves do any more processing, so we can
		// basically trust that the sniffed MIME type is sufficient, and return.
		pAnlRec.MimeType = sMime
		pAnlRec.MType = "bin/"
		L.L.Dbg("BINARY!")
		if S.HasPrefix(sMime, "image/") {
			hasEPS := S.Contains(sMime, "eps")
			hasTXT := S.Contains(sMime, "text") || S.Contains(sMime, "txt")
			if hasTXT || hasEPS {
				// TODO
				L.L.Warning("EPS/TXT confusion for MIME type: " + sMime)
				pAnlRec.MType = "txt/img/??!"
			}
			return pAnlRec, nil
		}
	}
	// =====================
	//  Quick check for XML
	//  based on MIME type
	// =====================
	hIsXml, hMsg := HttpContypeIsXml("http-stdlib", httpContype, filext)
	mIsXml, mMsg := HttpContypeIsXml("3p-mime-lib", mimeLibTreeS, filext)
	if hIsXml || mIsXml {
		L.L.Info("(isXml:%t) %s", hIsXml, hMsg)
		L.L.Info("(isXml:%t) %s", mIsXml, mMsg)
	} else {
		L.L.Info("XML not detected yet")
	}

	// ===================================
	//  MAIN XML PRELIMINARY ANALYSIS:
	//  Peek into file to look for root
	//  tag and other top-level XML stuff
	// ===================================

	// Peek also sets KeyElms (Root,Meta,Text)
	var peek *XM.XmlStructurePeek
	peek = XM.PeekAtStructure_xml(sCont)
	// NOTE! An error from peeking might be from, for
	// example, applying XML parsing to a Markdown file.
	// So, an error should not be fatal.
	var xmlParsingFailed bool
	if peek.HasError() {
		L.L.Info("XML parsing got error: " + peek.GetError().Error())
		xmlParsingFailed = true
	}
	// ===============================
	//  If it's DTD stuff, we're done
	// ===============================
	if peek.HasDTDstuff && SU.IsInSliceIgnoreCase(filext, XM.DTDtypeFileExtensions) {
		L.L.Okay("DTD content detected (& filext<%s>)", filext)
		pAnlRec.MimeType = "application/xml-dtd"
		pAnlRec.MType = "xml/sch/" + S.ToLower(filext[1:])
		L.L.Warning("DDT stuff: should allocate and fill field XmlInfo")
		return pAnlRec, nil
	}
	// ===============================
	//  Set bool variables, including
	//  supporting analysis by stdlib
	// ===============================
	gotRootElm := (peek.ContentityStructure.CheckXmlSections())
	gotDoctype := (peek.Doctype != "")
	gotPreambl := (peek.Preamble != "")
	gotSomeXml := (gotRootElm || gotDoctype || gotPreambl)
	// Note that stdlib assigns "text/html" to DITA maps :-/

	// ================================
	//  So if it's not XML, we're done
	// ================================
	if xmlParsingFailed || !gotSomeXml {
		if cheatXml {
			L.L.Panic("analyzefile: non-xml + xml")
		}
		pAnlRec.ContypingInfo = *DoContypingInfo_non_xml(httpContype, sCont, filext)
		L.L.Okay("Non-XML: " + pAnlRec.ContypingInfo.String())
		// Check for YAMNL metadata
		iEnd, e := SU.YamlMetadataHeaderRange(sCont)
		// if there is an error, it will mess up parsing the file, so just exit.
		if e != nil {
			L.L.Error("Metadata header YAML error: " + e.Error())
			return pAnlRec, fmt.Errorf("metadata header YAML error: %w", e)
		}
		// Default: no YAML metadata found
		pAnlRec.Text.Beg = *XM.NewFilePosition(0)
		pAnlRec.Text.End = *XM.NewFilePosition(len(sCont))
		pAnlRec.Meta.Beg = *XM.NewFilePosition(0)
		pAnlRec.Meta.End = *XM.NewFilePosition(0)
		// No YAML metadata found ?
		if iEnd <= 0 {
			pAnlRec.Meta.Beg = *XM.NewFilePosition(0)
			pAnlRec.Meta.End = *XM.NewFilePosition(0)
			pAnlRec.Text.Beg = *XM.NewFilePosition(0)
			pAnlRec.Text.End = *XM.NewFilePosition(len(sCont))
		} else {
			// Found YAML metadata
			s2 := SU.TrimYamlMetadataDelimiters(sCont[:iEnd])
			ps, e := SU.GetYamlMetadataAsPropSet(s2)
			if e != nil {
				L.L.Error("loading YAML: " + e.Error())
				return pAnlRec, fmt.Errorf("loading YAML: %w", e)
			}
			// SUCCESS! Set ranges.
			pAnlRec.Meta.Beg = *XM.NewFilePosition(0)
			pAnlRec.Meta.End = *XM.NewFilePosition(iEnd)
			pAnlRec.Text.Beg = *XM.NewFilePosition(iEnd)
			pAnlRec.Text.End = *XM.NewFilePosition(len(sCont))

			pAnlRec.MetaProps = ps
		}
		L.L.Dbg("|RAW|" + pAnlRec.Raw + "|END|")
		return pAnlRec, nil
	}

	// ======================================
	//  Handle a possible pathological case.
	// ======================================
	if xmlParsingFailed {
		L.L.Dbg("Does not seem to be XML")
		if hIsXml {
			L.L.Dbg("Although http-stdlib seems to think it is:", hMsg)
		}
		if mIsXml {
			L.L.Dbg("Although 3p-mime-lib seems to think it is:", mMsg)
		}
	}

	// =========================================
	//  So from this point onward, WE HAVE XML.
	// =========================================
	if cheatYaml {
		L.L.Panic("analyzefile: xml + yaml")
	}
	var sP, sD, sR, sDtd string
	if gotPreambl {
		sP = "<?xml..> "
	}
	if gotDoctype {
		sD = "<!DOCTYPE..> "
	}
	if gotRootElm {
		sR = "root<" + peek.Root.TagName + "> "
	}
	if peek.HasDTDstuff {
		sDtd = "<!DTD stuff> "
	}
	L.L.Okay("Is XML: found: %s%s%s%s", sP, sD, sR, sDtd)
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
	L.L.Progress("Now split the file")
	pAnlRec.ContentityStructure = peek.ContentityStructure
	pAnlRec.MakeXmlContentitySections(sCont)
	/*
		fmt.Printf("--> meta pos<%d>len<%d> text pos<%d>len<%d> \n",
			pAnlRec.Meta.Beg.Pos, len(pAnlRec.MetaRaw()),
			pAnlRec.Text.Beg.Pos, len(pAnlRec.TextRaw()))
		/*
			if !peek.IsSplittable() {
				println("--> Can't split into meta and text")
			}
	*/
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
		L.L.Dbg("Contyping: " + pCntpg.String())
		L.L.Dbg("DTDfields: " + pXDTF.String())

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
	rutag := S.ToLower(peek.Root.TagName)
	L.L.Info("XML without DOCTYPE: root<%s> filext<%s> MType<%s>",
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
	L.L.Dbg("Contyping: " + pAnlRec.ContypingInfo.String())

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
	pAnlRec.DitaFlavor = "TBS"
	pAnlRec.DitaContype = "TBS"

	L.L.Warning("fu.af: TODO set more XML info")
	// pAnlRec.XmlInfo = *new(XM.XmlInfo)

	L.L.Info("fu.af: MType<%s> xcntp<%s> ditaFlav<%s> ditaCntp<%s> DT<%s>",
		pAnlRec.MType, pAnlRec.XmlContype, // pAnlRec.XmlPreambleFields,
		pAnlRec.DitaFlavor, pAnlRec.DitaContype, pAnlRec.XmlDoctypeFields)
	// println("--> fu.af: MetaRaw:", pAnlRec.MetaRaw())
	// println("--> fu.af: TextRaw:", pAnlRec.TextRaw())

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

func HttpContypeIsXml(src, sContype, filext string) (isXml bool, msg string) {
	src += " contype-detection "

	if S.Contains(sContype, "xml") {
		return true, src + "got XML (in MIME type)"
	}
	if sContype == "text/html" {
		return true, src + "got XML (text/html, HDITA?)"
	}
	if S.HasPrefix(sContype, "text/") &&
		(filext == ".dita" || filext == ".ditamap" || filext == ".map") {
		return true, src + "got XML (text/dita-filext)"
	}
	if S.Contains(sContype, "ml") {
		return true, src + "got <ml>"
	}
	if S.Contains(sContype, "svg") {
		return true, src + "got <svg>"
	}
	return false, ""
}

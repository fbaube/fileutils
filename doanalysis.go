package fileutils

import (
	// "errors"
	"fmt"
	"net/http"
	FP "path/filepath"
	S "strings"

	"github.com/gabriel-vasile/mimetype"
	"golang.org/x/tools/godoc/util" // used once, L125

	L "github.com/fbaube/mlog"
	SU "github.com/fbaube/stringutils"
	XU "github.com/fbaube/xmlutils"
)

// <!ELEMENT  map     (topicmeta?, (topicref | keydef)*)  >
// <!ELEMENT topicmeta (navtitle?, linktext?, data*) >

// DoAnalysis is called only by NewContentityRecord(..).
// It has very different handling for XML content versus non-XML content.
// Most of the function is making several checks for the presence of XML.
// When a file is identified as XML, we have much more info available,
// so processing becomes both simpler and more complicated.
//
// Binary content is tagged as such and no further examination is done.
// So, the basic top-level classificaton of content is:
// (1) Binary
// (2) XML (when a DOCTYPE is detected)
// (3) Everything else (incl. plain text,
// Markdown, and XML/HTML that lacks DOCTYPE)
//
// The second argument "filext" can be any filepath; the Go stdlib
// is used to split off the file extension. It can also be "", if
// (for example) the content is entered interactively, without a
// file name or already-determined MIME type.
//
// If the first argument "sCont" (the content) is less than six bytes,
// return (nil, nil) to indicate that there is not enough content to
// work with.
// .
func DoAnalysis(sCont string, filext string) (*XU.AnalysisRecord, error) {

	// A trailing dot in the filename provides no filetype info.
	filext = FP.Ext(filext)
	if filext == "." {
		filext = ""
	}
	// ===========================
	//  Handle pathological case:
	//  Too short or non-existent
	// ===========================
	if len(sCont) < 6 {
		if sCont == "" {
			L.L.Progress("DoAnalysis: skipping zero-length content")
		} else {
			L.L.Warning("DoAnalysis: content too short (%d bytes)", len(sCont))
		}
		p := new(XU.AnalysisRecord)
		p.FileExt = filext
		// return nil, errors.New(fmt.Sprintf("content is too short (%d bytes) to analyse", len(sCont)))
		return p, nil
	}
	L.L.Dbg("(AF) filext<%s> len<%d> beg<%s>",
		filext, len(sCont), sCont[:5])

	// ========================
	//  Try a coupla shortcuts
	// ========================
	cheatYaml := S.HasPrefix(sCont, "---\n")
	cheatXml := S.HasPrefix(sCont, "<?xml ")
	// ========================
	//  Content type detection:
	//   "h_" = http stdlib
	//   (unreliable!)
	// ========================
	var h_contype string
	h_contype = http.DetectContentType([]byte(sCont))
	h_contype = S.TrimSuffix(h_contype, "; charset=utf-8")
	if S.Contains(h_contype, ";") {
		L.L.Warning("Content type from http stdlib " +
			"(still) has a semicolon: " + h_contype)
	}
	// ========================
	//  Content type detection:
	//  "m_" = 3rd party library
	//  (AUTHORITATIVE!)
	// ========================
	var m_contypeData *mimetype.MIME
	var m_contype string // *mimetype.MIME
	// MIME type ?
	m_contypeData = mimetype.Detect([]byte(sCont))
	m_contype = m_contypeData.String()
	m_contype = S.TrimSuffix(m_contype, "; charset=utf-8")
	if S.Contains(m_contype, ";") {
		L.L.Warning("Content type from 3P lib " +
			"(still) has a semicolon: " + m_contype)
	}
	// Authoritative MIME type string:
	// var sAuthtvMime = m_contype
	if h_contype != m_contype {
		L.L.Warning("(AF) MIME type per libs: "+
			"OK-3P-mime<%s> v barfy-stdlib-http<%s>",
			m_contype, h_contype)
	}
	L.L.Info("(AF) filext<%s> has 3P-snift-MIME-type: %s", filext, m_contype)

	// =====================================
	// INITIALIZE ANALYSIS RECORD:
	// pAnlRec is *xmlutils.AnalysisRecord is
	// basically all of our analysis results,
	// including ContypingInfo.
	// Don't forget to set the content!
	// (omitting this caused a lot of bugs)
	// =====================================
	var pAnlRec *XU.AnalysisRecord
	pAnlRec = new(XU.AnalysisRecord)
	pAnlRec.ContentityStructure.Raw = sCont
	pAnlRec.FileExt = filext
	pAnlRec.MimeType = h_contype // Junk this ?
	pAnlRec.MimeTypeAsSnift = m_contype

	// ===========================
	//  Check for & handle BINARY
	// ===========================
	var h_isBinary bool // Unreliable!
	var m_isBinary bool // Authoritative!
	h_isBinary = !util.IsText([]byte(sCont))
	m_isBinary = true
	for mime := m_contypeData; mime != nil; mime = mime.Parent() {
		if mime.Is("text/plain") {
			m_isBinary = false
		}
		// FIXME If "text/" here, is an error to sort out
	}
	if h_isBinary != m_isBinary {
		L.L.Warning("(AF) is-binary: "+
			"using OK-3P-mime-lib <%t> and not barfy-http-stdlib <%t> ",
			m_isBinary, h_isBinary)
	}
	if m_isBinary {
		if cheatYaml || cheatXml {
			L.L.Panic("(AF) both binary & yaml/xml")
		}
		return DoAnalysis_bin(pAnlRec)
	}
	// ======================================
	// We have text, but it might not be XML.
	// So use two libraries to check for XML
	// based on MIME type.
	// ======================================
	hIsXml, hMsg := contypeIsXml("http", h_contype, filext)
	mIsXml, mMsg := contypeIsXml("mime", m_contype, filext)
	if hIsXml || mIsXml {
		var hS, mS string
		if !hIsXml {
			hS = "not "
		}
		if !mIsXml {
			mS = "not "
		}
		L.L.Info("(AF) 3P (is-%sXML) %s ; stdlib (is-%sXML) %s",
			mS, mMsg, hS, hMsg)
	} else {
		L.L.Info("(AF) XML not detected by MIME libs")
	}

	// ===================================
	//  MAIN XML PRELIMINARY ANALYSIS:
	//  Peek into file to look for root
	//  tag and other top-level XML stuff
	// ===================================

	var xmlParsingFailed bool
	var peek *XU.XmlStructurePeek
	var e error

	// Peek also sets KeyElms (Root,Meta,Text)
	peek, e = XU.PeekAtStructure_xml(sCont)
	// NOTE! An error from peeking might be from, for
	// example, applying XML parsing to a Markdown file.
	// So, an error should NOT be fatal.
	if e != nil {
		L.L.Info("(AF) XML parsing got error: " + e.Error())
		xmlParsingFailed = true
	}
	// pathological
	if xmlParsingFailed && (hIsXml || mIsXml) {
		L.L.Panic("(AF) XML confusion (case #1) in DoAnalysis")
	}
	// Note that this next test dusnt always work for Markdown!
	/*
		if (!xmlParsingFailed) && (! (hIsXml || mIsXml)) {
			L.L.Panic("XML confusion (case #2) in AnalyzeFile")
		}
	*/

	// ===============================
	//  If it's DTD stuff, we're done
	// ===============================
	if peek.HasDTDstuff && SU.IsInSliceIgnoreCase(
		filext, XU.DTDtypeFileExtensions) {
		return DoAnalysis_sch(pAnlRec)
	}
	// ===============================
	//  Set bool variables, including
	//  supporting analysis by stdlib
	// ===============================
	gotRootElm := (peek.ContentityStructure.CheckXmlSections())
	gotDoctype := (peek.RawDoctype != "")
	gotPreambl := (peek.RawPreamble != "")
	gotSomeXml := (gotRootElm || gotDoctype || gotPreambl)
	// Note that stdlib assigns "text/html" to DITA maps :-/

	// ================================
	//  So if it's not XML, we're done
	// ================================
	if xmlParsingFailed || !gotSomeXml {
		if cheatXml {
			L.L.Panic("(AF) both non-xml & xml")
		}
		return DoAnalysis_txt(pAnlRec)
	}

	// =====================================
	//  Handle possible pathological cases.
	// =====================================
	if xmlParsingFailed {
		L.L.Dbg("(AF) Does not seem to be XML")
		if hIsXml {
			L.L.Dbg("(AF) Although http-stdlib seems to think it is:", hMsg)
		}
		if mIsXml {
			L.L.Dbg("(AF) Although 3p-mime-lib seems to think it is:", mMsg)
		}
	}
	if cheatYaml {
		L.L.Panic("(AF) both xml & yaml")
	}

	// =========================================
	//  So from this point onward, WE HAVE XML.
	// =========================================
	var sP, sD, sR, sDtd string
	if gotPreambl {
		sP = "<?xml..> "
	}
	if gotDoctype {
		sD = "<DOCTYPE..> "
	}
	if gotRootElm {
		sR = "root<" + peek.Root.TagName + "> "
	}
	if peek.HasDTDstuff {
		sDtd = "<DTD stuff> "
	}
	L.L.Progress("Is XML: found: %s%s%s%s", sP, sD, sR, sDtd)
	if !(gotRootElm || peek.HasDTDstuff) {
		L.L.Warning("(AF) XML file has no root tag (and is not DTD)")
	}

	var pPRF *XU.ParsedPreamble
	if gotPreambl {
		L.L.Dbg("(AF) got: %s", peek.RawPreamble)
		pPRF, e = XU.ParsePreamble(peek.RawPreamble)
		if e != nil {
			// println("xm.peek: preamble failure in:", peek.RawPreamble)
			return nil, fmt.Errorf("(AF) preamble failure: %w", e)
		}
		// print("--> Parsed XML preamble: " + pPRF.String())
	}
	// ================================
	//  Time to do some heavy lifting.
	// ================================
	L.L.Progress("(AF) Now split the file")
	if pAnlRec.ContentityStructure.Raw == "" {
		L.L.Error("(AF) XML has nil Raw")
	}
	pAnlRec.ContentityStructure = peek.ContentityStructure
	pAnlRec.ContentityStructure.Raw = sCont // naybe redundant ?
	pAnlRec.MakeXmlContentitySections()
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
	if peek.RawDoctype != "" {
		// We are here if we got a DOCTYPE; we also have a file extension,
		// and we should have a root tag (or else the DOCTYPE makes no sense !)
		var pPDT *XU.ParsedDoctype
		pPDT, e = pAnlRec.ContypingInfo.ParseDoctype(peek.RawDoctype)
		if e != nil {
			L.L.Panic("FIXME:" + e.Error())
		}
		pAnlRec.ParsedDoctype = pPDT
		L.L.Dbg("(AF) gotDT: MType: " + pAnlRec.MType)
		L.L.Dbg("(AF) gotDT: AnalysisRecord: " + pAnlRec.String())
		// L.L.Dbg("gotDT: DctpFlds: " + pPDT.String())

		if pAnlRec.MType == "" {
			L.L.Panic("(AF) no MType, L362")
		}
		pAnlRec.ParsedDoctype = pPDT

		if pAnlRec.MType == "" {
			L.L.Panic("(AF) no MType, L367")
		}
		L.L.Okay("(AF) Success: XML with DOCTYPE")
		// HACK ALERT
		if S.HasSuffix(pAnlRec.MType, "---") {
			rutag := S.ToLower(peek.Root.TagName)
			if pAnlRec.MType == "xml/map/---" {
				pAnlRec.MType = "xml/map/" + rutag
				L.L.Okay("(AF) Patched MType to: " + pAnlRec.MType)
			} else {
				panic("MType ending in \"---\" not fixed")
			}
		}
		return pAnlRec, nil
	}
	// =====================
	//  No DOCTYPE. Bummer.
	// =====================
	if !gotRootElm {
		return pAnlRec, fmt.Errorf("(AF) Got no XML root tag in file with ext <%s>", filext)
	}
	// ==========================================
	//  Let's at least try to set the MType.
	//  We have a root tag and a file extension.
	// ==========================================
	rutag := S.ToLower(peek.Root.TagName)
	L.L.Progress("(AF) XML without DOCTYPE: <%s> root<%s> MType<%s>",
		filext, rutag, pAnlRec.MType)
	// Do some easy cases
	if rutag == "html" && (filext == ".html" || filext == ".htm") {
		pAnlRec.MType = "html/cnt/html5"
	} else if rutag == "html" && S.HasPrefix(filext, ".xht") {
		pAnlRec.MType = "html/cnt/xhtml"
	} else if SU.IsInSliceIgnoreCase(rutag, XU.DITArootElms) &&
		SU.IsInSliceIgnoreCase(filext, XU.DITAtypeFileExtensions) {
		pAnlRec.MType = "xml/cnt/" + rutag
		if rutag == "bookmap" && S.HasSuffix(filext, "map") {
			pAnlRec.MType = "xml/map/" + rutag
		}
	}
	// pAnlRec.ContypingInfo = *pCntpg
	if pAnlRec.MType == "-/-/-" {
		pAnlRec.MType = "xml/???/" + rutag
	}
	// At this point, mt should be valid !
	L.L.Dbg("(AF) Contyping: " + pAnlRec.ContypingInfo.String())

	// Now we should fill in all the detail fields.
	pAnlRec.XmlContype = "RootTagData"

	if pPRF != nil {
		pAnlRec.ParsedPreamble = pPRF
	} else {
		// SKIP
		// pBA.XmlPreambleFields = XU.STD_PreambleFields
	}
	// pBA.DoctypeIsDeclared  =  true
	pAnlRec.DitaFlavor = "TBS"
	pAnlRec.DitaContype = "TBS"

	// L.L.Info("fu.af: MType<%s> xcntp<%s> ditaFlav<%s> ditaCntp<%s> DT<%s>",
	L.L.Progress("(AF) final: MType<%s> xcntp<%s> dita:TBS DcTpFlds<%s>",
		pAnlRec.MType, pAnlRec.XmlContype, // pAnlRec.XmlPreambleFields,
		// pAnlRec.DitaFlavor, pAnlRec.DitaContype,
		pAnlRec.ParsedDoctype)
	// println("--> fu.af: MetaRaw:", pAnlRec.MetaRaw())
	// println("--> fu.af: TextRaw:", pAnlRec.TextRaw())

	if pAnlRec.MType == "" {
		switch m_contype {
		case "image/svg+xml":
			pAnlRec.MType = "xml/cnt/svg"
		}
		if pAnlRec.MType != "" {
			L.L.Warning("(AF) Lamishly hacked the MType to: %s", pAnlRec.MType)
		}
	}
	L.L.Okay("(AF) Success: XML without DOCTYPE")
	return pAnlRec, nil
}

func CollectKeysOfNonNilMapValues(M map[string]*XU.FilePosition) []string {
	var ss []string
	for K, V := range M {
		if V != nil {
			ss = append(ss, K)
		}
	}
	return ss
}

func contypeIsXml(src, sContype, filext string) (isXml bool, msg string) {
	src += " "
	if S.Contains(sContype, "xml") {
		return true, src + "got XML (in MIME type)"
	}
	if sContype == "text/html" {
		return true, src + "got XML (cos is text/html)"
	}
	if S.HasPrefix(sContype, "text/") &&
		(filext == ".dita" || filext == ".ditamap" || filext == ".map") {
		return true, src + "got DITA XML (cos of text/dita-filext)"
	}
	if S.Contains(sContype, "ml") {
		return true, src + "got <ml>"
	}
	if S.Contains(sContype, "svg") {
		return true, src + "got <svg>"
	}
	return false, ""
}

package fileutils

import (
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
// func DoAnalysis(sCont string, filext string) (*PathAnalysis, error) {
func NewPathAnalysis(pPP *PathProps) /* sCont string, filext string) */ (*PathAnalysis, error) {

	sCont := pPP.Raw
	filext := FP.Ext(pPP.AbsFP.S())

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
		p := new(PathAnalysis)
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
	// ==============================
	//  Content type detection using
	//   http stdlib (unreliable!)
	// ==============================
	var stdlib_contype string
	stdlib_contype = http.DetectContentType([]byte(sCont))
	stdlib_contype = S.TrimSuffix(stdlib_contype, "; charset=utf-8")
	if S.Contains(stdlib_contype, ";") {
		L.L.Warning("Content type from stdlib " +
			"(still) has a semicolon: " + stdlib_contype)
	}
	// ========================
	//  Content type detection
	//  using 3rd party library
	//  (AUTHORITATIVE!)
	// ========================
	var lib_3p_contypeData *mimetype.MIME
	var lib_3p_contype string // *mimetype.MIME
	// MIME type ?
	lib_3p_contypeData = mimetype.Detect([]byte(sCont))
	lib_3p_contype = lib_3p_contypeData.String()
	lib_3p_contype = S.TrimSuffix(lib_3p_contype, "; charset=utf-8")
	if S.Contains(lib_3p_contype, ";") {
		L.L.Warning("Content type from lib_3p " +
			"(still) has a semicolon: " + lib_3p_contype)
	}
	// Authoritative MIME type string:
	// var sAuthtvMime = lib_3p_contype
	if stdlib_contype != lib_3p_contype {
		L.L.Warning("(AF) MIME type: lib_3p<%s> stdlib<%s>",
			lib_3p_contype, stdlib_contype)
	}
	L.L.Info("(AF) filext<%s> has 3P-snift-MIME-type: %s",
		filext, lib_3p_contype)

	// =====================================
	// INITIALIZE ANALYSIS RECORD:
	// pAnlRec is *xmlutils.AnalysisRecord is
	// basically all of our analysis results,
	// including ContypingInfo.
	// Don't forget to set the content!
	// (omitting this caused a lot of bugs)
	// =====================================
	var pAnlRec *PathAnalysis
	pAnlRec = new(PathAnalysis)
	pAnlRec.PathProps = new(PathProps)
	pAnlRec.PathProps.Raw = sCont
	pAnlRec.FileExt = filext
	pAnlRec.MimeType = stdlib_contype // Junk this ?
	pAnlRec.MimeTypeAsSnift = lib_3p_contype

	// ===========================
	//  Check for & handle BINARY
	// ===========================
	var stdlib_isBinary bool // Unreliable!
	var lib_3p_isBinary bool // Authoritative!
	stdlib_isBinary = !util.IsText([]byte(sCont))
	lib_3p_isBinary = true
	for mime := lib_3p_contypeData; mime != nil; mime = mime.Parent() {
		if mime.Is("text/plain") {
			lib_3p_isBinary = false
		}
		// FIXME If "text/" here, is an error to sort out
	}
	if stdlib_isBinary != lib_3p_isBinary {
		L.L.Warning("(AF) is-binary: using lib_3p <%t> not stdlib <%t> ",
			lib_3p_isBinary, stdlib_isBinary)
	}
	if lib_3p_isBinary {
		if cheatYaml || cheatXml {
			L.L.Panic("(AF) both binary & yaml/xml")
		}
		return pAnlRec, pAnlRec.DoAnalysis_bin()
	}
	// ======================================
	// We have text, but it might not be XML.
	// So process the MIME types returned by
	// the two libraries.
	// ======================================
	hIsXml, hMsg := contypeIsXml("stdlib", stdlib_contype, filext)
	mIsXml, mMsg := contypeIsXml("lib_3p", lib_3p_contype, filext)
	if hIsXml || mIsXml {
		var hS, mS string
		if !hIsXml {
			hS = "not "
		}
		if !mIsXml {
			mS = "not "
		}
		L.L.Info("(AF) lib_3p (is-%sXML) %s \n\t\t stdlib (is-%sXML) %s",
			mS, mMsg, hS, hMsg)
	} else {
		L.L.Info("(AF) XML not detected by either MIME lib")
	}
	// ===================================
	//  MAIN XML PRELIMINARY ANALYSIS:
	//  Peek into file to look for root
	//  tag and other top-level XML stuff
	// ===================================
	var xmlParsingFailed bool
	var pPeek *XU.XmlPeek
	var e error

	// ==============================
	//  Peek for XML; this also sets
	//  KeyElms (Root,Meta,Text)
	// ==============================
	pPeek, e = XU.Peek_xml(sCont)
	/*
		if pPeek.Raw == "" {
			panic("nil Raw")
		}
		if pAnlRec.PathProps.Raw != pPeek.Raw {
			panic("MISMATCH-1")
		}
	*/
	// NOTE! An error from peeking might be caused
	// by, for example, applying XML parsing to a
	// Markdown file. So, an error is NOT fatal.
	if e != nil {
		L.L.Info("(AF) XML parsing got error: " + e.Error())
		xmlParsingFailed = true
	}
	// ===============================
	//  If it's DTD stuff, we're done
	// ===============================
	if pPeek.HasDTDstuff && SU.IsInSliceIgnoreCase(
		filext, XU.DTDtypeFileExtensions) {
		return pAnlRec, pAnlRec.DoAnalysis_sch()
	}
	// Check for pathological cases
	var gotSomeXml bool
	gotSomeXml = pPeek.ContentityBasics.HasRootTag() ||
		(pPeek.RawDoctype != "") || (pPeek.RawPreamble != "")
	if xmlParsingFailed && (hIsXml || mIsXml) {
		L.L.Panic("(AF) XML confusion (case #1) in DoAnalysis")
	}
	// Note that this next test dusnt always work for Markdown!
	/*
		if (!xmlParsingFailed) && (! (hIsXml || mIsXml)) {
			L.L.Panic("XML confusion (case #2) in AnalyzeFile")
		}
	*/
	// =============================
	//  If it's not XML, we're done
	// =============================
	if xmlParsingFailed || !gotSomeXml {
		if cheatXml {
			L.L.Panic("(AF) both non-xml & xml")
		}
		return pAnlRec, pAnlRec.DoAnalysis_txt()
	}
	// ===========================================
	//  It's XML, so crank thru it and we're done
	// ===========================================
	return pAnlRec, pAnlRec.DoAnalysis_xml(pPeek)
}

/*
	// ===============================
	//  Set bool variables, including
	//  supporting analysis by stdlib
	// ===============================
	gotRootElm := (pPeek.ContentityTopLevelStructure.HasRootTag())
	gotDoctype := (pPeek.RawDoctype != "")
	gotPreambl := (pPeek.RawPreamble != "")
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
		sR = "root<" + pPeek.Root.TagName + "> "
	}
	if pPeek.HasDTDstuff {
		sDtd = "<DTD stuff> "
	}
	L.L.Progress("Is XML: found: %s%s%s%s", sP, sD, sR, sDtd)
	if !(gotRootElm || pPeek.HasDTDstuff) {
		L.L.Warning("(AF) XML file has no root tag (and is not DTD)")
	}

	var pPRF *XU.ParsedPreamble
	if gotPreambl {
		L.L.Dbg("(AF) got: %s", pPeek.RawPreamble)
		pPRF, e = XU.ParsePreamble(pPeek.RawPreamble)
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
	if pAnlRec.ContentityTopLevelStructure.Raw == "" {
		L.L.Error("(AF) XML has nil Raw")
	}
	pAnlRec.ContentityTopLevelStructure = pPeek.ContentityTopLevelStructure
	pAnlRec.ContentityTopLevelStructure.Raw = sCont // naybe redundant ?
	pAnlRec.MakeXmlContentitySections()
	/*
		fmt.Printf("--> meta pos<%d>len<%d> text pos<%d>len<%d> \n",
			pAnlRec.Meta.Beg.Pos, len(pAnlRec.MetaRaw()),
			pAnlRec.Text.Beg.Pos, len(pAnlRec.TextRaw()))
		/*
			if !peek.IsSplittable() {
				println("--> Can't split into meta and text")
			}
	* /
	// =================================
	//  If we have DOCTYPE,
	//  it is gospel (and we are done).
	// =================================
	if gotDoctype {
		// We are here if we got a DOCTYPE; we also have a file extension,
		// and we should have a root tag (or else the DOCTYPE makes no sense !)
		var pPDT *XU.ParsedDoctype
		pPDT, e = pAnlRec.ContypingInfo.ParseDoctype(pPeek.RawDoctype)
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
			rutag := S.ToLower(pPeek.Root.TagName)
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
	rutag := S.ToLower(pPeek.Root.TagName)
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
		switch lib_3p_contype {
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
*/

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

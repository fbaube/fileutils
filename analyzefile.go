package fileutils

import (
	"errors"
	"fmt"
	"net/http"
	FP "path/filepath"
	S "strings"

	MT "github.com/gabriel-vasile/mimetype"
	"golang.org/x/tools/godoc/util"

	L "github.com/fbaube/mlog"
	SU "github.com/fbaube/stringutils"
	XU "github.com/fbaube/xmlutils"
)

// <!ELEMENT  map     (topicmeta?, (topicref | keydef)*)  >
// <!ELEMENT topicmeta (navtitle?, linktext?, data*) >

// AnalyseFile is called only by dbutils.NewContentityRecord(..).
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
// The second argument "filext" can be any filepath; the Go stdlib is used
// to split off the file extension. It can also be "", if (for example) the
// content is entered interactively, without a file name or assigned MIME type.
//
// If the first argument "sCont" (the content) is less than six bytes, 
// return (nil, nil).
//
func AnalyseFile(sCont string, filext string) (*XU.AnalysisRecord, error) {

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
	// A trailing dot in the filename provides no filetype info.
	filext = FP.Ext(filext)
	if filext == "." {
		filext = ""
	}
	L.L.Dbg("Analysing file: ext<%s> len<%d>", filext, len(sCont))

	// ========================
	//  Try a coupla shortcuts
	// ========================
	cheatYaml := S.HasPrefix(sCont, "---\n")
	cheatXml := S.HasPrefix(sCont, "<?xml")
	// ========================
	//  Content type detection
	// ========================
	var httpStdlibContype string
	var mimeLibDatContype *MT.MIME
	var mimeLibStrContype string // *mimetype.MIME
	var mimeLibIsBinary, stdUtilIsBinary, isBinary bool
	// MIME type ?
	httpStdlibContype = http.DetectContentType([]byte(sCont))
	mimeLibDatContype = MT.Detect([]byte(sCont))
	mimeLibStrContype = mimeLibDatContype.String()
	// Binary ?
	stdUtilIsBinary = !util.IsText([]byte(sCont))
	mimeLibIsBinary = true
	for mime := mimeLibDatContype; mime != nil; mime = mime.Parent() {
		if mime.Is("text/plain") {
			mimeLibIsBinary = false
		}
	}
	httpStdlibContype = S.TrimSuffix(httpStdlibContype, "; charset=utf-8")
	mimeLibStrContype = S.TrimSuffix(mimeLibStrContype, "; charset=utf-8")
	if S.Contains(httpStdlibContype, ";") || 
	   S.Contains(mimeLibStrContype, ";") {
		L.L.Warning("Content type contains a semicolon")
	}
	// Authoritative MIME type string 
	var sAuthtvMime string
	if httpStdlibContype == mimeLibStrContype {
		sAuthtvMime = httpStdlibContype
	} else {
		sAuthtvMime = httpStdlibContype + "|" + mimeLibStrContype
		L.L.Warning("MIME type disagreement in libs: %s", sAuthtvMime)
	}
	L.L.Info("filext<%s> snift-MIME-type: %s", filext, sAuthtvMime)

	// =====================================
	// pAnlRec is *xmlutils.AnalysisRecord is 
	// basically all of our analysis results, 
	// including ContypingInfo
	// Don't forget to set the content
	// (omitting this caused a lot of bugs)
	// =====================================
	var pAnlRec *XU.AnalysisRecord
	pAnlRec = new(XU.AnalysisRecord)
	pAnlRec.ContentityStructure.Raw = sCont
	pAnlRec.FileExt = filext 
	pAnlRec.MimeType = httpStdlibContype
	pAnlRec.MimeTypeAsSnift = sAuthtvMime

	// ===========================
	//  Check for & handle BINARY
	// ===========================
	isBinary = mimeLibIsBinary
	if stdUtilIsBinary != mimeLibIsBinary {
		L.L.Warning("MIME disagreement re is-binary: "+
			"http-stdlib<%t> mime-lib<%t>",
			stdUtilIsBinary, mimeLibIsBinary)
	}
	if isBinary {
		if cheatYaml || cheatXml {
			L.L.Panic("analyzefile: binary + yaml/xml")
		}
		// For BINARY we won't ourselves do any more processing, 
		// so we can basically trust that the sniffed MIME type 
		// is sufficient, and return.
		pAnlRec.MimeType = sAuthtvMime
		pAnlRec.MType = "bin/"
		L.L.Dbg("BINARY!")
		if S.HasPrefix(sAuthtvMime, "image/") {
			hasEPS := S.Contains(sAuthtvMime, "eps")
			hasTXT := S.Contains(sAuthtvMime, "text") || S.Contains(sAuthtvMime, "txt")
			if hasTXT || hasEPS {
				// TODO
				L.L.Warning("EPS/TXT confusion for MIME type: " + sAuthtvMime)
				pAnlRec.MType = "txt/img/??!"
			}
		}
		return pAnlRec, nil
	}
	// ======================================
	// We have text, but it might not be XML.
	// So use two libraries to check for XML 
	// based on MIME type. 
	// ======================================
	hIsXml, hMsg := HttpContypeIsXml("http-stdlib",
		httpStdlibContype, filext)
	mIsXml, mMsg := HttpContypeIsXml("3p-mime-lib",
		mimeLibStrContype, filext)
	if hIsXml || mIsXml {
		hS, mS := "is-", "is-"
		if !hIsXml {
			hS = "not"
		}
		if !mIsXml {
			mS = "not"
		}
		L.L.Info("(%sXML) %s", hS, hMsg)
		L.L.Info("(%sXML) %s", mS, mMsg)
	} else {
		L.L.Info("XML not detected by MIME libraries")
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
	// So, an error should not be fatal.
	if e != nil {
		L.L.Info("XML parsing got error: " + e.Error())
		xmlParsingFailed = true
	}
	// pathological
	if xmlParsingFailed && (hIsXml || mIsXml) {
		L.L.Panic("XML confusion (case #1) in AnalyzeFile")
	}
	if (!xmlParsingFailed) && (! (hIsXml || mIsXml)) {
		L.L.Panic("XML confusion (case #2) in AnalyzeFile")
	}

	/* Reminder: 
	type ContypingInfo struct {
	FileExt         string
	MimeType        string
	MimeTypeAsSnift string
	MType           string
	IsLwDita        bool
	*/

	// ===============================
	//  If it's DTD stuff, we're done
	// ===============================
	if peek.HasDTDstuff && SU.IsInSliceIgnoreCase(filext, XU.DTDtypeFileExtensions) {
		L.L.Okay("DTD content detected (& filext<%s>)", filext)
		pAnlRec.MimeType = "application/xml-dtd"
		pAnlRec.MType = "xml/sch/" + S.ToLower(filext[1:])
		L.L.Warning("DTD stuff: should allocate and fill field XmlInfo")
		return pAnlRec, nil
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
			L.L.Panic("analyzefile: non-xml + xml")
		}
		pAnlRec.ContypingInfo = *DoContypingInfo_non_xml(
			httpStdlibContype, sCont, filext)
		L.L.Okay("Non-XML: " + pAnlRec.ContypingInfo.String())
		// Check for YAML metadata
		iEnd, e := SU.YamlMetadataHeaderRange(sCont)
		// if there is an error, it will mess up parsing the file, so just exit.
		if e != nil {
			L.L.Error("Metadata header YAML error: " + e.Error())
			return pAnlRec, fmt.Errorf("metadata header YAML error: %w", e)
		}
		// Default: no YAML metadata found
		pAnlRec.Text.Beg = *XU.NewFilePosition(0)
		pAnlRec.Text.End = *XU.NewFilePosition(len(sCont))
		pAnlRec.Meta.Beg = *XU.NewFilePosition(0)
		pAnlRec.Meta.End = *XU.NewFilePosition(0)
		// No YAML metadata found ?
		if iEnd <= 0 {
			pAnlRec.Meta.Beg = *XU.NewFilePosition(0)
			pAnlRec.Meta.End = *XU.NewFilePosition(0)
			pAnlRec.Text.Beg = *XU.NewFilePosition(0)
			pAnlRec.Text.End = *XU.NewFilePosition(len(sCont))
		} else {
			// Found YAML metadata
			s2 := SU.TrimYamlMetadataDelimiters(sCont[:iEnd])
			ps, e := SU.GetYamlMetadataAsPropSet(s2)
			if e != nil {
				L.L.Error("loading YAML: " + e.Error())
				return pAnlRec, fmt.Errorf("loading YAML: %w", e)
			}
			// SUCCESS! Set ranges.
			pAnlRec.Meta.Beg = *XU.NewFilePosition(0)
			pAnlRec.Meta.End = *XU.NewFilePosition(iEnd)
			pAnlRec.Text.Beg = *XU.NewFilePosition(iEnd)
			pAnlRec.Text.End = *XU.NewFilePosition(len(sCont))

			pAnlRec.MetaProps = ps
			L.L.Dbg("Got YAML metadata: " + s2)
		}
		s := SU.NormalizeWhitespace(pAnlRec.ContentityStructure.Raw)
		s = SU.TruncateTo(s, 56)
		L.L.Dbg("|RAW|" + s + "|END|")
		return pAnlRec, nil
	}

	// =====================================
	//  Handle possible pathological cases.
	// =====================================
	if xmlParsingFailed {
		L.L.Dbg("Does not seem to be XML")
		if hIsXml {
			L.L.Dbg("Although http-stdlib seems to think it is:", hMsg)
		}
		if mIsXml {
			L.L.Dbg("Although 3p-mime-lib seems to think it is:", mMsg)
		}
	}
	if cheatYaml {
		L.L.Panic("analyzefile: xml + yaml")
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
	L.L.Okay("Is XML: found: %s%s%s%s", sP, sD, sR, sDtd)
	if !(gotRootElm || peek.HasDTDstuff) {
		println("--> WARNING! XML file has no root tag (and is not DTD)")
	}

	var pPRF *XU.ParsedPreamble
	if gotPreambl {
		L.L.Dbg("Preamble: %s", peek.RawPreamble)
		pPRF, e = XU.ParsePreamble(peek.RawPreamble)
		if e != nil {
			println("xm.peek: preamble failure in:", peek.RawPreamble)
			return nil, fmt.Errorf("xu.af: preamble failure: %w", e)
		}
		// print("--> Parsed XML preamble: " + pPRF.String())
	}
	// ================================
	//  Time to do some heavy lifting.
	// ================================
	L.L.Progress("Now split the file")
	if pAnlRec.ContentityStructure.Raw == "" {
		L.L.Error("analyzeFile XML: nil Raw")
	}
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
	if peek.RawDoctype != "" {
		// We are here if we got a DOCTYPE; we also have a file extension,
		// and we should have a root tag (or else the DOCTYPE makes no sense !)
		var pPDT *XU.ParsedDoctype
		pPDT = pAnlRec.ContypingInfo.ParseDoctype(peek.RawDoctype)
		if pPDT.HasError() {
			L.L.Panic("FIXME:" + pPDT.Error())
		}
		pAnlRec.ParsedDoctype = pPDT 
		L.L.Dbg("gotDT: MType: " + pAnlRec.MType)
		L.L.Dbg("gotDT: AnalysisRecord: " + pAnlRec.String())
		// L.L.Dbg("gotDT: DctpFlds: " + pPDT.String())

		if pAnlRec.MType == "" {
			L.L.Panic("fu.af: no MType, L380")
		}
		pAnlRec.ParsedDoctype = pPDT

		if pAnlRec.MType == "" {
			L.L.Panic("fu.af: no MType, L385")
		}
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
	L.L.Info("XML without DOCTYPE: <%s> root<%s> MType<%s>",
		filext, rutag, pAnlRec.MType)
	// OBS // pCntpg.MType = pAnlRec.MType
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
	L.L.Dbg("Contyping: " + pAnlRec.ContypingInfo.String())

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
	L.L.Info("fu.af: final: MType<%s> xcntp<%s> dita:TBS DcTpFlds<%s>",
		pAnlRec.MType, pAnlRec.XmlContype, // pAnlRec.XmlPreambleFields,
		// pAnlRec.DitaFlavor, pAnlRec.DitaContype,
		pAnlRec.ParsedDoctype)
	// println("--> fu.af: MetaRaw:", pAnlRec.MetaRaw())
	// println("--> fu.af: TextRaw:", pAnlRec.TextRaw())

	// ## // ## // text/xml/image/svg + xml

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

func HttpContypeIsXml(src, sContype, filext string) (isXml bool, msg string) {
	src += " contype-det'n "

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

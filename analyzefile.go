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
// The return value is an XU.AnalysisRecord, which has a BUNCH of fields.
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
	// ===================
	//  Prepare variables
	// ===================
	// pAnlRec is AnalysisRecord is basicly all of our analysis results, incl. ContypingInfo
	var pAnlRec *XU.AnalysisRecord
	// pCntpg is ContypingInfo is all of: FileExt MimeType MType Doctype IsLwDita IsProcbl
	var pCntpg *XU.ContypingInfo
	pAnlRec = new(XU.AnalysisRecord)
	// A trailing dot in the filename provides no filetype info.
	filext = FP.Ext(filext)
	if filext == "." {
		filext = ""
	}
	L.L.Dbg("Analysing file: <%s>[%d]", filext, len(sCont))
	// =====================================
	//   Don't forget to set the content.
	// (Omitting this caused a lot of bugs.)
	// =====================================
	pAnlRec.Raw = sCont
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
	var sMime string
	if httpStdlibContype == mimeLibStrContype {
		sMime = httpStdlibContype
	} else {
		sMime = httpStdlibContype + "/" + mimeLibStrContype
	}
	L.L.Info("<%s> snift-MIME-type: %s", filext, sMime)
	sniftMimeType := sMime
	// pCntpg.MimeTypeAsSnift = sMime

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
		L.L.Info("XML not detected yet")
	}

	// ===================================
	//  MAIN XML PRELIMINARY ANALYSIS:
	//  Peek into file to look for root
	//  tag and other top-level XML stuff
	// ===================================

	// Peek also sets KeyElms (Root,Meta,Text)
	var peek *XU.XmlStructurePeek
	peek = XU.PeekAtStructure_xml(sCont)
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
	if peek.HasDTDstuff && SU.IsInSliceIgnoreCase(filext, XU.DTDtypeFileExtensions) {
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
		s := SU.NormalizeWhitespace(pAnlRec.Raw)
		s = SU.TruncateTo(s, 56)
		L.L.Dbg("|RAW|" + s + "|END|")
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
	pCntpg = new(XU.ContypingInfo)
	pCntpg.FileExt = filext
	pCntpg.MimeType = httpStdlibContype
	pCntpg.MimeTypeAsSnift = sniftMimeType
	var e error
	// pAnlRec.MType = ""
	// var isLwDita bool

	var pPRF *XU.XmlPreambleFields
	if gotPreambl {
		// println("preamble:", preamble)
		pPRF, e = XU.NewXmlPreambleFields(peek.Preamble)
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
	if pAnlRec.Raw == "" {
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
	if peek.Doctype != "" {
		// We are here if we got a DOCTYPE; we also have a file extension,
		// and we should have a root tag (or else the DOCTYPE makes no sense !)
		var pXDTF *XU.XmlDoctypeFields

		pXDTF = pCntpg.AnalyzeXmlDoctype(peek.Doctype)
		if pXDTF.HasError() {
			panic("FIXME:" + pXDTF.Error())
		}
		L.L.Dbg("gotDT: MType:     " + pCntpg.MType)
		L.L.Dbg("gotDT: Contyping: " + pCntpg.String())
		L.L.Dbg("gotDT: DTDfields: " + pXDTF.String())

		if pCntpg.MType == "" {
			panic("fu.af: no MType, L339")
		}
		pAnlRec.XmlDoctypeFields = pXDTF

		if pCntpg.MType == "" {
			panic("fu.af: no MType, L345")
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
	pCntpg.MType = pAnlRec.MType
	// Do some easy cases
	if rutag == "html" && (filext == ".html" || filext == ".htm") {
		pCntpg.MType = "html/cnt/html5"
	} else if rutag == "html" && S.HasPrefix(filext, ".xht") {
		pCntpg.MType = "html/cnt/xhtml"
	} else if SU.IsInSliceIgnoreCase(rutag, XU.DITArootElms) &&
		SU.IsInSliceIgnoreCase(filext, XU.DITAtypeFileExtensions) {
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
	// pAnlRec.XmlDoctype = XU.XmlDoctype("DOCTYPE " + Peek.Doctype)
	// ?? pAnlRec.DoctypeFields = pDF
	if pPRF != nil {
		pAnlRec.XmlPreambleFields = pPRF
	} else {
		// SKIP
		// pBA.XmlPreambleFields = XU.STD_PreambleFields
	}
	// pBA.DoctypeIsDeclared  =  true
	pAnlRec.DitaFlavor = "TBS"
	pAnlRec.DitaContype = "TBS"

	// pAnlRec.XmlInfo = *new(XU.XmlInfo)

	// L.L.Info("fu.af: MType<%s> xcntp<%s> ditaFlav<%s> ditaCntp<%s> DT<%s>",
	L.L.Info("fu.af: final: MType<%s> xcntp<%s> dita:TBS DcTpFlds<%s>",
		pAnlRec.MType, pAnlRec.XmlContype, // pAnlRec.XmlPreambleFields,
		// pAnlRec.DitaFlavor, pAnlRec.DitaContype,
		pAnlRec.XmlDoctypeFields)
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

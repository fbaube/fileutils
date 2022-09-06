package fileutils

import (
	// "errors"
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

// Analysis is called only by dbutils.NewContentityRecord(..).
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
// .
func Analysis(sCont string, filext string) (*XU.AnalysisRecord, error) {

	// A trailing dot in the filename provides no filetype info.
	filext = FP.Ext(filext)
	if filext == "." {
		filext = ""
	}
	// ===========================
	//  Handle pathological cases
	// ===========================
	if len(sCont) < 6 {
		if sCont == "" {
			L.L.Progress("AnalyseFile: skipping zero-length content")
		} else {
			L.L.Warning("AnalyseFile: content too short (%d bytes)", len(sCont))
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
	httpStdlibContype = S.TrimSuffix(httpStdlibContype, "; charset=utf-8")
	mimeLibStrContype = S.TrimSuffix(mimeLibStrContype, "; charset=utf-8")
	if S.Contains(httpStdlibContype, ";") ||
		S.Contains(mimeLibStrContype, ";") {
		L.L.Warning("Content type contains a semicolon")
	}
	// Authoritative MIME type string
	var sAuthtvMime = mimeLibStrContype
	if httpStdlibContype != mimeLibStrContype {
		L.L.Warning("(AF) MIME type per libs: http<%s> mime<%s>",
			httpStdlibContype, mimeLibStrContype)
		sAuthtvMime = mimeLibStrContype
	}
	L.L.Info("(AF) filext<%s> has snift-MIME-type: %s", filext, sAuthtvMime)

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
	stdUtilIsBinary = !util.IsText([]byte(sCont))
	mimeLibIsBinary = true
	for mime := mimeLibDatContype; mime != nil; mime = mime.Parent() {
		if mime.Is("text/plain") {
			mimeLibIsBinary = false
		}
	}
	isBinary = mimeLibIsBinary
	if stdUtilIsBinary != mimeLibIsBinary {
		L.L.Warning("(AF) is-binary: "+
			"use mime-lib <%t> not http-stdlib <%t> ",
			stdUtilIsBinary, mimeLibIsBinary)
	}
	if isBinary {
		if cheatYaml || cheatXml {
			L.L.Panic("(AF) both binary & yaml/xml")
		}
		// For BINARY we won't ourselves do any more processing,
		// so we can basically trust that the sniffed MIME type
		// is sufficient, and return.
		pAnlRec.MimeType = sAuthtvMime
		pAnlRec.MType = "bin/"
		if S.HasPrefix(sAuthtvMime, "image/") {
			hasEPS := S.Contains(sAuthtvMime, "eps")
			hasTXT := S.Contains(sAuthtvMime, "text") || S.Contains(sAuthtvMime, "txt")
			if hasTXT || hasEPS {
				// TODO
				L.L.Warning("(AF) EPS/TXT confusion for MIME type: " + sAuthtvMime)
				pAnlRec.MType = "txt/img/??!"
			}
		}
		L.L.Okay("(AF) Success: detected BINARY")
		return pAnlRec, nil
	}
	// ======================================
	// We have text, but it might not be XML.
	// So use two libraries to check for XML
	// based on MIME type.
	// ======================================
	hIsXml, hMsg := HttpContypeIsXml("http",
		httpStdlibContype, filext)
	mIsXml, mMsg := HttpContypeIsXml("mime",
		mimeLibStrContype, filext)
	if hIsXml || mIsXml {
		var hS, mS string
		if !hIsXml {
			hS = "not "
		}
		if !mIsXml {
			mS = "not "
		}
		L.L.Info("(AF) (is-%sXML) %s", hS, hMsg)
		L.L.Info("(AF) (is-%sXML) %s", mS, mMsg)
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
	// So, an error should not be fatal.
	if e != nil {
		L.L.Info("(AF) XML parsing got error: " + e.Error())
		xmlParsingFailed = true
	}
	// pathological
	if xmlParsingFailed && (hIsXml || mIsXml) {
		L.L.Panic("(AF) XML confusion (case #1) in AnalyzeFile")
	}
	// Note that this next test dusnt always work for Markdown!
	/*
		if (!xmlParsingFailed) && (! (hIsXml || mIsXml)) {
			L.L.Panic("XML confusion (case #2) in AnalyzeFile")
		}
	*/

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
		L.L.Okay("(AF) Success: DTD content detected (filext<%s>)", filext)
		pAnlRec.MimeType = "application/xml-dtd"
		pAnlRec.MType = "xml/sch/" + S.ToLower(filext[1:])
		L.L.Warning("(AF) DTD stuff: should allocate and fill fields")
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
			L.L.Panic("(AF) both non-xml & xml")
		}
		pAnlRec.ContypingInfo = *DoContypingInfo_non_xml(
			httpStdlibContype, sCont, filext)
		L.L.Okay("(AF) Non-XML: " + pAnlRec.ContypingInfo.MultilineString())
		// Check for YAML metadata
		iEnd, e := SU.YamlMetadataHeaderRange(sCont)
		// if there is an error, it will mess up parsing the file, so just exit.
		if e != nil {
			L.L.Error("(AF) Metadata header YAML error: " + e.Error())
			return pAnlRec, fmt.Errorf("(AF) metadata header YAML error: %w", e)
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
				L.L.Error("(AF) loading YAML: " + e.Error())
				return pAnlRec, fmt.Errorf("loading YAML: %w", e)
			}
			// SUCCESS! Set ranges.
			pAnlRec.Meta.Beg = *XU.NewFilePosition(0)
			pAnlRec.Meta.End = *XU.NewFilePosition(iEnd)
			pAnlRec.Text.Beg = *XU.NewFilePosition(iEnd)
			pAnlRec.Text.End = *XU.NewFilePosition(len(sCont))

			pAnlRec.MetaProps = ps
			L.L.Dbg("(AF) Got YAML metadata: " + s2)
		}
		s := SU.NormalizeWhitespace(pAnlRec.ContentityStructure.Raw)
		s = SU.TruncateTo(s, 56)
		L.L.Dbg("|RAW|" + s + "|END|")
		L.L.Okay("(AF) Success: detected non-XML")
		return pAnlRec, nil
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
		switch mimeLibStrContype {
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

func HttpContypeIsXml(src, sContype, filext string) (isXml bool, msg string) {
	src += " "
	if S.Contains(sContype, "xml") {
		return true, src + "got XML (in MIME type)"
	}
	if sContype == "text/html" {
		return true, src + "got XML (cos text/html: HDITA?)"
	}
	if S.HasPrefix(sContype, "text/") &&
		(filext == ".dita" || filext == ".ditamap" || filext == ".map") {
		return true, src + "got XML (cos text/dita-filext)"
	}
	if S.Contains(sContype, "ml") {
		return true, src + "got <ml>"
	}
	if S.Contains(sContype, "svg") {
		return true, src + "got <svg>"
	}
	return false, ""
}

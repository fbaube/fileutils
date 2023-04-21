package fileutils

import (
	"fmt"
	"net/http"
	FP "path/filepath"
	S "strings"

	"github.com/gabriel-vasile/mimetype"
	"golang.org/x/tools/godoc/util" // used once, L125

	CT "github.com/fbaube/ctoken"
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
func NewPathAnalysis(pPP *PathProps) (*PathAnalysis, error) {
	var sCont string
	sCont = string(pPP.Raw)
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
	cheatHtml := S.HasPrefix(sCont, "<DOCTYPE html")
	cheatXml := S.HasPrefix(sCont, "<?xml ")
	// =========================
	//  Content type detection
	//  using 3rd party library
	//  (this is AUTHORITATIVE!)
	// =========================
	var contypeData *mimetype.MIME
	var contype string // *mimetype.MIME
	// MIME type ?
	contypeData = mimetype.Detect([]byte(sCont))
	contype = contypeData.String()
	contype = S.TrimSuffix(contype, "; charset=utf-8")
	if S.Contains(contype, ";") {
		L.L.Warning("Content type from lib_3p " +
			"(still) has a semicolon: " + contype)
	}
	// ================================
	//  Also do content type detection
	//  using http stdlib (unreliable!)
	// ================================
	var stdlib_contype string
	stdlib_contype = http.DetectContentType([]byte(sCont))
	stdlib_contype = S.TrimSuffix(stdlib_contype, "; charset=utf-8")
	if S.Contains(stdlib_contype, ";") {
		L.L.Warning("Content type from stdlib " +
			"(still) has a semicolon: " + stdlib_contype)
	}
	// ================================
	//  Authoritative MIME type string
	// ================================
	if stdlib_contype != contype {
		L.L.Warning("(AF) MIME type: lib_3p<%s> v stdlib<%s>",
			contype, stdlib_contype)
	}
	L.L.Info("(AF) <%s> snift-as: %s", filext, contype)
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
	// 2023.02 take struct PathProps out of struct PathAnalysis
	// pAnlRec.PathProps = new(PathProps)
	// pAnlRec.PathProps.Raw = sCont
	pAnlRec.FileExt = filext
	pAnlRec.MimeType = stdlib_contype // Junk this ?
	pAnlRec.MimeTypeAsSnift = contype

	// ===========================
	//  Check for & handle BINARY
	// ===========================
	var isBinary bool        // Authoritative!
	var stdlib_isBinary bool // Unreliable!
	stdlib_isBinary = !util.IsText([]byte(sCont))
	isBinary = true
	for mime := contypeData; mime != nil; mime = mime.Parent() {
		if mime.Is("text/plain") {
			isBinary = false
		}
		// FIXME If "text/" here, is an error to sort out
	}
	if stdlib_isBinary != isBinary {
		L.L.Warning("(AF) is-binary: using lib_3p <%t> not stdlib <%t> ",
			isBinary, stdlib_isBinary)
	}
	if isBinary {
		if cheatYaml || cheatXml || cheatHtml {
			L.L.Panic("(AF) both binary & yaml/xml")
		}
		return pAnlRec, pAnlRec.DoAnalysis_bin()
	}
	// ======================================
	// We have text, but it might not be XML.
	// So process the MIME types returned by
	// the two libraries.
	// ======================================
	// hIsXml, hMsg := contypeIsXml("stdlib", stdlib_contype, filext)
	mIsXml, mMsg := contypeIsXml("lib_3p", contype, filext)
	if mIsXml { // || hIsXml {
		var mS string // hS
		// if !hIsXml {
		//	hS = "not "
		// }
		if !mIsXml {
			mS = "not "
		}
		L.L.Info("(AF) lib_3p (is-%sXML) %s", // \n\t\t stdlib (is-%sXML) %s",
			mS, mMsg) // , hS, hMsg)
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
	if xmlParsingFailed && mIsXml { // (hIsXml || mIsXml) {
		L.L.Panic("(AF) XML confusion (case #1) in DoAnalysis")
	}
	// Note that this next test dusnt always work for Markdown!
	// if (!xmlParsingFailed) && (! (hIsXml || mIsXml)) {
	//      L.L.Panic("XML confusion (case #2) in AnalyzeFile")
	// }
	var hasRootTag, gotSomeXml bool
	hasRootTag, _ = pPeek.ContentityBasics.HasRootTag()
	gotSomeXml = hasRootTag || (pPeek.DoctypeRaw != "") ||
		(pPeek.PreambleRaw != "")
	// =============================
	//  If it's not XML, we're done
	// =============================
	if xmlParsingFailed || !gotSomeXml {
		if cheatXml {
			// L.L.Panic("(AF) both non-xml & xml")
			L.L.Panic(fmt.Sprintf("WHOOPS xmlParsingFailed<%t> gotSomeXml<%t> \n",
				xmlParsingFailed, gotSomeXml))
		}
		return pAnlRec, pAnlRec.DoAnalysis_txt(sCont)
	}
	// ===========================================
	//  It's XML, so crank thru it and we're done
	// ===========================================
	return pAnlRec, pAnlRec.DoAnalysis_xml(pPeek, sCont)
}

func CollectKeysOfNonNilMapValues(M map[string]*CT.FilePosition) []string {
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

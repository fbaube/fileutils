package fileutils

import (
	"fmt"
	L "github.com/fbaube/mlog"
	SU "github.com/fbaube/stringutils"
	XU "github.com/fbaube/xmlutils"
	S "strings"
)

func (pAR *PathAnalysis) DoAnalysis_xml(pXP *XU.XmlPeek, sCont string) error {
	var filext string
	filext = pAR.FileExt
	// ===============================
	//  Set bool variables, including
	//  supporting analysis by stdlib
	// ===============================
	gotRootElm, rootMsg := (pXP.ContentityBasics.HasRootTag())
	gotDoctype := (pXP.RawDoctype != "")
	gotPreambl := (pXP.RawPreamble != "")
	// gotSomeXml := (gotRootElm || gotDoctype || gotPreambl)
	// Write a progress string
	if true {
		var sP, sD, sR, sDtd string
		if gotPreambl {
			sP = "<?xml..> "
		}
		if gotDoctype {
			sD = "<DOCTYPE..> "
		}
		if gotRootElm {
			sR = "root<" + pXP.Root.TagName + "> "
		}
		if pXP.HasDTDstuff {
			sDtd = "<DTD stuff> "
		}
		L.L.Progress("Is XML: found: %s%s%s%s", sP, sD, sR, sDtd)
		if rootMsg != "" {
			L.L.Warning("Is XML: " + rootMsg)
		}
	}
	if !(gotRootElm || pXP.HasDTDstuff) {
		L.L.Warning("(AF) XML file has no root tag (and is not DTD)")
	}
	var pPRF *XU.ParsedPreamble
	var e error
	if gotPreambl {
		L.L.Dbg("(AF) got: %s", pXP.RawPreamble)
		pPRF, e = XU.ParsePreamble(pXP.RawPreamble)
		if e != nil {
			// println("xm.peek: preamble failure in:", peek.RawPreamble)
			return fmt.Errorf("(AF) preamble failure: %w", e)
		}
		// print("--> Parsed XML preamble: " + pPRF.String())
	}
	// ================================
	//  Time to do some heavy lifting.
	// ================================
	L.L.Progress("(AF) Now split the file")
	if sCont == "" { // pAR.PathProps.Raw == "" {
		L.L.Error("(AF) XML has nil Raw")
	}
	pAR.ContentityBasics = pXP.ContentityBasics
	// L.L.Warning("SKIPPING call to pAR.MakeXmlContentitySections")
	// pAR.MakeXmlContentitySections()
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
	if gotDoctype {
		// We are here if we got a DOCTYPE; we also have a file extension,
		// and we should have a root tag (or else the DOCTYPE makes no sense !)
		var pPDT *XU.ParsedDoctype
		pPDT, e = pAR.ContypingInfo.ParseDoctype(pXP.RawDoctype)
		if e != nil {
			L.L.Panic("FIXME:" + e.Error())
		}
		pAR.ParsedDoctype = pPDT
		L.L.Dbg("(AF) gotDT: MType: " + pAR.MType)
		L.L.Dbg("(AF) gotDT: AnalysisRecord: " + pAR.String())
		// L.L.Dbg("gotDT: DctpFlds: " + pPDT.String())

		if pAR.MType == "" {
			L.L.Panic("(AF) no MType, L362")
		}
		pAR.ParsedDoctype = pPDT

		if pAR.MType == "" {
			L.L.Panic("(AF) no MType, L367")
		}
		L.L.Okay("(AF) Success: got XML with DOCTYPE")
		// HACK ALERT
		if S.HasSuffix(pAR.MType, "---") {
			rutag := S.ToLower(pXP.Root.TagName)
			if pAR.MType == "xml/map/---" {
				pAR.MType = "xml/map/" + rutag
				L.L.Okay("(AF) Patched MType to: " + pAR.MType)
			} else {
				panic("MType ending in \"---\" not fixed")
			}
		}
		return nil
	}
	// =====================
	//  No DOCTYPE. Bummer.
	// =====================
	if !gotRootElm {
		return fmt.Errorf("(AF) Got no XML root tag in file with ext <%s>", filext)
	}
	// ==========================================
	//  Let's at least try to set the MType.
	//  We have a root tag and a file extension.
	// ==========================================
	rutag := S.ToLower(pXP.Root.TagName)
	L.L.Progress("(AF) XML without DOCTYPE: <%s> root<%s> MType<%s>",
		filext, rutag, pAR.MType)
	// Do some easy cases
	if rutag == "html" && (filext == ".html" || filext == ".htm") {
		pAR.MType = "html/cnt/html5"
	} else if rutag == "html" && S.HasPrefix(filext, ".xht") {
		pAR.MType = "html/cnt/xhtml"
	} else if SU.IsInSliceIgnoreCase(rutag, XU.DITArootElms) &&
		SU.IsInSliceIgnoreCase(filext, XU.DITAtypeFileExtensions) {
		pAR.MType = "xml/cnt/" + rutag
		if rutag == "bookmap" && S.HasSuffix(filext, "map") {
			pAR.MType = "xml/map/" + rutag
		}
	}
	// pAnlRec.ContypingInfo = *pCntpg
	if pAR.MType == "-/-/-" {
		pAR.MType = "xml/???/" + rutag
	}
	// At this point, mt should be valid !
	L.L.Dbg("(AF) Contyping: " + pAR.ContypingInfo.String())

	// Now we should fill in all the detail fields.
	pAR.XmlContype = "RootTagData"

	if pPRF != nil {
		pAR.ParsedPreamble = pPRF
	} else {
		// SKIP
		// pBA.XmlPreambleFields = XU.STD_PreambleFields
	}
	// pBA.DoctypeIsDeclared  =  true
	pAR.DitaFlavor = "TBS"
	pAR.DitaContype = "TBS"

	// L.L.Info("fu.af: MType<%s> xcntp<%s> ditaFlav<%s> ditaCntp<%s> DT<%s>",
	L.L.Progress("(AF) final: MType<%s> xcntp<%s> dita:TBS DcTpFlds<%s>",
		pAR.MType, pAR.XmlContype, // pAnlRec.XmlPreambleFields,
		// pAnlRec.DitaFlavor, pAnlRec.DitaContype,
		pAR.ParsedDoctype)
	// println("--> fu.af: MetaRaw:", pAnlRec.MetaRaw())
	// println("--> fu.af: TextRaw:", pAnlRec.TextRaw())

	// HACK!
	if pAR.MType == "" {
		switch pAR.ContypingInfo.MimeTypeAsSnift { // m_contype {
		case "image/svg+xml":
			pAR.MType = "xml/img/svg"
		}
		if pAR.MType != "" {
			L.L.Warning("(AF) Lamishly hacked the MType to: %s", pAR.MType)
		}
	}
	L.L.Okay("(AF) Success: got XML without DOCTYPE")
	return nil
}

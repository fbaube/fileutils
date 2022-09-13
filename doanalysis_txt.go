package fileutils

import (
	"fmt"
	L "github.com/fbaube/mlog"
	SU "github.com/fbaube/stringutils"
	XU "github.com/fbaube/xmlutils"
	S "strings"
)

// DoAnalysis_txt is called when the content is identified
// as non-XML. It does not expect to see binary content.
// .
func (pAR *PathAnalysis) DoAnalysis_txt() error {
	sCont := pAR.PathProps.Raw
	/*
		// ===================================
		// pAR.ContypingInfo = *DoContypingInfo_non_xml(h_contype, sCont, filext)
		// func DoContypingInfo_non_xml(sniftMT, sCont, filext string) *XU.ContypingInfo {
		sniftMT := pAR.MimeTypeAsSnift
		filext := pAR.ContypingInfo.FileExt

		// theFileExt includes a leading "."
		var ret = new(XU.ContypingInfo)
		ret.FileExt = filext
		ret.MType = sniftMT
	*/
	// Markdown is a tough case, because it's basically a text file.
	// There is no string that definitively declares "This is Markdown",
	// and there might not be YAML metadata at the start of the file, and
	// at this point here we don't want to scan (and try to definitively
	// categorise) ALL the file content, at least not more than the first
	// few characters. So, about the best we can do is check for a known
	// file extensions.
	if S.HasPrefix(pAR.MimeTypeAsSnift, "text/") &&
		SU.IsInSliceIgnoreCase(pAR.ContypingInfo.FileExt, XU.MarkdownFileExtensions) {
		pAR.ContypingInfo.MimeType = "text/markdown"
		pAR.ContypingInfo.MimeTypeAsSnift = "text/markdown"
		pAR.ContypingInfo.MType = "mkdn/tpcOrMap/?fmt"
	} else {
		L.L.Warning("Reached no conclusion about type of non-XML content")
	}
	// L.L.Okay("(AF) Non-XML: " + pAR.ContypingInfo.MultilineString())
	// Check for YAML metadata
	iEnd, e := SU.YamlMetadataHeaderRange(sCont)
	// if there is an error, it will mess up parsing the file, so just exit.
	if e != nil {
		L.L.Error("(AF) Metadata header YAML error: " + e.Error())
		return fmt.Errorf("(AF) metadata header YAML error: %w", e)
	}
	// Default: no YAML metadata found
	pAR.Text.Beg = *XU.NewFilePosition(0)
	pAR.Text.End = *XU.NewFilePosition(len(sCont))
	pAR.Meta.Beg = *XU.NewFilePosition(0)
	pAR.Meta.End = *XU.NewFilePosition(0)
	// No YAML metadata found ?
	if iEnd <= 0 {
		pAR.Meta.Beg = *XU.NewFilePosition(0)
		pAR.Meta.End = *XU.NewFilePosition(0)
		pAR.Text.Beg = *XU.NewFilePosition(0)
		pAR.Text.End = *XU.NewFilePosition(len(sCont))
	} else {
		// Found YAML metadata
		s2 := SU.TrimYamlMetadataDelimiters(sCont[:iEnd])
		ps, e := SU.GetYamlMetadataAsPropSet(s2)
		if e != nil {
			L.L.Error("(AF) loading YAML: " + e.Error())
			return fmt.Errorf("loading YAML: %w", e)
		}
		// SUCCESS! Set ranges.
		pAR.Meta.Beg = *XU.NewFilePosition(0)
		pAR.Meta.End = *XU.NewFilePosition(iEnd)
		pAR.Text.Beg = *XU.NewFilePosition(iEnd)
		pAR.Text.End = *XU.NewFilePosition(len(sCont))

		pAR.MetaProps = ps
		L.L.Dbg("(AF) Got YAML metadata: " + s2)
	}
	s := SU.NormalizeWhitespace(pAR.PathProps.Raw)
	s = SU.TruncateTo(s, 56)
	L.L.Dbg("|RAW|" + s + "|END|")
	// L.L.Okay("(AF) Success: detected non-XML")
	return nil
}

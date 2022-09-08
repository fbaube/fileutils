package fileutils

import (
	L "github.com/fbaube/mlog"
	XU "github.com/fbaube/xmlutils"
	S "strings"
)

// DoAnalysis_sch will handle DTDs and related files,
// and the code is mostly written but not yet integrated,
// so this func doesn't really worry about it yet.
// .
func DoAnalysis_sch(pAR *XU.AnalysisRecord) (*XU.AnalysisRecord, error) {
	// L.L.Okay("(AF) Success: DTD-type content detected (filext<%s>)", filext)
	pAR.MimeType = "application/xml-dtd"
	pAR.MType = "xml/sch/" + S.ToLower(S.TrimPrefix(pAR.FileExt, "."))
	L.L.Warning("(AF) DTD stuff: should allocate and fill fields")
	return pAR, nil
}

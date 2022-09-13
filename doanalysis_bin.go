package fileutils

import (
	L "github.com/fbaube/mlog"
	S "strings"
)

// DoAnalysis_bin doesn't do any further processing for binary, cos we
// basically trust that the sniffed MIME type is sufficient, and return.
// .
func (pAR *PathAnalysis) DoAnalysis_bin() error {
	// pAnlRec.MimeType = m_contype
	pAR.MType = "bin/"
	m_contype := pAR.ContypingInfo.MimeTypeAsSnift
	if S.HasPrefix(m_contype, "image/") {
		det := S.TrimPrefix(m_contype, "image/")
		pAR.MType += "img/"
		hasEPS := S.Contains(m_contype, "eps")
		hasTXT := S.Contains(m_contype, "text") ||
			S.Contains(m_contype, "txt")
		if hasTXT || hasEPS {
			// TODO
			L.L.Warning("(AF) EPS/TXT confusion for MIME type: " + m_contype)
			pAR.MType = "txt/img/??!"
		} else {
			pAR.MType += det
			if !S.EqualFold(pAR.FileExt, "."+det) {
				L.L.Warning("Image: detMime<%s> filext<%s>", det, pAR.FileExt)
			}
		}
	} else {
		L.L.Warning("Image problem: mime<%s> filext<%s>",
			pAR.MimeTypeAsSnift, pAR.FileExt)
	}
	// L.L.Okay("(AF) Success: detected BINARY")
	return nil
}

package fileutils

import (
	CT "github.com/fbaube/ctoken"
	S "strings"
)

func CollectKeysOfNonNilMapValues(M map[string]*CT.FilePosition) []string {
	var ss []string
	for K, V := range M {
		if V != nil {
			ss = append(ss, K)
		}
	}
	return ss
}

// contypeIsXml returns true for HTML too.
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

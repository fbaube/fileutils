package fileutils

import (
	S "strings"
)

func (p *PathAnalysis) String() string {
	var sb S.Builder
	var sPDT string
	if p.ParsedDoctype != nil {
		sPDT = p.ParsedDoctype.String()
	}
	sb.WriteString("PathAnalysis: ")
	sb.WriteString("CntpgInfo: \n\t" + p.ContypingInfo.String() + "\n\t")
	sb.WriteString("XmlCntp<" + p.XmlContype + "> ")
	sb.WriteString("XmlDctp<" + sPDT + "> ")
	return sb.String()
}

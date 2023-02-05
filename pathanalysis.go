package fileutils

import (
	SU "github.com/fbaube/stringutils"
	XU "github.com/fbaube/xmlutils"
	"strings"
	S "strings"
)

type Doctype string
type MimeType string

// PathAnalysis is the results of content analysis
// on the contents of the embedded [PathProps].
// .
type PathAnalysis struct { // this has has Raw
	// PathProps is a ptr, so that we get a
	// NPE if it is not initialized properly.
	// WRONG WRONG WRONG *PathProps // this has Raw
	// ContypingInfo is simple fields:
	// FileExt MType MimeType's
	XU.ContypingInfo
	MarkdownFlavor string
	// ContentityBasics includes Raw
	// (the entire input content)
	XU.ContentityBasics // has Raw
	// KeyElms is: (Root,Meta,Text)ElmExtent
	// KeyElmsWithRanges
	// ContentitySections is: Text_raw, Meta_raw, MetaFormat; MetaProps SU.PropSet
	// ContentityRawSections
	// XmlInfo is: XmlPreambleFields, XmlDoctype, XmlDoctypeFields, ENTITY stuff
	/* XmlInfo */
	// XmlContype is an enum: "Unknown", "DTD", "DTDmod", "DTDent",
	// "RootTagData", "RootTagMixedContent", "MultipleRootTags", "INVALID"}
	XmlContype string
	// XmlPreambleFields is nil if no preamble - it can always
	// default to xmlutils.STD_PreambleFields (from stdlib)
	*XU.ParsedPreamble
	// XmlDoctypeFields is a ptr - nil if ContypingInfo.Doctype
	// is "", i.e. if there is no DOCTYPE declaration
	*XU.ParsedDoctype
	// DitaInfo
	DitaFlavor  string
	DitaContype string
}

// IsXML is true for all XML, including all HTML.
func (p PathAnalysis) IsXML() bool {
	s := p.MarkupType()
	return s == SU.MU_type_XML || s == SU.MU_type_HTML
}

func (p *PathAnalysis) String() string {
	/*
		// ContypingInfo is simple fields:
		// FileExt MimeType MType Doctype IsLwDita
		ContypingInfo
		MarkdownFlavor string
		// ContentityStructure includes Raw (the entire input content)
		ContentityStructure
		// KeyElms is: (Root,Meta,Text)ElmExtent
		// KeyElmsWithRanges
		// ContentitySections is: Text_raw, Meta_raw, MetaFormat; MetaProps SU.PropSet
		// ContentityRawSections
		// XmlInfo is: XmlPreambleFields, XmlDoctype, XmlDoctypeFields, ENTITY stuff
		/* XmlInfo * /
		// XmlContype is an enum: "Unknown", "DTD", "DTDmod", "DTDent",
		// "RootTagData", "RootTagMixedContent", "MultipleRootTags", "INVALID"}
		XmlContype string
		// XmlPreambleFields is nil if no preamble - it can always
		// default to xmlutils.STD_PreambleFields (from stdlib)
		*XmlPreambleFields
		// XmlDoctypeFields is a ptr - nil if ContypingInfo.Doctype
		// is "", i.e. if there is no DOCTYPE declaration
		*XmlDoctypeFields
		// DitaInfo
		DitaFlavor  string
		DitaContype string
	*/
	var sb strings.Builder
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

// MarkupType returns an enum with values of "XML",
// "MKDN", "HTML", "UNK", or future stuff TBD.
// .
func (p PathAnalysis) MarkupType() SU.MarkupType {
	// HTML is an exceptional case
	if S.HasPrefix(p.MType, "xml/html/") {
		return SU.MU_type_HTML
	}
	if S.HasPrefix(p.MimeType, "text/html") {
		return SU.MU_type_HTML
	}
	if S.HasPrefix(p.MimeType, "html/") {
		return SU.MU_type_HTML
	}

	if S.HasPrefix(p.MType, "xml/") {
		return SU.MU_type_XML
	}
	if S.HasPrefix(p.MType, "text/") ||
		S.HasPrefix(p.MType, "txt/") ||
		S.HasPrefix(p.MType, "md/") ||
		S.HasPrefix(p.MType, "mkdn/") {
		return SU.MU_type_MKDN
	}
	if S.HasPrefix(p.MType, "bin/") {
		return SU.MU_type_UNK // No BIN !!
	}
	return SU.MU_type_UNK
}

// XML, HTML, BIN, TXT, MD/MKDN

/*
	// Normal case
	// return S.ToUpper(MTypeSub(p.MType, 0))
	// Cut & Paste
	if p.MType == "" {
		return SU.MU_type_UNK
	}
	sUnk := string(SU.MU_type_UNK)
	i := S.Index(p.MType, "/") // not S.Cut(..)
	if i == -1 {
		return SU.MarkupType(sUnk + ":" + p.MType)
	}
	return SU.MarkupType(sUnk + S.ToUpper(p.MType[:i]))
}
*/

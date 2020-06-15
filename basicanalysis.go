package fileutils

import (
  "fmt"
  S  "strings"
  XM "github.com/fbaube/xmlmodels"
)

type BasicAnalysis struct { 
  baError     error
  FileIsOkay  bool
  FileExt     string
  IsXml       int
  MimeType    string
  MType       string // used to be []! (i.e. [3])
  RootTag     string // e.g. <html>, enclosing both <head> and <body>
  RootAtts    string // e.g. <html lang="en">
  // MARKDOWN
  MarkdownFlavor string
  // XML & HTML
  XM.XmlInfo
  XM.DitaInfo
}

func NewBasicAnalysis() *BasicAnalysis {
  p := new(BasicAnalysis)
  p.MType = "-/-/-"
  return p
}

func (pBA *BasicAnalysis) IsXML() bool {
  if pBA.IsXml == 0 { return false }
  return true
}

// FileType returns "XML", "MKDN", "HTML", or future stuff TBD.
func (p BasicAnalysis) FileType() string {
	// Exceptional case
	if S.HasPrefix(p.MType, "xml/html/") { return "HTML" }
	// Normal case
	return S.ToUpper(MTypeSub(p.MType, 0))
}

func (p BasicAnalysis) String() string {
  var x string
  if p.IsXML() { x = "IsXML.. " }
  return fmt.Sprintf("fu.basicAnls: \n   " + x +
    "MimeType<%s> " +
    "MType<%v> " +
    "\n   " +
    "XmlInfo<%+v> " +
    "\n   " +
    "DitaInfo<%+v>",
    p.MimeType, p.MType, p.XmlInfo, p.DitaInfo)
}

/*
// MType is modeled after Mime-type. It is set by our own code,
// based on `MimeType` and shallow analysis of file contents.
// Markdown is presumed to be MDITA, because in any case,
// any Markdown is sposta be compatible with MDITA.
//
// String possibilities in each byte:
// [0] XML, BIN, TXT, MKDN
// [1] IMG, CNT (Content), TOC (Map), SCH(ema)
// [2] XML: per-DTD; BIN: format/filext; MD: flavor; SCH: format/filext
*/

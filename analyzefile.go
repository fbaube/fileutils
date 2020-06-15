package fileutils

import(
  "fmt"
  "net/http"
  S "strings"
  FP "path/filepath"
  SU "github.com/fbaube/stringutils"
  XM "github.com/fbaube/xmlmodels"
)

// AnalyseFile has drastically different handling for XML content versus
// non-XML content, so for the sake of clarity and simplicity, it tries
// to first process non-XML content. For XML files we have much more
// info available, so processing is both simpler and more complicated.
//
// The second argument "filext" can be any filepath; the Go stdlib is used
// to split off the file extension. It can also be "", if (for example) the
// content is entered interactively, without a file name or assigned MIME type.
//
// If the first argument "sCont" (the content) is zero-length, return (nil, nil).
//
func AnalyseFile(sCont string, filext string) (*BasicAnalysis, error) {

  if sCont == "" {
    println("==>", "Cannot analyze zero-length content")
    return nil, nil
  }
  filext = FP.Ext(filext)
  if filext == "." { filext = "" }
  fmt.Printf("==> Content analysis: len<%d> filext<%s> \n", len(sCont), filext)

  // =======================================
	//  stdlib HTTP content detection library
	// =======================================
	// Note that it assigns "text/html" to DITA maps :-/
	httpContype := http.DetectContentType([]byte(sCont))
	httpContype = S.TrimSuffix(httpContype, "; charset=utf-8")
  println("-->", "HTTP stdlib:", httpContype)

  // =======================================
	//  Quick check for top-level XML stuff
	// =======================================
  preamble, doctype, rootTag, e := XM.Peek_xml(sCont)
  fmt.Printf("--> xm.peek: \n\t prmbl: %t \n\t DT: %s \n\t RootTag: %s \n",
    // RootTag.Name<%+v> RootTag.Attr<%+v> \n",
    (preamble != ""), doctype, XmlStartElmS(rootTag))
    // rootTag.Name, rootTag.Attr)

  var isXml bool
  // Note that this check means that if an error was encountered,
  // content that might be XML will instead be processed as non-XML.
  isXml = (e == nil && (preamble != "" || doctype != "" || rootTag.Name.Local != ""))
  // We can do a bit more checking for XML
  if !isXml {
    if S.Contains(httpContype, "xml") {
      println("--> Peek_xml oops: HTTP contype-detection got XML (in MIME type)")
      isXml = true
    }
    if httpContype == "text/html" {
      println("--> Peek_xml oops: HTTP contype-detection got XML (text/html) (HDITA?)")
      isXml = true
    }
    if S.HasPrefix(httpContype, "text/") &&
  		(filext == ".dita" || filext == ".ditamap" || filext == ".map") {
      println("--> Peek_xml oops: HTTP contype-detection got XML (text/dita-filext)")
      isXml = true
    }
    if S.Contains(httpContype, "ml") {
      println("--> Peek_xml oops: HTTP contype-detection got <ml>")
      isXml = true
  	}
    if S.Contains(httpContype, "svg") {
      println("--> Peek_xml oops: HTTP contype-detection got <svg>")
      isXml = true
  	}
    if SU.IsInSliceIgnoreCase(filext, DTDtypeFileExtensions) {
      println("--> Peek_xml oops: DTDtypeFileExtensions")
      isXml = true
      // RETURN "application/xml-dtd", "xml/sch/" + filext[1:]
    }
  }

  p := NewBasicAnalysis()
  p.FileExt  = filext

  // ===============
  //   NOT XML !!!
  // ===============
  if !isXml {
    var mimetype, mtype string
    mimetype, mtype = DoContentTypes_non_xml(httpContype, sCont, filext)
    fmt.Printf("--> NON-XML: filext|%s| mtype|%s| mimetype|%s| \n",
      filext, mtype, mimetype)
    p.FileExt  = filext
    p.MimeType = mimetype
    p.MType    = mtype
    return p, nil
  }

  // ===============
  //   YES XML !!!
  // ===============
  // var isLwDita bool
  p.IsXml = 1
  var pPR *XM.XmlPreambleFields
  var pDT *XM.XmlDoctypeFields
  if preamble != "" {
    // println("preamble:", preamble)
    pPR, e = XM.NewXmlPreambleFields(preamble)
    if e != nil {
      println("xm.peek: preamble failure")
    }
    print("--> XML preamble fields: " + pPR.String())
  }
  if doctype != "" {
    println("doctype:", doctype)
    mt,_ /* isLwdita */ := XM.GetMTypeByDoctype(doctype)
    println("-->", "search results:", mt)
    pDT, e = XM.NewXmlDoctypeFieldsInclMType(doctype)
    if e != nil {
      println("xm.peek: doctype failure")
    }
    println("-->", "XML doctype fields:", pDT.String())
  }
  if rootTag.Name.Local == "" { println("ROOT TAG OOPS") }

  // =========================
  //  NOW USE DOCTYPE INFO !!
  // =========================
  // rootTag xml.StartElement is valid
  // pDoctp is valid:
  // DoctypeMType string
  // TopTag string
  // XmlPublicIDcatalogRecord
  dmt := "none"
  if pDT != nil {
    fmt.Printf("    RootTag: DT<%s> text<%s> \n",
       pDT.TopTag, XmlStartElmS(rootTag))
    if pDT.TopTag != rootTag.Name.Local {
       panic("ROOT TAG MISMATCH")
      }
    dmt = pDT.DoctypeMType
  }
  fmt.Printf("    MimeTyp: contype<%s> \n", p.MimeType)
  fmt.Printf("     M-Type: DT<%s> contype<%s> \n", dmt, p.MType)

  fmt.Printf("DT-ptrs BFR: p.XmlInfo<%+v> p.DitaInfo<%+v> \n",
    p.XmlInfo, p.DitaInfo)
  /*
  type XmlInfo struct {
    XmlContype `db:"xmlcontype"`
    XmlDoctype `db:"xmldoctype"`
   *XmlDoctypeFields
   *XmlPreambleFields
    RootTagCt int
    DoctypeIsDeclared, DoctypeIsGuessed bool
  }
  type DitaInfo struct {
    DitaMarkupLg `db:"ditamarkuplg"`
    DitaContype  `db:"ditacontype"`
  } */
  p.XmlInfo.XmlContype = "TBS"
  p.XmlInfo.XmlDoctype = XM.XmlDoctype("DOCTYPE " + doctype)
  p.XmlDoctypeFields   = pDT
  p.XmlPreambleFields  = pPR
  p.DoctypeIsDeclared  = true
  p.DitaInfo.DitaMarkupLg = "TBS"
  p.DitaInfo.DitaContype  = "TBS"

  fmt.Printf("DT-ptrs AFR: p.XmlInfo<%+v> p.DitaInfo<%+v> \n",
    p.XmlInfo, p.DitaInfo)

  return p, nil
}

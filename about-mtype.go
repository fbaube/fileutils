package fileutils

// An MType is specific to this app and/but is modeled after
// the prior concept of Mime-type. An MType has three fields.
//
// Its value is generally based on two to four inputs:
//  - The Mime-type guess returned by Go stdlib
//    func net/http.DetectContentType(data []byte) string
//    (which is based on https://mimesniff.spec.whatwg.org/ )
//    (The no-op default return value is "application/octet-stream")
//  - Our own shallow analysis of file contents
//  - The file extension (it is normally present)
//  - The DOCTYPE (iff XML, incl. HTML)
//
// Note that
//  - a plain text file MAY be presumed to be Markdown, altho it
//    is not clear (yet) which (TXT or MKDN) should take precedence.
//  - a Markdown file CAN and WILL be presumed to be LwDITA MDITA.
//  - mappings can appear bogus, for example
//    HTTP stdlib "text/html" becomes MType "xml/html".
//
// String possibilities in each field:
//  [0] XML, HTML, BIN, TXT, MKDN [we keep XML and HTML distinct for
//      a number of reasons, but partly because in the Go stdlib, they
//      have quite different processing, and we take advantage of it to
//      keep HTML processing free of nasty surprises]
//  [1] CNT (Content), MAP (ToC), IMG, SCH(ema) [and maybe others TBD?]
//  [2] Depends on [0]: 
//       XML: per-DTD [and/or Pub-ID/filext];
//      HTML: per-DTD [and/or Pub-ID/filext];
//       BIN: format/filext;
//       SCH: format/filext [DTD,MOD,XSD,wotevs];
//       TXT: TBD
//      MKDN: flavor of Markdown (?) (note also AsciiDoc, RST, ...)
//
// FIXME: Let [2] (3rd) be version info (html5, lwdiat, dita13)
//        and then keep root tag info separately.
//
// type XmlDoctypeFamily string
//      XmlDoctypeFamilies are the broad groups of DOCTYPES.
//  var XmlDoctypeFamilies = []XmlDoctypeFamily {
//	"lwdita",
//	"dita13",
//	"dita",
//	"html5",
//	"html4",
//  "svg",
//  "mathml",
//	"other",
// }

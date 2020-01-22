package fileutils

import (
	"net/http"
	S "strings"
)

// BasicContent is normally a file, opened and loaded into `Raw`
// (by `func FileLoad()`), and at that point the content is fully
// decoupled from the file system.
// Mime type guessing is done using standard libraries (both Go's and
// a 3rd party's), so this code can still be considered very low-level.
type BasicContent struct {
	// BCerror - if non-nil - makes methods in a chain skip their own processing.
	bcError error
	// FileCt is >1 IFF this struct refers to a directory, and multifile
	// processing is needed. In the future we might also handle wildcards.
	FileCt int
	// Raw and Size apply to files only, not to directories or symlinks.
	Raw  string
	// MagicMimeType is set using a 3rd party binding to libmagic.
	MagicMimeType string
	// SniftMimeType is set using the Golang stdlib.
	SniftMimeType string
	// IsXML is set by our own code, using various heuristics of
	// our own fiendish device.
	IsXML bool
	// MType is modeled after Mime-type. It is set by our own code, based on
	// `MagicMimeType`, `SniftMimeType`, and shallow analysis of file contents.
	// Markdown is presumed to be MDITA, because in any case, any Markdown is
	// sposta be compatible with MDITA.
	//
	// [0] XML, BIN, TXT, MD
	// [1] IMG, CNT (Content), TOC (Map), SCH(ema)
	// [2] XML: per-DTD; BIN: fmt/filext; MD: flavor; SCH: fmt/filext
	//
	// Common values:
	// * Textual  image files:  image /  text / (svg|eps)
	// * Binary   image files:  image /  bin  / (fmt)
	// * DITA13 content files:   dita / (tech|..) / (task|..)
	// * LwDITA content files: lwdita / (xdita|hdita|mdita[xp]) / (map|topic|..)
	// *   HTML content files:   html / (5|4) [/TBD]
	// * Parsed  schema files: schema / dtd / (root elm)
	// * Indeterminate XML that hopefully will get
	//     DOCTYPE processing:    xml / xml [/TBD]
	// [0] = doc family = image/dita/lwdita/html/schema/xml
	// [1] = doc format = its format/dtd
	// [2] = specifics
	// NOTE how a text-based image file (i.e. SVG or EPS)
	// can be `image` but not `binary`.
	MType []string
}

// GetError is necessary cos `Error()`` dusnt tell you whether `error` is `nil`,
// which is the indication of no error. Therefore we need this function, which
// can actually return the telltale `nil`.`
func (p *BasicContent) GetError() error {
	return p.bcError
}

// Error satisfied interface `error`, but the
// weird thing is that `error` can be nil.
func (p *BasicContent) Error() string {
	if p.bcError != nil {
		return p.bcError.Error()
	}
	return ""
}

func (p *BasicContent) SetError(e error) {
	p.bcError = e
}

// FileType returns `XML`, `MKDN`, `HTML`, or future stuff TBD.
func (p *BasicContent) FileType() string {
	// Exceptional case
	if p.MType[0] == "xml" && p.MType[1] == "html" { return "HTML" }
	// Normal case 
	return S.ToUpper(p.MType[0])
}

// InspectFile comprises four steps:
//
// * use stdlib and third-party libraries to make initial guesses
// * dump those guesses for the purpose of evaluating those libraries
// * call custom code to evaluate more deeply XML and/or as mixed content
// * dump those results for the purpose of refining the code
//
// The fields of interest in `struct fileutiles.InputFile`:
//
// - Set using various heuristics of our own devising: IsXML bool
// - Set using Golang stdlib: SniftMimeType string
// - Set using 3rd-party lib: MagicMimeType string
// - Set by our own code, based on the results set
// in the preceding string fields: Mtype []string
//
func (p *BasicContent) InspectFile() *BasicContent {
	if p.bcError != nil {
		return p
	}
	p.MagicMimeType = GoMagic(p.Raw)
	// Trim long JPEG descriptions
	if s := p.MagicMimeType; S.HasPrefix(s, "JPEG") {
		if i := S.Index(s, "xres"); i > 0 {
			p.MagicMimeType = "JPEG, " + s[i:]
		}
	}
	// This next call assigns "text/html" to DITA maps :-/
	contyp := http.DetectContentType([]byte(p.Raw)) // (content))
	p.SniftMimeType = S.TrimSuffix(contyp, "; charset=utf-8")

	// fmt.Printf("    MIME: (%s) %s \n", p.SniftMimeType, p.MagicMimeType)
	return p
}

// NOTE A text-based image file (i.e. SVG or EPS) can be
// `IsImage` but `!IsBinary`.
//
/*
type SegmentedFile struct {
	CheckedPath
	// FileContent holds the file's entire contents (i.e. `GGFile.Raw`)
	// MINUS any content analysed and read into `Header`.
	FileContent
	// Header basically holds all the in-file metadata, no matter what
	// the format of the file. For MDITA (Markdown-XP) this is YAML.
	// For XDITA and HDITA ([X]HTML) this is head/meta elements.
	// It can be non-nil but have its sero value, which indicates
	// that metadata was checked for but none was found.
	// Storing metadata this way makes it easier to manage it in a
	// consistent and format-independent manner, and makes it easier
	// to add to it and modify it at runtime.
	// TODO: When in-file metadata is stored as JSON K/V pairs,
	// it can be accessed from the command line using Sqlite tools.
	*Header
}
*/

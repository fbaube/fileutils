package fileutils

import(
	"fmt"
	"os"
	"io/fs"
	"errors"
	S "strings"
	FP "path/filepath"
	SU "github.com/fbaube/stringutils"
)

// Filepaths shuld always have all three fields set, even if the third
// ([ShortFP]) is basically session-specific. Note that directories
// always have a slash (or OS sep) appended, and symlinks never should. 
//
// NOTE that the file name (aka [FP.Base], the part of the full path after
// the last directory separator) is not stored separately: it is stored in
// AbsFP *and* RelFP. Note also that all this path & name information
// duplicates what is stored in an instance of orderednodes.Nord .
// . 
type Filepaths struct {
     // RelFP is tipicly the path given (e.g.) on the command line and is
     // useful for resolving relative paths in batches of content items.
     // The value might be valid only for the current CLI invocation or
     // user session, but it is persistable to preserve relationships
     // among files in import batches. 
     RelFP string
     // AbsFP is the authoritative field when processing individual files. 
     AbsFP string
     // GotAbs (from [path/filepath/IsAbs]) says that this struct was
     // created using a relative FP, not an absolute FP, and so the
     // field [RelFP] is calculated. 
     GotAbs bool
     // Local (from [path/filepath/IsLocal]) is OK but not-Local might
     // be a security hole.
     Local bool
     // Valid (from [path/filepath/ValidPath]) fails for absolute paths,
     // but can be set to `true` for them. 
     Valid bool
     // ShortFP is the path shortened by using "." (CWD) or "~" (user's
     // home directory), so it might only be valid for the current CLI
     // invocation or user session and it is def not persistable. 
     ShortFP string
}

// NewFilepaths relies on the std lib, and accepts
// either an absolute or a relative filepath. It
// does, however, not accept an empty filepath.
// It 
// It takes care to remove a trailing slash (or OS sep) before calling functions in [path/filepath], so that symlinks are not unintentionally followed.
//
// Ref: type PathError struct {	Op string Path string Err error }
// .
func NewFilepaths(anFP string) (*Filepaths, error) {
     var pFPs *Filepaths 
     if anFP == "" {
     	return nil, errors.New("NewFilepaths: empty path")
	} 
     // Normalize it ("using only lexical analysis") 
     anFP = FP.Clean(anFP)
     // Allocate the storage now, to use the filag fields. 
     pFPs = new(Filepaths)
     // Validate it (altho we expect this 
     // call to fail if it is an absolute FP) 
     pFPs.Valid = fs.ValidPath(anFP)
     // Check whether it is local.
     // func FP.IsLocal(path string) bool
     // IsLocal reports whether path, using lexical
     // analysis only, has all of these properties:
     //  - is within the subtree rooted at the 
     //    directory in which path is evaluated
     //  - is not an absolute path
     //  - is not empty
     //  - on Windows, is not a reserved name such as "NUL"
     // If IsLocal(path) returns true, then
     //  - FP.Join(base, path) is always a path contained within base, and
     //  - FP. Clean(path) is always an unrooted path with no ".." path elements.
     // IsLocal is a purely lexical operation. In particular,
     // it does not account for the effect of any symbolic
     // links that may exist in the filesystem.
     pFPs.Local = FP.IsLocal(anFP) 
     pFPs.GotAbs = FP.IsAbs(anFP)
     fmt.Printf("<%s> Abs<%t> Local<%t> Valid <%t> \n",
     	 anFP, pFPs.GotAbs, pFPs.Local, pFPs.Valid)
     if pFPs.GotAbs {
     	if pFPs.Valid { println("Abs is valid ?!:", anFP) } else
	 { println("Abs is invalid, as expected ::", anFP) }
	} 
     if !(pFPs.Valid || pFPs.GotAbs) {
     	return nil, &os.PathError{
	       Op: "fs.ValidPath(RelFP)", Path: anFP, Err: fs.ErrInvalid }
	}
     // Maybe required somewhere near here for Windoze:
     // func Localize(path string) (string, error)
     // Localize converts a slash-separated path into an OS path.
     // The input path must be a valid path per io/fs.ValidPath.
     // Localize returns an error if the path cannot be represented
     // by the OS. For example, the path a\b is rejected on Windows,
     // cos \ is a separator character and cannot be part of a filename.

     // If got an abs.FP 
     if pFPs.GotAbs {
     	pFPs.AbsFP = anFP
	pFPs.RelFP = SU.Tildotted(anFP) // Calculated 
	pFPs.ShortFP = pFPs.RelFP 
     } else {
        pFPs.RelFP = anFP
	// If there is some exotic problem with 
	// the input path, it could surface here.
	var e error 
	pFPs.AbsFP, e = FP.Abs(anFP)
	if e != nil {
	   return nil, &fs.PathError{ Op:"FP.Abs", Err:e, Path:anFP }
	}
     	pFPs.ShortFP = SU.Tildotted(pFPs.AbsFP)
     }
     // Strip off any trailing slash (or OS sep), cos we do 
     // not want func [os.Lstat] to auto-follow symlinks. If 
     // the path is a directory and needs a traililng slash 
     // (or OS sep), it will be added later elsewhere. 
     pFPs.TrimPathSepSuffixes() 
     return pFPs, nil
}

// CreationPath is the path (abs or rel) used to create it.
// It can be "", if the item wasn't/isn't on disk.
func (p *Filepaths) CreationPath() string {
        if p.GotAbs { return p.AbsFP }
        return p.RelFP
}

func (p *Filepaths) TrimPathSepSuffixes() {
     if p.AbsFP != "" && p.AbsFP != "/" && p.AbsFP != PathSep {
     	p.AbsFP = trimPathSepSuffix(p.AbsFP)
	}
     if p.RelFP != "" && p.RelFP != "/" && p.RelFP != PathSep {
     	p.RelFP = trimPathSepSuffix(p.RelFP)
	}
     if p.ShortFP != "" && p.ShortFP != "/" && p.ShortFP != PathSep {
     	p.ShortFP = trimPathSepSuffix(p.ShortFP)
	}
}

func trimPathSepSuffix(s string) string {
     if	S.HasSuffix(s, "/")     { s = s[0:len(s)-1] }
     if	S.HasSuffix(s, PathSep) { s = s[0:len(s)-1] }
     return s
}

func (p *Filepaths) EnsurePathSepSuffixes() {
     if p.AbsFP != "" && p.AbsFP != "/" && p.AbsFP != PathSep {
     	p.AbsFP = ensurePathSepSuffix(p.AbsFP)
	}
     if p.RelFP != "" && p.RelFP != "/" && p.RelFP != PathSep {
     	p.RelFP = ensurePathSepSuffix(p.RelFP)
	}
     if p.ShortFP != "" && p.ShortFP != "/" && p.ShortFP != PathSep {
     	p.ShortFP = ensurePathSepSuffix(p.ShortFP)
	}
}

func ensurePathSepSuffix(s string) string {
     if	!(S.HasSuffix(s, "/") || S.HasSuffix(s, PathSep)) {
     	s += "/"
	}
     return s
}


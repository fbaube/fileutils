package fileutils

import(
	"fmt"
	"io/fs"
	"errors"
	S "strings"
	FP "path/filepath"
	SU "github.com/fbaube/stringutils"
)

// Path validity and path locality are very closely related.

// Notes about ValidPath, from https://pkg.go.dev/io/fs#ValidPath:
//  - func ValidPath(name string) bool
//  - It reports whether the given path name is valid for use in a call to Open.
//  - Path names passed to open are UTF-8-encoded, UNROOTED (i.e. relative), 
//    slash-separated sequences of path elements, like “x/y/z”.
//  - Path names must not contain an element that is “.” or “..” or the empty
//    string, except for the special case that the name "." may be used for
//    the [local] root directory. (Maybe use func path.Clean(path string) ?)
//  - Paths must not start or end with a slash: “/x” and “x/” are invalid.
//  - Note that paths are slash-separated on all systems, even Windows.
//    Paths containing other characters such as backslash and colon are
//    accepted as valid, but those characters must never be interpreted
//    by an FS implementation as path element separators.

// Notes about Local, from https://go.dev/blog/osroot and Go API docs:
//  - Func path/filepath.IsLocal reports whether a path is “local”,
//    i.e. one which:
//  - * Does not escape the directory in which it is evaluated
//      ("../etc/passwd" is not allowed), i.e. it is within the
//      subtree rooted at the directory in which path is evaluated
//  - * Is not an absolute path ("/etc/passwd" is not allowed);
//  - * Is not empty ("" is not allowed);
//  - * On Windows, is not a reserved name (“COM1” is not allowed).
//  - It use lexical processing only, i.e. it does not account for
//    the effect of any symbolic links in the filesystem.
//  - If IsLocal(path) returns true, then
//  - * FP.Join(base, path) is always a path contained within base, and
//  - * FP.Clean(path) is always an unrooted path with no ".." path elements.
//  - Func path/filepath.Localize converts a /-separated path into
//    a local operating system path; the input path must be a valid
//    path as reported by io/fs.ValidPath; it returns an error if
//    the path cannot be represented by the operating system.
//  - A program may defend against unintended symlink traversal 
//    by using func path/filepath.EvalSymlinks to resolve links in
//    untrusted names before validation, but this two-step process
//    is vulnerable to TOCTOU races.
//  - Func path/filepath.EvalSymlinks returns the path name after
//    evaluating of any symbolic links; if path is relative, the
//    result will be relative to the current directory, unlessc one
//    of the components is an absolute symbolic link; EvalSymlinks
//    calls Clean on the result.
//  - A program that takes potentially attacker-controlled paths
//    should almost always use filepath.IsLocal or filepath.Localize
//    to validate or sanitize those paths.
//  - os.Root ensures that symbolic links pointing outside the root
//    directory cannot be followed, providing additional security.


// Filepath has three paths, and tipicly all three are set, even if the 
// third([ShortFP]) is normally session-specific. Note that directories 
// always have a slash (or OS sep) appended, and symlinks never should.
//
// Input from os.Stdin probably uses a local file to capture the input,
// and that file will need special handling. 
//
// Note that the file name (aka [FP.Base], the part of the full path 
// after the last directory separator) is not stored separately: it 
// is stored in both AbsFP and RelFP. Note also that all this path 
// and name information duplicates what is stored in an instance of
// [orderednodes.Nord] .
// . 
type Filepaths struct {
     // RelFP is tipicly the path given (e.g.) on the command line, and is
     // useful for resolving relative paths in batches of content items.
     // The value might be valid only for the current CLI invocation or
     // user session, but it is persistable to preserve relationships
     // among files in import batches. 
     RelFP string
     // AbsFP is the authoritative field when processing individual files. 
     AbsFP string
     // GotAbs (from [path/filepath/IsAbs]) says that this struct was
     // created using an absolute FP, not a relative FP, and so the
     // field [RelFP] is calculated. 
     GotAbs bool
     // Local (from func [path/filepath/IsLocal]) is OK; not-Local might
     // be a security hole.
     Local bool
     // Valid (from func [path/filepath/ValidPath]) fails for absolute 
     // paths, but can be set to `true` for them. 
     Valid bool
     // ShortFP is the path shortened by using "." (CWD) or "~" (user's
     // home directory), so it might only be valid for the current CLI
     // invocation or user session and it is def not persistable. 
     ShortFP string
}

func (p *Filepaths) String() string {
     	var src = "rel"
	if p.GotAbs { src = "abs" }
	return fmt.Sprintf("FPs(%s)%s:%s", src, p.RelFP, p.AbsFP)
}

// NewFilepaths relies on the std lib, and accepts
// either an absolute or a relative filepath. It
// does not, however, accept an empty filepath.
// 
// It takes care to remove a trailing slash (or OS
// sep) before calling functions in [path/filepath],
// so that symlinks are not unintentionally followed.
//
// NOTE that the stdlib funcs called here (Valid, IsLocal)
// reject absolute filepaths, so it might be better to
// call this with a relative filepath when possible. 
//
// Possible error returns: input filepath is... 
//  - empty (0-length) 
//  - neither absolute nor [fs.ValidPath] 
//  - failing in a call to [path/filepath.Abs]
// 
// Ref: type PathError struct {	Op string Path string Err error }
// .
func NewFilepaths(anFP string) (*Filepaths, error) {
     var pFPs *Filepaths 
     if anFP == "" {
     	return nil, errors.New("newfilepaths: empty path")
	} 
     // Normalize it ("using only lexical analysis") 
     anFP = FP.Clean(anFP)
     // Allocate the storage now, to use the flag fields. 
     pFPs = new(Filepaths)
     // Validate it (altho we expect this 
     // call to fail if it is an absolute FP) 
     pFPs.Valid = fs.ValidPath(anFP)
     // Check whether it is local.
     pFPs.Local = FP.IsLocal(anFP) 
     pFPs.GotAbs = FP.IsAbs(anFP)
     // fmt.Fprintf(os.Stderr, "<%s> Abs<%t> Local<%t> Valid <%t> \n",
     //	 anFP, pFPs.GotAbs, pFPs.Local, pFPs.Valid)
     if pFPs.GotAbs {
     	if pFPs.Valid { println("Abs is valid ?!:", anFP) } /* else
	// Comment this out, cos it does not help. 
	 { println("Abs.FP is invalid, as expected:", anFP) } // ;panic("OOPS")}
	*/ } 
     if !(pFPs.Valid || pFPs.GotAbs) {
     	return nil, &fs.PathError{
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
	pFPs.AbsFP, e = FP.Abs(anFP) // need not be PathError 
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


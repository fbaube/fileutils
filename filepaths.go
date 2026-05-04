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

// Path validity and path locality are very closely related.

// Notes about ValidPath, from https://pkg.go.dev/io/fs#ValidPath:
//  - func ValidPath(name string) bool
//  - It says whether the given path name is valid for use in a call to Open.
//  - Path names passed to open are UTF-8-encoded, UNROOTED (i.e. relative), 
//    slash-separated sequences of path elements, like “x/y/z”.
//  - Path names must not contain an element that is “.” or “..” or the empty
//    string, except for the special case that the name "." may be used for
//    the [local] root directory. (Maybe use func path.Clean(path string) ?)
//  - Paths must not start or end with a slash: “/x” and “x/” are invalid.
//    (Note that the latter conflicts with our app's internal rules.) 
//  - Note that paths are slash-separated on all systems, even Windows.
//    Paths containing other characters such as backslash and colon are
//    accepted as valid, but those characters must never be interpreted
//    by an FS implementation as path element separators.

// Notes about Local (the concept is like [os.Root]), 
// from https://go.dev/blog/osroot and Go API docs:
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
// SHOULD always have a slash (or OS sep) appended, and that symlinks
// never should.
//
// It should execute fairly quickly, because it probably only needs
// to read the inode, not any disk sectors containing content.
// 
// Input from os.Stdin will probably normally use a local file to
// capture and store the input, and such a file needs special handling. 
//
// Note that the file name (aka [FP.Base], the part of the full path 
// after the last directory separator) is not stored separately: it 
// is stored in both AbsFP and RelFP.
//
// The truth values of fields `IsExist` and `IsDir` are subject to
// change by the actions of a caller, so it is up to the caller to
// use and maintain the fields properly. 
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
     // Errer helps sometimes.
     Errer 
     // GotAbs (from [path/filepath/IsAbs]) says that this struct 
     // was created using an absolute FP, not a relative FP, and 
     // so the field [RelFP] is calculated. 
     GotAbs bool
     // DoesNotExist is made very visible. If a path does not 
     // exist, then as noted in func `NewFilepaths`,
     //  - the error (a `*PathError`) is put in field `Errer`
     //  - the field `DoesNotExist` is set to `true`
     //  - no other fields in the struct are set, including paths
     DoesNotExist bool
     IsDir  bool
     IsFile  bool
     IsSymlink bool 
     IsDirlike bool // dir or symlink 
     // Local (from func [path/filepath/IsLocal]) means OK; 
     // not-Local might flag the possibility of a security hole.
     IsLocal bool
     // Valid (from func [path/filepath/ValidPath]) fails for 
     // absolute paths, but can be set to `true` for them. 
     IsValid bool
     // ShortFP is the path shortened by using "." (CWD) or "~" (user's
     // home directory), so it might only be valid for the current CLI
     // invocation or user session and it is def not persistable. 
     ShortFP string
}

func (p *Filepaths) String() string {
     	var src = "rel"
	if p.GotAbs { src = "abs" }
	return fmt.Sprintf("FPs(%s)%s:%s(ex:%s,dir:%s)",
	       src, p.RelFP, p.AbsFP, SU.Yn(
	       !p.DoesNotExist), SU.Yn(p.IsDir))
}

// NewFilepaths turns a filesystem path into a struct with
// abs & rel paths, and flags for existence, isaDirectory,
// isDirlike, isaFile, isaSymlink. It relies on the std lib,
// and accepts either an absolute or a relative filepath. It
// does not, however, accept an empty filepath. It probably
// accepts ".".
//
// NOTE that if the path does NOT exist then this func
// returns AN ERROR but it DOES provide precise info:
//  - it sets field [DoesNotExist] AND
//  - it puts the error [fs.ErrNotExist] in field [Errer]
//  - other errors that might be returned are for files 
//    & dirs that DO exist, and they are distinguished by
//     - [DoesNotExist] remains `false` (its zero value) AND
//     - the field [Errer] is NOT [fs.ErrNotExist] 
// 
// Therefore a caller that doesn't necessarily expect anything to
// be existing yet - at its target filepath - forcing an item to 
// be created - should be ready to check that for an error return,
// the flag field `DoesNoExist` is set, and then correctly handle
// (and comprehend) the error. 
//
// NOTE that if HasError(), 
//  - the error (a `*PathError`) is put in field `Errer`
//  - no other fields in the struct are set, including paths 
// 
// Possible errors: input filepath is...
//  - non-existent (or other error from `Lstat`) 
//  - neither absolute nor [fs.ValidPath] 
//  - failing in a call to [path/filepath.Abs]
// 
// It takes care to remove a trailing slash (or OS sep)
// before calling functions in [path/filepath], so that
// symlinks are not unintentionally followed.
//
// NOTE that some of the stdlib funcs called here (`Valid`,
// `IsLocal`) reject absolute filepaths, so it's probably 
// better to call this with a relative filepath when possible. 
//
// Ref: type PathError struct {	Op string; Path string; Err error }
// .
func NewFilepaths(anFP string) *Filepaths {
     var pFPs *Filepaths
     var pPE *os.PathError 
     // Normalize the FP ("using only lexical analysis") 
     anFP = FP.Clean(anFP)
     // Allocate a PathError now, just in case.
     pPE = new(os.PathError)
     pPE.Path = anFP
     pPE.Op = "newfilepaths: "
     // Allocate the storage now, so we can 
     // use the fields `Errer` and flags.
     pFPs = new(Filepaths)
     if anFP == "" {
     	pFPs.SetError(errors.New("empty path"))
     	return pFPs
	}
     // -----------
     //  Let's GO!     
     // -----------
     // We do not want to accidentally follow symlinks. 
     // If the path is a directory and needs a trailing 
     // slash (or OS sep), it will be added later. 
     pFPs.TrimPathSepSuffixes()
     // func Lstat(name string) (FileInfo, error)
     // returns a FileInfo describing the named file. 
     // If the file is a symlink, the returned FileInfo 
     // describes the symlink; Lstat does not try to 
     // follow the link. Any error is of type *PathError.
     pFI, e := os.Lstat(anFP)
     if e != nil {
	     // Do not set ANY other fields in pFPs.
	     pPE.Op += "Lstat" 
	     pFPs.SetError(e) 
	     if errors.Is(e, fs.ErrNotExist) {
		pFPs.DoesNotExist = true
	     } 
	     return pFPs 
     	}
     // ---------------------------
     //  Basic checks using stdlib 
     // ---------------------------
     // Validate it (altho we expect this 
     // call to fail if it is an absolute FP) 
     pFPs.IsValid = fs.ValidPath(anFP)
     // Check whether it is local.
     pFPs.IsLocal = FP.IsLocal(anFP) 
     pFPs.GotAbs = FP.IsAbs(anFP)
     // "It's an absolute path" and "It is a valid path"
     // should be mutually exclusive. I think. 
     if pFPs.GotAbs {
     	if pFPs.IsValid { println("Abs is valid ?!:", anFP) } /* else
	// Comment this out, cos it does not help. 
	 { println("Abs.FP is invalid, as expected:", anFP) } // ;panic("OOPS")}
	*/ } 
     if !(pFPs.IsValid || pFPs.GotAbs) {
     	pPE.Op = "fs.ValidPath(RelFP)"
	pPE.Err = fs.ErrInvalid
	pFPs.SetError(pPE)
	return pFPs
	}
     // --------------------------
     //  Use the FileInfo in *pFI 
     //  for setting other flags 
     // --------------------------
//   pFPs.IsNotExist = PathExists(string)
     pFPs.IsFile = pFI.Mode().IsRegular() && !pFI.IsDir()
     pFPs.IsDir  = pFI.IsDir()
     pFPs.IsSymlink = (0 != (pFI.Mode() & os.ModeSymlink))
     pFPs.IsDirlike = pFPs.IsDir || pFPs.IsSymlink
     // fmt.Fprintf(os.Stderr, "<%s> Abs<%t> Local<%t> Valid <%t> \n",
     //	 anFP, pFPs.GotAbs, pFPs.Local, pFPs.Valid)
     
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
	   pPE.Op = "FP.Abs"
	   pPE.Err = e
	   pFPs.SetError(pPE)
	   return pFPs
	}
     	pFPs.ShortFP = SU.Tildotted(pFPs.AbsFP)
     }
     return pFPs 
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


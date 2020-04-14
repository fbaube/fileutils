package fileutils

import (
	"fmt"
	"os"
	FP "path/filepath"
	S "strings"
	SU "github.com/fbaube/stringutils"
)

// NOT re-entrant
var gotOkayExts bool
// NOT re-entrant
var theOkayExts []string
// NOT re-entrant
var gotOkayName bool
// NOT re-entrant
var theOkayName string
// NOT re-entrant
var theOkayFiles []AbsFilePath

// FileWalkInfo records the arguments passed to every call
// to a "filepath.WalkFunc" that refers to a valid file.
//
// NOTE that if an error is passed in, something is pretty
// messed up. In principle we could still record the call,
// but the logic is complex, so instead we just print an
// error message to the console, return, and carry on with
// other calls.
//
// NOTE that thruout this package, the following
// are *invalid* files that are NOT recorded:
//  * directories
//  * files that fail `FileInfo.Mode().IsRegular()`
//  * emacs backup files (suffixed "~")
//  * dotfiles (prefixed ".")
//  * the contents (recursively downward) of dot folders (prefixed ".")
//
type FileWalkInfo struct {
	Path AbsFilePath
	Info os.FileInfo
	Errg error
}

// ListFilesUnder normally handles the case where "path" is a
// directory (but "path" may also be a simple file argument).
// It "error" is not nil, a message is printed to the user
// and the file in question is not added to the "FileSet".
//
func ListFilesUnder(path string, useSymLinks bool) (FS *FileSet, err error) {
	if path == "" {
		return nil, nil
	}
	FS = new(FileSet)
	FS.DirSpec = *NewBasicPath(path)
	FS.FilePaths = make([]string, 0, 10)
	FS.CheckedFiles = make([]BasicPath, 0, 10)
	// A single file ? If so, don't even bother to check it :)
	if !FS.DirSpec.IsOkayDir() { // PathType() != "DIR" { // !DirExists(FS.AbsFilePath) {
		println("==> Warning: not a directory:", path)
		pF, e := os.Open(FS.DirSpec.AbsFilePath.S())
		defer pF.Close()
		if e != nil {
			return nil, e
		}
		FS.FilePaths = append(FS.FilePaths, path)
		cp := NewBasicPath(path)
		FS.CheckedFiles = append(FS.CheckedFiles, *cp) // *NewCheckedPath(path))
		return FS, nil
	}
	// PROCESS THE DIRECTORY
	err = FP.Walk(path, func(P string, I os.FileInfo, E error) error {
		// println("WALKER:", P)
		// Don't let an error stop processing,
		// but OTOH let's notify the user.
		if E != nil {
			println(fmt.Sprintf("%s: %s", P, E.Error()))
			return nil
		}
		// Ignore odd stuff, except if we are following symlinks.
		if !I.Mode().IsRegular() {
			m := I.Mode()
			var isSymLink bool
			isSymLink = 0 != (m & os.ModeSymlink)
			if isSymLink {
				println("found symlink:", P)
			}
			if isSymLink && useSymLinks {
				// FIXME
				// println("... following it")
			} else {
				return nil
			}
		}
		// Ignore dot files, and completely skip dot folders.
		filnam := FP.Base(P)
		if S.HasPrefix(filnam, ".") {
			if AbsFilePath(P).DirExists() {
				return FP.SkipDir
			}
			return nil
		}
		// Ignore emacs backup files
		if S.HasSuffix(filnam, "~") {
			if AbsFilePath(P).DirExists() {
				return FP.SkipDir
			}
			return nil
		}
		FS.FilePaths = append(FS.FilePaths, P)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("fu.ListFilesUnder<%s>: %w", path, err)
	}
	return FS, nil
}

// GatherInputFiles handles the case where "afp" is a directory
// (but it can also handle "path" being a simple file argument).
// It always excludes dotfiles (filenames that begin with "."
// and are not the current "." or parent ".." directory) and
// emacs backups (filenames that end with "~").
//
// It includes only files that end with any extension in the slice
// arg "okayExts" (the check is case-insensitive). Each extension
// in the slice argument should include the period; the function
// will get additional functionality if & when the periods are not
// included. If "okayExts" is nil, ALL file extensions are permitted.
func (afp AbsFilePath) GatherInputFiles(okayExts []string) (okayFiles []AbsFilePath, err error) {
	if afp == "" {
		return nil, nil
	}
	theOkayExts = okayExts
	gotOkayExts = (okayExts != nil && len(okayExts) > 0)
	gotOkayName = false
	// NOTE Must clear theOkayFiles between calls !
	// "nil" is kosher and releases the contents to garbage collection.
	theOkayFiles = nil

	// A single file ? If so, just check the fie extension.
	spath := afp.S()
	if !afp.DirExists() {
		sfx := FP.Ext(spath)
		if SU.IsInSliceIgnoreCase(sfx, okayExts) || !gotOkayExts {
			abs, _ := FP.Abs(spath)
			theOkayFiles = append(theOkayFiles, AbsFilePath(abs))
		}
		return theOkayFiles, nil
	}
	// PROCESS THE DIRECTORY
	err = FP.Walk(spath, myWalkFunc)
	if err != nil {
		return nil, fmt.Errorf("fu.GatherInputFiles.walkTo<%s>: %w", afp, err)
	}
	return theOkayFiles, nil
}

// GatherNamedFiles handles the case where "afp" is a file name. Whether
// "afp" includes a dot and/or a file extension makes no difference.
func (afp AbsFilePath) GatherNamedFiles(name string) (okayFiles []AbsFilePath, err error) {
	if afp == "" || name == "" {
		return nil, nil
	}
	gotOkayName = true
	theOkayName = name
	gotOkayExts = false
	// NOTE Must clear theOkayFiles between calls !
	// "nil" is kosher and releases the contents to garbage collection.
	theOkayFiles = nil

	// A directory ?
	if !afp.DirExists() {
		panic(fmt.Sprintf("fu.GatherNamedFiles.walkTo<%s:%s>", afp, name))
	}
	// PROCESS THE DIRECTORY
	err = FP.Walk(afp.S(), myWalkFunc)
	if err != nil {
		return nil, fmt.Errorf("fu.GatherNamedFiles.walkTo<%s>: %w", afp, err)
	}
	return theOkayFiles, nil
}

func myWalkFunc(path string, finfo os.FileInfo, inerr error) error {
	var abspath AbsFilePath
	var f *os.File
	var e error
	// Do we get a REL or ABS ?
	if !S.HasPrefix(path, PathSep) {
		panic("myWalkFunc got rel not abs: " + path)
	}
	// print("path|" + path + "|finfo|" + finfo.Name() + "|\n")
	if !S.HasSuffix(path, finfo.Name()) &&
		!S.HasSuffix(path, finfo.Name()+PathSep) {
		panic("fu.myWalkFunc<" + path + ":" + finfo.Name() + ">")
	}
	if inerr != nil {
		return fmt.Errorf("fu.myWalkFunc<%s>: %w", path, inerr)
	}
	// Is it hidden, or an emacs backup ? If so, ignore.
	if S.HasPrefix(path, ".") || S.HasSuffix(path, "~") {
		if finfo.IsDir() {
			return FP.SkipDir
		} else {
			return nil
		}
	}
	// Is it a (non-hidden) directory ? If so, carry on.
	if finfo.IsDir() {
		return nil
	}
	// Now the real tests
	if gotOkayName {
		if !S.HasSuffix(path, PathSep+theOkayName) {
			return nil
		}
	} else if gotOkayExts && !SU.IsInSliceIgnoreCase(FP.Ext(path), theOkayExts) {
		return nil
	}
	apath, e := FP.Abs(path)
	abspath = AbsFilePath(apath)
	if e != nil {
		return fmt.Errorf("fu.myWalkFunc<%s>: %w", path, e)
	}
	f = Must(OpenRO(abspath.S()))
	defer f.Close() /*
		if e != nil {
			return errors.Wrapf(e, "fu.myWalkFunc.MustOpenRO<%s>", path)
		} */
	// fmt.Printf("(DD) Infile OK: %+v \n", abspath)
	theOkayFiles = append(theOkayFiles, abspath)
	return nil
}

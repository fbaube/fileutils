package fileutils

import (
	"fmt"
	"os"
	fp "path/filepath"
	S "strings"

	SU "github.com/fbaube/stringutils"
	"github.com/pkg/errors"
)

var gotOkayExts bool
var theOkayExts []string
var gotOkayName bool
var theOkayName string
var theOkayFiles []AbsFilePath

// FileWalkInfo records the arguments passed to every call
// to a `filepath.WalkFunc` that refers to a valid file.
//
// NOTE that if an error is passed in, something is pretty
// messed up. In principle we could still record the call,
// but the logic is complex, so instead we just print an
// error message to the console, return, and carry on with
// other calls.
//
// NOTE that thruout this package, the following
// are *invalid files* that are *not* recorded:
// * directories
// * files that fail `FileInfo.Mode().IsRegular()`
// * emacs backup files (suffixed "~")
// * dotfiles (prefixed ".")
// * the contents (recursively downward) of dot folders (prefixed ".")
//
type FileWalkInfo struct {
	Path AbsFilePath
	Info os.FileInfo
	Errg error
}

// ListFilesUnder normally handles the case where `path` is a
// directory (but `path` may also be a simple file argument).
// It `error` is not nil, a message is printed to the user
// and the file in question is not added to the `FileSet`.
//
func ListFilesUnder(path string, useSymLinks bool) (FS *FileSet, err error) {
	if path == "" {
		return nil, nil
	}
	FS = new(FileSet)
	FS.RelFilePath = RelFilePath(path)
	FS.AbsFilePath = FS.RelFilePath.AbsFP()
	FS.FilePaths = make([]string, 0, 10)
	// A single file ? If so, don't even bother to check it :)
	if !DirExists(FS.AbsFilePath) {
		pF, e := os.Open(string(FS.AbsFilePath))
		defer pF.Close()
		if e != nil {
			return nil, e
		}
		FS.FilePaths = append(FS.FilePaths, path)
		return FS, nil
	}
	// PROCESS THE DIRECTORY
	err = fp.Walk(path, func(P string, I os.FileInfo, E error) error {
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
		filnam := fp.Base(P)
		if S.HasPrefix(filnam, ".") {
			if DirExists(AbsFilePath(P)) {
				return fp.SkipDir
			}
			return nil
		}
		// Ignore emacs backup files
		if S.HasSuffix(filnam, "~") {
			if DirExists(AbsFilePath(P)) {
				return fp.SkipDir
			}
			return nil
		}
		FS.FilePaths = append(FS.FilePaths, P)
		return nil
	})
	if err != nil {
		return nil, errors.Wrapf(err, "fu.ListFilesUnder<%s>", path)
	}
	return FS, nil
}

// GatherInputFiles handles the case where `path` is a directory
// (but it can also handle `path` being a simple file argument).
// It always excludes dotfiles (filenames that begin with "."
// and are not the current "." or parent ".." directory) and
// emacs backups (filenames that end with "~").
//
// It includes only files that end with any extension in the slice
// `okayExts` (the check is case-insensitive). Each extension in the
// slice argument should include the period; the function will get
// additional functionality if & when the periods are not included.
// If `okayExts` is nil, *all* file extensions are permitted.
func GatherInputFiles(path AbsFilePath, okayExts []string) (okayFiles []AbsFilePath, err error) {
	if path == "" {
		return nil, nil
	}
	theOkayExts = okayExts
	gotOkayExts = (okayExts != nil && len(okayExts) > 0)
	gotOkayName = false
	// NOTE Must clear theOkayFiles between calls !
	// "nil" is kosher and releases the contents to garbage collection.
	theOkayFiles = nil

	// A single file ? If so, just check the fie extension.
	if !DirExists(path) {
		sfx := fp.Ext(string(path))
		if SU.IsInSliceIgnoreCase(sfx, okayExts) || !gotOkayExts {
			abs, _ := fp.Abs(string(path))
			theOkayFiles = append(theOkayFiles, AbsFilePath(abs))
		}
		return theOkayFiles, nil
	}
	// PROCESS THE DIRECTORY
	err = fp.Walk(string(path), myWalkFunc)
	if err != nil {
		return nil, errors.Wrapf(err, "fu.GatherInputFiles.walkTo<%s>", path)
	}
	return theOkayFiles, nil
}

// GatherNamedFiles handles the case where `path` is a file name.
// Whether `path` includes a dot and/or a file extension makes no
// difference.
func GatherNamedFiles(path AbsFilePath, name string) (okayFiles []AbsFilePath, err error) {
	if path == "" || name == "" {
		return nil, nil
	}
	gotOkayName = true
	theOkayName = name
	gotOkayExts = false
	// NOTE Must clear theOkayFiles between calls !
	// "nil" is kosher and releases the contents to garbage collection.
	theOkayFiles = nil

	// A directory ?
	if !DirExists(path) {
		panic(fmt.Sprintf("fu.GatherNamedFiles.walkTo<%s:%s>", path, name))
	}
	// PROCESS THE DIRECTORY
	err = fp.Walk(string(path), myWalkFunc)
	if err != nil {
		return nil, errors.Wrapf(err, "fu.GatherNamedFiles.walkTo<%s>", path)
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
		return errors.Wrapf(inerr, "fu.myWalkFunc<%s>", path)
	}
	// Is it hidden, or an emacs backup ? If so, ignore.
	if S.HasPrefix(path, ".") || S.HasSuffix(path, "~") {
		if finfo.IsDir() {
			return fp.SkipDir
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
	} else if gotOkayExts && !SU.IsInSliceIgnoreCase(fp.Ext(path), theOkayExts) {
		return nil
	}
	apath, e := fp.Abs(path)
	abspath = AbsFilePath(apath)
	if e != nil {
		return errors.Wrapf(e, "fu.myWalkFunc<%s>", path)
	}
	f, e = TryOpenRO(abspath)
	defer f.Close()
	if e != nil {
		return errors.Wrapf(e, "fu.myWalkFunc.MustOpenRO<%s>", path)
	}
	// fmt.Printf("(DD) Infile OK: %+v \n", abspath)
	theOkayFiles = append(theOkayFiles, abspath)
	return nil
}
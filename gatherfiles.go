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
	if !IsDirectory(path) {
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
	if !IsDirectory(path) {
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

package fileutils

import (
	S  "strings"
	FP "path/filepath"
)

// AbsFilePath is a new type, based on `string`. It serves three purposes:
// - clarify and bring correctness to the processing of absolute path arguments
// - permit the use of a clearly named struct field
// - permit the definition of methods on the type
//
// Note that when working with an `os.File`, `Name()` returns the name of the
// file as was passed to `Open(..)`, so it might be a relative filepath.
//
type AbsFilePath string

// Some prior overenthusiasm.
// type RelFilePath string
// type ArgFilePath string
// type FileContent string

// S is a utility method to keep code cleaner.
func (afp AbsFilePath) S() string {
	s := string(afp)
	if !FP.IsAbs(s) {
		// panic("FU.types: AbsFP is not abs: " + s)
		// FIXME? // println("==> fu.types: AbsFP not abs: " + s)
		s, e := FP.Abs(s)
		if e != nil { panic("su.afp.S") }
		return s
	}
	return s
}

// AbsFP is like filepath.Abs(..) except using our own types.
func AbsFP(relFP string) AbsFilePath {
	if FP.IsAbs(relFP) {
		return AbsFilePath(relFP)
	}
	afp, e := FP.Abs(relFP)
	if e != nil {
		panic("fu.AbsFP<" + relFP + ">: " + e.Error())
	}
	return AbsFilePath(afp)
}

func (afp AbsFilePath) DirPath() AbsFilePath {
	dp, _ := FP.Split(afp.S())
	return AbsFilePath(dp)
}

func (afp AbsFilePath) BaseName() string {
	_, fm := FP.Split(afp.S())
	return S.TrimPrefix(fm, FP.Ext(afp.S()))
}

func (afp AbsFilePath) FileExt() string {
	return FP.Ext(afp.S())
}

// Append is a convenience function to keep code cleaner.
func (afp AbsFilePath) Append(rfp string) AbsFilePath {
	return AbsFilePath(FP.Join(afp.S(), rfp))
}

// StartsWith is like strings.HasPrefix(..) but uses our types.
func (afp AbsFilePath) HasPrefix(beg AbsFilePath) bool {
	return S.HasPrefix(afp.S(), beg.S())
}

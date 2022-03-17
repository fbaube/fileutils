package fileutils

import (
	"errors"
	"fmt"
	"io/fs"
)

/* 
Reference:
type PathError struct {
	Op   string
	Path string
	Err  error }
*/
// Err can contain %w but must be set by caller to NewPathPropsError(..),
// and caller can also decide whether to set pkg.filename.methodname.Lnn 
// e.g. PathError { Err: fmt.Errorf("Zork failed: %w (fu.zork.L22)", e) }

// PathPropsError is
// PathProps + SrcLoc (in source code) + 
// PathError struct { Op, Path string; Err error }
//
// Maybe use the format pkg.filename.methodname.Lnn
//
// In code where package `mcfile` is available,
// use mcfile.ContentityError 
//
type PathPropsError struct {
	PE     fs.PathError
	*PathProps
}

// WrapAsPathPropsError SHOULD USE %w 
func WrapAsPathPropsError(e error, op string, pp *PathProps) PathPropsError {
	ce := PathPropsError{}
	ce.PE.Err = e
	ce.PE.Op  = op
	if pp == nil {
		ce.PE.Path = "(path not provided)"
	} else {
		ce.PE.Path = pp.AbsFP.S()
	}
	return ce
}

// NewPathPropsError TBD. 
func NewPathPropsError(ermsg string, op string, pp *PathProps) PathPropsError {
	ce := PathPropsError{}
	ce.PE.Err = errors.New(ermsg)
	ce.PE.Op  = op
	if pp == nil {
		ce.PE.Path = "(path not provided)"
	} else {
		ce.PE.Path = pp.AbsFP.S()
	}
	return ce
}

func (ce PathPropsError) Error() string {
	return ce.String()
}

func (ce *PathPropsError) String() string {
	var s string
	s = fmt.Sprintf("%s(%s): %s", ce.PE.Op, ce.PE.Path, ce.PE.Err.Error())
	return s
}

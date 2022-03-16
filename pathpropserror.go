package fileutils

import (
	"errors"
	"fmt"
	"io/fs"
)

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
	SrcLoc string
	*PathProps
}

func WrapAsPathPropsError(e error, op string, pp *PathProps, srcLoc string) PathPropsError {
	ce := PathPropsError{}
	ce.PE.Err = e
	ce.PE.Op  = op
	ce.SrcLoc = srcLoc 
	if pp == nil {
		ce.PE.Path = "(pathprops path not found!)"
	} else {
		ce.PE.Path = pp.AbsFP.S()
	}
	return ce
}

func NewPathPropsError(ermsg string, op string, pp *PathProps, srcLoc string) PathPropsError {
	ce := PathPropsError{}
	ce.PE.Err = errors.New(ermsg)
	ce.PE.Op  = op
	ce.SrcLoc = srcLoc
	if pp == nil {
		ce.PE.Path = "(pathprops path not found!)"
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
	if ce.SrcLoc != "" {
		s += fmt.Sprintf(" (in %s)", ce.SrcLoc)
	}
	return s
}

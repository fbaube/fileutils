package fileutils

import (
	"errors"
	"fmt"
	"io/fs"
)

/* Reference:
type PathError struct {
	Op   string
	Path string
	Err  error }
*/

// Err can contain %w but must be set by caller to NewFSItemError(..),
// and caller can also decide whether to set pkg.filename.methodname.Lnn
// e.g. PathError { Err: fmt.Errorf("Zork failed: %w (fu.zork.L22)", e) }

// FSItemError is
// FSItem + SrcLoc (in source code) +
// PathError struct { Op, Path string; Err error }
//
// # Maybe use the format pkg.filename.methodname.Lnn
//
// In code where package `mcfile` is available,
// use mcfile.ContentityError
type FSItemError struct {
	PE fs.PathError
	*FSItem
}

// WrapAsFSItemError is EOP (error,op,path) but like using Printf's "%w" 
func WrapAsFSItemError(e error, op string, pfsi *FSItem) FSItemError {
	pfsie := new(FSItemError)
	pfsie.PE.Err = e
	pfsie.PE.Op = op
	if pfsi == nil {
		pfsie.PE.Path = "(path not provided)"
	} else {
		pfsie.PE.Path = pfsi.FPs.AbsFP.S()
	}
	return *pfsie
}

// NewFSItemError is EOP (error,op,path) but like errors.New or fmt.Errorf
func NewFSItemError(ermsg string, op string, pfsi *FSItem) FSItemError {
	pfsie := new(FSItemError)
	pfsie.PE.Err = errors.New(ermsg)
	pfsie.PE.Op = op
	if pfsi == nil {
		pfsie.PE.Path = "(path not provided)"
	} else {
		pfsie.PE.Path = pfsi.FPs.AbsFP.S()
	}
	return *pfsie
}

func (p FSItemError) Error() string {
	return p.String()
}

func (p *FSItemError) String() string {
	var s string
	s = fmt.Sprintf("%s(%s): %s", p.PE.Op, p.PE.Path, p.PE.Err.Error())
	return s
}

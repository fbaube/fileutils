package fileutils

import "os"

type ContentPathError struct {
	// FuncID is freeform-ish, but like "su.file.procname"
	FuncID string
	os.PathError
	// MoreInfo is optional
	MoreInfo string
	// Filename is optional
	Filename string
	// Linenumber is optinoal
	Linenumber int
}

func NewContentPathError(fid, op, path, mi string, e error) *ContentPathError {
	p := new(ContentPathError)
	p.PathError = *new(os.PathError)
	p.FuncID = fid
	p.Op = op
	p.Path = path
	p.Err = e
	p.MoreInfo = mi
	return p
}

func NewContentPathErrorFlnr(fid, op, path string, e error, mi, fn string, ln int) *ContentPathError {
	p := new(ContentPathError)
	p.PathError = *new(os.PathError)
	p.FuncID = fid
	p.Op = op
	p.Path = path
	p.Err = e
	p.MoreInfo = mi
	p.Filename = fn
	p.Linenumber = ln
	return p
}

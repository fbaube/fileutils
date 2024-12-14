package fileutils

import "io/fs"

// Errer is a struct that can be used to embed 
// an error in another struct, when we want to 
// execute (pointer) methods on a struct in the 
// style of a data pipeline, i.e. chainable, and
// executed left-to-right.
//
// We make the error public so that it is easily
// set, and so that we can wrap errors easily
// using the "%w" printf format spec.
//
// Methods are on *Errer, not Errer, so that
// modification is possible.
//
// NOTE that this resembles a Result type, 
// so we would like the API to resemble
// https://pkg.go.dev/go.bytecodealliance.org@v0.4.0/cm#Result :
//
//  - type Result
//  - func (r *Result)   Err() *Err
//  - func (r *Result) IsErr() bool
//  - func (r *Result) IsOK()  bool
//  - func (r *Result)   OK()  *OK
//  - func (r Result) Result() (ok OK, err Err, isErr bool)
//
// TODO: This is in package fileutils, so it
// should be safe to assume that an error here
// is of type [*PathError].
// .
type Errer struct {
	Err  error
	isPE bool 
}

// HasError is a convenience function.
// Since Err is publicly visible, HasError is
// not really needed, but it seems appropriate
// given that we also have func Error()
// .
func (p *Errer) HasError() bool {
	return p.Err != nil
}

// HasPathError is a convenience function.
// Since Err is publicly visible, HasPathError is
// not really needed, but it seems appropriate
// given that we also have func Error()
// .
func (p *Errer) HasPathError() bool {
	return p.Err != nil && p.isPE
}

// Error is an NPE-proof improvement
// on the standard error.Error()
// .
func (p *Errer) Error() string {
	if p.Err == nil { // !p.HasError() {
		return ""
	}
	return p.Err.Error()
}

// SetError is a convenience func because setting
// Error.Err is ugly.
// .
func (p *Errer) SetError(e error) {
	p.Err = e
	_, p.isPE = e.(*fs.PathError)
}

// GetError is a convenience func because getting
// Error.Err is ugly.
// .
func (p *Errer) GetError() error {
	return p.Err
}

// GetPathError is a convenience func because getting
// Error.Err is ugly.
// .
func (p *Errer) GetPathError() *fs.PathError {
     	if !p.isPE { return nil }
	return p.Err.(*fs.PathError)
}

func (p *Errer) ClearError() {
	p.Err = nil
	p.isPE = false 
}

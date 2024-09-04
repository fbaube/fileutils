package fileutils

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
// .
type Errer struct {
	Err error
}

// HasError is a convenience function.
// Since Err is publicly visible, HasError is
// not really needed, but it seems appropriate
// given that we also have func Error()
// .
func (p *Errer) HasError() bool {
	return p.Err != nil
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

// SerError is a convenience func because setting
// Error.Err is ugly.
// .
func (p *Errer) SetError(e error) {
	p.Err = e
}

// GetError is a convenience func because getting
// Error.Err is ugly.
// .
func (p *Errer) GetError() error {
	return p.Err
}

func (p *Errer) ClearError() {
	p.Err = nil
}

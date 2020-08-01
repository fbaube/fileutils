package fileutils

// Errable is for structs that have an embedded field "error", not visible
// outside the package. This makes it possible to carry around (and access)
// the error long after it is generated.
type Errable interface {
	HasError() bool
	// Error satisfies interface "error", but the weird thing is that "error"
	// can be nil, which is why we need "GetError()" and "SetError()".
	Error() string
	// GetError is necessary cos "Error()"" dusnt tell you whether "error"
	// is "nil", which is the indication of no error. Therefore we need
	// this function, which can actually return the telltale "nil".
	GetError() error
	// SetError lets the caller set the field "error", not visible outside.
	// Make it optional, so that "error" can be invisible outside the package.
	// SetError(e error)
}

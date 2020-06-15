package fileutils

import (
)

// BasicContent is normally a file, opened and loaded into field "Raw"
// by "func FileLoad()"", and at that point the content is fully decoupled
// from the file system.
//
type BasicContent struct {
	// bcError - if non-nil - makes methods in a chain skip their own processing.
	bcError error
	// FileCt is >1 IFF this struct refers to a directory, and multifile
	// processing is needed. In the future we might also handle wildcards.
	FileCt int
	// Raw applies to files only, not to directories or symlinks.
	Raw  string
}

// GetError is a necessary evil cos `Error()`` dusnt tell you whether
// `error` is `nil`, which is the indication of no error. Therefore
// we need this function, which can actually return the telltale `nil`.`
func (p *BasicContent) GetError() error {
	return p.bcError
}

// Error satisfies interface `error`, but the odd thing is that `error` can
// be nil, but this possibility is masked by the standard error.Error().
func (p *BasicContent) Error() string {
	if p.bcError != nil {
		return p.bcError.Error()
	}
	return ""
}

// SetError sets the bcError.
func (p *BasicContent) SetError(e error) {
	p.bcError = e
}

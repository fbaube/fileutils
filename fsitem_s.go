package fileutils

import (
	"fmt"
)

func (p *FSItem) String() (s string) {
     return p.Info()
}

// Echo implements [Stringser].
func (p *FSItem) Echo() string {
	return p.FPs.AbsFP.S()
}

// Info implements [Stringser].
func (p *FSItem) Info() string {
        var s string 
	if p.IsFile() {
		s = fmt.Sprintf("File[L:%d] ", p.Size())
	} else if p.IsDir() {
		s = fmt.Sprintf("Dirr[L:%d] ", p.Size())
	} else if p.IsSymlink() {
		s = "Symlink "
	} else {
		s = "FSItem:?uninitialized "
	}
	s += p.FPs.ShortFP
	return s
}

// Debug implements [Stringser].
func (p *FSItem) Debug() string {
	return p.Info()
}


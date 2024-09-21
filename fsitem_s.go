package fileutils

import (
	"fmt"
)

func (p *FSItem) String() (s string) {
     return p.Info()
}

func (p *FSItem) StringWithPermissions() (s string) {
     return p.Info() + " " + p.Perms
}

// Echo implements [Stringser].
func (p *FSItem) Echo() string {
	return p.FPs.AbsFP
}

// Info implements [Stringser].
func (p *FSItem) Info() string {
        var s string 
	if p.IsFile() {
		s = fmt.Sprintf("File[len:%d] ", p.fi.Size())
		// panic("DERF")
	} else if p.fi.IsDir() {
		s = fmt.Sprintf("Dirr[len:%d] ", p.fi.Size())
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


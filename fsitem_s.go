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

// ListingString prints:
// rwx,rwx,rwx [or not exist] ... 
// Rawtype Len Abs/Rel Local? Valid? Name nLinks? Error? \n 
func (p *FSItem) ListingString() string {
     // If this gets an error, the error should
     // have been set already via interface Errer
     p.LoadContents()
     var rtp, nlinks, err string
     var local, valid = "-", "-"
     var absrel = "r"
     if p.FPs.GotAbs { absrel = "a" }
     if p.TypedRaw != nil { rtp = string(p.TypedRaw.Raw_type) }
     if p.NLinks > 1 { nlinks = fmt.Sprintf("%d", p.NLinks) } 
     err = p.Error()
     if p.FPs.Local { local = "L" }
     if p.FPs.Valid { valid = "V" }
     return fmt.Sprintf("%s %s %4d %s%s%s %s %s %s",
     	    p.Perms, rtp, p.FI.Size(), absrel, local,
	    valid, p.FI.Name(), nlinks, err)

}

// Echo implements [Stringser].
func (p *FSItem) Echo() string {
	return p.FPs.AbsFP
}

// Info implements [Stringser].
func (p *FSItem) Info() string {
        var s string 
	if p.IsFile() {
		s = fmt.Sprintf("File[len:%d] ", p.FI.Size())
		// panic("DERF")
	} else if p.FI.IsDir() {
		s = fmt.Sprintf("Dirr[len:%d] ", p.FI.Size())
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

func boolstring(b bool) string {
     if b { return "Y" }
     return "-"
}


package fileutils

import (
	"fmt"
	S "strings"
	HB "github.com/fbaube/humanbytes"
)

func (p *FSItem) String() (s string) {
     return p.Infos()
}

func (p *FSItem) StringWithPermissions() (s string) {
     return p.Infos() + " " + p.Perms
}

// ListingString prints:
// rwx,rwx,rwx [or not exist] ... 
// Rawtype (file)Len Name Error? \n 
func (p *FSItem) ListingString() string {

     // If this returns an error, it should 
     // also set the error via interface Errer,
     // so here we ignore the error return value. 
     elc := p.LoadContents()
     if elc != nil {
     	p.SetError(fmt.Errorf("LoadContents: %w", elc))
        return "ERROR:ListingString:" + elc.Error()
        }
     
     var fstp, size, err string
     fstp = S.ToUpper(string(p.FSItem_type)) 
     // if p.IsFile() { size = fmt.Sprintf("%4d", p.FI.Size()) }
     if p.IsFile() { size = fmt.Sprintf("%6s", HB.SizeSI(int(p.Size()))) }
     err = p.Error()
     return fmt.Sprintf("%s %s %s %s %s",
     	    p.Perms, fstp, size, p.Name(), err)

}

// Echo implements [Stringser].
func (p *FSItem) Echo() string {
	return p.FPs.AbsFP
}

// Infos implements [Stringser].
func (p *FSItem) Infos() string {
        var s string 
	if p.IsFile() {
		s = fmt.Sprintf("File[len:%d] ", p.Size())
		// panic("DERF")
	} else if p.IsDir() {
		s = fmt.Sprintf("Dirr[len:%d] ", p.Size())
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
	return p.Infos()
}

func boolstring(b bool) string {
     if b { return "Y" }
     return "-"
}


package fileutils

import (
	fp "path/filepath"

	SU "github.com/fbaube/stringutils"
)

// FilterInBySuffix is TBS.
func (p *FileSet) FilterInBySuffix(okayExts []string) (someOut bool) {
	// println("FilterInBySuffix")
	if okayExts == nil || len(okayExts) == 0 {
		return false
	}
	if p.Size() == 0 {
		return false
	}
	outs := make([]string, 0, p.Size())
	for _, instring := range p.FilePaths {
		sfx := fp.Ext(instring)
		// println("sfx:", sfx)
		if !SU.IsInSliceIgnoreCase(sfx, okayExts) {
			someOut = true
		} else {
			// println("OK input:", sfx, instring)
			outs = append(outs, instring)
		}
	}
	if someOut {
		p.FilePaths = outs
	}
	return someOut
}

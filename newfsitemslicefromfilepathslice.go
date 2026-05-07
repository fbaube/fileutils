package fileutils

import (
	"io/fs"
	S "strings"

	SU "github.com/fbaube/stringutils"
	CT "github.com/fbaube/ctoken"
)

// NewFSObjectSliceFromFilepathSlice creates a [*FSObject] for every input string,
// and itself has no error return value, because FSObject implements [Errer], so
// that every entry is non-nil and either is valid or has its own error message.
//
// It does tho also return summary statistics, so that is is easy to 
// check whether any errors at all were encountered. 
//
// The func assumes that entries have been "sanitized" by collecting 
// them using [os.Root], so it does check the security of symlinks, 
// or use funcs [io/fs.ValidPath] or [path/filepath.IsLocal].
//
// Accumulated NewContentity errors are counted
// in the field CotentityFS.nErrors 
// .
func NewFSObjectSliceFromFilepathSlice(FPs [] string) ([]*FSObject, *FSObjectSummaryStats) {
	var FSIs []*FSObject
	var pFSS = new(FSObjectSummaryStats)

	for _, sFP := range FPs {
	    pFSI := NewFSObject(sFP)
	    FSIs = append(FSIs, pFSI)
	    // Categorise the FSObject (file, dir, wotev).
	    // Note the hacks to TypedRaw.
	    if pFSI.IsDir() {
                if pFSI.TypedRaw == nil {
                   pFSI.TypedRaw = new(CT.TypedRaw)
                   } 
                pFSI.TypedRaw.Raw_type = SU.Raw_type_DIRLIKE
                pFSS.NrDirs++ 
            } else if pFSI.Type() == 0 { // regular file
                pFSS.NrFiles++ 
            } else if (pFSI.Type() & fs.ModeSymlink) != 0 { // Symlink
                if pFSI.TypedRaw == nil {
                   pFSI.TypedRaw = new(CT.TypedRaw)
                }
                pFSI.TypedRaw.Raw_type = SU.Raw_type_DIRLIKE 
                pFSS.NrSymLs++ 
                // L.L.Okay("Item (SYML) OK: what to do ?!")
            } else { // Some weirdness in the Mode bits 
                pFSI.TypedRaw.Raw_type = SU.Raw_type_DIRLIKE
                pFSS.NrMiscs++ 
                // pFSS.nErrors++
            }
	    if pFSI.HasError() {
	       pFSS.NrErrors++
	    } else
	      if pFSI.IsDir() && !S.HasSuffix(sFP, "/") {
	   // Make sure a dir has a trailing slash (assumed as path separator)
	   // inPath = FU.EnsureTrailingPathSep(inPath)
	      panic("Missing trlg path sep in NewFSObjectSliceFromFilepathSlice")
	    }
	}
	if len(FSIs) != len(FPs) {
	   panic("Mismatched lengths in NewFSObjectSliceFromFilepathSlice")
	}
	return FSIs, pFSS
}


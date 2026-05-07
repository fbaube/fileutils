package fileutils

// FSObjectSummaryStats is so that NrItems equals the sum of
// NrDirs + NrFiles + NrSymLs + NrMiscs; NrErrors is independent.
type FSObjectSummaryStats struct {
     NrItems, NrDirs, NrFiles, NrSymLs, NrMiscs, NrErrors int
}

/*
func (pSS *FSObjectSummaryStats) AddIn(pI *FSObject) {
	if pFSI.IsDir() {
                if pFSI.TypedRaw == nil {
                   pFSI.TypedRaw = new(CT.TypedRaw)
                   }
                pFSI.TypedRaw.Raw_type = SU.Raw_type_DIRLIKE
                pMFT.nDirs++ // just a simple counter
            } else if pFSI.Type() == 0 { // regular file
                pMFT.nFiles++ // just a simple counter
            } else if (pFSI.Type() & fs.ModeSymlink) != 0 { // Symlink
                if pFSI.TypedRaw == nil {
                   pFSI.TypedRaw = new(CT.TypedRaw)
                }
                pFSI.TypedRaw.Raw_type = SU.Raw_type_DIRLIKE // OOPS
                pMFT.nMiscs++ // just a simple counter
                L.L.Okay("Item (SYML) OK: what to do ?!")
            } else { // Some weirdness in the Mode bits
                pFSI.TypedRaw.Raw_type = SU.Raw_type_DIRLIKE
             // pMFT.nMiscs++ // just a simple counter
                pMFT.nErrors++
                L.L.E
}
*/


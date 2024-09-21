package fileutils

import(
	"os"
	"fmt"
	"errors"
	FP "path/filepath"
)

// ReadDir returns only errors from the initial step of opening the directory.
// An error returned on an individual directory item is attached to the item 
// via interface [Errer].
func ReadDir(inpath string) ([]FSItem, error) {
     if inpath == "" { return nil, errors.New("ReadDir: nil path") }
     var e error
     // Check the path
     var FPs *Filepaths
     // Ordinarily ´func NewFilepaths` should be called with
     // a relative filepath when possible, but here in this
     // use case it doesn't really bring any extra benefit. 
     FPs, e = NewFilepaths(inpath)
     if e != nil {
     	return nil, &os.PathError{ Op:"NewFilepaths", Path:inpath, Err:e }
	}
     var path = FPs.AbsFP
     var sAbsRel = "Rel" 
     if FPs.GotAbs { sAbsRel = "Abs" } 
     fmt.Printf("Readdir: %s: %s Local:%t \n", inpath, sAbsRel, FPs.Local)
     var theDir *os.File
     theDir, e = os.Open(path)
     if e != nil {
     	  return nil, &os.PathError{ Op:"Open", Path:path, Err:e }
	  }
     // [fs.FileInfo] and [fs.DirEntry] are useless as arguments 
     // to [NewFSItem], because they do not have path information.
     // Therefore we use this instead:
     //   func (f *File) Readdirnames(n int) (names []string, err error)
     // Readdirnames reads the contents of the directory associated with
     // (f *File) and returns a slice of names of files in the directory,
     // in directory order. 
     // Use with n <= 0, so that Readdirnames returns all the names from
     // the directory in a single slice. Then:
     //  - If it succeeds (reads all the way to the end of the directory),
     //    it returns the slice and a nil error.
     //  - If it encounters an error before the end of the directory,
     //    it returns the names read until that point and a non-nil error.
     var entries []string
     var FSIs []FSItem
     var pFSI  *FSItem
     entries, e = theDir.Readdirnames(-1)
     if e != nil {
     	return nil, &os.PathError{ Op: "Readdirnames", Path:path, Err:e }
	}
     for _, E := range entries {
     	    pFSI, e = NewFSItem(FP.Join(path, E))
	    if e != nil {
	       	 if pFSI == nil { pFSI = new(FSItem) }
	       	 pFSI.SetError(e)
		 }
	    FSIs = append (FSIs, *pFSI)
     }
     return FSIs, nil
}
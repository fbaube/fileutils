package fileutils

import(
	"os"
	"fmt"
	"errors"
	FP "path/filepath"
)

// ReadDir returns only errors from the initial step of opening the directory.
// An error returned on an individual directory item is attached to the items 
// via interface [Errer].
func ReadDir(inpath string) ([]FSItem, error) {
     if inpath == "" { return nil, errors.New("ReadDir: nil path") }
     var e error
     // Check the path
     var FPs *Filepaths
     FPs, e = NewFilepaths(inpath)
     if e != nil {
     	return nil, &os.PathError{ Op:"NewFilepaths", Path:inpath, Err:e }
	}
     var path = FPs.AbsFP
     var sAbsOrRel string 
     if FP.IsAbs(inpath) { sAbsOrRel = "Abs" } else { sAbsOrRel = "Rel" } 
     notLcl := !FPs.Local
     fmt.Printf("Readdir: Inpath<%s> type:%s Local:%t \n",
     		 inpath, sAbsOrRel, !notLcl)
     var theDir *os.File
     theDir, e = os.Open(path)
     if e != nil {
     	  return nil, &os.PathError{ Op:"Open", Path:path, Err:e }
	  }
     // ReadDir should not be used, cos an [fs.FileInfo] is useless as an
     // argument to [NewFSItem], cos it does not have path information.
     // Therefore we use this instead:
     //   func (f *File) Readdirnames(n int) (names []string, err error)
     // Readdirnames reads the contents of the directory associated with
     // (f *File) and returns a slice of names of files in the directory,
     // in directory order. 
     // Use with n <= 0, so that Readdirnames returns all the names from
     // the directory in a single slice. Then:
     //  - If Readdirnames succeeds (reads all the way to the end of the
     //    directory), it returns the slice and a nil error.
     //  - If it encounters an error before the end of the directory,
     //    Readdirnames returns the names read until that point and
     //    a non-nil error.
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
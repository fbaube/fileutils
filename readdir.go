
package fileutils

import(
	"os"
	"fmt"
	"errors"
	"io/fs"
	FP "path/filepath"
)

func ReadDirAsPtrs(inpath string) ([]*FSObject, error) {
     var rFSI []FSObject
     var outp []*FSObject
     rFSI, e := ReadDir(inpath)
     if rFSI == nil || len(rFSI) == 0 || e != nil { return outp, e }
     for _,fsi := range rFSI {
	// fmt.Printf("Appending: %d \n", ii)
     	outp = append(outp, &fsi) // ptr)
	}
     return outp, nil
}

// ReadDirAsMap assumes that file/dir names are case-sensitive. 
func ReadDirAsMap(inpath string) (map[string]*FSObject, error) {
     var theMap map[string]*FSObject
     var rFSI []FSObject
     rFSI, e := ReadDir(inpath)
     if rFSI == nil || len(rFSI) == 0 || e != nil { return nil, e }
     theMap = make(map[string]*FSObject)
     var p *FSObject
     for _,fsi := range rFSI {
	// fmt.Printf("Appending: %d \n", ii)
	p = new(FSObject)
	*p = fsi
     	theMap[fsi.Name()] = p
	}
     return theMap, nil
}

// ReadDir returns only errors from the initial step of opening the directory.
// An error returned on an individual directory item is attached to the item 
// via interface [Errer].
//
// It might be more useful in mamny use cases to return a slice of pointers,
// but the pattern in the stdlib is to return struct instances, not pointers.
// .
func ReadDir(inpath string) ([]FSObject, error) {
     if inpath == "" { return nil, errors.New("ReadDir: nil path") }
     var e error
     // Check the path
     var FPs *Filepaths
     // Ordinarily ´func newFilepaths` should be called with
     // a relative filepath when possible, but here in this
     // use case it doesn't really bring any extra benefit. 
     FPs = NewFilepaths(inpath)
     if FPs.HasError() {
     	return nil, &fs.PathError{
	       Op:"newFPs", Path:inpath, Err:FPs.GetError() }
	}
     var theAbsPath = FPs.AbsFP
     var sAbsRel = "Rel" 
     if FPs.GotAbs { sAbsRel = "Abs" } 
     fmt.Printf("Readdir: %s: %sPath Local:%t \n",
     		inpath, sAbsRel, FPs.IsLocal)
     var theDir *os.File
     theDir, e = os.Open(theAbsPath)
     if e != nil {
     	  return nil, &fs.PathError{ Op:"Open", Path:theAbsPath, Err:e }
	  }
     // [fs.FileInfo] and [fs.DirEntry] are useless as arguments 
     // to [NewFSObject], because they do not have path information.
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
     var FSIs []FSObject
     var pFSI  *FSObject
     entries, e = theDir.Readdirnames(-1)
     if e != nil {
     	return nil, &fs.PathError{ Op: "Readdirnames", Path:theAbsPath, Err:e }
	}
     for _, E := range entries {
     	    // NOTE this could probably be a relative path;
	    // it might or might not add value.
	    // If error, is in embedded struct Errer.
     	    pFSI = NewFSObject(FP.Join(theAbsPath, E))
	    FSIs = append (FSIs, *pFSI)
     }
     return FSIs, nil
}
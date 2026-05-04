package fileutils


import(
	"os"
	"io/fs"
)

// GetRootPaths is a convenience function. An error is a *PathError. 
func GetRootPaths(path string) (rt *os.Root, fp *Filepaths, e error) {
     	rt, e = os.OpenRoot(path)
        if e != nil {
           return nil, nil, &fs.PathError{ Path:path, Err:e,
                  Op:"fu.GetRootPaths: os.OpenRoot" }
                }
        fp = NewFilepaths(path)
        if fp.HasError() {
           return nil, nil, &fs.PathError{ Path:path, Err:fp.GetError(),
                  Op:"fu.getRootPaths.newFPs" }
        }
	return rt, fp, nil
}
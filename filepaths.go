package fileutils

import(
	"fmt"
	"io/fs"
	FP "path/filepath"
	SU "github.com/fbaube/stringutils"
)

// Filepaths shuld always have all three fields set, even if the third
// ([ShortFP]) is basically session-specific. Note that directories have
// a "/" appended. 
type Filepaths struct {
     // RelFP is tipicly the path given (e.g.) on the command line and is
     // useful for resolving relative paths in batches of content items.
     RelFP string
     // AbsFP is the authoritative field when processing individual files. 
     AbsFP AbsFilePath
     // ShortFP is the path shortened by using "." (CWD) or "~" (user's
     // home directory), so it might only be valid for the current CLI
     // invocation or user session and it is def not persistable. 
     ShortFP string
}

// NewFilepaths relies on the std lib, and accepts
// either an absolute or a relative filepath.
//
// Ref: type PathError struct {	Op string Path string Err error }
// .
func NewFilepaths(anFP string) (*Filepaths, error) {
     if anFP == "" {
     	println("NewFilepaths GOT NIL PATH")
	return nil, nil
	} 
     pFPs := new(Filepaths)
     fm, e := NewFSItemMeta(anFP)
     if e != nil {
     	return nil, fmt.Errorf("NewFilepaths<%s>: %w", anFP, e)
     }
     if fm.IsDir() { anFP = EnsureTrailingPathSep(anFP) }
     
     if FP.IsAbs(anFP) {
     	pFPs.AbsFP = AbsFilePath(anFP)
	pFPs.RelFP = SU.Tildotted(anFP) 
     } else {
        pFPs.RelFP = anFP
	// If there is a problem with the input path,
	// it should surface here.
	s, e := FP.Abs(anFP)
	if e != nil {
	   return nil, &fs.PathError{Op:"FP.AbsFP",Err:e,Path:anFP}
	}
	pFPs.AbsFP = AbsFilePath(s)
     }
     pFPs.ShortFP = SU.Tildotted(pFPs.AbsFP.S())
     return pFPs, nil
}


package fileutils

// FileSet groups a set of files that can and should be considered
// as a group. For example, when processing a multi-file document
// (LwDITA), or a multi-file DTD. It is assumed that they are
// related via a top-level directory, and it is the top-level
// directory that is contained in the field `DirSpec`.
//
// In the pathological case that this was called on a file not
// a directory, all data refer to the file path, rather than
// (say) just the directory portion.
//
type FileSet struct {
	// RelFilePath is a "short" argument such as supplied on the
	// command line; its absolute resolution is in the next field.
	// It may of course store an absolute (full) file path instead.
	// If this is "", it is not an error.
	// // RelFilePath string
	// AbsFilePath is the fully resolved counterpart to `RelFilePath`.
	// // AbsFilePath
	DirSpec PathProps
	// `filepath.WalkFunc` can provide relative filepaths, so we can't
	// say for sure whether this list will contain relative or absolute
	// paths. Therefore for convenience we use a bunch of strings.
	FilePaths []string
	// Then we process them.
	CheckedFiles []PathProps
}

// Size returns the number of files.
func (p *FileSet) Size() int {
	if p.FilePaths == nil {
		return 0
	}
	return len(p.FilePaths)
}

func NewOneFileSet(s string) *FileSet {
	pfs := new(FileSet)
	if s == "" {
		return pfs
	}
	pp, e := NewPathProps(s)
	if e != nil || pp == nil || !pfs.DirSpec.IsOkayFile() { // PathType() != "FILE" {
		panic("fu.FileSet.NewOneFS: bad file: " + s)
	}
	pfs.DirSpec = *pp
	pfs.FilePaths = make([]string, 0, 1)
	pfs.FilePaths = append(pfs.FilePaths, string(pfs.DirSpec.AbsFP))
	pfs.CheckedFiles = make([]PathProps, 0, 1)
	pfs.CheckedFiles = append(pfs.CheckedFiles, pfs.DirSpec)
	return pfs
}

package fileutils

// FileSet groups a set of files that can and shuld be considered
// as a unit. For example, when processing a multi-file document
// (LwDITA), or a multi-file DTD. It is assumed that thet are
// related via a top-level directory, and it is the top-level
// directory that is contained in the initial fields.
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
	RelFilePath
	// AbsFilePath is the fully resolved counterpart to `RelFilePath`.
	AbsFilePath
	// `filepath.WalkFunc` can provide relative filepaths, so we can't
	// say for sure whether this list will contain relative or absolute
	// paths. Therefore for convenience we use a bunch of strings.
	FilePaths []string
}

// Size returns the number of files.
func (p *FileSet) Size() int {
	if p.FilePaths == nil {
		return 0
	}
	return len(p.FilePaths)
}

func NewOneFileSet(s string) *FileSet {
	p := new(FileSet)
	p.FilePaths = make([]string, 0, 1)
	if s == "" || !Exists(s) {
		return p
	}
	p.RelFilePath = RelFilePath(s)
	p.AbsFilePath = p.RelFilePath.AbsFP()
	p.FilePaths = append(p.FilePaths, string(p.AbsFilePath))
	return p
}

package fileutils

import (
	// SU "github.com/fbaube/stringutils"
	MU "github.com/fbaube/miscutils"
	"os"
)

/* REF: os.FileInfo
Name() string       // base name of the file
Size() int64        // length in bytes for regular files; else TBS
Mode() FileMode     // file mode bits
ModTime() time.Time // modification time
IsDir() bool        // abbreviation for Mode().IsDir()
*/

/* REF: os.FileMode:
ModeDir        // d: is a directory
ModeAppend     // a: append-only
ModeExclusive  // l: exclusive use
ModeSymlink    // L: symbolic link
ModeDevice     // D: device file
ModeNamedPipe  // p: named pipe (FIFO)
ModeSocket     // S: Unix domain socket
ModeSetuid     // u: setuid
ModeSetgid     // g: setgid
ModeCharDevice // c: Unix char device, if also ModeDevice is set
ModeSticky     // t: sticky
ModeIrregular  // ?: non-regular file; nothing else is known
*/

// FileMeta is the most basic level of file system
// metadata: the results of a call to [os.Stat] (or
// the contents of a record in sqlar), lightly parsed.
//
// This is "mostly" applicable to contentful nodes
// in other hierarchical structures. For example,
// for non-files it still has size (& permissions).
//
// IsDir() is pass-thru.
// Size() is pass-thru, and/but might need mods
// for directories, and might be overridden when
// file content is available (and modifiable).
// .
type FileMeta struct {
	os.FileInfo
	exists bool
	// error
	MU.Errer
}

// NewFileMeta replaces a call to [os.LStat]. This is necessary because
// a call of the form NewFileMeta(FileInfo) won't work because an error
// return from [os.LStat] indicates whether the file or dir (or symlink)
// exists. However no further analysis of the path is performed in this
// func, because that is more properly done by the caller.
//
// Note that if the file/dir does not exist, [exists] is false and/but
// no error is indicated (i.e. [error] is nil).
// .
func NewFileMeta(path string) *FileMeta {
	var p *FileMeta
	var e error
	p = new(FileMeta)
	p.FileInfo, e = os.Lstat(path)
	p.SetError(e)
	if e == nil || !os.IsNotExist(e) {
		p.exists = true
		// Is this necessary ?
		p.exists = p.IsDir() || p.isFile() || p.isSymlink()
		// Make sure
		p.ClearError()
	}
	return p
}

// Exists is a convenience function.
func (p *FileMeta) Exists() bool {
	return p.exists
}

// IsEmpty is a convenience function
// for files (and directories too?).
// It can be overwritten when the file
// contents are loaded (and modifiable).
// .
func (p *FileMeta) IsEmpty() bool {
	return p.Size() == 0
}

// HasContents is the opposite of [IsEmpty].
func (p *FileMeta) HasContents() bool {
	return p.Size() != 0
}

// IsFile is a (somewhat foolproofed) convenience function.
func (p *FileMeta) IsFile() bool {
	return p.exists && p.isFile() && !p.IsDir() && !p.isSymlink()
}

/*
// IsOkayDir is a (somewhat foolproofed) convenience function.
func (p *FileMeta) IsOkayDir() bool {
	return p.exists && !p.isFile() && p.IsDir() && !p.isSymL()
}
*/

// IsSymlink is a (somewhat foolproofed) convenience function.
func (p *FileMeta) IsSymlink() bool {
	return p.exists && !p.isFile() && !p.IsDir() && p.isSymlink()
}

func (p *FileMeta) isFile() bool {
	return p.Mode().IsRegular() && !p.IsDir()
}

func (p *FileMeta) isSymlink() bool {
	return (0 != (p.Mode() & os.ModeSymlink))
}

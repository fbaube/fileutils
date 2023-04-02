package fileutils

import (
	// SU "github.com/fbaube/stringutils"
	MU "github.com/fbaube/miscutils"
	// "io"
	"os"
	// FP "path/filepath"
	// S "strings"
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

// BasicMeta is the most basic level of system metadata:
// the results of a call to [os.Stat] (or the contents
// of a record in sqlar), lightly parsed.
//
// IsDir() is pass-thru.
// Size() is pass-thru, and/but might need mods for directories.
// .
type BasicMeta struct {
	os.FileInfo
	exists bool
	// error
	MU.Errer
}

// NewBasicMeta replaces a call to [os.LStat]. This is necessary because
// a call of the form NewBasicMeta(FileInfo) won't work because an error
// return from [os.LStat] indicates whether the file or dir (or symlink)
// exists. However no further analysis of the path is performed in this
// func, because that is more properly done by the caller.
//
// Note that if the file/dir does not exist, [exists] is false and/but
// no error is indicated (i.e. [error] is nil).
// .
func NewBasicMeta(path string) *BasicMeta {
	var p *BasicMeta
	var e error
	p = new(BasicMeta)
	p.FileInfo, e = os.Lstat(path)
	p.SetError(e)
	if e == nil || !os.IsNotExist(e) {
		p.exists = true
		// Is this necessary ?
		p.exists = p.IsDir() || p.isFile() || p.isSymL()
		p.ClearError()
	}
	return p
}

// Exists is a convenience function.
func (p *BasicMeta) Exists() bool {
	return p.exists
}

// IsEmpty is a convenience function.
func (p *BasicMeta) IsEmpty() bool {
	return p.Size() == 0
}

// IsOkayFile is a convenience function.
func (p *BasicMeta) IsOkayFile() bool {
	return p.exists && p.isFile() && !p.IsDir() && !p.isSymL()
}

// IsOkayDir is a convenience function.
func (p *BasicMeta) IsOkayDir() bool {
	return p.exists && !p.isFile() && p.IsDir() && !p.isSymL()
}

// IsOkaySymlink is a convenience function.
func (p *BasicMeta) IsOkaySymlink() bool {
	return p.exists && !p.isFile() && !p.IsDir() && p.isSymL()
}

func (p *BasicMeta) isFile() bool {
	return p.Mode().IsRegular() && !p.IsDir()
}

func (p *BasicMeta) isSymL() bool {
	return (0 != (p.Mode() & os.ModeSymlink))
}

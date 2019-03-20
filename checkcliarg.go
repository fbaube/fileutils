package fileutils

import (
	"os"
	"path/filepath"
	S "strings"
)

type FilePathArg struct {
	ArgFilePath string
	RelFilePath
	AbsFilePath
	BaseFilePath string
	Exists       bool
	IsDir        bool
	IsFile       bool
	IsSymL       bool
	Size         int
}

// ProcessFilePathArg does nothing if the path argument is "".
func (p *FilePathArg) ProcessFilePathArg(path string) {
	if path == "" {
		return
	}
	// p = new(FilePathArg)
	p.ArgFilePath = path
	p.RelFilePath = RelFilePath(p.ArgFilePath)
	p.AbsFilePath = p.RelFilePath.AbsFP()
	p.BaseFilePath = S.TrimSuffix(
		p.AbsFilePath.S(), filepath.Ext(p.AbsFilePath.S()))

	FI, e := os.Lstat(path)
	if e != nil {
		// panic("fu.checkcliarg.ProcessFilePathArg.os.Lstat: " + path)
		// The file or directory does not exist.
		// Don't panic. Just return before any flags are set, such as Exists.
		return
	}
	p.IsDir = FI.IsDir()
	p.IsFile = FI.Mode().IsRegular()
	p.IsSymL = (0 != (FI.Mode() & os.ModeSymlink))
	p.Exists = p.IsDir || p.IsFile || p.IsSymL
	if p.IsFile {
		p.Size = int(FI.Size())
	}
	// return p
}

// func (p *FilePathArg)

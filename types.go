package fileutils

import (
	"os"
	"os/user"
	fp "path/filepath"
	S "strings"
)

// Motivations:
// (1) We want in general to use (and require) absolute filepaths.
// Therefore we clearly mark relative and short (as used in the CLI) filepaths.
// (2)

// Note that when working with an os.File, "Name() returns the name of
// the file as presented to Open()", so it might be a relative filepath.

type RelFilePath string
type AbsFilePath string
type FileContent string

const PathSep = string(os.PathSeparator)

var homedir AbsFilePath

func GetHomeDir() string {
	return string(homedir)
}

// ResolveToAbsolute relies on fp.Abs(path).
func (rpf RelFilePath) ResolveToAbsolute() AbsFilePath {
	if S.HasPrefix(string(rpf), PathSep) {
		return AbsFilePath(rpf)
	}
	abspath, e := fp.Abs(string(rpf))
	if e != nil {
		panic("ResolveRelToAbs: " + rpf)
	}
	return AbsFilePath(abspath)
}

func (afp AbsFilePath) MakeRelativeWRT(wrt AbsFilePath) RelFilePath {
	if !afp.StartsWith(homedir) {
		return RelFilePath(afp)
	}
	bytesToTrim := len(wrt) + 1
	return RelFilePath("~" + PathSep + string(afp)[bytesToTrim:])
}

func (afp AbsFilePath) ElideUserHome() RelFilePath {
	return afp.MakeRelativeWRT(homedir)
}

func (afp AbsFilePath) Append(rfp RelFilePath) AbsFilePath {
	return AbsFilePath(string(afp) + string(rfp))
}

func (afp AbsFilePath) StartsWith(beg AbsFilePath) bool {
	return S.HasPrefix(string(afp), string(beg))
}

func init() {
	username, e := user.Current()
	if e != nil {
		println("==> ERROR: Could not determine current user")
		return
	}
	homedir = AbsFilePath(username.HomeDir)
}

package fileutils

import (
	"github.com/nbio/xml"
	"fmt"
	"os"
	"os/user"
	FP "path/filepath"
	S "strings"

	SU "github.com/fbaube/stringutils"
	WU "github.com/fbaube/wasmutils"
)

func boolToInt(b bool) int {
	if !b {
		return 0
	}
	return 1
}

// A token nod to Windoze compatibility.
const PathSep = string(os.PathSeparator)

// These should end in the path separator!
// NOTE See init(), at bottom.
var currentWorkingDir, currentUserHomeDir string

var currentUser *user.User

// GetHomeDir is a convenience function, and
// refers to the invoking user's home directory.
func GetHomeDir() string {
	return currentUserHomeDir
}

func MTypeSub(mtype string, i int) string {
	if i < 0 || i > 2 {
		return ""
	}
	ss := S.Split(mtype, "/")
	return ss[i]
}

func XmlStartElmS(se xml.StartElement) string {
	// type StartElement struct { Name Name ; Attr []Attr }
	// type Name struct         { Space, Local string }
	// type Attr struct         { Name  Name ; Value string }
	// <space:local space:local=value ...>
	ret := "<" + XmlNameS(se.Name)
	for _, a := range se.Attr {
		ret += " " + XmlAttrS(a)
	}
	ret += ">"
	return ret
}

func XmlNameS(n xml.Name) string {
	// type Name struct         { Space, Local string }
	if n.Space == "" {
		return n.Local
	}
	return n.Space + ":" + n.Local
}

func XmlAttrS(a xml.Attr) string {
	// type Attr struct         { Name  Name ; Value string }
	return XmlNameS(a.Name) + "=\"" + a.Value + "\""
}

// Enhomed shortens a filepath by substituting "~".
func Enhomed(s string) string {
	// If it's missing, use assumed/default...
	if s == "" {
		return "."
	}
	// If it can't be normalised...
	if s == "." || s == "~" || s == PathSep {
		return s
	}
	// If it can't be further normalised...
	if S.HasPrefix(s, "~"+PathSep) {
		return s
	}
	// At this point, if it's not an absolute FP, it's a relative FP,
	// but let it slide and don't prepend "./".
	if !FP.IsAbs(s) {
		// panic("NiceFP barfs on: " + s)
		return s
	}
	// println("arg:", s)
	// println("cwd:", currentworkingdir)

	if S.HasPrefix(s, currentUserHomeDir) {
		return ("~" + PathSep + S.TrimPrefix(s, currentUserHomeDir))
		// bytesToTrim := len(currentUserHomeDir) + 1
		// return "~" + PathSep + s[bytesToTrim:]
	}
	// No luck
	return s
}

func init() {
	var e error
	if WU.IsWasm() {
		currentUserHomeDir = "?"
		currentWorkingDir = "."
		currentUser = &user.User{
			Uid:      "aUid",
			Gid:      "aGid",
			Username: "webuser",
			Name:     "webuser",
			HomeDir:  "~",
		}
		return
	}
	currentUser, e = user.Current()
	if e != nil {
		println("==> ERROR: Cannot determine current user")
		return
	}
	currentUserHomeDir, e = os.UserHomeDir()
	if e != nil {
		println("Cannot determine current user's home directory")
		return
	}
	homedir := currentUser.HomeDir
	if currentUserHomeDir != homedir {
		println("==> ERROR: Inconsistent values for current user's home directory")
		return
	}
	currentWorkingDir, e = os.Getwd()
	if e != nil {
		println("==> ERROR: Cannot determine current working directory")
		return
	}
	if !S.HasSuffix(currentWorkingDir, PathSep) {
		currentWorkingDir = currentWorkingDir + PathSep
	}
	if !S.HasSuffix(currentUserHomeDir, PathSep) {
		currentUserHomeDir = currentUserHomeDir + PathSep
	}
}

// SessionSummary can be called anytime.
func SessionSummary() string {
	return fmt.Sprintf("%s (uid:%s,gid:%s,fullname:%s) in dir: %s",
		currentUser.Username, currentUser.Uid, currentUser.Gid,
		currentUser.Name, SU.Tildotted(currentWorkingDir))
}

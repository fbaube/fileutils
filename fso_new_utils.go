package fileutils

import (
	"io/fs"
	"fmt"
	FP "path/filepath"
    SU "github.com/fbaube/stringutils"
)

func checkFP(anFP string) {
	// *** START FOR ROOT ***
    // isValid is returning odd results.
    // func ValidPath(name string) bool
    // reports whether "name" is valid for a call to Open.
	isValid := fs.ValidPath(anFP)
	isLocal := FP.IsLocal(anFP)
	isAbsol := FP.IsAbs(anFP)
    if isValid && isLocal && !isAbsol { return }
    fmt.Printf("****>> isValid:%s isLocal:%s isAbsol:%s / %s \n",
               SU.Yn(isValid), SU.Yn(isLocal), SU.Yn(isAbsol), anFP)
	// *** END FOR ROOT ***
}

func permString(pFI fs.FileInfo) string { 
	var perms, world, group, yuser int 
	perms = int(pFI.Mode().Perm()) // 0777 or 0x1ff
	world =  perms & 7
	group = (perms >> 3) & 7
	yuser = (perms >> 6) & 7
	var ww, gg, yu string
	ww = permStr(world)
	gg = permStr(group)
	yu = permStr(yuser)
	return fmt.Sprintf("%s,%s,%s", yu, gg, ww)
}

func permStr(i int) string {
     var s string
     if 0 != (i&4) { s  = "r" } else { s  = "-" }
     if 0 != (i&2) { s += "w" } else { s += "-" }
     if 0 != (i&1) { s += "x" } else { s += "-" }
     return s
}


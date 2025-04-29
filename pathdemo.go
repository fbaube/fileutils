package fileutils

import (
	"fmt"
	"os"
	SU "github.com/fbaube/stringutils"
	"errors"
	"io/fs"
	FP "path/filepath"
)

var R *os.Root
func init() {
     var e error
     R, e = os.OpenRoot(".")
     if e != nil { panic("OOPS Root") }
}

func fpt(path string) string {
     var A, V, L bool
     var sA, sL string
     var eA, eL error 
     A = FP.IsAbs(path)
     V = fs.ValidPath(path)
     L = FP.IsLocal(path)
     sA, eA = FP.Abs(path)
     sL, eL = FP.Localize(path)
     if eA == nil { eA = errors.New("OK") }
     if eL == nil { eL = errors.New("OK") }
     nF, nE := os.Open(path)
     rF, rE :=  R.Open(path) // this line barfs on symlink to ".."!
     nF.Close()
     rF.Close()
     return fmt.Sprintf("Path: %s \n" +
     	    "Rel:%s LV:%s%s Abs<%s:%s> Lcl<%s:%s> \n" +
     	    "norm.Open.error: %s \n" + "root.Open.error: %s \n", 
     	    path, SU.Yn(!A), SU.Yn(L), SU.Yn(V), sA, eA, sL, eL, nE, rE)
}

func PathDemo() {

        var e error 
     	R, e = os.OpenRoot(".")
	if e != nil { panic("OOPS Root") }
     	println(fpt(""))
     	println(fpt("."))
     	println(fpt(".."))
     	println(fpt("../"))
     	println(fpt("../../"))
     	println(fpt("/"))
     	println(fpt("/etc"))
     	println(fpt("/etc/"))
     	println(fpt("derf"))
     	println(fpt("derf/derf2"))
	println(fpt("/Users/fbaube/src/m5app/m5/m5"))
	println(fpt("/Users/fbaube/src/m5app/m5/m5/derf/"))
	println(fpt("tstat/L-etc"))
	println(fpt("tstat/L-file-Nexist"))
	println(fpt("tstat/L-file-OK"))
	println("=> tilde")
	println(fpt("tstat/L-tilde"))
	// println("=> double dot:") // panics/crashes 
	// println(fpt("tstat/L-par-dbldot"))
	
}


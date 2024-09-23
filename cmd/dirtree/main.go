package main

import (
	"fmt"
	"os"
	FP "path/filepath"
	FU "github.com/fbaube/fileutils"
	SU "github.com/fbaube/stringutils"
)

var pre = []string { "#" }
var mid = []string { "/#", "/." }
var pst = []string { "~" } 

func main() {
     arg := os.Args[1]
     
     FileSlc := FU.GatherDirTreeList(arg)
     for i, f := range FileSlc {
	fmt.Printf("%s [%02d] %s \n", arg, i, f)
	}
     fmt.Printf("%s :: total %d \n", arg, len(FileSlc))
     
     // Filter it
     FileSlc = SU.FilterStringList(FileSlc, pre, mid, pst) 
     for i, f := range FileSlc {
	fmt.Printf("%s [%02d] %s \n", abs, i, f)
	}
     fmt.Printf("%s :: total %d after filtering \n", arg, len(FileSlc))
}


package main

import (
	"fmt"
	"os"
	// FP "path/filepath"
	FU "github.com/fbaube/fileutils"
	SU "github.com/fbaube/stringutils"
)

var pre = []string { "#", ".git", ".DS_Store" }
var mid = []string { "/#", "/." } // incl .git, .DS_Store 
var pst = []string { "~" }

func main() {
     if len(os.Args) == 1 {
     	println("Provide an argument")
	os.Exit(1)
	}
     arg := os.Args[1]
     
     FileSlc := FU.GatherDirTreeList(arg)
     for i, f := range FileSlc {
	fmt.Printf("%s [%02d] %s \n", arg, i, f)
	}
     fmt.Printf("%s :: total %d BEFORE filtering \n", arg, len(FileSlc))

     fmt.Printf("===\nFILTERS: \n\t pre|%#v| \n\t mid|%#v| " +
     			      "\n\tpost|%#v| \n===\n", pre, mid, pst)
     // Filter it
     FileSlc = SU.FilterStringList(FileSlc, pre, mid, pst) 
     for i, f := range FileSlc {
	fmt.Printf("%s [%02d] %s \n", arg, i, f)
	}
     fmt.Printf("%s :: total %d AFTER filtering \n", arg, len(FileSlc))
}


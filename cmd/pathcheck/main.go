package main

import(
	"os"
	"fmt"
	FU "github.com/fbaube/fileutils"
	CA "github.com/fbaube/contentanalysis"
)

// chkerr should not fail if there is a portability issue, but it will. 
func chkerr(e error) {
     if e != nil {
     	fmt.Fprintf(os.Stderr, "%s: %s: %w \n", os.Args[0], os.Args[1], e)
	os.Exit(1)
	}
}

func main() {
     if len(os.Args) < 2 { fmt.Fprintf(os.Stderr,
     	"Usage: %s file-or-dir-name \n", os.Args[0]); os.Exit(0) }
     nam2chk := os.Args[1]
     // fmt.Printf("arg is: " + nam2chk + "\n")
     p, e := FU.NewFSItem(nam2chk)
     chkerr(e)
     if p == nil { fmt.Printf("%s: %s: does not exist \n",
     	  os.Args[0], os.Args[1]); os.Exit(0) }
     // fmt.Printf("Mode: " + p.Type().String() + "\n")
     // fmt.Printf(p.StringWithPermissions() + "\n")
     if p.IsFile() {
     	// p.LoadContents() 
     	PA, e := CA.NewPathAnalysis(p)
	if e != nil {
	   println("ERROR:main:NPAerr:", e.Error())
	   }
	if PA == nil {
	   println("ERROR:main:NPA PAAnil")
	   }
	if p.TypedRaw == nil { 
	   println("ERROR:mail:nil TypedRaw")
	   }
     	p.TypedRaw.Raw_type = PA.RawType()
	}
     fmt.Printf(p.ListingString()+ "\n")
     if p.IsSymlink() { 
     	s, _ := os.Readlink(nam2chk)
	fmt.Printf("Symlink to: %s \n", s)
	}
     if p.IsDir() {
     }
}
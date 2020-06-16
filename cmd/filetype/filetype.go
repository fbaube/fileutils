package main

import(
  // flag "github.com/spf13/pflag"
  "fmt"
  "os"
  FU "github.com/fbaube/fileutils"
)

var myAppName = "filetype"

// At the top level (i.e. in main()), we don't wrap errors
// and return them. We just complain and die. Simple!
func errorbarf(e error, s string) {
	if e == nil {
		return
	}
	if e.Error() == "" {
		return
	}
	fmt.Fprintf(os.Stderr, "%s failed: %s \n", myAppName, e)
	os.Exit(1)
}

func main() {
  if len(os.Args) < 2 {
    println(myAppName, ": Describe file using fbaube/fileutils")
    println("Usage:", myAppName, "myfilename")
    os.Exit(0)
  }
  filename := os.Args[1]
  fileinfo := FU.NewPathInfo(filename)
  println("File info:", fileinfo.String())
  chkdcont := FU.NewCheckedContent(fileinfo)
  fmt.Printf("%s \n", chkdcont)
  os.Exit(0)
}

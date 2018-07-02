Simple file utilities for golang, written for working with
markup files and other file types (like images) typically
associated with documentation.

Included are types for describing a file and configuring a
set of associated output files:

* `InputFile` describes a file: *full path, size, MIME type,
?IsXML, MMCtype*, and *contents* (up to 2 megabytes).

* `MMCtype` is meant to function like a MIME type and has three
fields. It can be set based on file name and contents, and later
updated if the file is XML and has a `DOCTYPE` declaration. Refer
to file `mmctype.go`

* `OutputFiles` makes it easier to create a group of like-named
files for an `InputFile`, in the same directory or optionally
in a like-named subdirectory.

### Known issues

* Tested only on macos (i.e. it's sure to fail on Windows)

### Example

```
$ cd /opt
$ ls example*
example.xml
```

```
import "github.com/fbaube/fileutils"

IF, _ := fileutils.NewInputFile("example.xml")

fmt.Fprintf(os.Stdout, "You opened: %s \n", IF)
// You opened: /opt/example.xml
println("i.e.", IF.DString())
// i.e. InputFile</opt/example.xml>sz<42>dir?<n>bin?<n>img?<n>mime<text/plain>

// Argument is not "": Creates a subdirectory for associated output files.
OF, _ := IF.NewOutputFiles("_myapp")

// Creates an associated file and returns the io.WriteCloser
w_diag, _ := OF.NewOutputExt("diag")
fmt.Fprintln(w_diag, "Lots of diagnostic info")
w_diag.Close()
```

```
$ ls example*
example.xml

example.xml_myapp:
example.diag
```

### Dependencies

* `github.com/hosom/gomagic` for MIME type analysis
* `github.com/pkg/errors` for wrapping errors
* `github.com/fbaube/stringutils` for various

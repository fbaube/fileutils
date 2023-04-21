package fileutils

import (
	"bufio"
	"fmt"
	CT "github.com/fbaube/ctoken"
	"os"
	S "strings"
)

// FileLine is a record (i.e. a line) in a LinesFile.
type FileLine struct {
	CT.Raw        // string
	RawLineNr int // source file line number
	error         // hey, why not an error per line ?
}

// LinesFile is for reading a file where each line is a record.
type LinesFile struct {
	*PathProps
	Lines []*FileLine
}

// NewLinesFile is pretty self-explanatory.
func (pPI *PathProps) NewLinesFile() (*LinesFile, error) {
	e := pPI.GoGetFileContents() // getContentBytes()
	if e != nil {
		panic(e)
	}
	pLF := new(LinesFile)
	pLF.Lines = make([]*FileLine, 0)
	var scnr bufio.Scanner
	scnr = *bufio.NewScanner(S.NewReader(pPI.TypedRaw.S())) // pPI.Raw))
	// Not actually needed since it’s a default split function.
	scnr.Split(bufio.ScanLines)
	var token string
	var linumber = 1
	for scnr.Scan() {
		token = scnr.Text()
		p := new(FileLine)
		p.Raw = CT.Raw(S.TrimSpace(token))
		p.RawLineNr = linumber
		pLF.Lines = append(pLF.Lines, p)
		fmt.Printf("L%02d<%s> \n", p.RawLineNr, p.Raw)
		linumber++
	}
	if err := scnr.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
		return pLF, err
	}
	return pLF, nil
}

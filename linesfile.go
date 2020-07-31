package fileutils

import (
	"bufio"
	"fmt"
	"os"
	S "strings"
)

// FileLine is a record (i.e. a line) in a LinesFile.
type FileLine struct {
	Raw       string
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
	var bb []byte
	bb = pPI.GetContentBytes()
	if pPI.HasError() {
		return nil, fmt.Errorf("fu.NewLF<%s> failed: %w", pPI.absFP, pPI.GetError())
	}
	pLF := new(LinesFile)
	pLF.Lines = make([]*FileLine, 0)
	var scnr bufio.Scanner
	scnr = *bufio.NewScanner(S.NewReader(string(bb)))
	// Not actually needed since it’s a default split function.
	scnr.Split(bufio.ScanLines)
	var token string
	var linumber = 1
	for scnr.Scan() {
		token = scnr.Text()
		p := new(FileLine)
		p.Raw = S.TrimSpace(token)
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

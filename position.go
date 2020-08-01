package fileutils

type FilePosition struct {
	Pos int // Position, from xml.Decoder
	Lnr int // Line number
	Col int // Column [number]
}


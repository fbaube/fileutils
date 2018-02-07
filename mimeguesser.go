package fileutils

import "github.com/pkg/errors"

// The hosom/gomagic library is licensed BSD-3.
// This file (fileutils.go) borrows heavily from it.
//
// Three different libraries for determining MIME types were evaluated.
// All three produced the same results on common files.
// "hosom" is the easiest to use because it takes care of its own cleanup.
import hosom "github.com/hosom/gomagic"

// MimeFlag is a friendlier version of the constants in the C library.
type MimeFlag int

// These constants determine the form of the output:
// File: /opt/REVIEW/OASISLogo.jpg
// TEXTUAL: JPEG image data, JFIF standard 1.01, resolution (DPI), ...
// TYPE: image/jpeg
// ENC:  binary
// FULL: image/jpeg; charset=binary
const (
	// MimeTextual returns a textual description
	MimeTextual MimeFlag = MimeFlag(0)
	// MimeType returns a MIME type string
	MimeType MimeFlag = MimeFlag(int(hosom.MAGIC_MIME_TYPE))
	// MimeEnc returns a MIME encoding
	MimeEncoding MimeFlag = MimeFlag(int(hosom.MAGIC_MIME_ENCODING))
	// MimeFull returns MIME-type-string ";" MIME-encoding
	MimeFull MimeFlag = MimeFlag(int(hosom.MAGIC_MIME))
)

// MimeFile returns MIME info about the file name.
// "mode" is one of the values Mime*
func MimeFile(filename string, mode MimeFlag) (string, error) {
	m, e := hosom.Open(hosom.Flag(mode))
	defer m.Close()
	if e != nil {
		return "", errors.Wrap(e, "MimeFile open-analyzer failed")
	}
	mt, e := m.File(filename)
	if e != nil {
		return "", errors.Wrap(e, "MimeFile analyze-file failed")
	}
	return mt, nil
}

// MimeBuffer returns MIME info.
// "mode" is one of the values Mime*
func MimeBuffer(buf []byte, mode int) (string, error) {
	m, e := hosom.Open(hosom.Flag(mode))
	defer m.Close()
	if e != nil {
		return "", errors.Wrap(e, "MimeBuffer open-analyzer failed")
	}
	mt, e := m.Buffer(buf)
	if e != nil {
		return "", errors.Wrap(e, "MimeBuffer analyze-buffer failed")
	}
	return mt, nil
}

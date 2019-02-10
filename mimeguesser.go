package fileutils

import (
	"github.com/pkg/errors"
)

// NOTE The hosom/gomagic library is licensed BSD-3,
// and this file (mimeguesser.go) borrows heavily from it.
//
// Three different libraries for determining MIME types were evaluated.
// All three produced the same results on common files.
// "hosom" is the easiest to use because it takes care of its own cleanup.
import hosom "github.com/hosom/gomagic"

// MimeFlag is a user-friendlier version
// of the constants in the C library.
type MimeFlag int

// These constants determine the form of the output:
// File: /opt/REVIEW/OASISLogo.jpg
// TEXTUAL: JPEG image data, JFIF standard 1.01, resolution (DPI), ...
// TYPE: image/jpeg
// ENC:  binary
// FULL: image/jpeg; charset=binary
const (
	// MimeTextual returns a textual description
	HgMimeTextual MimeFlag = MimeFlag(0)
	// MimeType returns a MIME type string
	HgMimeType MimeFlag = MimeFlag(int(hosom.MAGIC_MIME_TYPE))
	// MimeEnc returns a MIME encoding
	HgMimeEncoding MimeFlag = MimeFlag(int(hosom.MAGIC_MIME_ENCODING))
	// MimeFull returns MIME-type-string ";" MIME-encoding
	HgMimeFull MimeFlag = MimeFlag(int(hosom.MAGIC_MIME))
)

// MimeFile returns MIME info about the file name.
// "mode" is one of the values Mime*
func MimeFile(filename string, mode MimeFlag) (string, error) {
	m, e := hosom.Open(hosom.Flag(mode))
	defer m.Close()
	if e != nil {
		return "", errors.Wrapf(e, "fu.MimeFile.hosomOpen<%s:%x>", filename, mode)
	}
	mt, e := m.File(filename)
	if e != nil {
		return "", errors.Wrapf(e, "fu.MimeFile.hosomFile<%s>", filename)
	}
	return mt, nil
}

// MimeBuffer returns MIME info.
// "mode" is one of the values Mime*
func MimeBuffer(buf []byte, mode int) (string, error) {
	if len(buf) == 0 {
		return "", nil
	}
	m, e := hosom.Open(hosom.Flag(mode))
	defer m.Close()
	if e != nil {
		return "", errors.Wrap(e, "fu.MimeBuffer.hosomOpen")
	}
	mt, e := m.Buffer(buf)
	if e != nil {
		return "", errors.Wrap(e, "fu.MimeBuffer.hosomBuffer")
	}
	return mt, nil
}

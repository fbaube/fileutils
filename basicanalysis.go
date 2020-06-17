package fileutils

import (
  "fmt"
)

type BasicAnalysis struct {
  baError     error
  FileIsOkay  bool
  FileExt     string
  AnalysisRecord
}

func NewBasicAnalysis() *BasicAnalysis {
  p := new(BasicAnalysis)
  p.MType = "-/-/-"
  return p
}

func (p BasicAnalysis) String() string {
  var x string
  if p.IsXML() { x = "IsXML.. " }
  return fmt.Sprintf("fu.basicAnls: \n   " + x +
    "MimeType<%s> " +
    "MType<%v> " +
    "\n   " +
    "XmlInfo<%+v> " +
    "\n   " +
    "DitaInfo<%+v>",
    p.MimeType, p.MType, p.XmlInfo, p.DitaInfo)
}

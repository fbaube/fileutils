package fileutils

import(
	D "github.com/fbaube/dsmnd"
)

type FSO_type D.SemanticFieldType

const(
	FSO_type_DIRR FSO_type = FSO_type(D.SFT_FSDIR)
	FSO_type_FILE FSO_type = FSO_type(D.SFT_FSFIL)
	FSO_type_SYML FSO_type = FSO_type(D.SFT_FSYML)
	FSO_type_OTHR FSO_type = FSO_type(D.SFT_FSOTH)
	// tem_type_ FSO_type = FSO_type(D.SFT_FS)
)

/*
// FSYS: FILE SYSTEM ITEMS (4) 
{BDT_FSYS.DT(), SFT_FSDIR.S(), "Dir",  "FS directory (unord|ord)"}, 
{BDT_FSYS.DT(), SFT_FSFIL.S(), "File", "FS file (contentful)"}, 
{BDT_FSYS.DT(), SFT_FSYML.S(), "SymL", "FS symbolic link"},
{BDT_FSYS.DT(), SFT_FSOTH.S(), "Other","FS other (pipes, etc.)"},
*/
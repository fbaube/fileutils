package fileutils

import(
	D "github.com/fbaube/dsmnd"
)

type FSItem_type D.SemanticFieldType

const(
	FSItem_type_DIRR FSItem_type = FSItem_type(D.SFT_FSDIR)
	FSItem_type_FILE FSItem_type = FSItem_type(D.SFT_FSFIL)
	FSItem_type_SYML FSItem_type = FSItem_type(D.SFT_FSYML)
	FSItem_type_OTHR FSItem_type = FSItem_type(D.SFT_FSOTH)
	// tem_type_ FSItem_type = FSItem_type(D.SFT_FS)
)

/*
// FSYS: FILE SYSTEM ITEMS (4) 
{BDT_FSYS.DT(), SFT_FSDIR.S(), "Dir",  "FS directory (unord|ord)"}, 
{BDT_FSYS.DT(), SFT_FSFIL.S(), "File", "FS file (contentful)"}, 
{BDT_FSYS.DT(), SFT_FSYML.S(), "SymL", "FS symbolic link"},
{BDT_FSYS.DT(), SFT_FSOTH.S(), "Other","FS other (pipes, etc.)"},
*/
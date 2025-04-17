package fileutils 

// NOTE 1: Directory listings are most useful to 
// users when they rely primarily on files' NAMES. 
// However, using these exclusively would mean that
// we would miss simple file renamings. Therefore we
// also take file hashes, to identify file renamings.

// NOTE 1a: Hash collisions are likely, because there can 
// be (numerous?!) short files with identical content. So 
// in a map, hashes must be the values, not the keys. 

// NOTE 1b: We cannot define our own comparison function 
// for hashes (which are arrays, [16]byte), so instead we
// store them as hex strings. 

// NOTE 2: We distinguish between "contentful" names (names 
// that are files that have content length greater than zero, 
// and whose content can be hashed)) and "contentless" names
// (files with zero content length, and subdirectories). 

import (
	"fmt"
	"os"
	"io/fs"
	"encoding/hex"
)

// DirectoryDetails should NOT follow (or enforce) the convention 
// that a directory name always has a trailing slash, because this 
// struct is intended for comparison operations, where a name may 
// be a directory in one place but not the other. 
type DirectoryDetails struct {
	DirName 	string
	DirFileInfo	fs.FileInfo
	NamesToItems	map[string]*FSItem
	NamesToHashes	map[string]string
	ContentfulFileCount,
	ContentlessFileCount,
	DirCount, MiscCount  int 
}

// ReadDirectoryDetails includes "Read" in its name because it works 
// kinda like ReadDir: it does not recurse down into subdirectories.
func ReadDirectoryDetails(aPath string) (*DirectoryDetails, error) {

	var e error
	var anFI fs.FileInfo

	// The path argument has to be a directory
	anFI, e = os.Stat(aPath)
	if e != nil {	return nil, &fs.PathError {
			Op: "fu.readdirdetails.stat", Path: aPath, Err: e } }	
	if !anFI.IsDir() {
			return nil, &fs.PathError {
			Op: "fu.readdirdetails.isdir", Path: aPath, Err: e } }	
	var pDD = new(DirectoryDetails)
	pDD.DirName = aPath
	pDD.DirFileInfo = anFI
	// ======================
	//  Read directory items 
	// ======================
	pDD.NamesToItems, e = ReadDirAsMap(pDD.DirName)
	if e != nil {	return nil, &fs.PathError { 
			Op: "fu.readdirdetails.readdirasmap",
			Path: aPath, Err: e } }	
	fmt.Printf("%d items in M: %s \n", len(pDD.NamesToItems), pDD.DirName)
	// race condition ?
	if len(pDD.NamesToItems) == 0 {
	   return nil, nil 
	}
	for key,p := range pDD.NamesToItems {
	    	if p.IsDir()             { pDD.DirCount++ }  else
		if !p.Mode().IsRegular() { pDD.MiscCount++ } else
		if 0 == int(p.Size())    { pDD.ContentlessFileCount++ } else
		     			 { pDD.ContentfulFileCount++ } 
		e = p.LoadContents()
		// permissions problem ? 
		if e != nil {	return nil, &fs.PathError { 
			Op: "fu.readdirdetails.loadcontents",
			Path: key, Err: e } }	
	}
	// FIXME: What about file times ? 

	// =====================
	//  Map names to hashes
	// =====================
	pDD.NamesToHashes = mapNamesToHashes(pDD.NamesToItems)
	
	return pDD, nil 
}

func mapNamesToHashes(inMap map[string]*FSItem) map[string]string {
	var  outMap   map[string]string
	outMap = make(map[string]string)
	for inName,pFSI := range inMap {
		if pFSI == nil { fmt.Printf("nil *FSItem: %s \n",
			inName); continue  }
		if pFSI.TypedRaw == nil { fmt.Printf("no Raw: %s \n",
			inName); continue  }
		hashAsString := hex.EncodeToString(pFSI.TypedRaw.Hash[0:15])
		outMap[inName] = hashAsString
		fmt.Printf("%s: len[%d] Hash:%s \n",
			inName, pFSI.Size(), hashAsString)
	}
 	return outMap
}

/*
type fs.FileInfo interface {
	Name() string       // base name of the file
	Size() int64        // length in bytes for regular files; system-dependent for others
	Mode() FileMode     // file mode bits
	ModTime() time.Time // modification time
	IsDir() bool        // abbreviation for Mode().IsDir()
	Sys() any           // underlying data source (can return nil)
}
*/

func (pDD *DirectoryDetails) NamesByHash(hh string) []string {
    var keys []string
    for k, v := range pDD.NamesToHashes {
        if hh == v {
            keys = append(keys, k)
        }
    }
    return keys
}


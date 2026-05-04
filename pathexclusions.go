package fileutils

import (
	S "strings"
)

// m5ExcludePrefixes m5ExcludeContains m5ExcludeSuffixes 
// are dependencies on the app, but for simplicity we 
// allow them here in this utility package. Notes:
//  - special handling is required for a leading dot (".") 
//    or tilde ("~") that is by itself (i.e. is an entire 
//    filepath element), representing the current or home
//    directory, respectively 
//  - we don't necessarily want to exclude JSON suffixes, 
//    but we do "for now" because we auto-gen them as outputs
var m5ExcludePrefixes = []string { ".", "_" }
var m5ExcludeContains = []string { ".." }
var m5ExcludeSuffixes = []string {
    "~", ".env", ".sh", ".rc", ".bashrc", "gtk", "gtr",
    "_echo", "_tkns", "_tree", ".tmp.json" }

// ExcludeFilepath_m5 returns true (plus a reason) for a filepath
// that matches a blacklist for prefix or "midfix" or suffix.
//  - Excluded prefixes must follow a path separator; this rule should
//    allow "." and "./" (only) to pass thru unexcluded / unimpeded. 
//  - Excluded suffixes apply to all names, but will not apply to 
//    a directory name that has a path separator appended.
//  - The path separator is assumed to be slash ("/"), not os.Separator.
// .
func ExcludeFilepath_m5(s string) (bool, string) {
     var reason string
     if s == "." || s == "./" || s == "~" || s == "~/" {
     	return false, ""
	}
     for _, pfx := range m5ExcludePrefixes {
     	 if S.HasPrefix(s, pfx) {
	    reason += "leading-prefix<" + pfx + "> " 
	    }
	 }
     for _, sfx := range m5ExcludeSuffixes {
     	 if S.HasSuffix(s, sfx) {
	    reason += "trailing-suffix<" + sfx + "> " 
	    }
	 }
     for _, fpc := range m5ExcludeContains {
     	 if S.Contains(s, fpc) {
	    reason += "contains<" + fpc + "> " 
	    }
	 }
     // Check for prefixes and suffixes around path separators.
     for _, pfx := range m5ExcludePrefixes {
     	 if S.Contains(s, "/" + pfx) {
	    reason += "dir/+prefix</" + pfx + "> " 
	    }
	 }
     for _, sfx := range m5ExcludeSuffixes {
     	 if S.Contains(s, sfx + "/") {
	    reason += "suffix+/dir<" + sfx + "/> " 
	    }
	 }
     return (reason != ""), reason 
}


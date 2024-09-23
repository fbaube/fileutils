package fileutils

// NOTE these probably won't work on Windows.

var filterPrefixes = []string { "#", ".git", ".DS_Store" }
var filterMidfixes = []string { "/#", "/." } // incl .git, .DS_Store 
var filterSuffixes = []string { "~" }


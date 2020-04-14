package fileutils

// DTDtypeFileExtensions are all the file extensions that are automatically
// classified as being DTD-type.
var DTDtypeFileExtensions = []string{".dtd", ".mod", ".ent"}

// MarkdownFileExtensions are all the file extensions that are automatically
// classified as being Markdown-type, even tho we generally use a regex instead.
var MarkdownFileExtensions = []string{".md", ".mdown", ".markdown", ".mkdn"}

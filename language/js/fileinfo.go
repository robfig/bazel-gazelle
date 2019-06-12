package js

import (
	"io/ioutil"
	"log"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

// fileInfo holds information used to decide how to build a file. This
// information comes from the file's name, and from goog.require/provide/module
// declarations (in .js / .jsx files).
type fileInfo struct {
	path string
	name string

	// ext is the type of file, based on extension.
	ext ext

	// provides are the import paths that this file provides.
	provides []string

	// isTest is true if the file stem (the part before the extension)
	// ends with "_test". This may be true for js, jsx, or html files.
	isTest bool

	// isTestOnly is true if the file contains a goog.setTestOnly declaration.
	isTestOnly bool

	// isModule is true if this file provides via goog.module instead of
	// goog.provide.
	isModule bool

	// imports is a list of identifiers imported by a file.
	imports []string
}

// ext indicates how a file should be treated, based on extension.
type ext int

const (
	// unknownExt is applied files that aren't buildable with rules_closure
	unknownExt ext = iota

	// jsExt is applied to .js files.
	jsExt

	// jsxExt is applied to .jsx files.
	jsxExt

	// htmlExt is applied to .html files.
	htmlExt
)

// fileNameInfo returns information that can be inferred from the name of
// a file. It does not read data from the file.
func fileNameInfo(path_ string) fileInfo {
	name := filepath.Base(path_)
	var ext ext
	switch path.Ext(name) {
	case ".js":
		ext = jsExt
	case ".jsx":
		ext = jsxExt
	case ".html":
		ext = htmlExt
	default:
		ext = unknownExt
	}
	if strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_") {
		ext = unknownExt
	}

	var isTest bool
	l := strings.Split(name[:len(name)-len(path.Ext(name))], "_")
	if len(l) >= 2 && l[len(l)-1] == "test" {
		isTest = true
	}

	return fileInfo{
		path:   path_,
		name:   name,
		ext:    ext,
		isTest: isTest,
	}
}

// TODO: handle goog.module.get()
// TODO: handle ES6 modules

var (
	closureLibraryRepo = `com_google_javascript_closure_library`

	declRegexp = regexp.MustCompile(`(?m)^(?:(?:const|var) .* = )?goog\.(require|provide|module)\(['"]([^'"]+)`)

	testonlyRegexp = regexp.MustCompile(`^goog\.setTestOnly\(`)
)

// jsFileInfo returns information about a .js file.
// If the file can't be read, an error will be logged, and partial information
// will be returned.
func jsFileInfo(path string) fileInfo {
	info := fileNameInfo(path)
	b, err := ioutil.ReadFile(path)
	if err != nil {
		log.Printf("%s: error reading js file: %v", info.path, err)
		return info
	}
	for _, match := range declRegexp.FindAllSubmatch(b, -1) {
		var (
			declType   = string(match[1])
			identifier = string(match[2])
		)
		switch declType {
		case "provide", "module":
			info.isModule = declType == "module"
			info.provides = append(info.provides, identifier)
		case "require":
			info.imports = append(info.imports, identifier)
		default:
			panic("unhandled declType: " + declType)
		}
	}
	info.isTestOnly = testonlyRegexp.Match(b)
	return info
}

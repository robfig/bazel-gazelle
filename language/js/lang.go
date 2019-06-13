// Package js provides support for JS rules. It generates
// closure_js_library and closure_js_test rules.
//
// Rule generation
//
// Currently, Gazelle generates one rule per file.
//
// Dependency resolution
//
// JS libraries are indexed by their goog.module / goog.provide declarations.
//
// Gazelle has an index of the closure library and generates appropriate dependencies
// for imports.
package js

import "github.com/bazelbuild/bazel-gazelle/language"

const verbose = false

const jsName = "js"

type jsLang struct {
}

func (_ *jsLang) Name() string { return jsName }

func NewLanguage() language.Language {
	return &jsLang{}
}

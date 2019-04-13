// Package js provides support for JS rules. It generates
// closure_js_library and closure_js_test rules.
//
// Configuration
//
// JS rules support the flag -js_prefix and the directive # gazelle:js_prefix.
//
// Rule generation
//
// Currently, Gazelle generates one rule per file, in addition to a rule
// containing all js named for the directory.
//
// Dependency resolution
//
// JS libraries are indexed by their goog.module / goog.provide declarations. If
// an import doesn't match any known library, Gazelle guesses a name for it,
// locally (if the import path is under the current prefix).
//
// Gazelle is aware of closure library and generates appropriate dependencies
// for imports.
package js

import "github.com/bazelbuild/bazel-gazelle/language"

const jsName = "js"

type jsLang struct {
}

func (_ *jsLang) Name() string { return jsName }

func NewLanguage() language.Language {
	return &jsLang{}
}

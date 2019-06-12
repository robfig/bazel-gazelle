package js

import (
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/label"
	"github.com/bazelbuild/bazel-gazelle/language/js/closure_library"
	"github.com/bazelbuild/bazel-gazelle/repo"
	"github.com/bazelbuild/bazel-gazelle/resolve"
	"github.com/bazelbuild/bazel-gazelle/rule"
)

func (_ *jsLang) Imports(_ *config.Config, r *rule.Rule, f *rule.File) []resolve.ImportSpec {
	if !isJsLibrary(r.Kind()) {
		return nil
	}

	// TODO: See how badly this performs, whether or not we need to capture the
	// provides in the rule.
	var provides []resolve.ImportSpec
	for _, src := range r.AttrStrings("srcs") {
		srcFilename := filepath.Join(filepath.Dir(f.Path), src)
		fi := jsFileInfo(srcFilename)
		for _, provide := range fi.provides {
			provides = append(provides, resolve.ImportSpec{Lang: jsName, Imp: provide})
		}
	}

	fmt.Println("Imports:", f.Pkg+":"+r.Name(), "provides:", provides)
	return provides
}

func (_ *jsLang) Embeds(r *rule.Rule, from label.Label) []label.Label {
	return nil
}

func (gl *jsLang) Resolve(c *config.Config, ix *resolve.RuleIndex, rc *repo.RemoteCache, r *rule.Rule, importsRaw interface{}, from label.Label) {
	fmt.Println("Resolve:", from)
	if importsRaw == nil {
		// may not be set in tests.
		fmt.Println(".. no requires")
		return
	}

	fi := importsRaw.(fileInfo)
	r.DelAttr("deps")

	var deps []string
	// TODO: Use an in-place algorithm instead without a map
	var depMap = make(map[string]struct{})
	for _, imp := range fi.imports {
		l, err := resolveJs(c, ix, rc, r, imp, from)
		if err != nil {
			log.Print(err)
		}
		fmt.Println(".. requires ", imp, "=>", l)
		if l == label.NoLabel {
			continue
		}
		l = l.Rel(from.Repo, from.Pkg)
		//		deps = append(deps, l.String())
		depMap[l.String()] = struct{}{}
	}
	if len(depMap) > 0 {
		for k := range depMap {
			deps = append(deps, k)
		}
		r.SetAttr("deps", deps)
	}
}

var (
	skipImportError = errors.New("std or self import")
	notFoundError   = errors.New("rule not found")
)

func resolveJs(c *config.Config, ix *resolve.RuleIndex, rc *repo.RemoteCache, r *rule.Rule, imp string, from label.Label) (label.Label, error) {
	// gc := getJsConfig(c)

	if isClosureLibrary(imp) {
		// this probably has to do something more complicated like indexing
		// closure library and generating a file with that information.
		return resolveClosureLibrary(imp), nil
	}

	if l, ok := resolve.FindRuleWithOverride(c, resolve.ImportSpec{Lang: jsName, Imp: imp}, "js"); ok {
		return l, nil
	}

	if l, err := resolveWithIndexJs(ix, imp, from); err == nil || err == skipImportError {
		return l, err
	} else if err != notFoundError {
		return label.NoLabel, err
	}

	return label.NoLabel, nil
	// return resolveExternal(gc.moduleMode, rc, imp)
	// return resolveVendored(rc, imp)
}

func isClosureLibrary(imp string) bool {
	return strings.HasPrefix(imp, "goog.")
}

func resolveClosureLibrary(imp string) label.Label {
	if !strings.HasPrefix(imp, "goog.") {
		panic(fmt.Errorf("expected a closure library import: %v", imp))
	}
	if target, ok := closure_library.PROVIDE_TO_TARGET[imp]; ok {
		l, err := label.Parse("@io_bazel_rules_closure" + target)
		if err != nil {
			log.Printf("error parsing label %q from closure_library: %v", target, err)
			return label.NoLabel
		}
		return l
	}
	log.Println("closure library import not found:", imp)
	return label.NoLabel
}

func resolveWithIndexJs(ix *resolve.RuleIndex, imp string, from label.Label) (label.Label, error) {
	matches := ix.FindRulesByImport(resolve.ImportSpec{Lang: jsName, Imp: imp}, jsName)
	var bestMatch resolve.FindResult
	var matchError error

	for _, m := range matches {
		if bestMatch.Label.Equal(label.NoLabel) {
			// Current match is better
			bestMatch = m
			matchError = nil
		} else {
			// Match is ambiguous
			// TODO: consider listing all the ambiguous rules here.
			matchError = fmt.Errorf("rule %s imports %q which matches multiple rules: %s and %s. # gazelle:resolve may be used to disambiguate", from, imp, bestMatch.Label, m.Label)
		}
	}
	if matchError != nil {
		return label.NoLabel, matchError
	}
	if bestMatch.Label.Equal(label.NoLabel) {
		return label.NoLabel, notFoundError
	}
	if bestMatch.IsSelfImport(from) {
		return label.NoLabel, skipImportError
	}
	return bestMatch.Label, nil
}

// HACK: This might be called upon to recognize mapped kinds, which it does not
// have the information to do.
//
// For example:
//   gazelle:map_kind closure_js_library closure_js_thirdparty_library //tools/js/defs.bzl
//
// That causes Imports() to be invoked with r.Kind() == closure_js_thirdparty_library.
// This is a bug. Language plugins like this are supposed to only see their builtin rule kinds,
// and the surrounding code is supposed to handle the mappings.
//
// But, since this is unlikely to be merged, make the local (and easier) fix.
// Require any mapped kinds to begin with "closure_js".
func isJsLibrary(kind string) bool {
	//	return kind == "closure_js_library" || kind == "closure_jsx_library"
	return strings.HasPrefix(kind, "closure_js")
}

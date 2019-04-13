package js

import (
	"errors"
	"fmt"
	"log"
	"path"
	"path/filepath"
	"strings"

	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/label"
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
		fmt.Println(src, fi.provides)
		for _, provide := range fi.provides {
			provides = append(provides, resolve.ImportSpec{Lang: jsName, Imp: provide})
		}
	}
	return provides
}

func (_ *jsLang) Embeds(r *rule.Rule, from label.Label) []label.Label {
	return nil
}

func (gl *jsLang) Resolve(c *config.Config, ix *resolve.RuleIndex, rc *repo.RemoteCache, r *rule.Rule, importsRaw interface{}, from label.Label) {
	fmt.Println(r.Kind(), r.Name(), importsRaw)
	if importsRaw == nil {
		// may not be set in tests.
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
		// fmt.Println("Resolve:", imp, ":", label.New("io_bazel_rules_closure", path.Join("closure/library", fp), path.Base(fp)))
	}

	if l, ok := resolve.FindRuleWithOverride(c, resolve.ImportSpec{Lang: jsName, Imp: imp}, "js"); ok {
		fmt.Println("Resolve:", imp, "FindRuleWithOverride:", l)
		return l, nil
	}

	if l, err := resolveWithIndexJs(ix, imp, from); err == nil || err == skipImportError {
		fmt.Println("Resolve:", imp, "WithIndex:", l)
		return l, err
	} else if err != notFoundError {
		return label.NoLabel, err
	}

	fmt.Println("Resolve:", imp, "NotFound")
	return label.NoLabel, nil
	// return resolveExternal(gc.moduleMode, rc, imp)
	// return resolveVendored(rc, imp)
}

func isClosureLibrary(imp string) bool {
	return strings.HasPrefix(imp, "goog.")
}

// The packages under third_party don't match the pattern.
// For example: goog.require('goog.dom.query');
// is provided by:
//   @io_bazel_rules_closure//closure/library/third_party/goog/dojo/dom:query
//
// Not sure the best way to handle this in a way that doesn't require updating,
// but that set of code is small and slowly-changing enough that special cases
// seem ok.
//
// TODO: Generate the full set of current targets.

var thirdPartyClosureProvides = map[string]label.Label{
	"goog.dom.query": label.New("io_bazel_rules_closure",
		"third_party/closure/library/dojo/dom", "query"),
}

func resolveClosureLibrary(imp string) label.Label {
	if !strings.HasPrefix(imp, "goog.") {
		panic(fmt.Errorf("expected a closure library import: %v", imp))
	}
	if l, ok := thirdPartyClosureProvides[imp]; ok {
		return l
	}
	imp = imp[len("goog."):]
	imp = strings.Replace(imp, ".", "/", -1)
	imp = strings.ToLower(imp)

	var pkg, base string
	if !strings.Contains(imp, "/") {
		pkg = imp
		base = path.Base(imp)
	} else {
		pkg = path.Dir(imp)
		base = path.Base(imp)
	}

	return label.New("io_bazel_rules_closure", path.Join("closure/library", pkg), base)
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

func isJsLibrary(kind string) bool {
	return kind == "closure_js_library" || kind == "closure_jsx_library"
}

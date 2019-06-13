package js

import (
	"log"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bazelbuild/bazel-gazelle/language"
	"github.com/bazelbuild/bazel-gazelle/rule"
)

func (gl *jsLang) GenerateRules(args language.GenerateArgs) language.GenerateResult {
	var _ = getJsConfig(args.Config)

	// Handle existing rules that include multiple files by generating our
	// version of that same grouping.
	type multiRule struct {
		kind, name              string
		srcs, provides, imports []string
	}
	var existingRules []*rule.Rule
	if args.File != nil {
		existingRules = args.File.Rules
	}
	var multiFileRules = make(map[string]*multiRule)
	var multiFileRulesGen []*multiRule
	for _, r := range existingRules {
		if !isJsLibrary(r.Kind()) {
			continue
		}
		srcs := r.AttrStrings("srcs")
		if len(srcs) <= 1 {
			continue
		}
		genrule := &multiRule{kind: r.Kind(), name: r.Name()}
		for _, src := range srcs {
			multiFileRules[src] = genrule
			// if any of the srcs are in a sub-directory, add them to
			// RegularFiles to be processed.
			if strings.Contains(src, "/") {
				args.RegularFiles = append(args.RegularFiles, src)
			}
		}
		multiFileRulesGen = append(multiFileRulesGen, genrule)
	}

	// Loop through each file in this package, read their info (what they
	// provide & require), and generate a lib or test rule for it.
	var rules []*rule.Rule
	var imports []interface{}
	var testFileInfos = make(map[string][]fileInfo)
	sort.Strings(args.RegularFiles) // results in htmls first
	for _, filename := range args.RegularFiles {
		var fi = jsFileInfo(filepath.Join(args.Dir, filename))
		if fi.ext == unknownExt {
			continue
		}

		// Deal with tests separately since they involve multiple files.
		// TODO: Make this work with multi-file groupings.
		if fi.isTest {
			name := testBaseName(fi.name)
			testFileInfos[name] = append(testFileInfos[name], fi)
			continue
		}

		// If this file is part of a multi-file rule, merge in its properties
		// instead of creating a new rule.
		if r, ok := multiFileRules[filename]; ok {
			r.srcs = append(r.srcs, filename)
			r.provides = append(r.provides, fi.provides...)
			r.imports = append(r.imports, fi.imports...)
			continue
		}

		// Create one closure_js[x]_library rule per non-test source file.
		switch fi.ext {
		case jsExt, jsxExt:
			rules = append(rules, generateLib(filename))
			imports = append(imports, fi)
		}
	}

	// Group foo_test.js[x] with foo_test.html (if present) into test targets.
	for name, fis := range testFileInfos {
		switch len(fis) {
		case 1:
			rules = append(rules, generateTest(fis[0]))
			imports = append(imports, fis[0])
		case 2:
			rules = append(rules, generateCombinedTest(fis[1], fis[0]))
			imports = append(imports, fis[1])
		default:
			log.Println("unexpected number of test sources:", name)
		}
	}

	// Generate the multi-file rules.
	for _, r := range multiFileRulesGen {
		if len(r.srcs) == 0 {
			continue
		}
		rule := rule.NewRule(r.kind, r.name)
		rule.SetAttr("srcs", r.srcs)
		rules = append(rules, rule)
		imports = append(imports, fileInfo{
			name:     "rule:" + r.name,
			provides: r.provides,
			imports:  r.imports,
		})
	}

	return language.GenerateResult{
		Gen:     rules,
		Imports: imports,
	}
}

// testBaseName trims foo_test.[js|html] => "foo"
func testBaseName(name string) string {
	return name[:strings.Index(name, "_test.")]
}

func generateLib(filename string) *rule.Rule {
	jsOrJsx := filepath.Ext(filename)[1:]
	r := rule.NewRule("closure_"+jsOrJsx+"_library",
		filename[:len(filename)-len(filepath.Ext(filename))])
	r.SetAttr("srcs", []string{filename})
	r.SetAttr("visibility", []string{"//visibility:public"})
	return r
}

func generateCombinedTest(js, html fileInfo) *rule.Rule {
	r := generateTest(js)
	r.SetAttr("html", html.name)
	return r
}

func generateTest(js fileInfo) *rule.Rule {
	jsOrJsx := filepath.Ext(js.name)[1:]
	r := rule.NewRule("closure_"+jsOrJsx+"_test",
		js.name[:len(js.name)-len(filepath.Ext(js.name))])
	r.SetAttr("srcs", []string{js.name})
	r.SetAttr("compilation_level", "ADVANCED")
	r.SetAttr("entry_points", js.provides)
	r.SetAttr("visibility", []string{"//visibility:public"})
	return r
}

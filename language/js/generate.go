package js

import (
	"log"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bazelbuild/bazel-gazelle/language"
	"github.com/bazelbuild/bazel-gazelle/rule"
)

type importInfo struct {
	imports, deps []string
}

func newImportInfo(fi fileInfo) importInfo {
	return importInfo{
		imports: fi.imports,
		deps:    fi.deps,
	}
}

func (gl *jsLang) GenerateRules(args language.GenerateArgs) language.GenerateResult {
	var jsc = getJsConfig(args.Config)

	// For each JS rule, extract info from the srcs and emit our take on it.
	// Keep track of the src files that were covered.
	var (
		rules            []*rule.Rule
		imports          []interface{}
		empty            []*rule.Rule
		existingRules    []*rule.Rule
		existingRuleSrcs []string
	)
	if args.File != nil {
		existingRules = args.File.Rules
	}
	for _, r := range existingRules {
		if !isJsLibrary(r.Kind()) {
			continue
		}

		// A srcs attr with a glob appears as an empty srcs.
		// Just ignore those rules since there's nothing we can productively do.
		srcsattr := r.AttrStrings("srcs")
		if len(srcsattr) == 0 {
			continue
		}

		// Collect the combined srcs, requires, deps for this rule.
		var srcs, requires, deps []string
		for _, src := range srcsattr {
			// Ignore this src if it's a label.
			if strings.HasPrefix(src, ":") || strings.HasPrefix(src, "//") {
				srcs = append(srcs, src)
				continue
			}

			// Read the src and combine provides / imports.
			var fi, ok = jsFileInfo(args.Config.RepoRoot, jsc, filepath.Join(args.Dir, src))
			if !ok {
				continue
			}
			srcs = append(srcs, src)

			switch fi.ext {
			case jsExt, jsxExt:
				deps = append(deps, fi.deps...)
				requires = append(requires, fi.imports...)
			}
		}
		existingRuleSrcs = append(existingRuleSrcs, srcs...)
		if html := r.AttrString("html"); html != "" {
			existingRuleSrcs = append(existingRuleSrcs, html)
		}

		// Emit an empty rule if none of the srcs were present.
		if len(srcs) == 0 {
			empty = append(empty, existingLib(r, nil))
		} else {
			rules = append(rules, existingLib(r, srcs))
			imports = append(imports, importInfo{
				imports: requires,
				deps:    deps,
			})
		}
	}

	// Loop through each file in this package, read their info (what they
	// provide & require), and generate a lib or test rule for it.
	var testFileInfos = make(map[string][]fileInfo)
	sort.Strings(args.RegularFiles) // results in htmls first
	for _, filename := range args.RegularFiles {
		// Skip any that we already processed.
		if contains(existingRuleSrcs, filename) {
			continue
		}

		var fi, ok = jsFileInfo(args.Config.RepoRoot, jsc, filepath.Join(args.Dir, filename))
		if !ok {
			continue
		}
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

		// Create one closure_js[x]_library rule per non-test source file.
		switch fi.ext {
		case jsExt, jsxExt:
			rules = append(rules, generateLib(filename))
			imports = append(imports, newImportInfo(fi))
		}
	}

	// Group foo_test.js[x] with foo_test.html (if present) into test targets.
	for name, fis := range testFileInfos {
		switch len(fis) {
		case 1:
			if fis[0].ext == htmlExt {
				log.Println("unused test html:", path.Join(args.Rel, fis[0].name))
				continue
			}
			rules = append(rules, generateTest(fis[0]))
			imports = append(imports, newImportInfo(fis[0]))
		case 2:
			rules = append(rules, generateCombinedTest(fis[1], fis[0]))
			imports = append(imports, newImportInfo(fis[1]))
		default:
			log.Println("unexpected number of test sources:", name)
		}
	}

	return language.GenerateResult{
		Gen:     rules,
		Imports: imports,
		Empty:   empty,
	}
}

// testBaseName trims foo_test.[js|html] => "foo"
func testBaseName(name string) string {
	return name[:strings.Index(name, "_test.")]
}

func existingLib(existing *rule.Rule, srcs []string) *rule.Rule {
	r := rule.NewRule(existing.Kind(), existing.Name())
	if len(srcs) > 0 {
		r.SetAttr("srcs", srcs)
	}
	r.SetAttr("visibility", []string{"//visibility:public"})
	return r
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

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
	var rules []*rule.Rule
	var imports []interface{}
	//fmt.Println("GenerateRules:", args)
	var testFileInfos = make(map[string][]fileInfo)
	sort.Strings(args.RegularFiles) // results in htmls first
	for _, filename := range args.RegularFiles {
		var fi = jsFileInfo(filepath.Join(args.Dir, filename))
		if fi.ext == unknownExt {
			continue
		}

		// Deal with tests separately since they involve multiple files.
		if fi.isTest {
			name := testBaseName(fi.name)
			testFileInfos[name] = append(testFileInfos[name], fi)
			continue
		}

		// Create one closure_js[x]_library rule per non-test source file.
		switch fi.ext {
		case jsExt, jsxExt:
			// fmt.Println("GenerateRules:", filename, fi)
			// fmt.Println("	lib:", generateLib(filename))
			rules = append(rules, generateLib(filename))
			imports = append(imports, fi)
			//fmt.Println(filename, "provides", fi.provides, "imports", fi.imports)
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
		filepath.Base(filename)[:len(filename)-len(filepath.Ext(filename))])
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

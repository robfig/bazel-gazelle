package js

import "github.com/bazelbuild/bazel-gazelle/rule"

var jsKinds = map[string]rule.KindInfo{
	"filegroup": {
		NonEmptyAttrs:  map[string]bool{"srcs": true},
		MergeableAttrs: map[string]bool{"srcs": true},
	},
	"closure_js_library": {
		MatchAttrs: []string{"srcs"},
		NonEmptyAttrs: map[string]bool{
			"deps": true,
			"srcs": true,
		},
		MergeableAttrs: map[string]bool{
			"srcs": true,
		},
		ResolveAttrs: map[string]bool{"deps": true},
	},
	"closure_js_test": {
		MatchAttrs: []string{"srcs"},
		NonEmptyAttrs: map[string]bool{
			"deps": true,
			"srcs": true,
		},
		MergeableAttrs: map[string]bool{
			"srcs": true,
		},
		ResolveAttrs: map[string]bool{"deps": true},
	},
	"closure_jsx_library": {
		MatchAttrs: []string{"srcs"},
		NonEmptyAttrs: map[string]bool{
			"deps": true,
			"srcs": true,
		},
		MergeableAttrs: map[string]bool{
			"srcs": true,
		},
		ResolveAttrs: map[string]bool{"deps": true},
	},
	"closure_jsx_test": {
		MatchAttrs: []string{"srcs"},
		NonEmptyAttrs: map[string]bool{
			"deps": true,
			"srcs": true,
		},
		MergeableAttrs: map[string]bool{
			"srcs": true,
		},
		ResolveAttrs: map[string]bool{"deps": true},
	},
}

var jsLoads = []rule.LoadInfo{
	{
		Name: "@io_bazel_rules_closure//closure:defs.bzl",
		Symbols: []string{
			"closure_js_library",
			"closure_js_test",

			// NOTE: These rules do not actually exist, and there is no standard library
			// type for JSX. On the bright side, it's easy enough to write your own.
			// Users that need this functionality should use a map_kind directive to
			// direct Gazelle which rule type to use.
			"closure_jsx_library",
			"closure_jsx_test",
		},
	},
}

func (_ *jsLang) Kinds() map[string]rule.KindInfo { return jsKinds }
func (_ *jsLang) Loads() []rule.LoadInfo          { return jsLoads }

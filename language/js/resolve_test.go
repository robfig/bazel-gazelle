package js

import "testing"

func TestResolveClosureLibrary(t *testing.T) {
	var tests = []struct {
		in, out string
	}{
		{"goog.i18n.MessageFormat", "@io_bazel_rules_closure//closure/library/i18n:messageformat"},
	}
	for _, test := range tests {
		actual := resolveClosureLibrary(test.in).String()
		if actual != test.out {
			t.Errorf("%v: expected %v, got %v", test.in, test.out, actual)
		}
	}
}

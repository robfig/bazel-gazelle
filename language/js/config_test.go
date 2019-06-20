package js

import "testing"

func TestExternGrep(t *testing.T) {
	var tests = []struct {
		name               string
		file, token, label string
		matched            bool
	}{
		{
			name:  "no match",
			token: "$",
			label: "//js/vendor:jquery",
			file: `
corp.settings.Page = function($el) {
  this.$ctxt = $el;
};
`,
			matched: false,
		},
		{
			name:  "match",
			token: "$",
			label: "//js/vendor:jquery",
			file: `
corp.settings.Page = function() {
  this.$ctxt = $('.js-settings');
};
`,
			matched: true,
		},
		{
			name:  "match react",
			token: "React",
			label: "//js/externs:react",
			file: `
corp.settings.Page = function() {
  React.renderElement();
};
`,
			matched: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			eg := newExternGrep(test.token, test.label)
			actual := eg.matches([]byte(test.file))
			if actual != test.matched {
				t.Errorf("matched: (expected) %v != %v (actual)", test.matched, actual)
			}
		})
	}
}

package js

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestJsFileInfo(t *testing.T) {
	for _, tc := range []struct {
		desc, name, source string
		want               fileInfo
	}{
		{
			"empty file",
			"foo.js",
			"",
			fileInfo{
				imports:  nil,
				provides: []string{"es6:/foo.js", "es6:/foo"},
				ext:      jsExt,
			},
		},
		{
			"a provide",
			"foo.js",
			"goog.provide('corp.foo');",
			fileInfo{
				imports:  nil,
				provides: []string{"corp.foo"},
				ext:      jsExt,
			},
		},
		{
			"two provides",
			"foo.js",
			`goog.provide('corp.foo');
goog.provide('corp.foo2');
`,
			fileInfo{
				imports:  nil,
				provides: []string{"corp.foo", "corp.foo2"},
				ext:      jsExt,
			},
		},
		{
			"a module, jsx",
			"foo.jsx",
			"goog.module('corp.foo');",
			fileInfo{
				provides: []string{"corp.foo"},
				ext:      jsxExt,
				isModule: true,
			},
		},
		{
			"a require",
			"foo.js",
			`goog.provide('corp.foo');
goog.require('corp');`,
			fileInfo{
				imports:  []string{"corp"},
				provides: []string{"corp.foo"},
				ext:      jsExt,
			},
		},
		{
			"multiple requires",
			"foo.js",
			`goog.module('corp.foo');

goog.require('corp');
const str = goog.require('corp.string');
var dom = goog.require('corp.dom');
const {
  foo,
  bar
} = goog.require('corp.widgets');

`,
			fileInfo{
				provides: []string{"corp.foo"},
				imports:  []string{"corp", "corp.string", "corp.dom", "corp.widgets"},
				ext:      jsExt,
				isModule: true,
			},
		},
		{
			"test js",
			"foo_test.js",
			`goog.module('corp.foo')`,
			fileInfo{
				provides: []string{"corp.foo"},
				ext:      jsExt,
				isTest:   true,
				isModule: true,
			},
		},
		{
			"i18n.js from integration test",
			"i18n.js",
			`goog.provide("corp.i18n");
goog.provide('corp.msg');

goog.require('corp');
goog.require('goog.strings');
goog.require('goog.i18n.messageformat');
`,
			fileInfo{
				provides: []string{"corp.i18n", "corp.msg"},
				imports:  []string{"corp", "goog.strings", "goog.i18n.messageformat"},
				ext:      jsExt,
			},
		},
		{
			"es6modules",
			"path/to/app/ListEdit.jsx",
			`import {
  listDataShape,
} from '../../shapes.js';
import { IndeterminateValue } from '../../utils/display-utils.jsx';
import { FieldErrors } from '../../field-row/FieldErrors.jsx';

const { moveItem } = goog.require('goog.array');
goog.require('corp.i18n');
`,
			fileInfo{
				provides: []string{
					"es6:/path/to/app/ListEdit.jsx",
					"es6:/path/to/app/ListEdit",
				},
				imports: []string{
					"goog.array",
					"corp.i18n",
					"es6:/path/shapes.js",
					"es6:/path/utils/display-utils.jsx",
					"es6:/path/field-row/FieldErrors.jsx",
				},
				ext: jsxExt,
			},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			dir, err := ioutil.TempDir(os.Getenv("TEST_TEMPDIR"), "TestJsFileInfo")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(dir)
			path := filepath.Join(dir, tc.name)
			os.MkdirAll(filepath.Dir(path), 0777)
			if err := ioutil.WriteFile(path, []byte(tc.source), 0600); err != nil {
				t.Fatal(err)
			}

			var jsc jsConfig
			got, _ := jsFileInfo(dir, &jsc, path)
			// Clear fields we don't care about for testing.
			got = fileInfo{
				provides: got.provides,
				isTest:   got.isTest,
				isModule: got.isModule,
				imports:  got.imports,
				ext:      got.ext,
			}

			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("case %q:\n got %#v\nwant %#v", tc.desc, got, tc.want)
			}
		})
	}
}

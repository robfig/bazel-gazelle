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
				imports:    nil,
				provides:   []string{"/foo"},
				ext:        jsExt,
				moduleType: moduleTypeES6,
			},
		},
		{
			"a provide",
			"foo.js",
			"goog.provide('corp.foo');",
			fileInfo{
				imports:    nil,
				provides:   []string{"corp.foo"},
				ext:        jsExt,
				moduleType: moduleTypeGoogProvide,
			},
		},
		{
			"two provides",
			"foo.js",
			`goog.provide('corp.foo');
goog.provide('corp.foo2');
`,
			fileInfo{
				imports:    nil,
				provides:   []string{"corp.foo", "corp.foo2"},
				ext:        jsExt,
				moduleType: moduleTypeGoogProvide,
			},
		},
		{
			"a module, jsx",
			"foo.jsx",
			"goog.module('corp.foo');",
			fileInfo{
				provides:   []string{"corp.foo"},
				ext:        jsxExt,
				moduleType: moduleTypeGoogModule,
			},
		},
		{
			"a require",
			"foo.js",
			`goog.provide('corp.foo');
goog.require('corp');`,
			fileInfo{
				imports:    []string{"corp"},
				provides:   []string{"corp.foo"},
				ext:        jsExt,
				moduleType: moduleTypeGoogProvide,
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
const {A, B, C} =
  goog.require('corp.D');

`,
			fileInfo{
				provides:   []string{"corp.foo"},
				imports:    []string{"corp", "corp.string", "corp.dom", "corp.widgets", "corp.D"},
				ext:        jsExt,
				moduleType: moduleTypeGoogModule,
			},
		},
		{
			"test js",
			"foo_test.js",
			`goog.module('corp.foo')`,
			fileInfo{
				provides:   []string{"corp.foo"},
				ext:        jsExt,
				isTest:     true,
				moduleType: moduleTypeGoogModule,
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
				provides:   []string{"corp.i18n", "corp.msg"},
				imports:    []string{"corp", "goog.strings", "goog.i18n.messageformat"},
				ext:        jsExt,
				moduleType: moduleTypeGoogProvide,
			},
		},
		{
			"es6modules",
			"path/to/app/ListEdit.jsx",
			`import {
  listDataShape,
} from '../../shapes';
import { IndeterminateValue } from '../../utils/display-utils';
import { FieldErrors } from '../../field-row/FieldErrors';

const { moveItem } = goog.require('goog.array');
goog.require('corp.i18n');
`,
			fileInfo{
				provides: []string{
					"/path/to/app/ListEdit",
				},
				imports: []string{
					"goog.array",
					"corp.i18n",
					"/path/shapes",
					"/path/utils/display-utils",
					"/path/field-row/FieldErrors",
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
				provides:   got.provides,
				isTest:     got.isTest,
				moduleType: got.moduleType,
				imports:    got.imports,
				ext:        got.ext,
			}

			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("case %q:\n got %#v\nwant %#v", tc.desc, got, tc.want)
			}
		})
	}
}

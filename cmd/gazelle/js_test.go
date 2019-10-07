/* Copyright 2017 The Bazel Authors. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// This file contains integration tests for the JS language.

package main

import (
	"testing"

	"github.com/bazelbuild/bazel-gazelle/testtools"
)

// TestJS tests the JS rule generation.
func TestJS(t *testing.T) {
	files := []testtools.FileSpec{
		{
			Path: "WORKSPACE",
		}, {
			Path:    "BUILD.bazel",
			Content: "# gazelle:js_grep_extern React //js/externs:react",
		}, {
			Path:    "corp.js",
			Content: `goog.provide('corp')`,
		}, {
			Path: "i18n.js",
			Content: `goog.provide("corp.i18n");
goog.provide('corp.msg');

goog.require('corp');
goog.require('goog.string');
goog.require('goog.i18n.MessageFormat');
`,
		},
		{
			Path: "ui/widget.jsx",
			Content: `
goog.module('corp.ui.widget');

goog.require('corp.i18n');
goog.require('goog.ui.Component');
`,
		},
		{
			Path: "ui/widget_test.jsx",
			Content: `
goog.module('corp.ui.widget.test');

goog.require('corp');
const msg = goog.require('corp.msg');
goog.require('goog.ui.Component');
var testtools = goog.require('corp.ui.widget.testtools');
		`,
		},
		{
			Path:    "ui/widget_test.html",
			Content: `<!DOCTYPE html><html></html>`,
		},
		{
			Path: "app/app.js",
			Content: `
goog.module('corp.app.App');

const msg = goog.require('corp.msg');
goog.require('corp');
const widget = goog.require('corp.ui.widget');

React.renderElement(widget);
		`,
		},
		{
			Path: "app/app_test.js",
			Content: `
goog.module('corp.app.AppTest');

const app = goog.require('corp.app.App');
		`,
		},
		{
			Path: "existing/file1.js",
			Content: `goog.module('corp.existing.file1');
goog.require('corp.msg');
goog.require('goog.string');
`,
		},
		{
			Path: "existing/file2.js",
			Content: `goog.module('corp.existing.file1');
goog.require('corp.ui.widget');
goog.require('goog.dom.query');
const file1 = goog.require('corp.existing.file1');
`,
		},
		{
			Path: "existing/file3.js",
			Content: `goog.module('corp.existing.file3');
goog.require('goog.array');
const file1 = goog.require('corp.existing.file1');
`,
		},
		{
			Path:    "existing/file4.js",
			Content: `goog.module('corp.existing.file4');`,
		},
		{
			Path: "existing/BUILD.bazel",
			Content: `
load("@io_bazel_rules_closure//closure:defs.bzl", "closure_js_library")

closure_js_library(
    name = "existing",
    srcs = ["file1.js", "file2.js"],
    visibility = ["//visibility:public"],
)

closure_js_library(
    name = "existingfile3",
    srcs = ["file3.js"],
    visibility = ["//visibility:public"],
)

closure_js_library(
    name = "deleted",
    srcs = ["deleted.js"],
    visibility = ["//visibility:public"],
)

closure_js_library(
    name = "deleted_multifile",
    srcs = ["deleted1.js", "deleted2.js"],
    visibility = ["//visibility:public"],
)

closure_js_library(
    name = "deleted_some_multifile",
    srcs = ["file4.js", "deleted3.js"],
    visibility = ["//visibility:public"],
)

closure_js_library(
    name = "globbed",
    srcs = glob(["*.js"], exclude=["file1.js", "file2.js", "file3.js"]),
    visibility = ["//visibility:public"],
)
`,
		},
		{
			Path:    "map_kind/file1.js",
			Content: `goog.module('corp.map_kind.file1');`,
		},
		{
			Path:    "map_kind/file2.js",
			Content: `goog.module('corp.map_kind.file2');`,
		},
		{
			Path: "map_kind/BUILD.bazel",
			Content: `
load("//tools/js:defs.bzl", "closure_js_externs")

# gazelle:map_kind closure_js_library closure_js_externs //tools/js:defs.bzl

closure_js_externs(
    name = "existing",
    srcs = ["file1.js"],
    visibility = ["//visibility:public"],
)

closure_js_externs(
    name = "deleted",
    srcs = ["deleted.js"],
    visibility = ["//visibility:public"],
)
`,
		},

		{
			Path: "js/bugs/dupe_dep_not_caused_by_grep_extern_and_requires.js",
			Content: `
goog.require('React');

React.renderElement();
`,
		},
		{
			Path: "js/externs/react.js",
			Content: `/** @externs */
goog.provide('React');
var React = {};`,
		},

		// ES6 tests
		{
			Path:    "es6modules/utils/display-utils.jsx",
			Content: ``,
		},
		{
			Path: "es6modules/app/fields/widget.jsx",
			Content: `
import { IndeterminateValue } from '../../utils/display-utils';
const { moveItem } = goog.require('goog.array');
goog.require('corp.i18n');
`,
		},
		{
			Path: "es6modules/app/fields/widget_test.jsx",
			Content: `
import { Widget } from '/es6modules/app/fields/widget';
 `,
		},
	}
	dir, cleanup := testtools.CreateFiles(t, files)
	defer cleanup()

	if err := runGazelle(dir, []string{}); err != nil {
		t.Fatal(err)
	}

	testtools.CheckFiles(t, dir, []testtools.FileSpec{
		{
			Path: "BUILD.bazel",
			Content: `
load("@io_bazel_rules_closure//closure:defs.bzl", "closure_js_library")

# gazelle:js_grep_extern React //js/externs:react

closure_js_library(
    name = "corp",
    srcs = ["corp.js"],
    visibility = ["//visibility:public"],
)

closure_js_library(
    name = "i18n",
    srcs = ["i18n.js"],
    visibility = ["//visibility:public"],
    deps = [
        ":corp",
        "@io_bazel_rules_closure//closure/library/i18n:messageformat",
        "@io_bazel_rules_closure//closure/library/string",
    ],
)
`,
		},
		{
			Path: "ui/BUILD.bazel",
			Content: `
load("@io_bazel_rules_closure//closure:defs.bzl", "closure_jsx_library", "closure_jsx_test")

closure_jsx_library(
    name = "widget",
    srcs = ["widget.jsx"],
    visibility = ["//visibility:public"],
    deps = [
        "//:i18n",
        "@io_bazel_rules_closure//closure/library/ui:component",
    ],
)

closure_jsx_test(
    name = "widget_test",
    srcs = ["widget_test.jsx"],
    compilation_level = "ADVANCED",
    entry_points = ["corp.ui.widget.test"],
    html = "widget_test.html",
    visibility = ["//visibility:public"],
    deps = [
        "//:corp",
        "//:i18n",
        "@io_bazel_rules_closure//closure/library/ui:component",
    ],
)
`,
		},
		{
			Path: "app/BUILD.bazel",
			Content: `
load("@io_bazel_rules_closure//closure:defs.bzl", "closure_js_library", "closure_js_test")

closure_js_library(
    name = "app",
    srcs = ["app.js"],
    visibility = ["//visibility:public"],
    deps = [
        "//:corp",
        "//:i18n",
        "//js/externs:react",
        "//ui:widget",
    ],
)

closure_js_test(
    name = "app_test",
    srcs = ["app_test.js"],
    compilation_level = "ADVANCED",
    entry_points = ["corp.app.AppTest"],
    visibility = ["//visibility:public"],
    deps = [":app"],
)
`,
		},
		{
			Path: "existing/BUILD.bazel",
			Content: `
load("@io_bazel_rules_closure//closure:defs.bzl", "closure_js_library")

closure_js_library(
    name = "existing",
    srcs = [
        "file1.js",
        "file2.js",
    ],
    visibility = ["//visibility:public"],
    deps = [
        "//:i18n",
        "//ui:widget",
        "@io_bazel_rules_closure//closure/library/string",
        "@io_bazel_rules_closure//third_party/closure/library/dojo/dom:query",
    ],
)

closure_js_library(
    name = "existingfile3",
    srcs = ["file3.js"],
    visibility = ["//visibility:public"],
    deps = [
        ":existing",
        "@io_bazel_rules_closure//closure/library/array",
    ],
)

closure_js_library(
    name = "deleted_some_multifile",
    srcs = ["file4.js"],
    visibility = ["//visibility:public"],
)

closure_js_library(
    name = "globbed",
    srcs = glob(
        ["*.js"],
        exclude = [
            "file1.js",
            "file2.js",
            "file3.js",
        ],
    ),
    visibility = ["//visibility:public"],
)
`,
		},
		{
			Path: "map_kind/BUILD.bazel",
			Content: `
load("//tools/js:defs.bzl", "closure_js_externs")

# gazelle:map_kind closure_js_library closure_js_externs //tools/js:defs.bzl

closure_js_externs(
    name = "existing",
    srcs = ["file1.js"],
    visibility = ["//visibility:public"],
)

closure_js_externs(
    name = "file2",
    srcs = ["file2.js"],
    visibility = ["//visibility:public"],
)
`,
		},
		{
			Path: "js/externs/BUILD.bazel",
			Content: `
load("@io_bazel_rules_closure//closure:defs.bzl", "closure_js_library")

closure_js_library(
    name = "react",
    srcs = ["react.js"],
    visibility = ["//visibility:public"],
)
`,
		},

		{
			Path: "js/bugs/BUILD.bazel",
			Content: `
load("@io_bazel_rules_closure//closure:defs.bzl", "closure_js_library")

closure_js_library(
    name = "dupe_dep_not_caused_by_grep_extern_and_requires",
    srcs = ["dupe_dep_not_caused_by_grep_extern_and_requires.js"],
    visibility = ["//visibility:public"],
    deps = ["//js/externs:react"],
)
`,
		},

		{
			Path: "es6modules/utils/BUILD.bazel",
			Content: `
load("@io_bazel_rules_closure//closure:defs.bzl", "closure_jsx_library")

closure_jsx_library(
    name = "display-utils",
    srcs = ["display-utils.jsx"],
    visibility = ["//visibility:public"],
)
`,
		},
		{
			Path: "es6modules/app/fields/BUILD.bazel",
			Content: `
load("@io_bazel_rules_closure//closure:defs.bzl", "closure_jsx_library", "closure_jsx_test")

closure_jsx_library(
    name = "widget",
    srcs = ["widget.jsx"],
    visibility = ["//visibility:public"],
    deps = [
        "//:i18n",
        "//es6modules/utils:display-utils",
        "@io_bazel_rules_closure//closure/library/array",
    ],
)

closure_jsx_test(
    name = "widget_test",
    srcs = ["widget_test.jsx"],
    compilation_level = "ADVANCED",
    entry_points = ["/es6modules/app/fields/widget_test"],
    visibility = ["//visibility:public"],
    deps = [":widget"],
)
`,
		},
	})
}

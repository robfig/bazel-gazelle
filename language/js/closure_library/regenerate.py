#!/usr/bin/env python
# Copyright 2018 The Closure Rules Authors. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

"""Closure Library BUILD definition generator.

NOTE(robfig): This is a patched copy of @rules_closure/closure/library/regenerate.py
It generates a Go source file mapping each provide in closure library to its BUILD target.

Usage:
	regenerate.py (path to closure library base)

"""

import collections
import itertools
import os
import re
import subprocess
import sys

HEADER = '// DO NOT EDIT -- bazel run //language/js/closure_library:regenerate -- $PWD/closure_library.go\n\n'
REPO = 'com_google_javascript_closure_library'

PROVIDE_PATTERN = re.compile(r'^goog\.(?:provide|module)\([\'"]([^\'"]+)', re.M)
REQUIRE_PATTERN = re.compile(
    r'^(?:(?:const|var) .* = )?goog\.require\([\'"]([^\'"]+)', re.M)

TESTONLY_PATTERN = re.compile(r'^goog\.setTestOnly\(', re.M)
TESTONLY_PATHS_PATTERN = re.compile(r'(?:%s)' % '|'.join((
    r'^closure/goog/labs/testing/',  # forgot to use goog.setTestOnly()
    r'^closure/goog/testing/net/mockiframeio\.js$',
)))

IGNORE_PATHS_PATTERN = re.compile(r'(?:%s)' % '|'.join((
    r'_perf',
    r'_test',
    r'/demos/',
    r'/testdata/',
    r'^closure/goog/base\.js$',
    r'^closure/goog/deps\.js$',
    r'^closure/goog/transitionalforwarddeclarations\.js$',
    r'^closure/goog/transpile\.js$',
    r'^closure/goog/debug_loader_integration_tests/',
    r'^third_party/closure/goog/osapi',
)))

UNITTEST_PATTERN = re.compile('|'.join((
    r'goog\.require\(.goog\.testing\.testSuite',
    r'^function (?:setUp|tearDown)',
)), re.M)

MASTER = 'closure/library/BUILD'
MASTER_EXCLUDES = ('/ui', '/labs', 'third_party/')
MASTER_EXTRA = '''
filegroup(
    name = "base",
    srcs = [
        "@{0}//:closure/goog/base.js",
        "@{0}//:closure/goog/transitionalforwarddeclarations.js",
    ],
)
closure_js_library(
    name = "deps",
    srcs = ["@{0}//:closure/goog/deps.js"],
    lenient = True,
)
closure_js_library(
    name = "transpile",
    srcs = ["@{0}//:closure/goog/transpile.js"],
    lenient = True,
)
closure_css_library(
    name = "css",
    srcs = ["@{0}//:css_files"],
)
py_binary(
    name = "regenerate",
    srcs = ["regenerate.py"],
    args = ["$(location @{0}//:closure/goog/base.js)"],
    data = [
        "@{0}",
        "@{0}//:closure/goog/base.js",
    ],
    tags = [
        "local",
        "manual",
    ],
    visibility = ["//visibility:private"],
)
'''.format(REPO)


def mkdir(path):
  try:
    os.makedirs(path)
  except OSError as e:
    if e.errno != 17:  # File exists
      raise


def find(prefix):
  for base, _, names in os.walk(prefix):
    for name in names:
      yield os.path.join(base, name)


def normalize(path):
  return path.replace('closure/goog', 'closure/library')


def file2build(path):
  return os.path.join(os.path.dirname(normalize(path)), 'BUILD')


def file2name(path):
  return os.path.splitext(os.path.basename(path))[0]


def file2dep(path):
  path = normalize(path)
  return '//%s:%s' % (os.path.dirname(path), file2name(path))


def main(basejs, out):
  assert out.startswith('/')

  # cd @com_google_javascript_closure_library//
  os.chdir(os.path.join(os.path.dirname(basejs), '../..'))

  # files=$(find {third_party,}closure/goog | sort)
  files = sorted(itertools.chain(find('closure/goog'),
                                 find('third_party/closure/goog')))

  # Find JavaScript sources and determine their relationships.
  jslibs = []
  jstestlibs = set()  # jslibs with goog.setTestOnly()
  jsrawlibs = set()  # jslibs without goog.provide() or goog.module()
  file2requires = {}  # e.g. closure/goog/array/array.js -> goog.asserts
  provide2file = {}  # e.g. goog.asserts -> closure/goog/asserts/asserts.js
  for f in files:
    if IGNORE_PATHS_PATTERN.search(f) is not None:
      continue
    file2requires[f] = []
    if f.endswith('.js'):
      with open(f) as fh:
        data = fh.read()
        provides = [m.group(1) for m in PROVIDE_PATTERN.finditer(data)]
      if provides:
        if (TESTONLY_PATHS_PATTERN.search(f) is not None or
            TESTONLY_PATTERN.search(data) is not None):
          if UNITTEST_PATTERN.search(data) is not None:
            continue
          jstestlibs.add(f)
        for provide in provides:
          provide2file[provide] = f
        file2requires[f] = sorted(set(
            m.group(1) for m in REQUIRE_PATTERN.finditer(data)))
      else:
        jsrawlibs.add(f)
      jslibs.append(f)

  # PATCH(robfig): Print Go source code for a mapping from each `goog.provide`
  # to the BUILD target providing that source file.

  with open(out, 'w') as fh:
      fh.write(HEADER)
      fh.write("package closure_library\n\nvar PROVIDE_TO_TARGET = map[string]string{\n")
      for provide, f in sorted(provide2file.iteritems()):
          fh.write('\t"%s": "%s",\n' % (provide, file2dep(f)))
      fh.write("}")

  return subprocess.call(['go', 'fmt', out])


if __name__ == '__main__':
  sys.exit(main(*sys.argv[1:]))

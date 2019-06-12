load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

# Use this to re-generate closure_library.go
py_binary(
    name = "regenerate",
    srcs = ["regenerate.py"],
    args = ["$(location @com_google_javascript_closure_library//:base)"],
    data = [
        "@com_google_javascript_closure_library//:all",
        "@com_google_javascript_closure_library//:base",
    ],
    tags = [
        "local",
        "manual",
    ],
    visibility = ["//visibility:private"],
)

go_library(
    name = "closure_library",
    srcs = ["closure_library.go"],
    importpath = "github.com/bazelbuild/bazel-gazelle/language/js/closure_library",
    visibility = ["//visibility:public"],
)

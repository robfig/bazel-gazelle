workspace(name = "bazel_gazelle")

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
    name = "io_bazel_rules_go",
    sha256 = "313f2c7a23fecc33023563f082f381a32b9b7254f727a7dd2d6380ccc6dfe09b",
    urls = [
        "https://storage.googleapis.com/bazel-mirror/github.com/bazelbuild/rules_go/releases/download/0.19.3/rules_go-0.19.3.tar.gz",
        "https://github.com/bazelbuild/rules_go/releases/download/0.19.3/rules_go-0.19.3.tar.gz",
    ],
)

load("@io_bazel_rules_go//go:deps.bzl", "go_register_toolchains", "go_rules_dependencies")

go_rules_dependencies()

go_register_toolchains(nogo = "@bazel_gazelle//:nogo")

load("//:deps.bzl", "gazelle_dependencies")

gazelle_dependencies()

load("@io_bazel_rules_go//tests:bazel_tests.bzl", "test_environment")

test_environment()

http_archive(
    name = "com_google_javascript_closure_library",
    urls = [
        "https://mirror.bazel.build/github.com/google/closure-library/archive/v20190415.tar.gz",
        "https://github.com/google/closure-library/archive/v20190415.tar.gz",
    ],
    sha256 = "b92f9c2b81adeb06b16e0a5d748baf115eeb58884fe864bec3fe8e23a7cf7d14",
    strip_prefix = "closure-library-20190415",
    build_file_content = """
filegroup(name="all", srcs=glob(["**/*"]), visibility=["//visibility:public"])

filegroup(name="base", srcs=["closure/goog/base.js"], visibility=["//visibility:public"])
""",
)

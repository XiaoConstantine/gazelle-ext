load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")
load("@bazel_gazelle//:def.bzl", "gazelle", "gazelle_binary")

package(default_visibility = ["//visibility:public"])

go_library(
    name = "gazelle_ext",
    srcs = ["extension.go"],
    embedsrcs = ["analysis.scm"],
    importpath = "gazelle_ext",
    deps = [
        "@bazel_gazelle//config:go_default_library",
        "@bazel_gazelle//label:go_default_library",
        "@bazel_gazelle//language:go_default_library",
        "@bazel_gazelle//repo:go_default_library",
        "@bazel_gazelle//resolve:go_default_library",
        "@bazel_gazelle//rule:go_default_library",
        "@com_github_smacker_go_tree_sitter//:go-tree-sitter",
        "@com_github_smacker_go_tree_sitter//java",
        "@org_golang_x_exp//maps",
    ],
)

gazelle_binary(
    name = "gazelle_ext_binary",
    languages = [
        ":gazelle_ext",
    ],
)

gazelle(
    name = "ext",
    gazelle = ":gazelle_ext_binary",
)

# gazelle:prefix gazelle_ext
gazelle(
    name = "gazelle",
)

gazelle(
    name = "gazelle_update_repos",
    args = [
        "-from_file=go.mod",
        "-to_macro=go.bzl%deps",
        "-prune",
    ],
    command = "update-repos",
)

go_test(
    name = "gazelle_ext_test",
    srcs = ["extension_test.go"],
    embed = [":gazelle_ext"],
    deps = [
        "@bazel_gazelle//config:go_default_library",
        "@bazel_gazelle//label:go_default_library",
        "@bazel_gazelle//language:go_default_library",
        "@bazel_gazelle//repo:go_default_library",
        "@bazel_gazelle//resolve:go_default_library",
    ],
)
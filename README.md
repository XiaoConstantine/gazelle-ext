# gazelle-ext
---

Attempt to build a gazelle extension for Java langauge, using treesitter as parser. As a part of learning go and
gazelle process

__References__:

[ugazelle](https://github.com/sluongng/ugazelle)

[rules_jvm](https://github.com/bazel-contrib/rules_jvm/tree/main)

---

## Getting started

In your WORKSPACE file, add following rules:

```python

git_repository(
    name = "extension",
    branch = "main",
    remote = "https://github.com/XiaoConstantine/gazelle-ext.git"
)
load("@a//:go.bzl", "deps")

# install deps for the extension
deps()
```

and in the top level BUILD file:

```python
load("@bazel_gazelle//:def.bzl", "gazelle", "gazelle_binary")

gazelle(
    name = "ext",
    gazelle = "@extension//:ext",
)
```


```bash
  bazel run //:ext
```

## TODO

- [ ] Smarter usage of tree sitter
- [ ] Make `Resolve` actually look up current package as well `Maven` dependency
- [ ] Add `java_binary` target




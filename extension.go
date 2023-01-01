package gazelle_ext

import (
	"context"
	_ "embed"
	"errors"
	"flag"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/label"
	"github.com/bazelbuild/bazel-gazelle/language"
	"github.com/bazelbuild/bazel-gazelle/repo"
	"github.com/bazelbuild/bazel-gazelle/resolve"
	"github.com/bazelbuild/bazel-gazelle/rule"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/java"
	"golang.org/x/exp/maps"
)

const (
	languageName  = "java"
	fileType      = ".java"
	packagesKey   = "_java_packages"
	full_package  = "full_package"
	local_package = "local_package"
	full_import   = "full_import"
	class_name    = "class_name"
)

type Extension struct {
	parser *sitter.Parser
}

func NewLanguage() language.Language {
	parser := sitter.NewParser()
	parser.SetLanguage(java.GetLanguage())

	return &Extension{parser}
}

// RegisterFlags registers command-line flags used by the extension. This
// method is called once with the root configuration when Gazelle
// starts. RegisterFlags may set an initial values in Config.Exts. When flags
// are set, they should modify these values.
func (e *Extension) RegisterFlags(fs *flag.FlagSet, cmd string, c *config.Config) {
}

// CheckFlags validates the configuration after command line flags are parsed.
// This is called once with the root configuration when Gazelle starts.
// CheckFlags may set default values in flags or make implied changes.
func (e *Extension) CheckFlags(fs *flag.FlagSet, c *config.Config) error {
	return nil
}

// KnownDirectives returns a list of directive keys that this Configurer can
// interpret. Gazelle prints errors for directives that are not recoginized by
// any Configurer.
func (e *Extension) KnownDirectives() []string {
	return nil
}

// Configure modifies the configuration using directives and other information
// extracted from a build file. Configure is called in each directory.
//
// c is the configuration for the current directory. It starts out as a copy
// of the configuration for the parent directory.
//
// rel is the slash-separated relative path from the repository root to
// the current directory. It is "" for the root directory itself.
//
// f is the build file for the current directory or nil if there is no
// existing build file.
func (e *Extension) Configure(c *config.Config, rel string, f *rule.File) {
}

// Name returns the name of the language. This should be a prefix of the
// kinds of rules generated by the language, e.g., "go" for the Go extension
// since it generates "go_library" rules.
func (e *Extension) Name() string {
	return languageName
}

func javaLibrary(name string) bool {
	return name == "java_library"
}

// Imports returns a list of ImportSpecs that can be used to import the rule
// r. This is used to populate RuleIndex.
//
// If nil is returned, the rule will not be indexed. If any non-nil slice is
// returned, including an empty slice, the rule will be indexed.
func (e *Extension) Imports(c *config.Config, r *rule.Rule, f *rule.File) []resolve.ImportSpec {
	log.Println("calling imports")
	if !javaLibrary(r.Kind()) {
		return nil
	}
	var out []resolve.ImportSpec
	log.Printf("%#v", r)

	if pkgs := r.PrivateAttr(packagesKey); pkgs != nil {
		for _, pkg := range pkgs.([]string) {
			out = append(out, resolve.ImportSpec{Lang: languageName, Imp: pkg})
		}

	}
	log.Printf("Out %#v", out)
	return out
}

// Embeds returns a list of labels of rules that the given rule embeds. If
// a rule is embedded by another importable rule of the same language, only
// the embedding rule will be indexed. The embedding rule will inherit
// the imports of the embedded rule.
func (e *Extension) Embeds(r *rule.Rule, from label.Label) []label.Label {
	log.Println("calling embeds")
	panic("not implemented") // TODO: Implement
}

// Resolve translates imported libraries for a given rule into Bazel
// dependencies. Information about imported libraries is returned for each
// rule generated by language.GenerateRules in
// language.GenerateResult.Imports. Resolve generates a "deps" attribute (or
// the appropriate language-specific equivalent) for each import according to
// language-specific rules and heuristics.
func (e *Extension) Resolve(c *config.Config, ix *resolve.RuleIndex, rc *repo.RemoteCache, r *rule.Rule, imports interface{}, from label.Label) {
	log.Println("calling resolve")
	imps := imports.([]string)
	r.DelAttr("deps")
	if len(imps) == 0 {
		log.Println("empty imports")
		return
	}
	deps := make([]string, 0, len(imps))

	for _, imp := range imps {
		// Skip built-in dependencies
		if strings.HasPrefix(imp, "import java.") {
			continue
		}

		impLabel, err := label.Parse(imp)
		if err != nil {
			log.Printf("%s: import of %q is invalid: %v", from.String(), imp, err)
			continue
		}

		impLabel = impLabel.Abs(from.Repo, from.Pkg)
		log.Printf("%v", impLabel)

		if impLabel.Repo != "" || !c.IndexLibraries {
			// This is a dependency that is external to the current repo, or indexing
			// is disabled so take a guess at what hte target name should be.
			deps = append(deps, strings.TrimSuffix(imp, fileType))
			continue
		}

		res := resolve.ImportSpec{
			Lang: languageName,
			Imp:  impLabel.String(),
		}
		matches := ix.FindRulesByImportWithConfig(c, res, languageName)

		if len(matches) == 0 {
			log.Printf(
				"%s: %q (%s) was not found in dependency index. Skipping. This may result in an incomplete deps section and require manual BUILD file intervention.\n",
				from.String(),
				imp,
				impLabel.String(),
			)
		}

		for _, m := range matches {
			depLabel := m.Label
			depLabel = depLabel.Rel(from.Repo, from.Pkg)
			deps = append(deps, depLabel.String())
		}
	}
	log.Printf("Deps: %#v", deps)
	sort.Strings(deps)
	if len(deps) > 0 {
		r.SetAttr("deps", deps)
	}

}

var kinds = map[string]rule.KindInfo{
	"java_library": {
		NonEmptyAttrs:  map[string]bool{"srcs": true, "deps": true},
		MergeableAttrs: map[string]bool{"srcs": true},
		ResolveAttrs: map[string]bool{
			"deps":         true,
			"runtime_deps": true,
		},
	},
}

// Kinds returns a map of maps rule names (kinds) and information on how to
// match and merge attributes that may be found in rules of those kinds. All
// kinds of rules generated for this language may be found here.
func (e *Extension) Kinds() map[string]rule.KindInfo {
	return kinds
}

// Loads returns .bzl files and symbols they define. Every rule generated by
// GenerateRules, now or in the past, should be loadable from one of these
// files.
func (e *Extension) Loads() []rule.LoadInfo {
	return []rule.LoadInfo{}
}

// GenerateRules extracts build metadata from source files in a directory.
// GenerateRules is called in each directory where an update is requested
// in depth-first post-order.
//
// args contains the arguments for GenerateRules. This is passed as a
// struct to avoid breaking implementations in the future when new
// fields are added.
//
// A GenerateResult struct is returned. Optional fields may be added to this
// type in the future.
//
// Any non-fatal errors this function encounters should be logged using
// log.Print.

//go:embed analysis.scm
var analysisQuery []byte

type javaPackageSummary struct {
	full_pkg string
	deps     map[string][]string
	files    []string
	pkg      string
}

func (e *Extension) GenerateRules(args language.GenerateArgs) language.GenerateResult {
	log.Println("calling generateRules")

	// workspaceRoot := args.Config.RepoRoot
	// log.Printf("%#v", workspaceRoot)
	// summarize out of all java files
	// group by full package
	// a.b.c: {
	//	  files: [d, e, f],
	//
	// }
	m := make(map[string]*javaPackageSummary)
	for _, f := range args.RegularFiles {
		if !javaSourceFile(f) {
			continue
		}
		fullPath := filepath.Join(args.Dir, f)
		log.Printf("%#v", fullPath)
		loads, err := e.getTreeSitterJavaFileLoads(fullPath)
		if v, ok := m[loads.full_pkg]; ok {
			maps.Copy(m[loads.full_pkg].deps, v.deps)
			m[loads.full_pkg].files = append(m[loads.full_pkg].files, loads.file_name)
		} else {
			m[loads.full_pkg] = &javaPackageSummary{
				full_pkg: loads.full_pkg,
				deps:     loads.deps,
				files:    []string{loads.file_name},
				pkg:      loads.pkg,
			}
		}
		if err != nil {
			log.Printf("%s: contains syntax errors: %v", fullPath, err)
		}
	}

	rules, imports := generate(m)

	return language.GenerateResult{
		Gen:     rules,
		Imports: imports,
	}
}

func generate(m map[string]*javaPackageSummary) (rules []*rule.Rule, imports []interface{}) {
	var a []*rule.Rule
	var b []interface{}

	for key, summary := range m {
		log.Println(key)
		r := rule.NewRule("java_library", summary.pkg)
		r.SetAttr("srcs", summary.files)

		d := maps.Keys(summary.deps)
		a = append(a, r)
		b = append(b, d)
	}

	return a, b
}

func javaSourceFile(f string) bool {
	return strings.HasSuffix(f, fileType)
}

type javaFile struct {
	// is this needed?
	pkg       string
	deps      map[string][]string
	full_pkg  string
	file_name string
}

// The query will provide us information (if parse succeed) as following format:
// package a.b.(c)
// import a.b.(c)
// import a.(*)
// the parsing result looks like:
// { pkg: c
//   full_pkg: a.b.c
//   deps: {a.b: [c]}
//  }
//
func (e *Extension) getTreeSitterJavaFileLoads(path string) (*javaFile, error) {
	f, err := os.ReadFile(path)

	if err != nil {
		return nil, errors.New("error opening file")
	}

	tree, err := e.parser.ParseCtx(context.Background(), nil, f)

	if err != nil {
		return nil, errors.New("parse tree error")
	}

	q, err := sitter.NewQuery(analysisQuery, java.GetLanguage())
	if err != nil {
		return nil, errors.New("query init failure")
	}
	qc := sitter.NewQueryCursor()
	qc.Exec(q, tree.RootNode())

	var full_pkg, pkg string
	var imports []string
	deps := make(map[string][]string)
	for {
		m, ok := qc.NextMatch()
		if !ok {
			break
		}
		for _, c := range m.Captures {
			switch q.CaptureNameForId(c.Index) {
			case full_package:
				full_pkg = c.Node.Content(f)[7:]
			case full_import:
				imports = append(imports, strings.TrimSuffix(c.Node.Content(f)[6:], ";"))
			case class_name:
				clsName := c.Node.Content(f)
				if len(imports) > 0 && strings.HasSuffix(imports[len(imports)-1], clsName) {
					prev := imports[len(imports)-1]
					key := strings.TrimSuffix(prev[:len(prev)-len(clsName)], ".")
					deps[key] = append(deps[key], clsName)
				}
			case local_package:
				pkg = c.Node.Content(f)
			}
		}
	}

	item := &javaFile{
		pkg:       pkg,
		full_pkg:  full_pkg,
		deps:      deps,
		file_name: path,
	}
	log.Printf("%#v", item)
	return item, nil
}

// Fix repairs deprecated usage of language-specific rules in f. This is
// called before the file is indexed. Unless c.ShouldFix is true, fixes
// that delete or rename rules should not be performed.
func (e *Extension) Fix(c *config.Config, f *rule.File) {
}
